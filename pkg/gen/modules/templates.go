package modules

import (
	"bytes"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/embeds"
	"github.com/maxzaleski/codegen/internal/fs"
	"github.com/maxzaleski/codegen/internal/lib/slice"
	"github.com/maxzaleski/codegen/pkg/gen/partials"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

type (
	ITemplateProcessor interface {
		Exec(tts []core.ScopeJobTemplate, dtt bool, pkg *core.Package, dest, ext string, fm template.FuncMap) error
	}

	templateProcessor struct {
		pkgs []*core.Package
	}
)

func NewTemplateProcessor(pkgs []*core.Package) ITemplateProcessor {
	return &templateProcessor{
		pkgs: pkgs,
	}
}

func (tp *templateProcessor) Exec(
	tts []core.ScopeJobTemplate, dtt bool, pkg *core.Package, dest, ext string, fm template.FuncMap,
) error {
	// [dev] Execute an empty template.
	if dtt {
		tt, err := template.ParseFS(embeds.FS, "templates/empty.tmpl")
		if err != nil {
			panic("binary corrupted")
		}
		return tp.write(tt, pkg, dest, ext)
	}

	// 1. Define primary and secondary templates.
	parsable, ptt := make([]string, len(tts)), ""
	parsable = append(parsable,
		slice.Map(
			// Filter out the primary template.
			slice.Filter(tts, func(t core.ScopeJobTemplate) bool { return !t.Primary }),
			// Map to template names.
			func(t core.ScopeJobTemplate) string { return t.Name })...,
	)

	// 2. Parse primary template; defines base for all future inclusions.
	tt, err := template.ParseFiles(ptt)
	if err != nil {
		return errors.Wrapf(err, "failed to parse primary template '%s'", ptt)
	}
	// -> Include user-defined secondary templates.
	if tt, err = tt.ParseFiles(parsable...); err != nil {
		return errors.Wrap(err, "failed to parse secondary templates")
	}

	// 3. Append custom functions to template.
	tt.Funcs(fm)
	defer func() {
		if r := recover(); r != nil {
			panic(errors.Wrap(r.(error), "failed to append custom functions to template"))
		}
	}()

	// 4. Write template to disk.
	return tp.write(tt, pkg, dest, ext)
}

func (tp *templateProcessor) write(tt *template.Template, pkg *core.Package, dest, ext string) error {
	funcs := partials.GetByExtension(ext)

	var (
		buf  bytes.Buffer
		data any
	)
	data = tp.pkgs
	if pkg != nil {
		data = pkg
	}

	if err := tt.Funcs(funcs).Execute(&buf, data); err != nil {
		ts := strings.Join(slice.Map(tt.Templates(), func(t *template.Template) string { return t.Name() }), ", ")
		return errors.Wrapf(err, "failed to execute templates '%s'", ts)
	}
	if err := fs.CreateFile(dest, buf.Bytes()); err != nil {
		return err
	}
	return nil
}
