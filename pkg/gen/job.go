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
	"github.com/codegen/internal/fs"
	"github.com/codegen/internal/presets"
	"github.com/pkg/errors"
)

type jobGenerator struct {
	*state

	pkg     *core.Package
	fileExt string
}

func (sg *jobGenerator) Execute(_ context.Context, s *core.GenerationJob) error {
	ext := sg.fileExt
	if s.Lang != "" {
		ext = s.Lang
	}
	destPath := fmt.Sprintf("%s/%s/%s.%s", sg.paths.PkgOutPath, sg.pkg.Name, s.Destination, ext)

	mi := sg.metrics.NewIntent(sg.pkg.Name, &Measurement{FileName: s.Key, Path: destPath})

	// [1] Check for the presence of the job file.
	//
	// The aim is not to overwrite existing files.
	if _, err := os.Stat(destPath); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to check presence of job file '%s'", s.Destination)
		}
	} else {
		mi.Measure(&Measurement{Created: false})
		return nil
	}

	// The first template must always be the primary template.
	//
	// We instantiate the slice with a single element to ensure that the primary template is always
	// the first element; the second being the embeds template if available.
	tmpls := make([]string, 1, 2)

	pfs, err := presets.GetFS(sg.fileExt)
	if err != nil {
		return errors.WithMessage(err, "failed to get template (presets) filesystem")
	}

	// [2] Check for the presence of the 'embeds' template.
	//
	// This template is used to provide utility functions/layouts (via the "template" keyword) to the
	// primary template. The tool aims to package these utilities by default if available.
	if !s.DisableEmbeds {
		embedsTmpl, err := sg.checkEmbeds(pfs)
		if err != nil {
			return err
		}
		if embedsTmpl != "" {
			tmpls = append(tmpls, embedsTmpl)
		}
	}

	// [3] Parse the templates into a single executable set.
	//
	// The use of internal templates is determined by the presence of the "preset." prefix.
	var ts *template.Template
	if tmpl := s.Template; strings.HasPrefix(tmpl, presets.SpecPfx+".") {
		tmpl = strings.TrimPrefix(tmpl, presets.SpecPfx+".")
		ts, err = sg.withPresetTmpl(pfs, tmpl, tmpls)
	} else {
		ts, err = sg.withCustomTmpl(pfs, tmpl, tmpls)
	}
	if err != nil {
		return err
	}

	// [4] Execute the template set.
	if err := sg.write(ts, destPath); err != nil {
		return err
	}
	mi.Measure(&Measurement{Created: true})

	return nil
}

// viaPreset returns a template set with the specified local template.
func (sg *jobGenerator) withCustomTmpl(fs embed.FS, tmpl string, tmpls []string) (*template.Template, error) {
	tmpls[0] = fmt.Sprintf("%s/templates/%s.tmpl", sg.paths.CodegenPath, tmpl)

	ts, err := template.ParseFiles(tmpls[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse local template")
	}
	ts, err = ts.ParseFS(fs, tmpls[1])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse embeds template")
	}

	return ts, nil
}

// withPresetTmpl returns a template set with the specified preset template.
func (sg *jobGenerator) withPresetTmpl(fs embed.FS, tmpl string, tmpls []string) (*template.Template, error) {
	tmpls[0] = fmt.Sprintf("templates/%s/%s.tmpl", sg.fileExt, tmpl)
	ts, err := template.ParseFS(fs, tmpls...)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to parse templates '[%s]'", strings.Join(tmpls, ", "))
	}
	return ts, nil
}

// write writes the generated code to the destination file.
func (sg *jobGenerator) write(ts *template.Template, dest string) error {
	var buf bytes.Buffer
	if err := ts.Execute(&buf, sg.pkg); err != nil {
		return errors.Wrapf(err, "failed to execute template '%s'", ts.Name())
	}
	if err := fs.CreateFile(dest, bytes.TrimSpace(buf.Bytes())); err != nil {
		return err
	}
	return nil
}

// checkEmbeds checks if the embeds template exists for the current file extension.
//
// If it does, it returns the path to the template.
func (sg *jobGenerator) checkEmbeds(fs embed.FS) (string, error) {
	path := fmt.Sprintf("templates/%s/%s.tmpl", sg.fileExt, presets.EmbedsPfx)
	if b, err := fs.ReadFile(path); err != nil {
		return "", errors.Wrapf(err, "failed to read '%s' template", presets.EmbedsPfx)
	} else if len(b) == 0 {
		return "", nil
	}
	return path, nil
}
