package main

import (
	"github.com/jakemakesstuff/structuredhttp"
	"time"
)

// HandleUserCensus is used to handle the user census.
func HandleUserCensus() {
	// Check if the install ID exists.
	if InstallID == "" {
		return
	}

	// Handle the user census timeout.
	r, err := structuredhttp.POST(ConfigInitialised.Endpoint).
		Header("Ironfist-Action", "User-Census-Sleep-Time").
		Header("Ironfist-Version", "1.0.0").
		Header("Ironfist-Install-ID", InstallID).
		Run()
	if err != nil {
		return
	}
	if r.RaiseForStatus() != nil {
		return
	}
	j, err := r.JSON()
	if err != nil {
		return
	}
	i, ok := j.(int)
	if !ok {
		return
	}

	// Handle the user census POST request.
	for {
		time.Sleep(time.Second * time.Duration(i))
		_, _ = structuredhttp.POST(ConfigInitialised.Endpoint).
			Header("Ironfist-Action", "User-Census").
			Header("Ironfist-Version", "1.0.0").
			Header("Ironfist-Install-ID", InstallID).
			Run()
	}
}
