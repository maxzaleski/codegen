package core

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

const (
	DomainDir   = ".codegen"
	domainEntry = "config.yaml"
)

var validate = newValidator()

// NewSpec parses the .codegen directory and returns a `Spec`.
func NewSpec(loc string) (spec *Spec, err error) {
	spec = newSpec()

	// Establish presence of configuration directory.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if loc != "" {
		cwd += "/" + loc
	}
	cfgDirPath := cwd + "/" + DomainDir
	if _, err = os.Stat(cfgDirPath); os.IsNotExist(err) {
		err = errors.Wrapf(err, "failed to locate '%s' directory", DomainDir)
		return
	}
	spec.Metadata.CodegenDir = cfgDirPath
	spec.Metadata.Cwd = cwd

	// Parse generator specification.
	if err = unmarshal(cfgDirPath+"/config.yaml", spec.Config, true); err != nil {
		return
	}

	// Parse all packages.
	//
	// /!\ Assumes a flat directory structure.
	err = filepath.Walk(cfgDirPath+"/pkg", func(path string, info os.FileInfo, err error) error {
		// Handle unexpected error.
		if err != nil {
			return errors.Wrapf(err, "unexpected error during dir walk at file '%s'", path)
		}
		// Skip non-YAML files.
		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		pkg := &Package{}
		if err := unmarshal(path, pkg, false); err != nil {
			return err
		}
		// Order method arguments by `index` field.
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

	// Validate the resulting struct.
	if err = validate.Struct(spec.Config); err != nil {
		return
	}

	return
}

// unmarshal wraps `yaml.Unmarshal`.
//
// `checkPresence` is used to determine whether to return an error if the file is not found.
func unmarshal(path string, dest interface{}, checkPresence bool) error {
	bs, err := os.ReadFile(path)
	if err != nil {
		if checkPresence && os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to locate file at '%s'", path)
		}
		return err
	}
	if err := yaml.Unmarshal(bs, dest); err != nil {
		return errors.Wrapf(err, "failed to unmarshal file at '%s'", path)
	}
	return nil
}
