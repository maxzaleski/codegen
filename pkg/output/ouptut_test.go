package output

import (
	"github.com/maxzaleski/codegen/internal/core"
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

	//t.Run("final reporting", func(t *testing.T) {
	//	m := metrics.New()
	//	m.CaptureScope("Scope1", "pkg1", metrics.FileOutcome{
	//		AbsolutePath: "path/to/file",
	//		Created:      true,
	//	})
	//	m.CaptureScope("Scope1", "pkg2", metrics.FileOutcome{
	//		AbsolutePath: "path/to/file",
	//		Created:      true,
	//	})
	//	m.CaptureScope("Scope2", "pkg1", metrics.FileOutcome{
	//		AbsolutePath: "path/to/file",
	//		Created:      true,
	//	})
	//	o.PrintFinalReport(m)
	//})
	//
	//t.Run("final reporting – no change", func(t *testing.T) {
	//	m := metrics.New()
	//	m.CaptureScope("Scope1", "pkg1", metrics.FileOutcome{
	//		AbsolutePath: "path/to/file",
	//		Created:      false,
	//	})
	//	m.CaptureScope("Scope1", "pkg2", metrics.FileOutcome{
	//		AbsolutePath: "path/to/file",
	//		Created:      false,
	//	})
	//	m.CaptureScope("Scope2", "pkg1", metrics.FileOutcome{
	//		AbsolutePath: "path/to/file",
	//		Created:      false,
	//	})
	//	o.PrintFinalReport(m)
	//})
}
