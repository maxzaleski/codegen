package gen

import (
	"context"
	"os"
	"sync"

	"github.com/codegen/internal/core"
	"github.com/codegen/internal/fs"
	"github.com/pkg/errors"
)

// pkgGenerator encapsulates the logic required to generate a pkg.
type pkgGenerator struct {
	ext    string
	layers []*core.Layer

	wg      *sync.WaitGroup
	errChan chan<- error
}

func (g *pkgGenerator) Execute(ctx context.Context, pkg *core.Pkg) {
	defer g.wg.Done()

	state := getStateFromContext(ctx)

	// Check the presence of the specified package directory.
	pkgPath := state.paths.PkgOutPath + "/" + pkg.Name
	if _, err := os.Stat(pkgPath); err != nil {
		if !(os.IsNotExist(err)) {
			g.error(err, "failed to check presence of pkg directory '%s'", pkg.Name)
			return
		}
		// If the directory doesn't exist, create it.
		if err := fs.CreateDir(pkgPath); err != nil {
			g.error(err, "")
		}
	}

	fileExt := g.ext
	if e := pkg.Extension; e != "" {
		fileExt = e
	}

	// Execute each layer.
	lg := &layerGenerator{
		state:   state,
		pkg:     pkg,
		fileExt: fileExt,
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
		err = errors.Wrapf(err, msg, args...)
	} else if msg != "" {
		err = errors.Wrap(err, msg)
	}
	pg.errChan <- err
}
