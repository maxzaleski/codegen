package gen

import (
	"context"
	"fmt"
	"github.com/codegen/internal/core"
	"github.com/codegen/internal/core/slog"
	"github.com/codegen/internal/fs"
	"github.com/codegen/internal/utils/slice"
	"sync"
)

type (
	// ISetup is the interface that wraps the setup methods.
	ISetup interface {
		WalkDirectoryStructure(md *core.Metadata, pkgs []*core.Package)
		ExtractJobs(spec *core.Spec)
		StartWorkers()
		WaitQueueReadiness()
		WaitAndCleanup()
		WgAdd(delta int)
		GetErrChannel() <-chan error
	}

	setup struct {
		ctx    context.Context
		config Config

		errChan chan error
		wg      *sync.WaitGroup

		queue  IQueue
		logger ILogger
	}
)

// newSetup returns a new instance of ISetup.
func newSetup(ctx context.Context, rl slog.ILogger, c Config) ISetup {
	ctx2, cancel := context.WithCancel(ctx)
	s := &setup{
		ctx:    ctx2,
		config: c,

		errChan: make(chan error, 1),
		wg:      &sync.WaitGroup{},

		queue:  newQueue(rl, c),
		logger: newLogger(rl, "setup"),
	}
	go s.listenForError(cancel)

	return s
}

func (s *setup) WalkDirectoryStructure(md *core.Metadata, pkgs []*core.Package) {
	for _, as := range aggrScopes {
		for _, j := range as.Jobs {
			as.AbsoluteOutput = md.Cwd + "/" + as.Output
			if as.Inline || j.Unique {
				logCreatingDirectory(s.logger, as.AbsoluteOutput)

				if err := fs.CreateDirINE(as.AbsoluteOutput); err != nil {
					s.errChan <- err
					return
				}
				continue
			}

			for _, p := range pkgs {
				path := as.AbsoluteOutput + "/" + p.Name
				logCreatingDirectory(s.logger, as.AbsoluteOutput)

				if err := fs.CreateDirINE(path); err != nil {
					s.errChan <- err
					return
				}
			}
		}
	}
}

func (s *setup) ExtractJobs(spec *core.Spec) {
	defer func() {
		s.wg.Done()

		s.queue.Close()
		s.queue.Ready()
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			sJs := map[string]bool{}  // Seen jobs; populated off `Unique` jobs only.
			gJs := make([]*genJob, 0) // Final gen. jobs to be executed.

			for _, as := range aggrScopes {
				j := genJob{
					Metadata: Metadata{
						AbsolutePath: as.AbsoluteOutput,
						Inline:       as.Inline,
						DomainType:   as.DomainType,
						Metadata:     *spec.Metadata,
						ScopeKey:     as.Key,
					},
					DisableTemplates: s.config.DisableTemplates,
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
			s.logger.Log("feed", "msg", "enqueuing jobs", "count", len(gJs))
			for _, j := range gJs {
				if !s.config.DebugMode {
					// Case: len(jobs) > len(queue)
					// Mark queue as ready once the queue is almost full.
					q := s.queue
					if capacity := q.GetCapacity(); q.GetSize() == capacity-1 && len(gJs) > capacity {
						q.Ready()
					}
				}
				s.queue.Enqueue(j)
			}

			return
		}
	}
}

func (s *setup) StartWorkers() {
	defer s.wg.Done()

	s.logger.Log("workers", "msg", "starting workers", "count", s.config.WorkerCount)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			for i := 0; i < s.config.WorkerCount; i++ {
				s.wg.Add(1)
				go worker(s.ctx, i+1, s.wg, s.queue, s.errChan)
			}
			return
		}
	}
}

func (s *setup) WaitAndCleanup() {
	s.wg.Wait()
	close(s.errChan)
}

func (s *setup) GetErrChannel() <-chan error {
	return s.errChan
}

func (s *setup) WaitQueueReadiness() {
	// No values are ever passed to this channel. The only indicator is the channel being closed.
	_, ok := <-s.queue.ReadyListener()
	if !ok {
		return
	}
}

func (s *setup) WgAdd(delta int) {
	s.wg.Add(delta)
}

func (s *setup) listenForError(cancel func()) {
	// Waits for the first value to be sent to the error channel, subsequently terminates.
	err, ok := <-s.errChan
	// Channel is closed.
	if !ok {
		return
	}

	// Handle error.
	if err != nil {
		s.logger.Log("error<-", "msg", "received an error, cancelling the context", "err", err)

		// Cancel the context, killing all workers.
		cancel()

		// Forward the error to the caller.
		s.errChan <- err
	}
}
