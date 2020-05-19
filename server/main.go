package main

import (
	"github.com/AubSs/fasthttplogger"
	"log"

	"github.com/valyala/fasthttp"
)

func main() {
	s := &fasthttp.Server{
		Handler: fasthttplogger.Tiny(Router.Handler),
		// TODO: When fasthttp supports doing this per request, we should stop doing this.
		MaxRequestBodySize: 2000 * 1024 * 1024,
	}
	log.Fatal(s.ListenAndServe(":8000"))
}
