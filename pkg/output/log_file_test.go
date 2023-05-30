package output

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/codegen/internal/fs"
)

func TestWriteToErrorLog(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("file does not exists; create", func(t *testing.T) {
		err = WriteToErrorLog(tmpDir, fmt.Errorf("error message 1"))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !fs.FileExists(tmpDir + logFile) {
			t.Error("File was not created")
		}
	})

	t.Run("file exists; append", func(t *testing.T) {
		err = WriteToErrorLog(tmpDir, fmt.Errorf("error message 2"))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		fileBytes, err := os.ReadFile(tmpDir + logFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expectedBytes := []byte("error message 1\n\nerror message 2")
		if !bytes.Equal(fileBytes, expectedBytes) {
			t.Errorf("Unexpected file contents:\n%s", fileBytes)
		}
	})

	t.Run("error", func(t *testing.T) {
		err = WriteToErrorLog("/non-existent", fmt.Errorf("error message 3"))
		if err == nil {
			t.Error("Expected error, but none occurred")
		}
	})
}
