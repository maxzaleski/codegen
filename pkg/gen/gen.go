package gen

import (
	"context"
	"sync"

	"github.com/codegen/internal/core"
)

// generatorState represents the shared generatorState state.
type generatorState struct {
	paths   *paths
	metrics *metrics
}

type paths struct {
	// CodegenPath is the path to the `.codegen` directory.
	CodegenPath string
	// OutputPath is the path to the output directory.
	OutputPath string
}

// Execute generates the code for the given spec.
func Execute(spec *core.Spec) (_ Metrics, err error) {
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
		seen: make(map[string][]*measurement),
	}
	ps := &paths{
		CodegenPath: spec.Paths.DirPath,
		OutputPath:  spec.Paths.Cwd + "/" + spec.Global.Pkg.Output,
	}

	// Generate each package concurrently.
	for _, pkg := range spec.Pkgs {
		wg.Add(1)

		g := &pkgGenerator{
			generatorState: &generatorState{
				paths:   ps,
				metrics: ms,
			},

			metrics: ms,
			gcExt:   spec.Global.Pkg.Extension,
			layers:  spec.Global.Pkg.Layers,

			wg:      wg,
			errChan: errChan,
		}
		go g.Execute(ctx, pkg)
	}
	wg.Wait()
	close(errChan)

	return ms, err
}
