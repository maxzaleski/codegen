package gen

import (
	"os"

	"github.com/pkg/errors"
)

// createFile creates a file at the specified destination and writes the given bytes to it.
func createFile(dest string, b []byte) error {
	f, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "failed to create file at '%s'", dest)
	}
	if _, err := f.Write(b); err != nil {
		return errors.Wrapf(err, "failed to write to file at '%s'", dest)
	}
	return nil
}
