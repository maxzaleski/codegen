package gen

import (
	"context"
	"fmt"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/fs"
	"github.com/maxzaleski/codegen/internal/metrics"
	"github.com/maxzaleski/codegen/internal/utils/slice"
	"sync"
)

type (
	// IConcierge is the interface that wraps the generation concierge.
	IConcierge interface {
		WalkDirectoryStructure(md *core.Metadata, pkgs []*core.Package)
		ExtractJobs(spec *core.Spec)
		StartWorkers()
		WaitQueueReadiness()
		WaitAndCleanup()
		WgAdd(delta int)
		GetErrChannel() <-chan error
	}

	runtimeConcierge struct {
		ctx    context.Context
		config Config

		errChan chan error
		wg      *sync.WaitGroup

		queue      IQueue
		logger     ILogger
		metrics    metrics.IMetrics
		aggrScopes []*domainScope
	}
)

// newConcierge returns a new instance of IConcierge.
func newConcierge(ctx1 context.Context, rl slog.ILogger, c Config, as []*domainScope) IConcierge {
	ms := ctx1.Value(contextKeyMetrics).(*metrics.Metrics)

	ctx2, cancel := context.WithCancel(ctx1)
	s := &runtimeConcierge{
		ctx:    ctx2,
		config: c,

		errChan: make(chan error, 1),
		wg:      &sync.WaitGroup{},

		queue:      newQueue(rl, ms, c),
		logger:     newLogger(rl, "concierge", slog.Pink),
		metrics:    ms,
		aggrScopes: as,
	}
	go s.listenForError(cancel)

	return s
}

func (c *runtimeConcierge) WalkDirectoryStructure(md *core.Metadata, pkgs []*core.Package) {
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
					c.errChan <- err
					return
				}
				continue
			}

			for _, p := range pkgs {
				if err := createDirINE(as.Output+"/"+p.Name, as.AbsoluteOutput+"/"+p.Name); err != nil {
					c.errChan <- err
					return
				}
			}
		}
	}
}

func (c *runtimeConcierge) ExtractJobs(spec *core.Spec) {
	defer func() {
		c.wg.Done()

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
					Metadata: Metadata{
						AbsolutePath: as.AbsoluteOutput,
						Inline:       as.Inline,
						DomainType:   as.DomainType,
						Metadata:     *spec.Metadata,
						ScopeKey:     as.Key,
					},
					DisableTemplates: c.config.DisableTemplates,
				}

				for _, sj := range as.Jobs {
					// This check prevents the job from being executed for each package.
					if as.DomainType == core.DomainTypeHttp && sj.Unique {
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

func (c *runtimeConcierge) StartWorkers() {
	defer c.wg.Done()

	c.logger.Log("workers", "msg", "starting workers", "count", c.config.WorkerCount)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			for i := 0; i < c.config.WorkerCount; i++ {
				c.wg.Add(1)
				go worker(c.ctx, i+1, c.wg, c.queue, c.errChan)
			}
			return
		}
	}
}

func (c *runtimeConcierge) WaitAndCleanup() {
	defer c.logger.Log("exit", "msg", "returning to main thread – Goodbye 👋")

	c.wg.Wait()
	close(c.errChan)
}

func (c *runtimeConcierge) GetErrChannel() <-chan error {
	return c.errChan
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

func (c *runtimeConcierge) listenForError(cancel func()) {
	// Waits for the first value to be sent to the error channel, subsequently terminates.
	err, ok := <-c.errChan
	// Channel is closed.
	if !ok {
		return
	}

	// Handle error.
	if err != nil {
		c.logger.Log("error<-", "msg", "received an error, cancelling the context", "err", err)

		// Cancel the context, killing all workers.
		cancel()

		// Forward the error to the caller.
		c.errChan <- err
	}
}
