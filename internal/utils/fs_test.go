package utils_test

import (
	"os"
	"strings"
	"testing"

	"github.com/toolctl/toolctl/internal/utils"
)

var (
	content = "this is a dummy file"
)

func createOriginFile() (*os.File, error) {
	file, err := os.CreateTemp("/tmp", "toolctl_origin_")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	_, err = file.WriteString(content)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func createEmptyfile() (*os.File, error) {
	file, err := os.CreateTemp("/tmp", "toolctl_dest_")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return file, nil

}

// TestCopyFile tests if a file is correctly copied from source to destination
func TestCopyFile(t *testing.T) {
	// Create origin file
	sourceFilePath, err := createOriginFile()
	if err != nil {
		t.Errorf("failed to create the source file")
	}
	defer os.Remove(sourceFilePath.Name())

	// Create destination file and copy content from the original file
	destFilePath, _ := createEmptyfile()
	err = utils.CopyFile(sourceFilePath.Name(), destFilePath.Name())
	if err != nil {
		t.Errorf("failed to copy file to destination")
	}
	defer os.Remove(destFilePath.Name())

	// Read destination file content
	destFileContent, err := os.ReadFile(destFilePath.Name())
	if err != nil {
		t.Errorf("failed to open destination file")
	}

	// Compare the two files content
	if strings.Compare(content, string(destFileContent)) != 0 {
		t.Errorf("copied file content doesn't match the original")
	}
}

// TestMoveFile tests if a file is correctly moved from source to destination
func TestMoveFile(t *testing.T) {
	// Create origin file
	sourceFilePath, err := createOriginFile()
	if err != nil {
		t.Errorf("failed to create the source file")
	}
	defer os.Remove(sourceFilePath.Name())

	// Create destination file
	destFilePath, _ := createEmptyfile()

	// Move the original file to the destination path
	err = utils.MoveFile(sourceFilePath.Name(), destFilePath.Name())
	if err != nil {
		t.Errorf("failed to move file to destination")
	}

	// Check if the original file was removed
	if _, err := os.Stat(sourceFilePath.Name()); os.IsExist(err) {
		t.Error("original file was note removed")
	}

	// Read destination file content
	destFileContent, err := os.ReadFile(destFilePath.Name())
	if err != nil {
		t.Errorf("failed to open destination file")
	}

	// Compare the two files content
	if strings.Compare(content, string(destFileContent)) != 0 {
		t.Errorf("copied file content doesn't match the original")
	}
}

func TestSetPermissions(t *testing.T) {
	file, err := createEmptyfile()
	if err != nil {
		t.Errorf("failed to create the file")
	}
	err = utils.SetPermissions(file.Name())
	if err != nil {
		t.Errorf("failed to set file permissions")
	}

	stats, err := os.Stat(file.Name())
	if err != nil {
		t.Errorf("failed to read file permissions")
	}

	if stats.Mode() != 0755 {
		t.Errorf("permissions not set correctly")
	}
}
