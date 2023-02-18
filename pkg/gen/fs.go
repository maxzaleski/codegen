package gen

import (
	"os"

	"github.com/pkg/errors"
)

// createFile creates a file at the given path and writes the given bytes to it.
func createFile(path string, b []byte) error {
	out, err := os.Create(path)
	if err != nil {
		return errors.WithMessagef(err, "failed to create file at '%s'", path)
	}
	if _, err := out.Write(b); err != nil {
		return errors.WithMessagef(err, "failed to write to file at '%s'", path)
	}
	return nil
}
