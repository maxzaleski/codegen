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
	*state

	pkg     *core.Pkg
	fileExt string
}

func (lg *layerGenerator) Execute(ctx context.Context, l *core.Layer) error {
	// [1] Check for the presence of the layer file.
	//
	// The aim is not to overwrite existing files.
	fileName := fmt.Sprintf("%s.%s", l.FileName, lg.fileExt)
	destPath := fmt.Sprintf("%s/%s/%s", lg.paths.OutputPath, lg.pkg.Name, fileName)
	if _, err := os.Stat(destPath); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to check presence of layer file '%s'", l.FileName)
		}
	} else {
		lg.metrics.Measure(lg.pkg.Name, &measurement{Key: fileName, Created: false})
		return nil
	}

	// The first template must always be the primary template.
	//
	// We instantiate the slice with a single element to ensure that the primary template is always
	// the first element; the second being the embeds template if available.
	tmpls := make([]string, 1, 2)

	fs, err := presets.GetFileSystem(lg.fileExt)
	if err != nil {
		return errors.Wrapf(err, "failed to get template FS")
	}

	// [2] Check for the presence of the embeds template.
	//
	// This template is used to provide utility functions/layouts (via the "template" keyword) to the
	// primary template. The tool aims to package these utilities by default if available.
	if !l.DisableEmbeds {
		embedsTmpl, err := lg.checkEmbeds(fs)
		if err != nil {
			return err
		}
		if embedsTmpl != "" {
			tmpls = append(tmpls, embedsTmpl)
		}
	}
	// [3] Parse the various templates into a single executable template set.
	//
	// The use of internal templates is determined by the presence of the "preset." prefix.
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
	// [4] Execute the template set.
	if err := lg.write(ts, destPath); err != nil {
		return err
	}
	lg.metrics.Measure(lg.pkg.Name, &measurement{Key: fileName, Created: true})

	return nil
}

// viaPreset returns a template set with the specified local template.
func (lg *layerGenerator) viaCustom(fs embed.FS, tmpl string, tmpls []string) (*template.Template, error) {
	tmpls[0] = fmt.Sprintf("%s/templates/%s.tmpl", lg.paths.CodegenPath, tmpl)

	ts, err := template.ParseFiles(tmpls[0])
	if err != nil {
		return nil, errors.Wrap(err, "viaCustom: failed to parse local templates")
	}
	ts, err = ts.ParseFS(fs, tmpls[1])
	if err != nil {
		return nil, errors.Wrap(err, "viaCustom: failed to parse embeds template")
	}

	return ts, nil
}

// viaPreset returns a template set with the specified preset template.
func (lg *layerGenerator) viaPreset(fs embed.FS, tmpl string, tmpls []string) (*template.Template, error) {
	tmpls[0] = fmt.Sprintf("templates/%s/%s.tmpl", lg.fileExt, tmpl)
	ts, err := template.ParseFS(fs, tmpls...)
	if err != nil {
		return nil, errors.Wrapf(err,
			"viaPreset: failed to parse templates '[%s]'", strings.Join(tmpls, ", "))
	}
	return ts, nil
}

// write writes the generated code to the destination file.
func (lg *layerGenerator) write(ts *template.Template, dest string) error {
	var buf bytes.Buffer
	if err := ts.Execute(&buf, lg.pkg); err != nil {
		return errors.Wrapf(err, "failed to execute template '%s'", ts.Name())
	}
	if err := createFile(dest, bytes.TrimSpace(buf.Bytes())); err != nil {
		return err
	}
	return nil
}

// checkEmbeds checks if the embeds template exists for the current file extension.
//
// If it does, it returns the path to the template.
func (lg *layerGenerator) checkEmbeds(fs embed.FS) (string, error) {
	path := fmt.Sprintf("templates/%s/%s.tmpl", lg.fileExt, presets.EmbedsPfx)
	if b, err := fs.ReadFile(path); err != nil {
		return "", errors.Wrapf(err, "failed to read '%s' template", presets.EmbedsPfx)
	} else if len(b) == 0 {
		return "", nil
	}
	return path, nil
}
