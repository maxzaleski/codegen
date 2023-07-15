package gen

import (
	"context"
	"fmt"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/fs"
	"github.com/maxzaleski/codegen/internal/metrics"
	"github.com/maxzaleski/codegen/internal/utils/slice"
	"github.com/maxzaleski/codegen/pkg/gen/queue"
	"golang.org/x/sync/errgroup"
	"sync"
)

type (
	// IConcierge is the interface that wraps the generation concierge.
	IConcierge interface {
		WalkDirectoryStructure(md *core.Metadata, pkgs []*core.Package) error
		ExtractJobs(spec *core.Spec) error
		StartWorkers() error
		WaitQueueReadiness()
		WaitAndCleanup() error
		WgAdd(delta int)
	}

	runtimeConcierge struct {
		ctx    context.Context
		errg   *errgroup.Group
		config Config

		wg *sync.WaitGroup

		queue      queue.IQueue[genJob]
		logger     ILogger
		metrics    metrics.IMetrics
		aggrScopes []*core.DomainScope
	}
)

// newConcierge returns a new instance of IConcierge.
func newConcierge(ctx context.Context, rl slog.ILogger, errg *errgroup.Group, c Config, as []*core.DomainScope) IConcierge {
	ms := ctx.Value(contextKeyMetrics).(*metrics.Metrics)

	s := &runtimeConcierge{
		ctx:    ctx,
		errg:   errg,
		config: c,

		wg: &sync.WaitGroup{},

		queue:      newGenQueue(rl, ms, c),
		logger:     newLogger(rl, "concierge", slog.Pink),
		metrics:    ms,
		aggrScopes: as,
	}
	return s
}

func (c *runtimeConcierge) WalkDirectoryStructure(md *core.Metadata, pkgs []*core.Package) error {
	const event = "dirwalk"

	c.logger.Log(event, "msg", "starting directory structure walk")

	spm := map[string]any{}
	createDirINE := func(key, path string) error {
		// If the directory has already been seen, exit.
		if _, ok := spm[key]; ok {
			c.logger.Log(event, "status", "seen", "path", path)
			return nil
		}
		spm[key] = nil

		c.logger.Log(event, "status", "creating directory", "path", path)

		return fs.CreateDirINE(path)
	}

	for _, as := range c.aggrScopes {
		for _, j := range as.Jobs {
			as.AbsoluteOutput = md.Cwd + "/" + as.Output
			if as.Inline || j.Unique {
				if err := createDirINE(as.Output, as.AbsoluteOutput); err != nil {
					return err
				}
				continue
			}

			for _, p := range pkgs {
				if err := createDirINE(as.Output+"/"+p.Name, as.AbsoluteOutput+"/"+p.Name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *runtimeConcierge) ExtractJobs(spec *core.Spec) (err error) {
	defer func() {
		c.queue.Close()
		c.queue.Ready()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			sJs := map[string]bool{}  // Seen jobs; populated off `Unique` jobs only.
			gJs := make([]*genJob, 0) // Final gen. jobs to be executed.

			for _, as := range c.aggrScopes {
				j := genJob{
					Metadata: metadata{
						AbsolutePath: as.AbsoluteOutput,
						Inline:       as.Inline,
						DomainType:   as.Type,
						Metadata:     *spec.Metadata,
						ScopeKey:     as.Key,
					},
					DisableTemplates: c.config.DisableTemplates,
				}

				for _, sj := range as.Jobs {
					// This check prevents the job from being executed for each package.
					if as.Type == core.DomainTypeHttp && sj.Unique {
						if ok, _ := sJs[as.Key]; ok {
							return
						}
						sJs[as.Key] = true

						j.ScopeJob = sj
						gJs = append(gJs, &j)
					} else {
						jPkgs := slice.Map(spec.Pkgs, func(p *core.Package) *genJob {
							jPkg := j
							jPkg.ScopeJob = sj.Copy()
							jPkg.ScopeJob.Key = fmt.Sprintf("%s-%s", p.Name, jPkg.ScopeJob.Key)
							jPkg.Package = p
							return &jPkg
						})
						gJs = append(gJs, jPkgs...)
					}
				}
			}

			// Enqueue jobs.
			const event = "feed"
			c.logger.Log(event, "msg", "enqueuing jobs", "count", len(gJs))
			for _, j := range gJs {
				if err = j.Fill(); err != nil {
					return err
				}

				// Case: len(jobs) > len(queue)
				// Mark queue as ready once the queue is almost full.
				q := c.queue
				if capacity := q.GetCapacity(); q.GetSize() == capacity-1 && len(gJs) > capacity {
					if !q.GetReady() {
						c.logger.Log(event, "msg", "queue is almost full, marking as ready")
					}
					q.Ready()
				}
				c.queue.Enqueue(j)
			}

			return
		}
	}
}

func (c *runtimeConcierge) StartWorkers() (_ error) {
	c.logger.Log("workers", "msg", "starting workers", "count", c.config.WorkerCount)

	fn := func(wID int) (err error) {
		if err = worker(c.ctx, wID, c.queue); err != nil {
			switch err {
			case errFileAlreadyPresent, context.Canceled:
			default:
				c.logger.Log("worker:error<-",
					"worker_id", wID,
					"msg", "received an error",
					"err", err,
				)
			}
		}
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			for i := 0; i < c.config.WorkerCount; i++ {
				wID := i + 1 // `i` is captured by the closure.
				c.errg.Go(func() error { return fn(wID) })
			}
			return
		}
	}
}

func (c *runtimeConcierge) WaitAndCleanup() error {
	c.logger.Log("wait", "msg", "blocking until all processes have completed")
	defer c.logger.Log("exit", "msg", "goodbye 👋")

	err := c.errg.Wait()
	if err != nil {
		c.logger.Log("main:error<-", "msg", "received an error", "err", err)
	}
	return err
}

func (c *runtimeConcierge) WaitQueueReadiness() {
	// No values are ever passed to this channel. The only indicator is the channel being closed.
	_, ok := <-c.queue.ReadyListener()
	if !ok {
		return
	}
}

func (c *runtimeConcierge) WgAdd(delta int) {
	c.wg.Add(delta)
}
