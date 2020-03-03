package main

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// AppContentsHash is the hash of AppContents.
var AppContentsHash string

// Initialises the hash.
func init() {
	// Load the app contents bytes.
	AppContents := Assets.Bytes("app_contents.zip")

	// Create the hash.
	HashFixed := sha1.Sum(AppContents)
	Hash := make([]byte, 20)
	for i, v := range HashFixed {
		Hash[i] = v
	}
	AppContentsHash = base64.StdEncoding.EncodeToString(Hash)
	AppContentsHash = strings.ReplaceAll(AppContentsHash, "/", "_")
	AppContentsHash = strings.ReplaceAll(AppContentsHash, "+", "-")
}

// ExtractInternalBundle is used to extract an internal bundle.
func ExtractInternalBundle() string {
	// Load the app contents bytes.
	AppContents := Assets.Bytes("app_contents.zip")

	// Make the directory for the folder.
	PathCreate := path.Join(FolderPath, AppContentsHash)
	err := os.Mkdir(PathCreate, 0700)
	if err != nil {
		panic(err)
	}

	// Write the .zip file.
	ZIPPath := path.Join(FolderPath, AppContentsHash+".zip")
	_ = ioutil.WriteFile(ZIPPath, AppContents, 0700)
	defer func() {
		_ = os.Remove(ZIPPath)
	}()

	// Extract the ZIP file.
	r, err := zip.OpenReader(ZIPPath)
	if err != nil {
		panic(err)
	}
	HandleExtraction := func(f *zip.File) {
		x, err := f.Open()
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := x.Close(); err != nil {
				panic(err)
			}
		}()

		p := filepath.Join(PathCreate, f.Name)

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(p, f.Mode())
		} else {
			_ = os.MkdirAll(filepath.Dir(p), f.Mode())
			f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				panic(err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, x)
			if err != nil {
				panic(err)
			}
		}
	}
	for _, f := range r.File {
		HandleExtraction(f)
	}

	// Write the version.
	err = ioutil.WriteFile(path.Join(FolderPath, "version"), []byte(AppContentsHash), 0666)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path.Join(FolderPath, "launcher_version"), []byte(AppContentsHash), 0666)
	if err != nil {
		panic(err)
	}

	// Return the path.
	return PathCreate
}
