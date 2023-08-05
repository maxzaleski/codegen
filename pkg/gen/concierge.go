package gen

import (
	"context"
	"fmt"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/db"
	"github.com/maxzaleski/codegen/internal/lib"
	"github.com/maxzaleski/codegen/internal/lib/datastructure"
	"github.com/maxzaleski/codegen/internal/lib/slice"
	"github.com/maxzaleski/codegen/internal/slog"
	"github.com/maxzaleski/codegen/pkg/gen/modules"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"os"
	"strings"
)

type (
	// IConcierge is the interface that wraps the generation concierge.
	IConcierge interface {
		Start(c Config, spec *core.Spec, scopes []*core.DomainScope)
		Wait() error
	}

	concierge struct {
		ctx    IContext
		config Config
		logger ILogger

		errg *errgroup.Group
		ds   []*core.DomainScope

		// Modules.
		queue       datastructure.IQueue[genJob]
		metrics     modules.IMetrics
		diagnostics modules.IDiagnostics
		ttProcessor modules.ITemplateProcessor
	}
)

// newConcierge returns a new instance of IConcierge.
func newConcierge(
	errg *errgroup.Group, ctx IContext, c Config, logger slog.ILogger, db db.IDatabase, ds []*core.DomainScope,
) IConcierge {
	metrics := ctx.GetMetrics()
	s := &concierge{
		ctx:    ctx,
		errg:   errg,
		config: c,

		metrics:     metrics,
		ds:          ds,
		queue:       newQueue(logger, c),
		logger:      newLogger(logger, "concierge", slog.Pink),
		diagnostics: modules.NewDiagnostics(logger, db),
		ttProcessor: modules.NewTemplateProcessor(ctx.GetPackages()),
	}
	return s
}

func (rc *concierge) Wait() (err error) {
	logger := rc.logger
	logger.Log("wait", "msg", "blocking until all processes have completed")
	defer func(err error) {
		if err == nil {
			logger.Log("exit", "msg", "goodbye 👋")
		}
	}(err)

	if err = rc.errg.Wait(); err != nil {
		logger.Log("main:error<-", "msg", "received an error", "err", err)

		if rc.config.DebugVerbose {
			p := logger.Parent()
			p.LogEnv("Debug verbose", "debugVerbose", "logging verbose error")
			p.Logf("verboseErr=\n%+v", err)
		}
	}
	return
}

func (rc *concierge) Start(c Config, spec *core.Spec, scopes []*core.DomainScope) {
	{
		log := func(fields ...any) { rc.logger.Log("preflight", fields...) }
		log("msg", "executing preflight functions")
		defer log("msg", "done")
	}

	// [1] Prepare diagnostics module.
	if err := rc.diagnostics.Prepare(spec); err != nil {
		panic(errors.Wrap(err, "concierge: failed to prepare diagnostics module"))
	}

	// [2] Extract jobs and feed the queue.
	rc.errg.Go(func() error { return rc.feedQueue(c, *spec.Metadata, scopes, spec.Pkgs) })
	<-rc.queue.ReadySignal()

	// [3] Start workers.
	rc.errg.Go(func() error { return rc.startWorkers() })
}

func (rc *concierge) feedQueue(c Config, sm core.Metadata, scopes []*core.DomainScope, pkgs []*core.Package) error {
	{
		log := func(fields ...any) { rc.logger.Log("preflight:extract", fields...) }
		log("msg", "starting job extraction")
		defer log("msg", "done")
	}

	// [1] Parse executable jobs.
	//
	// For each scope, we extract the jobs and enqueue them:
	// • (1) If domain = 'http' && j.Unique, we only enqueue the job once
	// • (2) Otherwise, we enqueue a copy of the job for each package (default)
	fJs := make([]*genJob, 0)
	for _, scope := range rc.ds {
		newJob := func(sj *core.ScopeJob) *genJob {
			return &genJob{
				Metadata: metadata{
					Inline:     scope.Inline,
					DomainType: scope.ParentType,
					Metadata:   sm,
					ScopeKey:   scope.Key,
				},
				OutputFile: &genJobFile{
					AbsoluteDirPath: sm.Cwd + "/" + scope.Output,
				},
				DisableTemplates: c.IgnoreTemplates,
				ScopeJob:         sj,
			}
		}

		for _, sJob := range scope.Jobs {
			// -> Scenario (1)
			if scope.ParentType == core.DomainTypeHttp && sJob.Unique {
				fJs = append(fJs, newJob(sJob))
				continue
			}

			// -> Scenario (2)
			js := slice.Map(pkgs,
				func(p *core.Package) *genJob {
					jPkg := newJob(sJob.Copy())
					jPkg.ScopeJob.Key = fmt.Sprintf("%s-%s", p.Name, jPkg.ScopeJob.Key)
					jPkg.Package = p
					return jPkg
				})
			fJs = append(fJs, js...)
		}
	}

	// [2] Feed the queue.
	return rc.enqueue(fJs)
}

