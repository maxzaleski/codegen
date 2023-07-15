package gen

import (
	"bytes"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/embeds"
	"github.com/maxzaleski/codegen/internal/fs"
	"github.com/maxzaleski/codegen/internal/utils/slice"
	"github.com/maxzaleski/codegen/pkg/gen/partials"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

type (
	templateFactory struct {
		j       *genJob
		pkgs    []*core.Package
		funcMap template.FuncMap
	}
)

func (tf templateFactory) ExecuteTemplate() error {
	jts := tf.j.Templates

	// [dev] Execute an empty template.
	if tf.j.DisableTemplates {
		tt, err := template.ParseFS(embeds.FS, "templates/empty.tmpl")
		if err != nil {
			panic("binary corrupted")
		}
		return tf.write(tt)
	}

	// 1. Define primary and secondary templates.
	ts, pt := make([]string, 1, len(jts)), ""

	// 2. Parse primary template; defines base for all future inclusions.
	tt, err := template.ParseFiles(pt)
	if err != nil {
		return errors.Wrapf(err, "failed to parse primary template '%s'", pt)
	}
	// -> Include user-defined secondary templates.
	if tt, err = tt.ParseFiles(ts...); err != nil {
		return errors.Wrap(err, "failed to parse secondary templates")
	}

	return tf.write(tt)
}

func (tf templateFactory) write(tt *template.Template) error {
	funcs := partials.GetByExtension(tf.j.OutputFile.Ext)

	var (
		buf  bytes.Buffer
		data any
	)
	data = tf.pkgs
	if tf.j.Package != nil {
		data = tf.j.Package
	}

	if err := tt.Funcs(funcs).Execute(&buf, data); err != nil {
		ts := strings.Join(slice.Map(tt.Templates(), func(t *template.Template) string { return t.Name() }), ", ")
		return errors.Wrapf(err, "failed to execute templates '%s'", ts)
	}
	if err := fs.CreateFile(tf.j.OutputFile.AbsolutePath, buf.Bytes()); err != nil {
		return err
	}
	return nil
}
