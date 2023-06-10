package gen

import (
	"context"
	"github.com/codegen/internal/metrics"
	"github.com/codegen/pkg/slog"
	"github.com/pkg/errors"
	"sync"

	"github.com/codegen/internal/core"
	"github.com/codegen/internal/fs"
)

var aggrScopes []*core.DomainScope

func Execute(sl string, workers int, logger slog.ILogger) (md *core.Metadata, _ metrics.IMetrics, err error) {
	spec, err := core.NewSpec(sl) // Parse configuration via `.codegen` directory.
	md = spec.Metadata
	if err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	c := spec.Config
	tScopes, pScopes := c.HttpDomain.Scopes, c.PkgDomain.Scopes
	aggrScopes = make([]*core.DomainScope, 0, len(tScopes)+len(pScopes))
	aggrScopes = append(aggrScopes, pScopes...)
	aggrScopes = append(aggrScopes, tScopes...)

	wg, errChan := &sync.WaitGroup{}, make(chan error, 1)
	q, ms := newPool(workers), metrics.New(nil)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "logger", logger)
	ctx = context.WithValue(ctx, "metrics", ms)

	// Listen for the first error if any.
	go func(errChan <-chan error, cancel func()) {
		for err = range errChan {
			if err != nil {
				cancel()
				break
			}
		}
	}(errChan, cancel)

	// Make sure the output directories exist.
	wg.Add(1)
	go walkDirectoryStructure(ctx, spec.Metadata, wg, errChan)

	// Feed the queue with jobs.
	wg.Add(1)
	go feedQueue(ctx, wg, spec, q)

	// Start workers.
	wg.Add(1)
	go startWorkers(ctx, workers, wg, q, errChan)

	wg.Wait()
	close(errChan) // Consequently, the err-listening goroutine will terminate.

	return md, ms, err
}

func walkDirectoryStructure(_ context.Context, md *core.Metadata, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	for _, s := range aggrScopes {
		s.AbsoluteOutput = md.Cwd + "/" + s.Output
		if err := fs.CreateDirINE(s.AbsoluteOutput); err != nil {
			errChan <- err
			return
		}
	}
}

func feedQueue(_ context.Context, wg *sync.WaitGroup, spec *core.Spec, q IQueue) {
	defer wg.Done()

	for _, s := range aggrScopes {
		for _, p := range spec.Pkgs {
			for _, j := range s.Jobs {
				q.Enqueue(&job{
					ScopeJob: j,
					Package:  p,
					Metadata: Metadata{
						Metadata:     *spec.Metadata,
						ScopeKey:     s.Key,
						AbsolutePath: s.AbsoluteOutput,
						Inline:       s.Inline,
					},
				})
			}
		}
	}
	q.Close()
}

func startWorkers(ctx context.Context, count int, wg *sync.WaitGroup, q IQueue, errChan chan<- error) {
	defer wg.Done()

	for i := 0; i < count; i++ {
		wg.Add(1)
		go worker(ctx, wg, q, errChan)
	}
}
