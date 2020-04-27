package main

import (
	"encoding/json"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

var router = fasthttprouter.New()

func init() {
	// Get the length of the database.
	router.GET("/len", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.SetStatusCode(200)
		ctx.Response.Header.Set("Content-Type", "application/json")
		b, err := json.Marshal(updateLen())
		if err != nil {
			panic(err)
		}
		ctx.Response.SetBody(b)
	})

	// Pushes a item into the database.
	router.POST("/push", func(ctx *fasthttp.RequestCtx) {
		var x map[string]interface{}
		err := json.Unmarshal(ctx.Request.Body(), &x)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte("Invalid request."))
			return
		}
		err = updatePush(x)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		ctx.Response.SetStatusCode(204)
	})

	// Remove a item by the hash.
	router.GET("/rm/:hash", func(ctx *fasthttp.RequestCtx) {
		Hash := ctx.UserValue("hash").(string)
		err := updateRm(Hash)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		ctx.Response.SetStatusCode(204)
	})

	// Get the items before a hash that meets a filter.
	router.GET("/before/:hash", func(ctx *fasthttp.RequestCtx) {
		Filter := map[string]interface{}{}
		var err error
		ctx.QueryArgs().VisitAll(func(key, value []byte) {
			if err != nil {
				return
			}
			var i interface{}
			err = json.Unmarshal(value, &i)
			if err != nil {
				return
			}
			Filter[string(key[0])] = i
		})
		Hash := ctx.UserValue("hash").(string)
		arr, err := updatesBefore(Hash, Filter)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		b, err := json.Marshal(&arr)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody(b)
	})

	// Get the items after a hash that meets a filter.
	router.GET("/after/:hash", func(ctx *fasthttp.RequestCtx) {
		Filter := map[string]interface{}{}
		var err error
		ctx.QueryArgs().VisitAll(func(key, value []byte) {
			if err != nil {
				return
			}
			var i interface{}
			err = json.Unmarshal(value, &i)
			if err != nil {
				return
			}
			Filter[string(key[0])] = i
		})
		Hash := ctx.UserValue("hash").(string)
		arr, err := updatesAfter(Hash, Filter)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		b, err := json.Marshal(&arr)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody(b)
	})

	// Gets information about the hash provided.
	router.GET("/info/:hash", func(ctx *fasthttp.RequestCtx) {
		Hash := ctx.UserValue("hash").(string)
		arr, err := updateInfo(Hash)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		b, err := json.Marshal(&arr)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody(b)
	})

	// Gets the latest hash.
	router.GET("/latest", func(ctx *fasthttp.RequestCtx) {
		r, err := latestUpdate()
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		b, err := json.Marshal(&r)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.SetBody([]byte(err.Error()))
			return
		}
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody(b)
	})
}