func (rc *concierge) enqueue(js []*genJob) error {
	defer func() {
		rc.queue.Ready()
		rc.queue.Close()
	}()

	log := func(fields ...any) { rc.logger.Log("preflight:enqueue", fields...) }
	{
		log("msg", "starting enqueue", "count", len(js))
		defer log("msg", "done")
	}

	for _, j := range js {
		if err := j.Prepare(); err != nil {
			return err
		}

		// Ensures that if len(queue.remaining) < len(jobs.remaining), the queue is marked as 'ready' (for consumption),
		// thereby allowing the remaining jobs to be enqueued.
		q := rc.queue
		if capacity := q.GetCapacity(); q.GetSize() == capacity-1 && len(js) > capacity {
			if !q.GetReady() {
				log("msg", "queue is almost full, marking as ready")
			}
			q.Ready()
		}
		q.Enqueue(j)
	}
	return nil
}

var errAllJobsProcessed = errors.New("all jobs processed")

func (rc *concierge) startWorkers() (_ error) {
	{
		log := func(fields ...any) { rc.logger.Log("preflight:workers", fields...) }
		log("msg", "starting workers", "count", rc.config.WorkerCount)
		defer log("msg", "done")
	}

	for {
		select {
		case <-rc.ctx.GetUnderlying().Done():
			return
		default:
			for i := 0; i < rc.config.WorkerCount; i++ {
				wID := i + 1 // `i` is captured by the closure.
				rc.errg.Go(func() (err error) {
					if err = rc.worker(wID); err != nil {
						switch lib.Unwrap(err) {
						case errAllJobsProcessed, context.Canceled:
							err = nil
						default:
							// This may be hit multiple times for different workers; this is normal.
							//
							// If a prevalent error is encountered, more than one job will be affected,
							// hence the multiple logs.
							rc.logger.Log("worker:error<-",
								"worker_id", wID,
								"msg", "worker exited with an error",
								"err", err,
							)
						}
					}
					return
				})
			}
			return
		}
	}
}

func (rc *concierge) worker(id int) error {
	ctx := rc.ctx
	metrics, logger :=
		ctx.GetMetrics(),
		newLogger(
			ctx.GetLogger(),
			fmt.Sprintf("worker_%d", id),
			slog.None,
		)

	logger.Log("start", "msg", "starting worker")
	defer logger.Log("exit", "msg", "worker exiting")

	exec := func(j *genJob) (err error) {
		logOutcome := func(o jobOutcome) {
			fn := strings.Replace(j.OutputFile.AbsolutePath, j.Metadata.Cwd, "", 1)
			logger.Ack("file", j, "status", string(o), "file", fn)
		}

		// [1] Setup metric capture.
		//
		// -> Parse scope key: `domain/scope | domain` -> `domain`.
		sk := j.Metadata.ScopeKey
		if s := strings.Split(sk, "/"); len(s) == 2 {
			sk = s[0]
		}
		// -> Define metric job and package key; if the job is unique, we use an alias.
		mj, pk :=
			&modules.MetricJob{FileAbsolutePath: j.OutputFile.AbsolutePath},
			core.UniquePkgAlias
		if p := j.Package; p != nil {
			pk = p.Name
		}
		defer func() { metrics.CaptureJob(sk, pk, *mj) }() // deferred as to allow mutation.

		// [2] Evaluate whether to proceed.
		//
		// • Override: true, always run job
		// • OverrideOn: should run iff the '.codegen/pkg' contents have changed per the `OverrideOn`'s specificities.
		if !j.Override {
			// [EXPERIMENTAL]: overrideOn: uncompleted.
			//if len(j.OverrideOn) != 0 {
			//	logger.Log("eval", "msg", "verifying configuration files")
			//
			//	// -> Check whether the configuration files have changed.
			//	var ok bool
			//	ok, err = rc.diagnostics.Verify(j.OverrideOn)
			//	if err != nil {
			//		return errors.Wrap(err, "failed to verify configuration files")
			//	}
			//
			//	// -> ok: no change was detected since last use; exit.
			//	if ok {
			//		defer logOutcome(fileOutcomeIgnored)
			//		return
			//	}
			//}

			if _, err = os.Stat(j.OutputFile.AbsolutePath); err != nil {
				if !os.IsNotExist(err) {
					return errors.WithMessagef(err, "failed presence check at '%s'", j.OutputFile.AbsolutePath)
				}
			} else {
				defer logOutcome(fileOutcomeIgnored)
				return
			}
		}

		// [3] Execute templates.
		if err = rc.ttProcessor.Exec(
			j.Templates,
			j.DisableTemplates,
			j.Package,
			j.OutputFile.AbsolutePath,
			j.OutputFile.Ext,
			rc.config.TemplateFuncMap,
		); err == nil {
			mj.FileCreated = true
			defer logOutcome(fileOutcomeSuccess)
		}

		return
	}

	for {
		if err := ctx.GetUnderlying().Err(); err != nil {
			return err
		}

		// [1] Dequeue job.
		j := rc.queue.Dequeue()
		if j == nil {
			return errAllJobsProcessed // queue is empty; all jobs processed.
		}
		// -> Capture work unit.
		metrics.CaptureWorkUnit(modules.MetricWorkUnit{WorkerID: id})

		// [2] Execute job.
		if err := exec(j); err != nil {
			return err
		}
	}
}
