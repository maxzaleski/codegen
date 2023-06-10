package output

import (
	"fmt"
	"os"

	"github.com/codegen/internal/fs"
)

const logFile = "/codegen_error.log"

// ErrorFile appends the given error to a log file in the current working directory.
//
// If the file does not exist, ErrorFile will create it.
func ErrorFile(cwd string, err error) error {
	dest, stackTrace := cwd+logFile, fmt.Sprintf("%+v", err)

	// Check if the file already exists. If so, append the bytes.
	if fs.FileExists(dest) {
		file, err := os.OpenFile(dest, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer func(file *os.File) { _ = file.Close() }(file)

		if _, err = file.Write([]byte("\n\n" + stackTrace)); err != nil {
			return err
		}
	} else {
		// Otherwise, create the file.
		if err = fs.CreateFile(dest, []byte(stackTrace)); err != nil {
			return err
		}
	}

	return nil
}
