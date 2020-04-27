package main

import "os"

var host string

func init() {
	host = os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1:7000"
	}
}
