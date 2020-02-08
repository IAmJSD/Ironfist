package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// IsIronfistInitialised is used to check if Ironfist is initialised.
// If it is, return the folder.
func IsIronfistInitialised() *string {
	CurrentVersionPath := filepath.Join(FolderPath, "version")
	if _, err := os.Stat(CurrentVersionPath); os.IsNotExist(err) {
		// The version does not exist.
		return nil
	}
	CurrentVersion, err := ioutil.ReadFile(CurrentVersionPath)
	if err != nil {
		// Cannot read the version.
		return nil
	}
	Path := filepath.Join(FolderPath, string(CurrentVersion))
	if _, err := os.Stat(Path); os.IsNotExist(err) {
		// The path does not exist.
		return nil
	}
	return &Path
}
