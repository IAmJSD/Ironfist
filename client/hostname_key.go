package main

import "github.com/google/uuid"

// Hostname is the main hostname that Ironfist is using.
var Hostname string

// ApplicationKey is the application key for Ironfist.
var ApplicationKey = uuid.Must(uuid.NewUUID()).String()
