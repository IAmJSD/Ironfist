package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

// TODO: Add endpoint to edit redis key "c:user_census_sleep"
// DefaultUserCensusSleep is the default time that user census's sleep.
var DefaultUserCensusSleep = 10

// RequestActions are all possible actions.
var RequestActions = map[string]func(ctx *fasthttp.RequestCtx){
	"Generate-Install-ID":    GenerateInstallID,
	"User-Census-Sleep-Time": InstallIDMiddleware(UserCensusSleepTime),
	"User-Census":            InstallIDMiddleware(UserCensus),
	"Set-Version-Hash":       InstallIDMiddleware(SetVersionHash),
	"Rollback-Required":      InstallIDMiddleware(RollbackRequired),
	"Join-Update-Channel":    InstallIDMiddleware(JoinUpdateChannel),
	"Leave-Update-Channel":   InstallIDMiddleware(LeaveUpdateChannel),
	"Get-Update-Chunk-Info":  InstallIDMiddleware(GetUpdateChunkInfo),
	"Get-Latest-Update":      InstallIDMiddleware(GetLatestUpdate),
	"Get-Previous-Versions":  InstallIDMiddleware(GetPreviousVersions),
	"Update-Pending":         InstallIDMiddleware(UpdatePending),
}

// RollbackRequired is used for the application to be able to rollback.
func RollbackRequired(ctx *fasthttp.RequestCtx, InstallID string) {
	// TODO: Handle rollbacks!
}

// JoinUpdateChannel is used to join the update channel.
func JoinUpdateChannel(ctx *fasthttp.RequestCtx, InstallID string) {
	// Get the update channel.
	var UpdateChannel string
	err := json.Unmarshal(ctx.Request.Body(), &UpdateChannel)
	if err != nil {
		ctx.Response.SetStatusCode(400)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte("\"Update channel is invalid.\""))
		return
	}

	// Join the specific channel.
	RedisClient.SAdd("u:"+InstallID, UpdateChannel)

	// Set the status to 204.
	ctx.Response.SetStatusCode(204)
}

// LeaveUpdateChannel is used to leave the update channel.
func LeaveUpdateChannel(ctx *fasthttp.RequestCtx, InstallID string) {
	// Get the update channel.
	var UpdateChannel string
	err := json.Unmarshal(ctx.Request.Body(), &UpdateChannel)
	if err != nil {
		ctx.Response.SetStatusCode(400)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte("\"Update channel is invalid.\""))
		return
	}

	// Leave the specific channel.
	RedisClient.SRem("u:"+InstallID, UpdateChannel)

	// Set the status to 204.
	ctx.Response.SetStatusCode(204)
}

// GetUpdateChunkInfo is used to get update chunk information.
func GetUpdateChunkInfo(ctx *fasthttp.RequestCtx, InstallID string) {
	// TODO: Handle getting update chunks!
}

// GetLatestUpdate is used to get the latest update.
func GetLatestUpdate(ctx *fasthttp.RequestCtx, InstallID string) {
	// TODO: Handle getting the latest update!
}

// GetPreviousVersions is used to get the previous versions.
func GetPreviousVersions(ctx *fasthttp.RequestCtx, InstallID string) {
	// TODO: Handle getting the previous versions!
}

// UpdatePending is used to check if updates are pending.
func UpdatePending(ctx *fasthttp.RequestCtx, InstallID string) {
	// TODO: Handle checking if updates are pending.
}

// SetVersionHash is used to set the version hash.
func SetVersionHash(ctx *fasthttp.RequestCtx, InstallID string) {
		// Get the version hash.
		var VersionHash string
		err := json.Unmarshal(ctx.Request.Body(), &VersionHash)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody([]byte("\"Version hash is invalid.\""))
			return
		}

	// Return a 204.
	ctx.Response.SetStatusCode(204)

	// Create a thread to handle the creation of the Redis key.
	go func() {
		// Set the Redis key.
		RedisClient.Set("h:"+InstallID, VersionHash, 0)
	}()
}

// UserCensus is used to get a census of the user counts on a specific hash.
func UserCensus(ctx *fasthttp.RequestCtx, InstallID string) {
	// Return a 204. There's no way for this to error.
	ctx.Response.SetStatusCode(204)

	// Get the user census sleep time.
	Sleep := DefaultUserCensusSleep
	s, err := RedisClient.Get("c:user_census_sleep").Result()
	if err == nil {
		i, err := strconv.Atoi(s)
		if err == nil {
			Sleep = i
		}
	}

	// Launch a thread to handle the user census.
	go func() {
		// Add to the census set.
		_ = RedisClient.SAdd("census", InstallID).Err()

		// Create the census item.
		_ = RedisClient.Set("C:"+InstallID, time.Now().Unix(), time.Duration(Sleep+(Sleep/2))).Err()

		// Wait for the census sleep + census sleep / 2 to check if the user is still using the application.
		time.Sleep(time.Duration(Sleep+(Sleep/2)) * time.Second)
		err := RedisClient.Get("C:" + InstallID).Err()
		if err != nil {
			RedisClient.SRem("census", InstallID)
		}
	}()
}

// UserCensusSleepTime is used to return the user census sleep time.
func UserCensusSleepTime(ctx *fasthttp.RequestCtx, _ string) {
	// Get the user census sleep time.
	Sleep := DefaultUserCensusSleep
	s, err := RedisClient.Get("c:user_census_sleep").Result()
	if err == nil {
		i, err := strconv.Atoi(s)
		if err == nil {
			Sleep = i
		}
	}

	// Return the user census sleep time.
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBody([]byte(strconv.Itoa(Sleep)))
}

// InstallIDMiddleware is used to handle the install ID.
func InstallIDMiddleware(WrappedFunc func(ctx *fasthttp.RequestCtx, InstallID string)) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		// Handle a invalid installation ID.
		HandleInvalid := func() {
			ctx.Response.SetStatusCode(403)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody([]byte("\"Install ID is invalid.\""))
		}

		// Check if the install ID is valid.
		InstallID := string(ctx.Request.Header.Peek("Ironfist-Install-ID"))
		if len(InstallID) == 0 {
			HandleInvalid()
			return
		}
		err := RedisClient.Get("e:" + InstallID).Err()
		if err != nil {
			HandleInvalid()
			return
		}

		// Run the wrapped function.
		WrappedFunc(ctx, InstallID)
	}
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

	// Check if there is a machine ID already.
	i, err := RedisClient.Get("m:" + MachineID).Result()
	if err == nil {
		ctx.Response.SetStatusCode(200)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte("\"" + i + "\""))
		return
	}

	// Create the install ID.
	i = uuid.Must(uuid.NewUUID()).String()

	// Insert the install ID.
	_ = RedisClient.Set("m:"+MachineID, i, 0).String()
	_ = RedisClient.Set("e:"+i, "y", 0).String()
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBody([]byte("\"" + i + "\""))
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

// Initialises the file.
func init() {
	Router.POST("/serverspec", ServerSpecRequest)
}
