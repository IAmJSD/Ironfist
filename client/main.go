package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	PathPtr := IsIronfistInitialised()
	if PathPtr == nil {
		// The path does not exist - extract the internal contents.
		println("[IRONFIST] New installation detected. Extracting the internal bundle!")
		Path := ExtractInternalBundle()
		PathPtr = &Path
	}

	// Ensure the install ID exists and is loaded.
	EnsureInstallID(*PathPtr)

	// Handle the user census.
	go HandleUserCensus()

	// Start the HTTP server for handling Ironfist requests.
	StartHTTPServer()

	// Launch the application.
	LaunchApplication(*PathPtr)

	// Handle CTRL+C.
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	err := syscall.Kill(*CurrentlyRunningPid, syscall.SIGINT)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}
