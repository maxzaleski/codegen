package fs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dest := filepath.Join(tmpDir, "test_file.txt")
	data := []byte("hello, world")

	// Create file.
	if err = CreateFile(dest, data); err != nil {
		t.Errorf("CreateFile() returned an error: %v", err)
	}

	// Verify that the file was created and its contents match what we wrote.
	actualData, err := os.ReadFile(dest)
	if err != nil {
		t.Errorf("Failed to read file at '%s': %v", dest, err)
	}
	if !bytes.Equal(actualData, data) {
		t.Errorf("File contents don't match: expected '%s', got '%s'", string(data), string(actualData))
	}
}
