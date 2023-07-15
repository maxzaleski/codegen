package core

import (
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

const (
	DomainDir   = ".codegen"
	domainEntry = "config.yaml"

	event = "parsing"
)

var validate = newValidator()

// NewSpec parses the .codegen directory and returns a `Spec`.
func NewSpec(rl slog.ILogger, src string) (spec *Spec, err error) {
	spec = newSpec() // Always returned.

	l := slog.NewNamed(rl, "core", "")
	l.Log(event, "msg", "parsing configuration")
	defer func() {
		if err == nil {
			l.Log(event, "msg", "parsing complete", "packages", len(spec.Pkgs))
		}
	}()

	// Establish presence of configuration directory.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if src != "" {
		cwd += "/" + src
	}
	cdp := cwd + "/" + DomainDir

	l.Log(event, "msg", "locating "+DomainDir, "location", cdp)

	if _, err = os.Stat(cdp); os.IsNotExist(err) {
		err = errors.Wrapf(err, "failed to locate '%s' directory", DomainDir)
		return
	}
	spec.Metadata.CodegenDir = cdp
	spec.Metadata.Cwd = cwd

	// Parse generator specification.
	path := cdp + "/config.yaml"
	l.Log(event, "msg", "parsing primary configuration", "path", path)
	if err = unmarshal(path, spec.Config, true); err != nil {
		return
	}

	// Parse all packages.
	//
	// /!\ Assumes a flat directory structure.
	err = filepath.Walk(cdp+"/pkg", func(path string, info os.FileInfo, err error) error {
		// Handle unexpected error.
		if err != nil {
			return errors.Wrapf(err, "unexpected error during dir walk at file '%s'", path)
		}
		// Skip non-YAML files.
		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		l.Log(event, "msg", "parsing package", "path", path)

		pkg := &Package{}
		if err = unmarshal(path, pkg, false); err != nil {
			return err
		}
		// Primary method arguments by `index` field.
		if len(pkg.Models) != 0 {
			for _, m := range pkg.Models {
				if len(m.Methods) != 0 {
					for _, m := range m.Methods {
						m.SortParams()
					}
				}
			}
			for _, m := range pkg.Interface.Methods {
				m.SortParams()
			}
		}
		spec.Pkgs = append(spec.Pkgs, pkg)

		return nil
	})
	if err != nil {
		err = errors.Wrap(err, "failed to walk configuration dir")
		return
	}

	// Assign domain types.
	for _, s := range spec.Config.PkgDomain.Scopes {
		s.Type = DomainTypePkg
	}
	for _, s := range spec.Config.HttpDomain.Scopes {
		s.Type = DomainTypeHttp
	}

	// Validate the resulting struct.
	err = validate.Struct(spec.Config)
	return
}

// unmarshal wraps `yaml.Unmarshal`.
//
// Param: `checkPresence` determines whether to return an error if the file is not found.
func unmarshal(path string, dest interface{}, checkPresence bool) error {
	bs, err := os.ReadFile(path)
	if err != nil {
		if checkPresence && os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to locate file at '%s'", path)
		}
		return err
	}
	if err = yaml.Unmarshal(bs, dest); err != nil {
		return errors.Wrapf(err, "failed to unmarshal file at '%s'", path)
	}
	return nil
}
