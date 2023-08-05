package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dirPath := filepath.Join(tmpDir, "test_dir")

	// Create a new directory.
	if err = CreateDir(dirPath); err != nil {
		t.Errorf("CreateDir(%q) returned error: %v", dirPath, err)
	}
	// Verify that the directory was created.
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Errorf("CreateDir(%q) failed to create directory", dirPath)
	}
}

func TestCreateDirINE(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer func(path string) { _ = os.RemoveAll(path) }(tmpDir)

	dirPath := filepath.Join(tmpDir, "test_dir")

	// Create a directory that already exists.
	if err = os.Mkdir(dirPath, 0777); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if _, err = CreateDirINE(dirPath); err != nil {
		t.Errorf("CreateDirINE(%q) returned error: %v", dirPath, err)
	}

	// Create a new directory.
	newDirPath := filepath.Join(tmpDir, "new_dir")
	if _, err = CreateDirINE(newDirPath); err != nil {
		t.Errorf("CreateDirINE(%q) returned error: %v", newDirPath, err)
	}
	if _, err := os.Stat(newDirPath); os.IsNotExist(err) {
		t.Errorf("CreateDirINE(%q) failed to create directory", newDirPath)
	}
}
