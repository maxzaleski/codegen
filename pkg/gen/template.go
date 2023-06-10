package gen

import (
	"bytes"
	"github.com/codegen/internal/embeds"
	"github.com/codegen/internal/fs"
	"github.com/codegen/internal/utils"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

type (
	templateFactory struct {
		j *job
	}
)

func (tf templateFactory) ExecuteTemplate() error {
	jts, ext := tf.j.Templates, tf.j.FileName.Extension

	// 1. Define primary and secondary templates.
	ts, pt := make([]string, 1, len(jts)), ""
	for _, t := range jts {
		if t.Primary {
			pt = t.Name
		} else {
			ts = append(ts, t.Name)
		}
	}
	if pt == "" {
		return errors.New("failed to specify provide a primary template")
	}

	// 2. Parse primary template; defines base for all future inclusions.
	tt, err := template.ParseFiles(pt)
	if err != nil {
		return errors.Wrapf(err, "failed to parse primary template '%s'", pt)
	}
	// -> Include internal secondary template.
	if !tf.j.DisableEmbeds {
		if tt, err = embeds.Link(tt, ext); err != nil {
			return err
		}
	}
	// -> Include user-defined secondary templates.
	if tt, err = tt.ParseFiles(ts...); err != nil {
		return errors.Wrap(err, "failed to parse secondary templates")
	}

	return tf.write(tt)
}

func (tf templateFactory) write(tt *template.Template) error {
	var buf bytes.Buffer
	if err := tt.Execute(&buf, tf.j.Package); err != nil {
		ts := strings.Join(utils.Map(tt.Templates(), func(t *template.Template) string { return t.Name() }), ", ")
		return errors.Wrapf(err, "failed to execute templates '%s'", ts)
	}
	if err := fs.CreateFile(tf.j.FileAbsolutePath, buf.Bytes()); err != nil {
		return err
	}
	return nil
}
