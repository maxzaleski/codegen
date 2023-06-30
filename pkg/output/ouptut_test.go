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

	t.Run("scope", func(t *testing.T) {
		printScope("Name")
	})

	t.Run("package", func(t *testing.T) {
		printPkg("Name")
	})

	t.Run("file created", func(t *testing.T) {
		printFile("Name", true)
	})

	t.Run("file ignored", func(t *testing.T) {
		printFile("Name", false)
	})

	t.Run("info", func(t *testing.T) {
		o.PrintInfo("Line one", "Line two")
	})

	t.Run("error", func(t *testing.T) {
		o.PrintError(errors.WithStack(errors.New("this is an error")))
	})

	t.Run("final reporting", func(t *testing.T) {
		m := metrics.New(map[string]interface{}{
			"go": map[string][]*metrics.Measurement{
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
			"java": map[string][]*metrics.Measurement{
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
		o.PrintFinalReport(m)
	})

	t.Run("final reporting – no change", func(t *testing.T) {
		m := metrics.New(map[string]interface{}{
			"Scope1": map[string][]*metrics.Measurement{
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
		o.PrintFinalReport(m)
	})
}
