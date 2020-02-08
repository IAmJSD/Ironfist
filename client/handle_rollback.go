package main

import (
	"github.com/jakemakesstuff/structuredhttp"
	"os"
)

// HandleRollback is used to handle rolling back a release.
func HandleRollback() {
	// Handle fails accessing the endpoint.
	HandleRollbackFail := func() {
		println("[IRONFIST] Application failed and we cannot rollback! Throwing error.")
		os.Exit(1)
		return
	}

	// Handle no install ID.
	if InstallID == "" {
		HandleRollbackFail()
		return
	}

	// Handle getting previous versions.
	r, err := structuredhttp.POST(ConfigInitialised.Endpoint).
		Header("Ironfist-Action", "Get-Previous-Versions").
		Header("Ironfist-Version", "1.0.0").
		Header("Ironfist-Install-ID", InstallID).
		Run()
	if err != nil {
		HandleRollbackFail()
		return
	}
	if r.RaiseForStatus() != nil {
		HandleRollbackFail()
		return
	}

	// Get the JSON.
	j, err := r.JSON()
	if err != nil {
		HandleRollbackFail()
		return
	}
	l := j.([]map[string]interface{})
	if len(l) == 0 {
		HandleRollbackFail()
		return
	}
	last := l[len(l)-1]["hash"].(string)

	// Execute the rollback.
	println("[IRONFIST] The application crashed! Rolling back to " + last + "!")
	_, _ = structuredhttp.POST(ConfigInitialised.Endpoint).
		Header("Ironfist-Action", "Rollback-Required").
		Header("Ironfist-Version", "1.0.0").
		Header("Ironfist-Install-ID", InstallID).
		Run()
	ok := UpDowngradeRelease(last)
	if !ok {
		println("[IRONFIST] Rollback failed. Crashing here.")
		os.Exit(1)
	}
}
