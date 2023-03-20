package core

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	configDir      = ".codegen"
	configFileName = "config.yaml"
)

// NewSpec parses the .codegen directory and returns a `Spec`.
func NewSpec(loc string) (*Spec, error) {
	// Establish presence of configuration directory.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if loc != "" {
		cwd += "/" + loc
	}
	cfgDirPath := cwd + "/" + configDir
	if _, err := os.Stat(cfgDirPath); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "failed to locate '%s' directory", configDir)
	}

	// Parse generator specification.
	spec := &Spec{
		Global: &GlobalConfig{},
		Pkgs:   make([]*Pkg, 0),
		Paths: &SpecPaths{
			Cwd:     cwd,
			DirPath: cfgDirPath,
		},
	}
	if err := unmarshal(cfgDirPath+"/config.yaml", spec.Global, true); err != nil {
		return nil, err
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

		pkg := &Pkg{}
		if err := unmarshal(path, pkg, false); err != nil {
			return err
		}
		// Order method arguments by `index` field.
		if len(pkg.Models) != 0 {
			for _, m := range pkg.Models {
				if len(m.Methods) != 0 {
					for _, m := range m.Methods {
						m.SortArguments()
					}
				}
			}
			for _, m := range pkg.Interface.Methods {
				m.SortArguments()
			}
		}
		spec.Pkgs = append(spec.Pkgs, pkg)

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk configuration dir")
	}

	return spec, nil
}

// unmarshal wraps `yaml.Unmarshal`.
//
// `checkPresence` is used to determine whether or not to return an error if the file is not found.
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
