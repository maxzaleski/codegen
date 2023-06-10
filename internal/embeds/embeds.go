package embeds

import (
	"embed"
	"github.com/pkg/errors"
	"text/template"
)

//go:embed templates/*.tmpl
var fs embed.FS

// Link injects the internal utility template for a given extension into the primary one.
func Link(pt *template.Template, ext string) (t *template.Template, err error) {
	t, err = pt.ParseFS(fs, ext+"_embeds.tmpl")
	if err != nil {
		err = errors.Wrapf(err, "failed to parse embeds for ext='%s'", ext)
	}
	return
}
