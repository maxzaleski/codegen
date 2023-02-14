package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// NewSpec parses the .codegen directory and returns a `Spec`.
func NewSpec() (*Spec, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	root += "/.codegen"
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, errors.New("missing .codegen directory")
	}

	spec := &Spec{
		GlobalConfig: &GlobalConfig{},
		Pkgs:         make([]*Pkg, 0),
	}
	if err := unmarshal(root+"/config.yaml", spec.GlobalConfig, true); err != nil {
		return nil, err
	}
	// Parse all packages.
	err = filepath.Walk(root+"/pkg", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		pkg := &Pkg{}
		if err := unmarshal(path, pkg, false); err != nil {
			return err
		}
		spec.Pkgs = append(spec.Pkgs, pkg)

		return nil
	})
	if err != nil {
		return nil, err
	}

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
