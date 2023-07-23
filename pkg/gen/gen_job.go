package gen

import (
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/fs"
	"github.com/maxzaleski/codegen/internal/lib/moddedstring"
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
		Name            string
		Ext             string
		AbsolutePath    string
		AbsoluteDirPath string
	}

	metadata struct {
		core.Metadata

		ScopeKey   string
		DomainType core.DomainType
		Inline     bool
	}
)

const tokenPkg = "pkg"

// Prepare prepares the job for execution by filling-in missing fields, and verifying output directory structure.
func (j *genJob) Prepare() (err error) {
	if err = j.fill(); err != nil {
		return
	}
	if err = j.checkOutputDirPresence(map[string]any{}); err != nil {
		return
	}
	return
}

func (j *genJob) fill() (err error) {
	// [1] Set output file name and extension.
	cs, f := strings.Split(j.FileName, "."), j.OutputFile
	f.Ext = cs[len(cs)-1]

	tm := map[string]string{}
	if j.Package != nil {
		tm[tokenPkg] = j.Package.Name
	}
	if f.Name, err = moddedstring.New(j.FileName, tm); err != nil {
		return
	}

	// [2] Set output file absolute path.
	fn, md, pkg := f.Name, j.Metadata, j.Package
	f.AbsolutePath = f.AbsoluteDirPath + "/"

	// (i) Inline: files are generated within the same directory space (e.g. models > User.Java, Car.Java).
	// (i) Unique: job is only to be performed once for the specified output.
	if md.Inline || j.Unique {
		f.AbsolutePath += fn
	} else if pkg != nil {
		f.AbsoluteDirPath += "/" + pkg.Name
		f.AbsolutePath += pkg.Name + "/" + fn
	}

	return
}

func (j *genJob) checkOutputDirPresence(seenMap map[string]any) error {
	if j.Unique || j.Metadata.Inline {
		key := j.Metadata.ScopeKey + "/" + j.OutputFile.Name
		if _, ok := seenMap[key]; ok {
			return nil
		}
		seenMap[key] = nil
	}

	return fs.CreateDirINE(j.OutputFile.AbsoluteDirPath)
}
