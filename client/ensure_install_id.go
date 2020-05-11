package main

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/jakemakesstuff/structuredhttp"
	"io/ioutil"
	"path"
)

// InstallID is the installation ID of the user.
var InstallID string

// EnsureInstallID is used to ensure the install ID has been configured.
func EnsureInstallID(Path string) {
	// Attempt to read the file.
	b, err := ioutil.ReadFile(path.Join(FolderPath, "install_id"))
	if err == nil {
		InstallID = string(b)
		return
	}

	// Throws the error.
	throw := func() {
		println("[IRONFIST] Failed to generate install ID. Ironfist's functionality will be disabled!")
	}

	// Create the machine ID.
	id, err := machineid.ID()
	if err != nil {
		panic(err)
	}

	// Run the generation request.
	r, err := structuredhttp.POST(ConfigInitialised.Endpoint).
		Header("Ironfist-Action", "Generate-Install-ID").
		Header("Ironfist-Version", "1.0.0").
		JSON(id).
		Run()
	if err != nil {
		throw()
		return
	}
	if r.RaiseForStatus() != nil {
		throw()
		return
	}

	// Set the install ID.
	j, err := r.JSON()
	if err != nil {
		throw()
		return
	}
	InstallID = j.(string)
	_ = ioutil.WriteFile(path.Join(FolderPath, "install_id"), []byte(InstallID), 0666)
}
