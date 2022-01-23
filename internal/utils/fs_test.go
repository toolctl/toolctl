package utils_test

import (
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toolctl/toolctl/internal/utils"
)

var (
	content = "this is a dummy file"
)

// Generates a random file path for the tests
func generateFilePath() string {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return "/tmp/toolctl_" + hex.EncodeToString(b) + ".test"
}

// Creates a dummy test file for the tests
func createDummyFile() (string, error) {
	path := generateFilePath()
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.WriteString(content)
	return path, nil
}

// removes all created files in the tests
func cleanup() {
	files, err := filepath.Glob("/tmp/toolctl_*.test")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
}

// TestCopyFile tests if a file is correctly copied from source to destination
func TestCopyFile(t *testing.T) {
	sourceFilePath, err := createDummyFile()
	if err != nil {
		t.Errorf("failed to create the source file")
	}
	destFilePath := generateFilePath()
	err = utils.CopyFile(sourceFilePath, destFilePath)
	if err != nil {
		t.Errorf("failed to copy file to destination")
	}

	destFileContent, err := ioutil.ReadFile(destFilePath)
	if err != nil {
		t.Errorf("failed to open destination file")
	}
	if strings.Compare(content, string(destFileContent)) != 0 {
		t.Errorf("copied file content doesn't match the original")
	}
	cleanup()
}

// TestMoveFile tests if a file is correctly moved from source to destination
func TestMoveFile(t *testing.T) {
	sourceFilePath, err := createDummyFile()
	if err != nil {
		t.Errorf("failed to create the source file")
	}
	destFilePath := generateFilePath()
	err = utils.MoveFile(sourceFilePath, destFilePath)
	if err != nil {
		t.Errorf("failed to move file to destination")
	}

	if _, err := os.Stat(sourceFilePath); os.IsExist(err) {
		t.Error("original file was note removed")
	}

	destFileContent, err := ioutil.ReadFile(destFilePath)
	if err != nil {
		t.Errorf("failed to open destination file")
	}
	if strings.Compare(content, string(destFileContent)) != 0 {
		t.Errorf("copied file content doesn't match the original")
	}

	cleanup()
}

func TestSetPermissions(t *testing.T) {
	file, err := createDummyFile()
	if err != nil {
		t.Errorf("failed to create the file")
	}
	utils.SetPermissions(file)

	stats, err := os.Stat(file)
	if err != nil {
		t.Errorf("failed to read file permissions")
	}

	if stats.Mode() != 0755 {
		t.Errorf("permissions not set correctly")
	}

	cleanup()
}
