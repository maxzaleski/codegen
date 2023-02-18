package gen

import (
	"context"
	"fmt"
	"sync"

	"github.com/codegen/internal/core"
)

// Execute generates the code for the specified spec.
func Execute(spec *core.Spec) error {
	wg := &sync.WaitGroup{}
	errChan := make(chan error, len(spec.Pkgs))

	ctx, cancel := context.WithCancel(context.Background())

	var err error
	go func(errChan chan error, cancel func()) {
		for err = range errChan {
			if err != nil {
				fmt.Println("Cancelling context due to error:", err)
				cancel()
				break
			}
		}
	}(errChan, cancel)

	// Generate each package concurrently.
	for _, pkg := range spec.Pkgs {
		wg.Add(1)

		g := &pkgGenerator{
			wg:      wg,
			errChan: errChan,

			gcExt:      spec.Global.Pkg.Extension,
			ourDirPath: spec.DirPath,
			outPath:    spec.Cwd + "/" + spec.Global.Pkg.Output,
			layers:     spec.Global.Pkg.Layers,
		}
		go g.Execute(ctx, pkg)
	}
	wg.Wait()
	close(errChan)

	return nil
}
