# Ironfist Server API Specification

This API only has one HTTP endpoint which accepts POST requests and we use headers to describe what Ironfist actions we want to perform. This is made to allow the API to be easily ported over to other languages/server configurations. The following headers will be in every Ironfist request:
- `Ironfist-Action` - The action which the client is trying to preform with the server. These are described below.
- `Ironfist-Version` - The Ironfist API version. This is currently `1.0.0`.

Additionally, every action except for `Generate-Install-ID` will have a `Ironfist-Install-ID` header which contains the installation ID.

## Response Codes
The following response codes are valid for the server API:
- `403` - Missing/invalid install ID (this will happen if you do not supply a `Ironfist-Install-ID` header for everything except `Generate-Install-ID`).
- `5XX` - Server failure. Don't deploy on Christmas Day or Fridays!
- `400` - Missing/broken action specific stuff or `Ironfist-Version`/`Ironfist-Action` missing.

## Actions

Here are all of the actions which are supported by the Ironfist server API:

### `Generate-Install-ID`
**Body:** JSON string containing the machine ID.

**Returns:** JSON string containing the install ID.

**Documentation:** Used to generate a install ID. You should store the install ID, a hash of the machine ID (to stop multiple install ID's from the same users after reinstalls), an array to store update channels and a (right now null) string to set the version hash in your database after the user ran this request.

### `User-Census-Sleep-Time`
**Body:** Not required.

**Returns:** JSON integer containing the sleep time for a user census in seconds.

**Documentation:** Used to tell Ironfist how long it should sleep before making another `User-Census` request in seconds.

### `User-Census`
**Body:** Not required.

**Returns:** 204 (no content)

**Documentation:** Used to get a census on how many people are actively using your software. Useful for stats and software updates.

### `Set-Version-Hash`
**Body:** JSON string containing the version hash.

**Returns:** 204 (no content)

**Documentation:** Used to set the version hash for the current install ID.

### `Rollback-Required`
**Body:** Not required.

**Returns:** 204 (no content)

**Documentation:** Used to notify the server that the application rolled back. The server can then used the stored update hash to make a decision when deploying.

### `Join-Update-Channel`
**Body:** JSON string containing the channel name.

**Returns:** 204 (no content) if channel exists or 400 with the body `"Update channel not found."`.

**Documentation:** Used to join a update channel.

### `Leave-Update-Channel`
**Body:** JSON string containing the channel name.

**Returns:** 204 (no content) if channel exists or 400 with the body `"Update channel not found."`.

**Documentation:** Used to leave a update channel.

### `Get-Update-Chunk-Info`
**Body:** JSON string containing the version hash.

**Returns:** Either 400 with the body `"Failed to find update information."` or a 200 with an array of the following object:
- `hash` - The update chunk hash.
- `url` - The URL for the update chunk. **This can be HTTP since we are comparing the hash and this can be cached by system administrators.**

Each chunk of the application should be ordered in the array in the order that it is needed to make up the application.

**Documentation:** Used to get the chunks of a update. The chunks should be gunzipped, should probably be under 10MB and go together in the order of the array to make a zip file containing the update.

### `Get-Latest-Update`
**Body:** None required.

**Returns:** A JSON object with the latest update information or `null` if the user is already on the latest release. It should contain the following:
- `name` - The name of the update.
- `version` - The version of the update.
- `changelogs` - The changelogs of the version.
- `hash` - The hash of the version. We will need this for downloading the update.

**Documentation:** Used to get the latest update from the Ironfist API. Note that the latest version should **NOT** contain releases that have been previously rolled back by the user.

### `Get-Previous-Versions`
**Body:** None required.

**Returns:** An array of previous updates. Each update should be a JSON object which contains the following:
- `name` - The name of the update.
- `version` - The version of the update.
- `changelogs` - The changelogs of the version.
- `hash` - The hash of the version. We will need this for downloading the update.

**Documentation:** Used to get previous updates from the Ironfist API which is mainly used to roll back updates. This should **NOT** contain releases that have been previously rolled back by the user.

### `Update-Pending`
**Body:** None required.

**Returns:** JSON boolean repersenting if a update is required.

**Documentation:** Check if a update is pending.
