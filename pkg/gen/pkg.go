package gen

import (
	"context"
	"os"
	"sync"

	"github.com/codegen/internal/core"
	"github.com/pkg/errors"
)

const internalTemplatesPath = "internal/templates/"

// pkgGenerator encapsulates the logic required to generate a pkg.
type pkgGenerator struct {
	*generatorState

	metrics *metrics
	gcExt   string
	layers  []*core.Layer

	wg      *sync.WaitGroup
	errChan chan<- error
}

func (g *pkgGenerator) Execute(ctx context.Context, pkg *core.Pkg) {
	defer g.wg.Done()

	// Check the presence of the specified package directory.
	pkgPath := g.paths.OutputPath + "/" + pkg.Name
	if _, err := os.Stat(pkgPath); err != nil {
		if !(os.IsNotExist(err)) {
			g.error(err, "failed to check presence of pkg '$%s'", pkg.Name)
			return
		}
		// If the directory doesn't exist, create it.
		os.Mkdir(pkgPath, 0777)
	}

	fileExt := g.gcExt
	if e := pkg.Extension; e != "" {
		fileExt = e
	}

	// Generate each layer.
	lg := &layerGenerator{
		generatorState: g.generatorState,
		pkg:            pkg,
		fileExt:        fileExt,
	}
	for _, l := range g.layers {
		if err := lg.Execute(ctx, l); err != nil {
			g.error(err, "failed to execute layer '%s'", l.Name)
			return
		}
	}
}

func (pg *pkgGenerator) error(err error, msg string, args ...interface{}) {
	if len(args) > 0 {
		err = errors.WithMessagef(err, msg, args...)
	} else {
		err = errors.WithMessage(err, msg)
	}
	pg.errChan <- err
}
