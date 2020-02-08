package main

import (
	"log"

	"github.com/valyala/fasthttp"
)

func main() {
	log.Fatal(fasthttp.ListenAndServe(":8000", Router.Handler))
}
