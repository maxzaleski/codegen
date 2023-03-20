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
	outPath := spec.Paths.Cwd + "/" + spec.Global.Pkg.Output
	if err := fs.CreateDirINE(outPath); err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	errChan := make(chan error, len(spec.Pkgs))

	ctx, cancel := context.WithCancel(context.Background())

	// Listen for the first error if any.
	go func(errChan <-chan error, cancel func()) {
		for err = range errChan {
			if err != nil {
				cancel()
				break
			}
		}
	}(errChan, cancel)

	ms := &metrics{
		mu:   &sync.Mutex{},
		seen: make(map[string][]*Measurement),
	}
	ctx = context.WithValue(ctx, stateInContextKey, &state{
		paths: &paths{
			CodegenPath: spec.Paths.DirPath,
			PkgOutPath:  outPath,
		},
		metrics: ms,
	})

	// Generate each package concurrently.
	for _, pkg := range spec.Pkgs {
		wg.Add(1)

		g := &pkgGenerator{
			ext:    spec.Global.Pkg.Extension,
			layers: spec.Global.Pkg.Layers,

			wg:      wg,
			errChan: errChan,
		}
		go g.Execute(ctx, pkg)
	}
	wg.Wait()
	close(errChan)

	return ms, err
}
