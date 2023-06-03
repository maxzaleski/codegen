package core

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-playground/assert"
)

func TestNewSpec(t *testing.T) {
	t.Run("no configuration directory", func(t *testing.T) {
		_, err := NewSpec("")
		assert.NotEqual(t, err, nil)
	})

	t.Run("valid configuration directory", func(t *testing.T) {
		reset, err := test_setupTestDir("")
		if err != nil {
			t.Fatal(err)
		}
		defer reset()

		_, err = NewSpec("")
		assert.Equal(t, err, nil)
	})

	t.Run("valid configuration directory with specified location", func(t *testing.T) {
		specified := "subdir"

		reset, err := test_setupTestDir(specified)
		if err != nil {
			t.Fatal(err)
		}
		defer reset()

		_, err = NewSpec(specified)
		assert.Equal(t, err, nil)
	})
}

func test_setupTestDir(subDir string) (func(), error) {
	var wd string
	if subDir != "" {
		wd = subDir + "/"
	}
	wd += DomainDir

	pkgDir := wd + "/pkg"
	if err := os.MkdirAll(pkgDir, 0777); err != nil {
		return nil, err
	}

	files := []struct {
		name string
		data []byte
	}{
		{
			wd + "/" + domainEntry,
			[]byte(codegenCfg)},
		{
			pkgDir + "/test.yaml",
			[]byte(testPkgCfg),
		},
	}
	for _, f := range files {
		if err := os.WriteFile(f.name, f.data, 0777); err != nil {
			return nil, fmt.Errorf("failed to write configuration file '%s': %v", f.name, err)
		}
	}

	reset := func() {
		from := DomainDir
		if subDir != "" {
			from = subDir
		}
		_ = os.RemoveAll(from)
	}
	return reset, nil
}

const codegenCfg = `pkg:
  scopes:
    # Go Example
    - key: go-server-pkgs
      output: core/pkg
      jobs:
        - key: models
          file-name: models
          template: presets.service

        - key: service
          file-name: service
          template: go/custom

        - key: repository
          file-name: repository_logger
          template: go/custom

    # Java Example
    - key: java-server-models
      output: java/main/src/models
      jobs:
        - key: models
          file-name: \{pkg.asTitle\}.java
          template: java/model

    - key: java-server-services
      output: java/main/src/services
      jobs:
        - key: service
          file-name: \{pkg.asTitle.asCamel\}Service.java
          template: java/service


http:
  scopes:
    - key: server
      output: cmd/api/http
      jobs:
        - key: controller
          file-name: \{pkg.asLower.asSnake\}_controller
          template: go/custom/controller

    - key: flutter-client
      output: apps/flutter/lib/api
      jobs:
        - key: api
          file-name: api.dart
          template: dart/api
          concat: true

    - key: nextjs-client
      output: apps/nextjs/lib/api
      jobs:
        - key: controller
          file-name: api.ts
          template: dart/api
          concat: true

`

const testPkgCfg = `name: user 

models:
  - name: User
    extends: internal.PublicEntity
    props:
      - name: Key
        type: string
        description: The user's name
        addons:
          tags:
              gorm: type:varchar(102);not null
              json: name,omitempty
      - name: Email 
        type: string
        scope: private
        addons:
          tags:
              gorm: type:varchar(102);not null
              json: name,omitempty
    methods:
      - name: TableName
        params:
          - name: _
            type: \*gorm.DB
        returns:
          - type: string

interface:
  methods:
    - name: Create
      description: Create creates a new user.
      params:
        - name: u
          type: User
    - name: Get
      description: Get retrieves an existing user.
      params:
        - name: uID
          type: string 
      returns:
        - type: User

`
