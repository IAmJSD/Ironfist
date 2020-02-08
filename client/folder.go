package main

import "strings"
import "os"
import "path/filepath"

// FolderPath is the path which Ironfist is using.
var FolderPath string

// ExactPath is used to get the exact path.
func ExactPath(Path string) string {
	h, err := os.UserHomeDir()
	if err != nil {
		println("[IRONFIST] Failed to get home directory.")
		os.Exit(1)
	}
	Path = strings.ReplaceAll(Path, "~", h)
	fp, err := filepath.Abs(Path)
	if err != nil {
		panic(err)
	}
	return fp
}
