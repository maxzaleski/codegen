package output

import (
	"github.com/codegen/internal/core"
	"github.com/codegen/internal/metrics"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestOutput(t *testing.T) {
	// t.Skip("Visual inspection only")
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(cwd)
	}
	var o = &client{
		began: time.Now(),
		Metadata: core.Metadata{
			Cwd: cwd,
		},
	}

	//t.Run("scope", func(t *testing.T) {
	//	.scope("Scope")
	//})
	//
	//t.Run("package", func(t *testing.T) {
	//	o.Package("Package", 0)
	//})
	//
	//t.Run("package – final", func(t *testing.T) {
	//	o.Package("Package", -1)
	//})
	//
	//t.Run("file created", func(t *testing.T) {
	//	o.File("foobar", true)
	//})
	//
	//t.Run("file ignored", func(t *testing.T) {
	//	o.File("foobar", false)
	//})

	t.Run("info", func(t *testing.T) {
		o.Info("Line one", "Line two")
	})

	t.Run("error", func(t *testing.T) {
		o.Error(errors.WithStack(errors.New("this is an error")))
	})

	t.Run("final reporting", func(t *testing.T) {
		m := metrics.New(map[string]map[string][]*metrics.Measurement{
			"go": {
				"pkg1": {
					{
						FileAbsolutePath: "path/to/file",
						Created:          false,
					},
					{
						FileAbsolutePath: "path/to/file",
						Created:          true,
					},
				},
			},
			"java": {
				"pkg1": {
					{
						FileAbsolutePath: "path/to/file",
						Created:          false,
					},
					{
						FileAbsolutePath: "path/to/file",
						Created:          true,
					},
				},
			},
		})
		o.FinalReport(m)
	})

	t.Run("final reporting – no change", func(t *testing.T) {
		m := metrics.New(map[string]map[string][]*metrics.Measurement{
			"Scope1": {
				"pkg1": {
					{
						FileAbsolutePath: "path/to/file",
						Created:          false,
					},
				},
				"pkg2": {
					{
						FileAbsolutePath: "path/to/file",
						Created:          false,
					},
				},
			},
		})
		o.FinalReport(m)
	})
}
