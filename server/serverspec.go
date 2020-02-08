package main

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

// RequestActions are all possible actions.
var RequestActions = map[string]func(ctx *fasthttp.RequestCtx){
	"Generate-Install-ID": GenerateInstallID,
}

// GenerateInstallID is used to generate a install ID.
func GenerateInstallID(ctx *fasthttp.RequestCtx) {
	// Get the machine ID.
	var MachineID string
	err := json.Unmarshal(ctx.Request.Body(), &MachineID)
	if err != nil {
		ctx.Response.SetStatusCode(400)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte("\"Machine ID is invalid.\""))
		return
	}

	// TODO: Create the install ID, return it!
}

// Initialises the file.
func init() {
	Router.POST("/serverspec", ServerSpecRequest)
}

// ServerSpecRequest is used to handle a request using the server spec.
func ServerSpecRequest(ctx *fasthttp.RequestCtx) {
	// Set the action/install ID.
	Action := string(ctx.Request.Header.Peek("Ironfist-Action"))
	if len(Action) == 0 {
		ctx.Response.SetStatusCode(400)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte("\"No action specified.\""))
		return
	}
	InstallID := string(ctx.Request.Header.Peek("Ironfist-Install-ID"))
	if Action != "Generate-Install-ID" && len(InstallID) == 0 {
		ctx.Response.SetStatusCode(403)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte("\"No install ID specified.\""))
		return
	}

	// Handle the request action.
	ActionHandler, ok := RequestActions[Action]
	if !ok {
		ctx.Response.SetStatusCode(404)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte("\"Action not found.\""))
		return
	}

	// Run the action handler.
	ActionHandler(ctx)
}
