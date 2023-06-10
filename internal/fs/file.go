package fs

import (
	"bytes"
	"os"

	"github.com/pkg/errors"
)

// CreateFile creates a file at the specified destination and writes the given bytes to it.
func CreateFile(dest string, b []byte) error {
	f, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "failed to create file at '%s'", dest)
	}
	if _, err := f.Write(bytes.TrimSpace(b)); err != nil {
		return errors.Wrapf(err, "failed to write to file at '%s'", dest)
	}
	return nil
}

// FileExists returns true if the file at the given path exists.
func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
