package gen

import (
	"context"
	"sync"

	"github.com/codegen/internal/core"
	"github.com/codegen/internal/fs"
)

// Execute generates the code for the given spec.
func Execute(spec *core.Spec, debug int) (_ Metrics, err error) {
	// Create the output directory if it doesn't exist.
	outPath := spec.GetOutPath()
	if err := fs.CreateDirINE(outPath); err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	errChan := make(chan error, len(spec.Pkgs))

	// Allocate a new (worker) pool with limited seats.
	// This is to address the potential use case of a large number of packages to parse.
	pool := newPool(20, wg)
	ms := newMetrics(nil)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, stateInContextKey, &state{
		paths: &paths{
			CodegenPath: spec.Metadata.DirPath,
			PkgOutPath:  outPath,
		},
		metrics: ms,
	})

	// Listen for the first error if any.
	wg.Add(1)
	go func(wg *sync.WaitGroup, errChan <-chan error, cancel func()) {
		defer wg.Done()

		for err = range errChan {
			if err != nil {
				cancel()
				break
			}
		}
	}(wg, errChan, cancel)

	// Execute package concurrently.
	g := &pkgGenerator{
		ext:    spec.Config.Extension,
		layers: spec.Config.Layers,

		pool:    pool,
		errChan: errChan,
	}
	for _, pkg := range spec.Pkgs {
		pool.Acquire()

		g.Execute(ctx, pkg)
	}

	wg.Wait()
	close(errChan)

	return ms, err
}
