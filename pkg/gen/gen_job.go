package gen

import (
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/utils/moddedstring"
	"strings"
)

type (
	genJob struct {
		*core.ScopeJob

		OutputFile       *genJobFile
		Metadata         metadata
		Package          *core.Package
		DisableTemplates bool
	}

	genJobFile struct {
		Name         string
		Ext          string
		AbsolutePath string
	}

	metadata struct {
		core.Metadata

		ScopeKey     string
		DomainType   core.DomainType
		AbsolutePath string
		Inline       bool
	}
)

// Fill sets the output file name and extension, and the output file absolute path.
func (j *genJob) Fill() (err error) {
	if j.OutputFile == nil {
		j.OutputFile = &genJobFile{}
	}

	// [1] Set output file name and extension.
	cs, of := strings.Split(j.FileName, "."), j.OutputFile
	of.Ext = cs[len(cs)-1]

	tm := map[string]string{}
	if j.Package != nil {
		tm["pkg"] = j.Package.Name
	}
	if of.Name, err = moddedstring.New(j.FileName, tm); err != nil {
		return
	}

	// [2] Set output file absolute path.
	fn, md, pkg := of.Name, j.Metadata, j.Package
	of.AbsolutePath = md.AbsolutePath + "/"

	// (i) Inline: files are generated within the same directory space (e.g. models > User.Java, Car.Java).
	// (i) Unique: job is only to be performed once for the specified output.
	if md.Inline || j.Unique {
		of.AbsolutePath += fn
	} else if pkg != nil {
		of.AbsolutePath += pkg.Name + "/" + fn
	}

	return
}
