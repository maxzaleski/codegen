package gen

import (
	"context"
	"fmt"
	"github.com/codegen/internal/core/slog"
	"github.com/codegen/internal/metrics"
	"github.com/codegen/internal/utils/slice"
	"github.com/pkg/errors"
	"sync"
	"time"

	"github.com/codegen/internal/core"
	"github.com/codegen/internal/fs"
)

var aggrScopes []*domainScope

type (
	// Config represents the configuration of the code generator.
	Config struct {
		Location         string
		WorkerCount      int
		DisableTemplates bool
		DebugMode        bool
	}

	domainScope struct {
		*core.DomainScope
		DomainType core.DomainType
	}
)

func Execute(c Config, began time.Time) (md *core.Metadata, _ metrics.IMetrics, _ error) {
	rl := slog.New(c.DebugMode, began)

	// Parse configuration via `.codegen` directory.
	spec, err := core.NewSpec(rl, c.Location)
	md = spec.Metadata
	if err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	sc := spec.Config
	hScopes, pScopes := sc.HttpDomain.Scopes, sc.PkgDomain.Scopes
	aggrScopes = make([]*domainScope, 0, len(hScopes)+len(pScopes))
	aggrScopes = append(aggrScopes, mapToDomainScope(pScopes, core.DomainTypePkg)...)
	aggrScopes = append(aggrScopes, mapToDomainScope(hScopes, core.DomainTypeHttp)...)

	wg, errChan := &sync.WaitGroup{}, make(chan error, 1)
	q, ms := newQueue(rl, c), metrics.New(nil)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "began", began)
	ctx = context.WithValue(ctx, "logger", rl)
	ctx = context.WithValue(ctx, "metrics", ms)

	l := newLogger(rl, "main")

	// Listen for the first error if any.
	go listenForError(l, errChan, cancel)

	// Make sure the output directories exist.
	walkDirectoryStructure(l, spec.Metadata, spec.Pkgs, errChan)

	// Feed the queue with jobs.
	wg.Add(1)
	go feedQueue(ctx, l, wg, spec, q, c)

	q.WaitReadiness()

	//Start workers.
	wg.Add(1)
	go startWorkers(ctx, l, c.WorkerCount, wg, q, errChan)

	wg.Wait()
	close(errChan) // Consequently, the err-listening goroutine will terminate.

	return md, ms, <-errChan
}

func listenForError(l ILogger, errChan chan error, cancel func()) {
	if err := <-errChan; err != nil {
		l.Log("error<-", "msg", "received an error, cancelling the context", "err", err)

		// Cancel the context, killing all workers.
		cancel()

		// Forward the error to the caller.
		errChan <- err
	}
}

func walkDirectoryStructure(l ILogger, md *core.Metadata, pkgs []*core.Package, errChan chan<- error) {
	const event = "dirwalk"

	for _, s := range aggrScopes {
		for _, j := range s.Jobs {
			s.AbsoluteOutput = md.Cwd + "/" + s.Output
			if s.Inline || j.Unique {
				l.Log(event, "msg", "creating directory if not exist", "path", s.AbsoluteOutput)

				if err := fs.CreateDirINE(s.AbsoluteOutput); err != nil {
					errChan <- err
					return
				}
				continue
			}

			for _, p := range pkgs {
				path := s.AbsoluteOutput + "/" + p.Name
				l.Log(event, "msg", "creating directory if not exist", "path", path)

				if err := fs.CreateDirINE(path); err != nil {
					errChan <- err
					return
				}
			}
		}
	}
}

func feedQueue(ctx context.Context, l ILogger, wg *sync.WaitGroup, spec *core.Spec, q IQueue, c Config) {
	defer func() {
		wg.Done()
		q.Close()
		q.Ready()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			sJs := map[string]bool{}  // Seen jobs; populated off `Unique` jobs only.
			gJs := make([]*genJob, 0) // Final gen. jobs to be executed.

			for _, s := range aggrScopes {
				j := genJob{
					Metadata: Metadata{
						AbsolutePath: s.AbsoluteOutput,
						Inline:       s.Inline,
						DomainType:   s.DomainType,
						Metadata:     *spec.Metadata,
						ScopeKey:     s.Key,
					},
					DisableTemplates: c.DisableTemplates,
				}

				for _, sj := range s.Jobs {
					// This small check prevents the job from being executed for each package.
					if s.DomainType == core.DomainTypeHttp && sj.Unique {
						if ok, _ := sJs[s.Key]; ok {
							return
						}
						sJs[s.Key] = true

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
			l.Log("feed", "msg", "enqueuing jobs", "count", len(gJs))
			for _, j := range gJs {
				if !c.DebugMode {
					// Case: len(jobs) > len(queue)
					// Mark queue as ready once the queue is almost full.
					if q.Size() == c.WorkerCount-1 && len(gJs) > c.WorkerCount {
						q.Ready()
					}
				}
				q.Enqueue(j)
			}

			return
		}
	}
}

func startWorkers(ctx context.Context, l ILogger, count int, wg *sync.WaitGroup, q IQueue, errChan chan<- error) {
	defer wg.Done()

	l.Log("workers", "msg", "starting workers", "count", count)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			for i := 0; i < count; i++ {
				wg.Add(1)
				go worker(ctx, i+1, wg, q, errChan)
			}
			return
		}
	}
}
