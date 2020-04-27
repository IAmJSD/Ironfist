package main

import (
	"io/ioutil"
	"os"
)

var filePath string

func init() {
	filePath = os.Getenv("FILE_PATH")
	if filePath == "" {
		filePath = "./updates.json"
	}
}

func readDatabaseFile() ([]byte, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func writeDatabaseFile(data []byte) {
	err := ioutil.WriteFile(filePath, data, 0666)
	if err != nil {
		panic(err)
	}
}
