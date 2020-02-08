package main

import (
	"fmt"
	"os"
)

// DeathByIronfist is when an application was closed by Ironfist.
var DeathByIronfist = make(chan bool)

// HandleProcessExit is used to handle when a process exits on its own.
func HandleProcessExit(p *os.ProcessState) {
	// There is no longer a running process.
	CurrentlyRunningPid = nil

	if p.Success() {
		// Exit with code 0.
		os.Exit(0)
	} else {
		// The process failed! Time to roll back if possible.
		fmt.Println("[IRONFIST] Application crashed with error code", p.ExitCode())
		HandleRollback()
	}
}

// HandleProcessWatching is used to watch the process.
func HandleProcessWatching() {
	// Find the process with the os wrapper.
	process, err := os.FindProcess(*CurrentlyRunningPid)
	if err != nil {
		panic(err)
	}

	// Make a channel for the process state.
	ProcessStatePtr := make(chan *os.ProcessState)
	go func() {
		p, _ := process.Wait()
		ProcessStatePtr <- p
	}()

	// Create a switch for the channels.
	select {
	case <-DeathByIronfist:
		// Ignore this.
		return
	case p := <-ProcessStatePtr:
		// Handle this.
		HandleProcessExit(p)
		return
	}
}
