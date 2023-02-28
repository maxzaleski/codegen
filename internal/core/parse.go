package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// NewSpec parses the .codegen directory and returns a `Spec`.
func NewSpec(loc string) (*Spec, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if loc != "" {
		cwd += "/" + loc
	}
	dirPath := cwd + "/.codegen"
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, errors.WithStack(errors.New("missing .codegen directory"))
	}

	spec := &Spec{
		Global: &GlobalConfig{},
		Pkgs:   make([]*Pkg, 0),
		Paths: &SpecPaths{
			Cwd:     cwd,
			DirPath: dirPath,
		},
	}
	if err := unmarshal(dirPath+"/config.yaml", spec.Global, true); err != nil {
		return nil, errors.WithStack(err)
	}
	// Parse all packages.
	err = filepath.Walk(dirPath+"/pkg", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		pkg := &Pkg{}
		if err := unmarshal(path, pkg, false); err != nil {
			return errors.WithStack(err)
		}
		spec.Pkgs = append(spec.Pkgs, pkg)

		return nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Order arguments by index.
	sortArguments(spec.Pkgs)

	return spec, nil
}

func unmarshal(path string, dest interface{}, presence bool) error {
	bs, err := os.ReadFile(path)
	if err != nil {
		if presence && os.IsNotExist(err) {
			return fmt.Errorf("core.unmarshal: missing file '%s'", path)
		}
		return fmt.Errorf("core.unmarshal: %w", err)
	}
	if err := yaml.Unmarshal(bs, dest); err != nil {
		return fmt.Errorf("core.unmarshal: %w", err)
	}
	return nil
}

func sortArguments(pkgs []*Pkg) {
	for _, pkg := range pkgs {
		for _, m := range pkg.Models {
			if len(m.Methods) == 0 {
				continue
			}
			for _, fn := range m.Methods {
				fn.SortArguments()
			}
		}
		for _, fn := range pkg.Interface.Methods {
			fn.SortArguments()
		}
	}
}
