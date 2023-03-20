package output

import (
	"testing"

	"github.com/pkg/errors"
)

func TestOutput(t *testing.T) {
	t.Skip("Visual inspection only")

	t.Run("package", func(t *testing.T) {
		Package("foobar")
	})

	t.Run("file created", func(t *testing.T) {
		File("foobar", true)
	})

	t.Run("file ignored", func(t *testing.T) {
		File("foobar", false)
	})

	t.Run("report", func(t *testing.T) {
		Report(4, 2)
	})

	t.Run("report none generated", func(t *testing.T) {
		Report(0, 2)
	})

	t.Run("info", func(t *testing.T) {
		Info("this is information about something")
	})

	t.Run("error", func(t *testing.T) {
		Error(errors.WithStack(errors.New("this is an error")))
	})
}
