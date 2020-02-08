package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jakemakesstuff/structuredhttp"
)

// UpDowngradeRelease is used to up/downgrade a release.
func UpDowngradeRelease(VersionHash string) bool {
	// Create the zip file.
	fp := path.Join(FolderPath, VersionHash+".zip")
	file, err := os.Create(fp)
	if err != nil {
		return false
	}

	// Get all of the hash objects.
	r, err := structuredhttp.POST(ConfigInitialised.Endpoint).
		Header("Ironfist-Action", "Get-Update-Chunk-Info").
		Header("Ironfist-Version", "1.0.0").
		Header("Ironfist-Install-ID", InstallID).
		JSON(VersionHash).
		Run()
	if err != nil {
		return false
	}
	if r.RaiseForStatus() != nil {
		return false
	}
	j, err := r.JSON()
	if err != nil {
		return false
	}
	x, ok := j.([]map[string]interface{})
	if !ok {
		return false
	}
	for _, v := range x {
		URL, ok := v["url"].(string)
		if !ok {
			return false
		}
		Hash, ok := v["hash"].(string)
		if !ok {
			return false
		}
		var GunzippedChunk []byte
		for {
			// Get the chunk and ensure its integrity.
			r, err := structuredhttp.GET(URL).Run()
			if err == nil {
				b, err := r.Bytes()
				if err == nil {
					bytearr := sha1.Sum(b)
					byteslice := make([]byte, 20)
					for i, v := range bytearr {
						byteslice[i] = v
					}
					B64Encoded := base64.StdEncoding.EncodeToString(byteslice)
					B64Encoded = strings.ReplaceAll(B64Encoded, "/", "_")
					B64Encoded = strings.ReplaceAll(B64Encoded, "+", "-")
					if B64Encoded == Hash {
						GunzippedChunk = b
						break
					}
				}
			}
		}
		r, err := gzip.NewReader(bytes.NewBuffer(GunzippedChunk))
		if err != nil {
			return false
		}
		buf := make([]byte, 1024)
		for {
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				return false
			}
			if n == 0 {
				break
			}
			if _, err := file.Write(buf[:n]); err != nil {
				return false
			}
		}
	}

	// Close the file.
	_ = file.Close()
	defer func() {
		_ = os.Remove(fp)
	}()

	// Extract the ZIP file.
	PathCreate := path.Join(FolderPath, VersionHash)
	_ = os.MkdirAll(PathCreate, 0700)
	z, err := zip.OpenReader(fp)
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
	for _, f := range z.File {
		HandleExtraction(f)
	}

	// Run the application.
	LaunchApplication(PathCreate)

	// Print that we are running a new release.
	println("[IRONFIST] Now running release hash " + VersionHash + ".")

	// Delete the current version and swapping with the new one.
	CurrentVersionPath := filepath.Join(FolderPath, "version")
	CurrentVersion, err := ioutil.ReadFile(CurrentVersionPath)
	if err == nil {
		_ = os.RemoveAll(path.Join(FolderPath, string(CurrentVersion)))
	}
	_ = ioutil.WriteFile(CurrentVersionPath, []byte(VersionHash), 0666)

	// Return true.
	return true
}
