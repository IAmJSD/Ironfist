package main

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
)

func main() {
	// Reads the database file.
	b, err := readDatabaseFile()
	if err != nil {
		db = updateStructure{
			Updates:  []map[string]interface{}{},
			KeyIndex: map[string]int{},
			Len: 0,
		}
	} else {
		err = json.Unmarshal(b, &db)
		if err != nil {
			panic(err)
		}
	}

	// Starts the server.
	println("Serving on " + host)
	log.Fatal(fasthttp.ListenAndServe(host, router.Handler))
}
