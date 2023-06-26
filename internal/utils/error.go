package utils

import "github.com/pkg/errors"

// Unwrap unwraps an error.
func Unwrap(err error) error {
	for {
		_, ok := err.(interface{ Unwrap() error })
		if !ok {
			return err
		}
		err = errors.Unwrap(err)
	}
}
