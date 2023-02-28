package gen

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/codegen/internal/core"
)

// Execute generates the code for the given spec.
func Execute(spec *core.Spec) (_ Metrics, err error) {
	// Create the output directory if it doesn't exist.
	outPath := spec.Paths.Cwd + "/" + spec.Global.Pkg.Output
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		os.Mkdir(outPath, 0777)
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
		seen: make(map[string][]*measurement),
	}
	ctx = context.WithValue(ctx, stateInContextKey, &state{
		paths: &paths{
			CodegenPath: spec.Paths.DirPath,
			OutputPath:  outPath,
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

	// Create error log file.
	if err != nil {
		createFile(spec.Paths.Cwd+"/codegen_error.log", []byte(fmt.Sprintf("%+v", err)))
	}

	return ms, err
}
