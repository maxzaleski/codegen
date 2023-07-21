package gen

import (
	"context"
	"fmt"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/lib/datastructure"
	"github.com/maxzaleski/codegen/internal/lib/slice"
	"github.com/maxzaleski/codegen/internal/metrics"
	"golang.org/x/sync/errgroup"
)

type (
	// IConcierge is the interface that wraps the generation concierge.
	IConcierge interface {
		Start(spec *core.Spec, c Config)
		Wait() error
	}

	runtimeConcierge struct {
		ctx    context.Context
		errg   *errgroup.Group
		config Config

		queue   datastructure.IQueue[genJob]
		logger  ILogger
		metrics metrics.IMetrics
	}
)

// newConcierge returns a new instance of IConcierge.
func newConcierge(ctx context.Context, rl slog.ILogger, errg *errgroup.Group, c Config) IConcierge {
	ms := ctx.Value(contextKeyMetrics).(*metrics.Metrics)

	s := &runtimeConcierge{
		ctx:    ctx,
		errg:   errg,
		config: c,

		queue:   newGenQueue(rl, ms, c),
		logger:  newLogger(rl, "concierge", slog.Pink),
		metrics: ms,
	}
	return s
}

func (rc *runtimeConcierge) Wait() (err error) {
	rc.logger.Log("wait", "msg", "blocking until all processes have completed")
	defer func(err error) {
		if err == nil {
			rc.logger.Log("exit", "msg", "goodbye 👋")
		}
	}(err)

	if err = rc.errg.Wait(); err != nil {
		rc.logger.Log("main:error<-", "msg", "received an error", "err", err)

		if rc.config.DebugVerbose {
			p := rc.logger.Parent()
			p.LogEnv("Debug verbose", "debugVerbose", "logging verbose error")
			p.Logf("err=\n%+v", err)
		}
	}
	return
}

func (rc *runtimeConcierge) Start(spec *core.Spec, c Config) {
	{
		log := func(fields ...any) { rc.logger.Log("preflight", fields...) }
		log("msg", "executing preflight functions")
		defer log("msg", "done")
	}

	// -> Extract jobs and feed the queue.
	rc.errg.Go(func() error { return rc.feedQueue(spec, c) })

	<-rc.queue.ReadyObservable()

	// -> Start workers.
	rc.errg.Go(func() error { return rc.startWorkers() })
}

func (rc *runtimeConcierge) feedQueue(s *core.Spec, c Config) error {
	{
		log := func(fields ...any) { rc.logger.Log("preflight:extract", fields...) }
		log("msg", "starting job extraction")
		defer log("msg", "done")
	}

	// -> Aggregate all scopes.
	hScopes, pScopes := s.Config.HttpDomain.Scopes, s.Config.PkgDomain.Scopes
	ds := make([]*core.DomainScope, 0, len(hScopes)+len(pScopes))
	ds = append(ds, hScopes...)
	ds = append(ds, pScopes...)

	// -> Extract final jobs.
	//
	// For each scope, we extract the jobs and enqueue them:
	// • (1) If the scope is of HTTP type with the `Unique` flag set, we only enqueue the job once
	// • (2) Otherwise, we enqueue a copy of the job for each package (default)
	fJs := make([]*genJob, 0)
	for _, scope := range ds {
		newJob := func(sj *core.ScopeJob) *genJob {
			return &genJob{
				Metadata: metadata{
					Inline:     scope.Inline,
					DomainType: scope.Type,
					Metadata:   *s.Metadata,
					ScopeKey:   scope.Key,
				},
				OutputFile: &genJobFile{
					AbsoluteDirPath: s.Metadata.Cwd + "/" + scope.Output,
				},
				DisableTemplates: c.IgnoreTemplates,
				ScopeJob:         sj,
			}
		}

		for _, sJob := range scope.Jobs {
			// Scenario (1): unique
			if scope.Type == core.DomainTypeHttp && sJob.Unique {
				fJs = append(fJs, newJob(sJob))
				continue
			}

			// Scenario (2):
			js := slice.Map(s.Pkgs,
				func(p *core.Package) *genJob {
					jPkg := newJob(sJob.Copy())
					jPkg.ScopeJob.Key = fmt.Sprintf("%s-%s", p.Name, jPkg.ScopeJob.Key)
					jPkg.Package = p
					return jPkg
				})
			fJs = append(fJs, js...)
		}
	}

	// -> Feed the queue.
	return rc.enqueue(fJs)
}

func (rc *runtimeConcierge) enqueue(js []*genJob) error {
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

		q := rc.queue
		if capacity := q.GetCapacity(); q.GetSize() == capacity-1 && len(js) > capacity {
			// Mark queue scope ready once the queue is almost full.
			if !q.GetReady() {
				log("msg", "queue is almost full, marking as ready")
			}
			q.Ready()
		}
		q.Enqueue(j)
	}
	return nil
}

func (rc *runtimeConcierge) startWorkers() (_ error) {
	{
		log := func(fields ...any) { rc.logger.Log("preflight:workers", fields...) }
		log("msg", "starting workers", "count", rc.config.WorkerCount)
		defer log("msg", "done")
	}

	for {
		select {
		case <-rc.ctx.Done():
			return
		default:
			for i := 0; i < rc.config.WorkerCount; i++ {
				wID := i + 1 // `i` is captured by the closure.
				rc.errg.Go(func() (err error) {
					if err = worker(rc.ctx, wID, rc.queue); err != nil {
						switch err {
						case errFileAlreadyPresent, context.Canceled:
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
