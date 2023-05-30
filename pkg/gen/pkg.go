package gen

import (
	"context"
	"os"

	"github.com/codegen/internal/core"
	"github.com/codegen/internal/fs"
	"github.com/pkg/errors"
)

// pkgGenerator encapsulates the logic required to generate a package.
type pkgGenerator struct {
	ext  string
	jobs []*core.GenerationJob

	errChan chan<- error
	pool    Pool
}

func (pg *pkgGenerator) Execute(ctx context.Context, pkg *core.Package) {
	defer pg.pool.Release()

	state := getStateFromContext(ctx)

	// Check the presence of the specified package directory.
	pkgPath := state.paths.PkgOutPath + "/" + pkg.Name
	if _, err := os.Stat(pkgPath); err != nil {
		if !(os.IsNotExist(err)) {
			pg.error(err, "failed to check presence of pkg directory '%s'", pkg.Name)
			return
		}
		// If the directory doesn't exist, create it.
		if err := fs.CreateDir(pkgPath); err != nil {
			pg.error(err, "")
		}
	}

	// Execute each job.
	g := &jobGenerator{
		state:   state,
		pkg:     pkg,
		fileExt: pg.ext,
	}
	for _, job := range pg.jobs {
		if job.Exclude.Contains(pkg.Name) {
			continue
		}
		if err := g.Execute(ctx, job); err != nil {
			pg.error(err, "failed to execute job '%s'", job.Key)
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
