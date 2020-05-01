package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

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

func sendClientError(ctx *fasthttp.RequestCtx, err error) {
	ctx.Response.SetStatusCode(400)
	ctx.Response.Header.Set("Content-Type", "application/json")
	b, _ := json.Marshal(err.Error())
	ctx.Response.SetBody(b)
}

func sendServerError(ctx *fasthttp.RequestCtx, err error) {
	ctx.Response.SetStatusCode(500)
	ctx.Response.Header.Set("Content-Type", "application/json")
	b, _ := json.Marshal(err.Error())
	ctx.Response.SetBody(b)
}

// RollbackRequired is used for the application to be able to rollback.
func RollbackRequired(ctx *fasthttp.RequestCtx, InstallID string) {
	err := RedisClient.Set("rollback:"+InstallID, "t", 0).Err()
	if err != nil {
		sendServerError(ctx, err)
		return
	}
	VersionHash, err := RedisClient.Get("h:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}
	err = RedisClient.SAdd("b:"+InstallID, VersionHash).Err()
	if err != nil {
		sendServerError(ctx, err)
		return
	}
	ctx.Response.SetStatusCode(204)
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
func GetUpdateChunkInfo(ctx *fasthttp.RequestCtx, _ string) {
	// Get the update hash.
	var UpdateHash string
	err := json.Unmarshal(ctx.Request.Body(), &UpdateHash)
	if err != nil {
		sendClientError(ctx, err)
		return
	}
	update, err := GetUpdateInfo(UpdateHash)
	if err != nil {
		sendClientError(ctx, err)
		return
	}
	b, err := json.Marshal(update.Chunks)
	if err != nil {
		sendServerError(ctx, err)
		return
	}
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBody(b)
}

func itemInArray(x string, a []string) bool {
	for _, v := range a {
		if x == v {
			return true
		}
	}
	return false
}

func getLastUpdate(VersionHash string, BlacklistedUpdates []string, Channels []string) *Update {
	updates, err := GetUpdatesBeforeAfter(true, VersionHash, Channels)
	if err != nil {
		return nil
	}
	for i, j := 0, len(updates)-1; i < j; i, j = i+1, j-1 {
		updates[i], updates[j] = updates[j], updates[i]
	}
	for _, v := range updates {
		if !itemInArray(v.UpdateHash, BlacklistedUpdates) {
			return &v
		}
	}
	return nil
}

func getNextUpdate(VersionHash, InstallID string, BlacklistedUpdates []string, Channels []string) *Update {
	// Check the version hash against the latest hash.
	LatestHash, err := GetLatestHash()
	if err != nil {
		return nil
	}

	// Check if we should rollback.
	_, err = RedisClient.Get("rollback:"+InstallID).Result()
	if LatestHash == VersionHash {
		// Check if we are meant to roll back.
		if err == nil {
			// We are not. We can just return no version here.
			return nil
		} else {
			// Hmmmm, check for previous updates. If not, return nothing.
			return getLastUpdate(VersionHash, BlacklistedUpdates, Channels)
		}
	}

	// Ok, try and look ahead for updates.
	updates, err := GetUpdatesBeforeAfter(false, VersionHash, Channels)
	if err != nil {
		return nil
	}
	for i, j := 0, len(updates)-1; i < j; i, j = i+1, j-1 {
		updates[i], updates[j] = updates[j], updates[i]
	}
	for _, v := range updates {
		if !itemInArray(v.UpdateHash, BlacklistedUpdates) {
			return &v
		}
	}

	// Ok, so we don't need to rollback (this is a new version - rollbacks only apply if there's not a newer version than the client).
	// However, there are no newer updates which have not been blacklisted. We will return nil here.
	return nil
}

// GetLatestUpdate is used to get the latest update.
func GetLatestUpdate(ctx *fasthttp.RequestCtx, InstallID string) {
	// Get the current version hash.
	VersionHash, err := RedisClient.Get("h:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Get the update channels.
	Channels, err := RedisClient.SMembers("u:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}
	Channels = append(Channels, "default")

	// Get all blacklisted updates.
	BlacklistedUpdates, err := RedisClient.SMembers("b:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Get the latest update.
	latest := getNextUpdate(VersionHash, InstallID, BlacklistedUpdates, Channels)

	// Serialize the update.
	var b []byte
	if latest == nil {
		b, err = json.Marshal(latest)
	} else {
		b, err = json.Marshal(&map[string]interface{}{
			"name": latest.Name,
			"version": latest.Version,
			"changelogs": latest.Changelogs,
			"hash": latest.UpdateHash,
		})
	}
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Send the response.
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBody(b)
}

// GetPreviousVersions is used to get the previous versions.
func GetPreviousVersions(ctx *fasthttp.RequestCtx, InstallID string) {
	// Get the current version hash.
	VersionHash, err := RedisClient.Get("h:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Get the update channels.
	Channels, err := RedisClient.SMembers("u:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}
	Channels = append(Channels, "default")

	// Get all blacklisted updates.
	BlacklistedUpdates, err := RedisClient.SMembers("b:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Get the past updates.
	p, err := GetUpdatesBeforeAfter(true, VersionHash, Channels)
	if err != nil {
		sendServerError(ctx, err)
		return
	}
	past := make([]map[string]interface{}, 0, len(p))
	for _, v := range p {
		if !itemInArray(v.UpdateHash, BlacklistedUpdates) {
			past = append(past, map[string]interface{}{
				"name": v.Name,
				"version": v.Version,
				"changelogs": v.Changelogs,
				"hash": v.UpdateHash,
			})
		}
	}

	// Serialize the update.
	b, err := json.Marshal(&past)
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Send the response.
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBody(b)
}

// UpdatePending is used to check if updates are pending.
func UpdatePending(ctx *fasthttp.RequestCtx, InstallID string) {
	// Get the current version hash.
	VersionHash, err := RedisClient.Get("h:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Get the update channels.
	Channels, err := RedisClient.SMembers("u:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}
	Channels = append(Channels, "default")

	// Get all blacklisted updates.
	BlacklistedUpdates, err := RedisClient.SMembers("b:"+InstallID).Result()
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Get the latest update.
	latest := getNextUpdate(VersionHash, InstallID, BlacklistedUpdates, Channels)

	// Serialize the update.
	var b []byte
	if latest == nil {
		b, err = json.Marshal(false)
	} else {
		b, err = json.Marshal(false)
	}
	if err != nil {
		sendServerError(ctx, err)
		return
	}

	// Send the response.
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBody(b)
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

	// If it's a blacklisted release, un-blacklist it.
	err = RedisClient.SRem("b:"+InstallID, VersionHash).Err()
	if err != nil {
		sendServerError(ctx, err)
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
	Router.POST("/client", ServerSpecRequest)
}
