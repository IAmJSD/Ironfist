package main

import "encoding/json"

type updateStructure struct {
	Updates []map[string]interface{} `json:"updates"`
	KeyIndex map[string]int `json:"key_index"`
	Len int `json:"len"`
}

// Write is used to write the update structure.
func (s *updateStructure) Write() {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	writeDatabaseFile(b)
}

var db updateStructure
