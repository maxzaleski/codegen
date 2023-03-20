package presets

import (
	"embed"

	"github.com/pkg/errors"
)

//go:embed templates/go/*.tmpl
var goTemplates embed.FS

// EmbedsPfx is the universal name give to the embeds template.
// This file, if present for the given extension, will always be included.
const EmbedsPfx = "_embeds"

// SpecPfx is the prefix used to retrieve internal templates from the spec file.
const SpecPfx = "presets"

// GetFS returns the template file system for the given extension.
func GetFS(ext string) (fs embed.FS, err error) {
	switch ext {
	case "go":
		fs = goTemplates
	default:
		err = errors.Wrapf(err, "unsupported template extension: '%s'", ext)
	}
	return
}
