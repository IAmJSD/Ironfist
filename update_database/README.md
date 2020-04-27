# Update Database
A simple database which is built for the single purpose of handling Ironfist updates. This is because traditional databases such as Redis and MongoDB had their own set of problems. This database is designed to be blocking and handle a very specific data structure which is required for Ironfist.

Note that whilst this does not support sharding, I'm not overly concerned. Go is quick enough where this specifically should not require sharding.

The following routes are supported:
- `/len` (GET) - Returns a JSON integer containing the length of the database.
- `/push` (POST) - Pushes a item to the database. The body of this request should be a JSON object which contains a `update_hash` attribute. This will return a 204 on success.
- `/before/<hash>?<filters>` (GET) - Gets updates before a specific hash. This will be a JSON array of objects. Filters should be key -> JSON serialized/URL encoded value.
- `/after/<hash>?<filters>` (GET) - Gets updates after a specific hash. This will be a JSON array of objects. Filters should be key -> JSON serialized/URL encoded value.
- `/latest` (GET) - Get the latest version hash. This will be a JSON string with the latest version.
- `/rm/<hash>` (GET) - Removes an update. This will return a 204 on success.
- `/info/<hash>` (GET) - Returns the JSON object of a specific hash.
