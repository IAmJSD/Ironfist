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
	LauncherCurrentVersionPath := filepath.Join(FolderPath, "launcher_version")
	if _, err := os.Stat(LauncherCurrentVersionPath); os.IsNotExist(err) {
		// The launcher version does not exist.
		return nil
	}
	CurrentVersion, err := ioutil.ReadFile(CurrentVersionPath)
	if err != nil {
		// Cannot read the version.
		return nil
	}
	LauncherCurrentVersion, err := ioutil.ReadFile(LauncherCurrentVersionPath)
	if err != nil {
		// Cannot read the launcher version.
		return nil
	}
	if string(LauncherCurrentVersion) != AppContentsHash {
		// The launcher hash is different.
		return nil
	}
	Path := filepath.Join(FolderPath, string(CurrentVersion))
	if _, err := os.Stat(Path); os.IsNotExist(err) {
		// The path does not exist.
		return nil
	}
	return &Path
}
