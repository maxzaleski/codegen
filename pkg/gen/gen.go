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

type domainScope struct {
	*core.DomainScope
	DomainType core.DomainType
}

func mapToDomainScope(ts []*core.DomainScope, dt core.DomainType) []*domainScope {
	return slice.Map(ts, func(ds *core.DomainScope) *domainScope {
		return &domainScope{
			DomainScope: ds,
			DomainType:  dt,
		}
	})
}

var aggrScopes []*domainScope

type Config struct {
	Location         string
	WorkerCount      int
	DisableTemplates bool
	DebugMode        bool
}

func Execute(rc Config, began time.Time) (md *core.Metadata, _ metrics.IMetrics, _ error) {
	spec, err := core.NewSpec(rc.Location) // Parse configuration via `.codegen` directory.
	md = spec.Metadata
	if err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	c := spec.Config
	hScopes, pScopes := c.HttpDomain.Scopes, c.PkgDomain.Scopes
	aggrScopes = make([]*domainScope, 0, len(hScopes)+len(pScopes))
	aggrScopes = append(aggrScopes, mapToDomainScope(pScopes, core.DomainTypePkg)...)
	aggrScopes = append(aggrScopes, mapToDomainScope(hScopes, core.DomainTypeHttp)...)

	wg, errChan := &sync.WaitGroup{}, make(chan error, 1)
	q, ms := newPool(rc.WorkerCount), metrics.New(nil)

	l := slog.New(rc.DebugMode)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "began", began)
	ctx = context.WithValue(ctx, "logger", l)
	ctx = context.WithValue(ctx, "metrics", ms)

	// Listen for the first error if any.
	go listenForError(errChan, cancel)

	// Make sure the output directories exist.
	walkDirectoryStructure(spec.Metadata, spec.Pkgs, errChan)

	// Feed the queue with jobs.
	wg.Add(1)
	go feedQueue(ctx, wg, spec, q, rc)

	q.WaitReadiness()

	//Start workers.
	wg.Add(1)
	go startWorkers(ctx, rc.WorkerCount, wg, q, errChan)

	wg.Wait()
	close(errChan) // Consequently, the err-listening goroutine will terminate.

	return md, ms, <-errChan
}

func listenForError(errChan chan error, cancel func()) {
	if err := <-errChan; err != nil {
		// Cancel the context, killing all workers.
		cancel()

		// Forward the error to the caller.
		errChan <- err
	}
}

func walkDirectoryStructure(md *core.Metadata, pkgs []*core.Package, errChan chan<- error) {
	for _, s := range aggrScopes {
		for _, j := range s.Jobs {
			s.AbsoluteOutput = md.Cwd + "/" + s.Output
			if s.Inline || j.Unique {
				if err := fs.CreateDirINE(s.AbsoluteOutput); err != nil {
					errChan <- err
					return
				}
				continue
			}

			for _, p := range pkgs {
				if err := fs.CreateDirINE(s.AbsoluteOutput + "/" + p.Name); err != nil {
					errChan <- err
					return
				}
			}
		}
	}
}

func feedQueue(ctx context.Context, wg *sync.WaitGroup, spec *core.Spec, q IQueue, c Config) {
	defer func() {
		wg.Done()
		q.Close()
		q.Ready()
	}()

	// Seen jobs; populated off `ScopeJob.Unique` jobs only.
	sjs := map[string]bool{}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			gJs := make([]*genJob, 0)

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
						if ok, _ := sjs[s.Key]; ok {
							return
						}
						sjs[s.Key] = true

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
			for _, j := range gJs {
				if c.WorkerCount > 1 {
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

func startWorkers(ctx context.Context, count int, wg *sync.WaitGroup, q IQueue, errChan chan<- error) {
	defer wg.Done()

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
