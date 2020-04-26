# Ironfist Server
This is all of the code for a official server-side implementation of Ironfist. The following endpoints are implemented:
- `/serverspec` - This endpoint follows the specification in `ironfist-server-api-spec.md` from the project root.

The following keys can be configured in Redis to handle various things relating to updates:
- `c:user_census_sleep` - Configures how long the user sleeps for before they phone home with their census.
