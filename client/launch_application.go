package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jakemakesstuff/structuredhttp"
)

// CurrentlyRunningPid is the currently running PID.
var CurrentlyRunningPid *int

// LaunchApplication is used to launch the base application and handle any errors/goroutines from it.
func LaunchApplication(Path string) {
	// Ensure the application is only running one PID at a time.
	if CurrentlyRunningPid != nil {
		DeathByIronfist <- true
		err := syscall.Kill(*CurrentlyRunningPid, syscall.SIGINT)
		if err != nil {
			panic(err)
		}
		CurrentlyRunningPid = nil
	}

	// Set the version hash.
	if InstallID != "" {
		CurrentVersionPath := filepath.Join(FolderPath, "version")
		b, err := ioutil.ReadFile(CurrentVersionPath)
		if err == nil {
			_, _ = structuredhttp.POST(ConfigInitialised.Endpoint).
				Header("Ironfist-Action", "Set-Version-Hash").
				Header("Ironfist-Version", "1.0.0").
				Header("Ironfist-Install-ID", InstallID).
				JSON(string(b)).
				Run()
		}
	}

	// Start the application.
	DeathByIronfist = make(chan bool)
	args := strings.Fields(ConfigInitialised.Exec)
	Env := os.Environ()
	for i, v := range Env {
		if strings.HasPrefix(v, "PATH=") {
			Env[i] = "PATH=" + Path
		}
	}
	Env = append(Env, "IRONFIST_HOSTNAME="+Hostname, "IRONFIST_KEY="+ApplicationKey)
	pid, err := syscall.ForkExec(args[0], args, &syscall.ProcAttr{
		Env: Env,
		Dir: Path,
		Sys: new(syscall.SysProcAttr),
		Files: []uintptr{
			uintptr(syscall.Stdin),
			uintptr(syscall.Stdout),
			uintptr(syscall.Stderr),
		},
	})
	if err != nil {
		// Handle rolling back.
		fmt.Print("[IRONFIST] Application launch error: ", err)
		HandleRollback()
		return
	}
	CurrentlyRunningPid = &pid
	go HandleProcessWatching()
}
