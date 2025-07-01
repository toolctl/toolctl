package sysutil

import (
	"fmt"
	"io"
	"os"
)

// CopyFile copies the file from the source to the destination
func CopyFile(src, dest string) error {
	// open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", err)
	}
	defer srcFile.Close()

	// open destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error opening destination file: %s", err)
	}
	defer destFile.Close()

	// copy file
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("writing to destination file failed: %s", err)
	}

	return nil
}

// MoveFile moves the file from the source to the destination
func MoveFile(src, dest string) error {
	// copy file
	err := CopyFile(src, dest)
	if err != nil {
		return fmt.Errorf("failed copying file: %s", err)
	}

	// delete source file
	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}

// SetPermissions sets a file permissions to be executable
func SetPermissions(src string) error {
	err := os.Chmod(src, 0755)
	if err != nil {
		return fmt.Errorf("failed changing file permissions: %s", err)
	}
	return nil
}
