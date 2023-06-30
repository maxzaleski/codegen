package fs

import (
	"os"

	"github.com/pkg/errors"
)

// CreateDir creates a directory at the given path.
func CreateDir(path string) error {
	if err := os.MkdirAll(path, 0777); err != nil {
		return errors.Wrapf(err, "failed to create directory at '%s'", path)
	}
	return nil
}

// CreateDirINE creates a directory at the given path if it doesn't exist.
func CreateDirINE(path string) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// If the directory doesn't exist, create it.
		if err := CreateDir(path); err != nil {
			return err
		}
	}
	return nil
}
