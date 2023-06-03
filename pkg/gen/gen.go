package gen

import (
	"context"
	"github.com/codegen/pkg/slog"
	"github.com/pkg/errors"
	"sync"

	"github.com/codegen/internal/core"
	"github.com/codegen/internal/fs"
)

// Execute generates the code for the given spec.
func Execute(loc string, logger slog.Logger) (md *core.Metadata, _ Metrics, err error) {
	// Parse configuration via `.codegen` directory.
	spec, err := core.NewSpec(loc)
	md = spec.Metadata
	if err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	// Create the output directory if it doesn't exist.
	outPath := spec.GetPkgDomainOutPath()
	if err = fs.CreateDirINE(outPath); err != nil {
		return
	}

	errChan := make(chan error, len(spec.Pkgs))
	ms := newMetrics(nil)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, stateInContextKey, &state{
		paths: &paths{
			CodegenPath: spec.Metadata.CodegenDir,
			PkgOutPath:  outPath,
		},
		metrics: ms,
	})

	// Listen for the first error if any.
	go func(errChan <-chan error, cancel func()) {
		for err = range errChan {
			if err != nil {
				cancel()
				break
			}
		}
	}(errChan, cancel)

	wg := &sync.WaitGroup{}
	// Allocate a new (worker) pool with limited seats.
	// This is to address the potential use case of a large number of packages to parse.
	pool := newPool(20, wg)

	// Execute package concurrently.
	g := &pkgGenerator{
		ext:  spec.Config.Lang,
		jobs: spec.Config.Jobs,

		pool:    pool,
		errChan: errChan,
	}
	for _, pkg := range spec.Pkgs {
		pool.Acquire()

		g.Execute(ctx, pkg)
	}

	wg.Wait()
	close(errChan) // Consequently, the err-listening goroutine will terminate.

	return md, ms, err
}
