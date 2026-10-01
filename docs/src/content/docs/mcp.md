---
title: Connect an inventory agent
description: Give an MCP client read-only access using your existing Stuff Stash identity.
---

Stuff Stash can expose a read-only [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)
endpoint. An agent can find your belongings, inspect a location, and check loan
history using your existing account permissions. It cannot create, move, edit,
archive, or check out anything through this endpoint.

## Enable the endpoint

MCP is off by default. Add these settings to the API environment, then restart
the API. In the bundled self-host setup, put them in `.env` and recreate the
`app` service with `docker compose -f compose.selfhost.yaml up -d app`.

```dotenv
STUFF_STASH_MCP_ENABLED=true
STUFF_STASH_MCP_AUTH_MODE=oidc
STUFF_STASH_MCP_PUBLIC_URL=https://your-api-host:8080/mcp
```

Use the **API** origin, including its port, followed by `/mcp`. The public URL
must match the host clients actually use. The authentication mode must explicitly
match `STUFF_STASH_AUTH_MODE`. Local development may use `local-dev` with a
loopback HTTP URL; production uses OIDC and HTTPS.

The endpoint uses the API's existing rate limit, JSON body limit and request
budget. `STUFF_STASH_HTTP_WRITE_TIMEOUT` also bounds MCP application work;
`STUFF_STASH_HTTP_MAX_JSON_BODY_BYTES` bounds incoming requests. Keep the default
limits unless a measured workload needs a change.

## Connect with your account

Use a client that supports **Streamable HTTP** and a configured **Bearer** header.
Its token must be an OIDC **ID token** from the same issuer as the API, with an
audience accepted by `STUFF_STASH_OIDC_CLIENT_IDS`. Obtain and refresh that token
through your registered OIDC client's normal sign-in flow. See
[Dex users and clients](./dex-users/) when using the bundled provider.

Stuff Stash does not issue a separate MCP API key. An arbitrary provider access
token will not work, and publishing the issuer metadata does not add automatic
MCP client registration or token exchange. Clients that only support their own
automatic OAuth flow may need a separately registered client or configured-token
support.

The endpoint supports protocol revisions `2026-07-28` and `2025-11-25`. A quick
protocol check with an already acquired token is:

```sh
curl --request POST "$STUFF_STASH_MCP_PUBLIC_URL" \
  --header "Authorization: Bearer $STUFF_STASH_ID_TOKEN" \
  --header 'MCP-Protocol-Version: 2025-11-25' \
  --header 'Accept: application/json, text/event-stream' \
  --json '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
```

Keep tokens out of committed configuration and shared transcripts. Browser clients
also need their exact origin in `STUFF_STASH_CORS_ALLOWED_ORIGINS`.

## Available reads

Start with `list_tenants`, then `list_inventories` to find the scope IDs your
account can access. Inventory tools require both IDs.

| Tool | What it reads |
| --- | --- |
| `search_assets` | Exact or fuzzy inventory search |
| `get_asset` | An asset's current details, tags and custom fields |
| `list_root_assets` | Active assets at the inventory's top level |
| `list_location_assets` | Active direct children of a location or container |
| `list_checked_out_assets` | Current loans |
| `list_asset_checkout_history` | An asset's checkout and return history |

Lists include pagination metadata. Continue with `nextCursor` when `hasMore` is
true; keep the same tool and scope. A location listing is not recursive. Results
include recorded text and metadata, not photo bytes, storage keys or signed file
links.

Viewers, editors and owners can read the inventory they are allowed to view.
Every request checks the current identity and permissions, so changing a token
or losing access cannot reuse an earlier client's access.

## Connection problems

- **401:** the token is missing, expired, or has the wrong issuer or audience.
  Sign in again through the registered OIDC client. The response's
  `WWW-Authenticate` header points to public resource metadata.
- **403:** check the public endpoint's host and the browser's allowed origin.
  Tool-level **Resource unavailable** means the account cannot read that scope or
  resource, or it no longer exists.
- **Invalid tool arguments:** use the published schema, the correct scope IDs,
  and a cursor from the same listing. No write tools are published.
- **429:** the shared API rate limit was reached. Slow down and retry later.
