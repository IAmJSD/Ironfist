package main

import "os"

import "encoding/json"

// Config is the struct for the configuration.
type Config struct {
	Exec     string `json:"exec"`
	Endpoint string `json:"endpoint"`
	Folder   string `json:"folder"`
}

// ConfigInitialised is the configuration.
var ConfigInitialised *Config

// Initialises the config.
func init() {
	// Initialise the configuration.
	b := Assets.Bytes(".ironfist.json")
	if len(b) == 0 {
		println("[IRONFIST] No configuration found!")
		os.Exit(1)
	}

	// Loads in the JSON.
	var c Config
	err := json.Unmarshal(b, &c)
	if err != nil {
		println("[IRONFIST] Config load error: ", err.Error())
		os.Exit(1)
	}
	ConfigInitialised = &c

	// Set the folder path and ensure it exists.
	FolderPath = ExactPath(ConfigInitialised.Folder)
	err = os.MkdirAll(FolderPath, 0700)
	if err != nil {
		println("[IRONFIST] Folder create error: ", err.Error())
		os.Exit(1)
	}
}
