# CLI public route coverage

The OpenAPI operation manifest does not include every handler registered directly
by the HTTP server. Keep these routes separate from its 192-operation count.
This inventory is based on `httpserver/server.go` at 21f088f98.

| Route | Purpose | CLI disposition |
| --- | --- | --- |
| `GET /healthz` | Public service/status health response, without an envelope | No CLI command yet; command inclusion awaits the user's decision. |
| `GET /.well-known/stuff-stash/mobile-auth` | Public issuer, client ID, redirect URI and scopes; unavailable configuration returns 503 | No CLI command yet; command inclusion awaits the user's decision. |
| `GET /` | API index/documentation links | Documentation, not a resource command. |
| `/docs`, `/openapi.*` | Generated API reference and schema documents | Documentation, not resource commands. |
| Realtime voice WebSocket handler | Microphone/conversation transport | Excluded by approved CLI scope. |
| MCP handler | Agent protocol transport | Excluded by approved CLI scope. |

`server show` and `server auth-config` already cover the OpenAPI instance and CLI
authentication metadata endpoints. They do not cover the two additional public
handlers above. Do not claim that a successful OpenAPI manifest check proves
coverage of these routes. No implementation or transport exception is authorized
by this inventory alone.
