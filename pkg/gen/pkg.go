package gen

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/codegen/internal/core"
	"github.com/pkg/errors"
)

const internalTemplatesPath = "internal/templates/"

// pkgGenerator encapsulates the logic required to generate a pkg.
type pkgGenerator struct {
	gcExt      string
	ourDirPath string
	outPath    string
	layers     []*core.Layer

	wg      *sync.WaitGroup
	errChan chan<- error
}

func (pg *pkgGenerator) Execute(_ context.Context, pkg *core.Pkg) {
	defer pg.wg.Done()

	// Check the presence of the specified package directory.
	pkgPath := pg.outPath + "/" + pkg.Name
	if f, err := os.Stat(pkgPath); err != nil {
		if !(os.IsNotExist(err)) {
			pg.error(err, "failed to check presence of pkg '$%s'", pkg.Name)
		}
		// if the directory doesn't exist, create it.
		os.Mkdir(pkgPath, 0777)
	} else if f.IsDir() {
		// We don't override existing files.
		pg.errChan <- nil
		return
	}

	fileExt := pg.gcExt
	if e := pkg.Extension; e != "" {
		fileExt = e
	}

	// Generate each layer.
	for _, l := range pg.layers {
		if err := pg.executeLayer(l, fileExt, pkg); err != nil {
			pg.error(err, "failed to execute layer '$%s'", l.Name)
		}
	}
}

func (g *pkgGenerator) executeLayer(l *core.Layer, ext string, data *core.Pkg) error {
	// [setup]
	// • Establish layer template.
	tmpl := l.Template
	if strings.HasPrefix(tmpl, "default.") { // internal template prefix.
		tmpl = internalTemplatesPath + ext + "/" + strings.TrimPrefix(tmpl, "default.")
	} else {
		tmpl = g.ourDirPath + "/templates/" + l.Template
	}

	tmpls := make([]string, 0, 2)
	tmpls = append(tmpls, tmpl+".tmpl")

	// [embeds]
	// • Check that the (internal) 'embeds' template exists.
	embedsTmplPath := internalTemplatesPath + ext + "/embeds.tmpl"
	if _, err := os.Stat(embedsTmplPath); err != nil && !os.IsNotExist(err) {
		return errors.WithMessage(err, "failed to check presence of 'embeds' template")
	}
	// • Template exists; location is safe to open.
	tmpls = append(tmpls, embedsTmplPath)

	// [execution]
	// • Parse the templates.
	ts, err := template.ParseFiles(tmpls...)
	if err != nil {
		return errors.WithMessage(err, "failed to parse templates")
	}
	// • Execute the templates.
	var buf bytes.Buffer
	if err := ts.Execute(&buf, data); err != nil {
		return errors.WithMessage(err, "failed to execute templates")
	}
	// • Write the generated code.
	// current: service.go
	path := fmt.Sprintf("%s/%s/%s.%s", g.outPath, data.Name, l.FileName, ext)
	if err := createFile(path, bytes.TrimSpace(buf.Bytes())); err != nil {
		return errors.WithMessage(err, "failed to create layer file")
	}

	return nil
}

func (pg *pkgGenerator) error(err error, msg string, args ...interface{}) {
	if len(args) > 0 {
		err = errors.WithMessagef(err, msg, args...)
	} else {
		err = errors.WithMessage(err, msg)
	}
	pg.errChan <- err
}
