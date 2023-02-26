package gen

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/codegen/internal/core"
	"github.com/codegen/internal/presets"
	"github.com/pkg/errors"
)

type layerGenerator struct {
	*generatorState

	pkg     *core.Pkg
	fileExt string
}

func (lg *layerGenerator) Execute(_ context.Context, l *core.Layer) error {
	fileName := fmt.Sprintf("%s.%s", l.FileName, lg.fileExt)
	destPath := fmt.Sprintf("%s/%s/%s", lg.paths.OutputPath, lg.pkg.Name, fileName)

	// We only care to continue if the file doesn't exist.
	if _, err := os.Stat(destPath); err != nil {
		if !os.IsNotExist(err) {
			return errors.WithMessagef(err, "failed to check presence of layer file '%s'", l.FileName)
		}
	} else {
		lg.metrics.Measure(lg.pkg.Name, &measurement{File: fileName, Created: false})
		return nil
	}

	// /!\
	// The first [0] template must be the primary template (the one rendered).
	tmpls := make([]string, 1, 2)

	fs, err := presets.GetTemplateFS(lg.fileExt)
	if err != nil {
		return errors.WithMessagef(err, "failed to get template FS")
	}
	embedsTmpl, err := lg.checkEmbeds(fs)
	if err != nil {
		return err
	}
	if embedsTmpl != "" {
		tmpls = append(tmpls, embedsTmpl)
	}

	var ts *template.Template
	if tmpl := l.Template; strings.HasPrefix(tmpl, presets.SpecPfx+".") {
		tmpl = strings.TrimPrefix(tmpl, presets.SpecPfx+".")
		ts, err = lg.viaPreset(fs, tmpl, tmpls)
	} else {
		ts, err = lg.viaCustom(fs, tmpl, tmpls)
	}
	if err != nil {
		return err
	}

	if err := lg.write(ts, destPath); err != nil {
		return err
	}
	lg.metrics.Measure(lg.pkg.Name, &measurement{File: fileName, Created: true})

	return nil
}

// viaPreset returns a template set with the specified local template.
func (lg *layerGenerator) viaCustom(fs embed.FS, tmpl string, tmpls []string) (*template.Template, error) {
	tmpls[0] = fmt.Sprintf("%s/templates/%s.tmpl", lg.paths.CodegenPath, tmpl)

	ts, err := template.ParseFiles(tmpls[0])
	if err != nil {
		return nil, errors.WithMessage(err, "viaCustom: failed to parse local templates")
	}
	ts, err = ts.ParseFS(fs, tmpls[1])
	if err != nil {
		return nil, errors.WithMessage(err, "viaCustom: failed to parse embeds template")
	}

	return ts, nil
}

// viaPreset returns a template set with the specified preset.
func (lg *layerGenerator) viaPreset(fs embed.FS, tmpl string, tmpls []string) (*template.Template, error) {
	tmpls[0] = fmt.Sprintf("templates/%s/%s.tmpl", lg.fileExt, tmpl)
	ts, err := template.ParseFS(fs, tmpls...)
	if err != nil {
		return nil, errors.WithMessage(err, "viaPreset: failed to parse templates")
	}
	return ts, nil
}

// write writes the generated code to the destination file.
func (lg *layerGenerator) write(ts *template.Template, dest string) error {
	var buf bytes.Buffer
	if err := ts.Execute(&buf, lg.pkg); err != nil {
		return errors.WithMessage(err, "failed to execute template")
	}
	if err := createFile(dest, bytes.TrimSpace(buf.Bytes())); err != nil {
		return errors.WithMessagef(err, "failed to create layer file at '%s'", dest)
	}
	return nil
}

// checkEmbeds checks if the embeds template exists for the current file extension.
//
// If it does, it returns the path to the template.
func (lg *layerGenerator) checkEmbeds(fs embed.FS) (string, error) {
	path := fmt.Sprintf("templates/%s/%s.tmpl", lg.fileExt, presets.EmbedsPfx)
	if b, err := fs.ReadFile(path); err != nil {
		return "", errors.WithMessage(err, "failed to read embeds template")
	} else if len(b) == 0 {
		return "", nil
	}
	return path, nil
}
