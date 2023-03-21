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
		t.Errorf("CreateFile(%s) returned an error: %v", dest, err)
	}

	// Verify that the file was created and its contents match what we wrote.
	actualData, err := os.ReadFile(dest)
	if err != nil {
		t.Errorf("ReadFile(%s) returned an error: %v", dest, err)
	}
	if !bytes.Equal(actualData, data) {
		t.Errorf("File contents don't match: expected '%s', got '%s'", string(data), string(actualData))
	}
}

func TestFileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file.
	path := filepath.Join(tmpDir, "test_file.txt")
	if err = CreateFile(path, []byte("hello world")); err != nil {
		t.Errorf("CreateFile(%s) returned an error: %v", path, err)
	}

	// Verify that the file exists.
	if !FileExists(path) {
		t.Errorf("FileExists(%s) should have returned 'true'", path)
	}
}
