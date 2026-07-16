---
title: Configuration Reference
description: Environment variables and runtime settings for Stuff Stash.
---

Use this page when you are wiring Stuff Stash into a real environment or
checking which values local Compose sets for you.

## API Parsing Rules

The API reads configuration from environment variables at startup.

| Type | Accepted values |
| --- | --- |
| Boolean | `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off` |
| Duration | Go duration strings such as `500ms`, `10s`, `2h` |
| Integer | Positive base-10 integers |
| List | Comma-separated values, trimmed and deduplicated |

Invalid primitive values fall back to their defaults during API environment
parsing. Unsupported adapter modes, missing required adapter settings, and
invalid enabled provider settings fail startup.

## API: HTTP

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_HTTP_ADDR` | `:8080` | Address the API listens on. |
| `STUFF_STASH_HTTP_READ_HEADER_TIMEOUT` | `5s` | Maximum time to read request headers. |
| `STUFF_STASH_HTTP_READ_TIMEOUT` | `15s` | Maximum time to read a request. |
| `STUFF_STASH_HTTP_WRITE_TIMEOUT` | `30s` | Maximum time to write a response. |
| `STUFF_STASH_HTTP_IDLE_TIMEOUT` | `60s` | Keep-alive idle timeout. |
| `STUFF_STASH_HTTP_MAX_JSON_BODY_BYTES` | `1048576` | Maximum JSON request body size. |
| `STUFF_STASH_CORS_ALLOWED_ORIGINS` | empty | Comma-separated browser origins allowed to call the API. |

## API: Rate Limiting

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_HTTP_RATE_LIMIT_ENABLED` | `true` | Enables API HTTP rate limiting. |
| `STUFF_STASH_HTTP_RATE_LIMIT_REQUESTS` | `1200` | Request budget per rate-limit window. |
| `STUFF_STASH_HTTP_RATE_LIMIT_WINDOW` | `1m` | Rate-limit accounting window. |
| `STUFF_STASH_HTTP_RATE_LIMIT_BURST` | `600` | Burst allowance. |

## API: Authentication

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_AUTH_MODE` | `local-dev` | Authentication adapter. Use `local-dev` or `oidc`. |
| `STUFF_STASH_OIDC_ISSUER` | empty | OIDC issuer URL when `STUFF_STASH_AUTH_MODE=oidc`. |
| `STUFF_STASH_OIDC_CLIENT_ID` | empty | Primary expected OIDC audience/client ID. |
| `STUFF_STASH_OIDC_CLIENT_IDS` | empty | Additional accepted OIDC client IDs, comma-separated. |
| `STUFF_STASH_OIDC_MOBILE_CLIENT_ID` | empty | Public native mobile OIDC client ID advertised to the mobile app. |
| `STUFF_STASH_OIDC_MOBILE_REDIRECT_URI` | `stuffstash://auth/callback` | Native redirect URI advertised to the mobile app. |
| `STUFF_STASH_OIDC_MOBILE_SCOPES` | `openid,email,profile,offline_access` | Comma-separated scopes requested by mobile sign-in. |

`STUFF_STASH_OIDC_CLIENT_ID` is included in the accepted client ID set even when
`STUFF_STASH_OIDC_CLIENT_IDS` is also configured.

## API: Authorization

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_AUTHZ_MODE` | `memory` | Authorization adapter. Use `memory` or `spicedb`. |
| `STUFF_STASH_SPICEDB_ENDPOINT` | empty | SpiceDB gRPC endpoint. |
| `STUFF_STASH_SPICEDB_PRESHARED_KEY` | empty | SpiceDB preshared key. Empty is only suitable for local `serve-testing`. |
| `STUFF_STASH_SPICEDB_TLS_ENABLED` | `true` | Enables TLS for SpiceDB connections. |
| `STUFF_STASH_SPICEDB_CA_PATH` | empty | Optional CA certificate path for self-signed or private SpiceDB TLS. |
| `STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA` | `false` | Bootstraps the checked-in SpiceDB schema on startup. |
| `STUFF_STASH_SPICEDB_SCHEMA_PATH` | `deploy/spicedb/schema.zed` | Schema file used when bootstrapping. |

## API: Persistence

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_REPOSITORY_MODE` | `memory` | Repository adapter. Use `memory`, `postgres`, or `sqlite`. |
| `STUFF_STASH_DATABASE_DSN` | empty | Postgres DSN or SQLite file path/DSN when using durable repository modes. |

`postgres` and `sqlite` require `STUFF_STASH_DATABASE_DSN`. SQLite will create
the parent directory for file-backed database paths.

SQLite is an API runtime mode, not the current Docker Compose self-hosting
topology. The Compose stack uses the Postgres migration job and Postgres service
unless a future SQLite-specific Compose file wires schema setup and a durable
SQLite file mount explicitly.

## API: Workers And Pagination

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_LIMIT` | `25` | Authorization outbox events claimed per drain. |
| `STUFF_STASH_AUTHORIZATION_OUTBOX_DRAIN_INTERVAL` | `10s` | Background authorization outbox drain interval. |
| `STUFF_STASH_AUTHORIZATION_OUTBOX_CLAIM_LEASE` | `30s` | Lease duration for claimed authorization outbox events. |
| `STUFF_STASH_BLOB_DELETION_OUTBOX_DRAIN_LIMIT` | `25` | Blob-deletion outbox events claimed per drain. |
| `STUFF_STASH_BLOB_DELETION_OUTBOX_DRAIN_INTERVAL` | `10s` | Background blob-deletion outbox drain interval. |
| `STUFF_STASH_BLOB_DELETION_OUTBOX_CLAIM_LEASE` | `30s` | Lease duration for claimed blob-deletion outbox events. |
| `STUFF_STASH_BLOB_DELETION_OUTBOX_MAX_ATTEMPTS` | `5` | Attempts before blob-deletion work is treated as terminal. |
| `STUFF_STASH_IMPORT_JOB_TIMEOUT_SECONDS` | `900` | Seconds before stored import source credentials are considered expired. |
| `STUFF_STASH_IMPORT_CREDENTIAL_VACUUM_INTERVAL_SECONDS` | `60` | Seconds between background cleanup passes for expired import source credentials. |
| `STUFF_STASH_INVITATION_TTL` | `168h` | Default inventory invitation token lifetime. |
| `STUFF_STASH_INVITATION_PUBLIC_BASE_URL` | empty | Public web URL used to create clickable inventory invitation links. Required to create invitations; use HTTPS outside explicit local development. |
| `STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP` | `false` | Allows a loopback or private RFC 1918 LAN HTTP invitation URL for explicit local development only. Do not enable in deployed environments. |
| `STUFF_STASH_DEFAULT_PAGE_LIMIT` | `50` | Default API collection page size. |
| `STUFF_STASH_MAX_PAGE_LIMIT` | `100` | Maximum accepted API collection page size. |

## API: Media Storage

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_BLOB_STORAGE_MODE` | `filesystem` | Blob storage adapter. Use `filesystem` or `s3`. |
| `STUFF_STASH_BLOB_STORAGE_PATH` | `.stuffstash/blobs` | Local filesystem blob path. |
| `STUFF_STASH_MAX_ATTACHMENT_BYTES` | `26214400` | Maximum attachment size in bytes. |
| `STUFF_STASH_S3_ENDPOINT` | empty | S3-compatible endpoint host and port, without scheme. |
| `STUFF_STASH_S3_PUBLIC_ENDPOINT` | empty | Browser-reachable S3-compatible endpoint for direct uploads. Defaults to `STUFF_STASH_S3_ENDPOINT` when empty. |
| `STUFF_STASH_S3_ACCESS_KEY` | empty | S3 access key. |
| `STUFF_STASH_S3_SECRET_KEY` | empty | S3 secret key. |
| `STUFF_STASH_S3_BUCKET` | empty | S3 bucket name. |
| `STUFF_STASH_S3_REGION` | `garage` | S3 region value. |
| `STUFF_STASH_S3_SECURE` | `true` | Uses HTTPS for S3-compatible storage when true. |

Set `STUFF_STASH_S3_SECURE=false` only for trusted local plain-HTTP storage,
such as local Garage verification.

When browser clients upload directly to S3-compatible storage, the public
endpoint must be reachable from the browser and the bucket must allow CORS for
the web origin. For local Garage this usually means:

- `STUFF_STASH_S3_ENDPOINT=garage:3900` for API-to-Garage traffic.
- `STUFF_STASH_S3_PUBLIC_ENDPOINT=localhost:3900` or
  `<server-lan-ip>:3900` for browser-to-Garage traffic.
- a bucket CORS rule that allows the web origin to use `GET` and `POST` and
  exposes `ETag`.

## API: Realtime Voice Providers

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_VOICE_DEV_FAKE_ENABLED` | `false` | Enables local development fake speech, language, and speech output providers. |
| `STUFF_STASH_VOICE_GOOGLE_ENABLED` | `false` | Enables Google realtime voice provider adapters. |
| `STUFF_STASH_VOICE_PROVIDER_HTTP_TIMEOUT` | `60s` | HTTP timeout for configured realtime voice provider calls. |
| `STUFF_STASH_GOOGLE_CLOUD_PROJECT` | empty | Google Cloud project ID. Required when Google voice providers are enabled. |
| `STUFF_STASH_GOOGLE_CLOUD_LOCATION` | `us-central1` | Google Cloud location for Gemini. |
| `STUFF_STASH_GOOGLE_GEMINI_MODEL` | `gemini-2.5-flash-lite` | Gemini model name. |
| `STUFF_STASH_GOOGLE_TTS_LANGUAGE_CODE` | `en-US` | Google Text-to-Speech language code. |
| `STUFF_STASH_GOOGLE_TTS_VOICE_NAME` | `en-US-Standard-C` | Google Text-to-Speech voice name. |
| `STUFF_STASH_GOOGLE_CREDENTIAL_MODE` | `adc` | Google credential source. Use `adc` or `access_token`. |
| `STUFF_STASH_GOOGLE_ACCESS_TOKEN` | empty | Static Google access token when credential mode is `access_token`. |

When `STUFF_STASH_VOICE_GOOGLE_ENABLED=true`, Google configuration is validated
at startup. Google voice providers take precedence over development fakes when
both are enabled.

## API: Provider Credential Sealing

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_PROVIDER_CREDENTIAL_KEY_ID` | empty | Identifier stored with sealed tenant provider credentials. |
| `STUFF_STASH_PROVIDER_CREDENTIAL_KEY` | empty | Base64-encoded 32-byte AES-GCM key for sealing provider credentials and temporary import source material. |

If active provider credentials exist, startup fails closed unless a valid
provider credential sealing key is configured.

Generate a key with `openssl rand -base64 32`. Keep the key stable for a given
deployment. Rotating or losing it can make stored provider credentials
unreadable.

## Web Runtime Config

The web app reads `/config.json` at runtime. It is not configured through
`STUFF_STASH_*` environment variables inside the browser bundle.

| Field | Required | Purpose |
| --- | --- | --- |
| `apiBaseUrl` | yes | Public API base URL. Trailing slashes are trimmed. |
| `oidcIssuer` | yes | Browser-visible OIDC issuer URL. Trailing slashes are trimmed. |
| `oidcClientId` | yes | Browser OIDC client ID. |
| `oidcRedirectUri` | yes | Browser redirect URI after OIDC sign-in. |
| `mediaUploadPolicy.supportedContentTypes` | no | Allowed upload content types. Defaults to JPEG, PNG, WebP, and PDF. |
| `mediaUploadPolicy.maxBytes` | no | Client upload limit. Defaults to `5242880`. |

Example:

```json
{
  "apiBaseUrl": "https://api.example.test",
  "oidcIssuer": "https://accounts.example.test",
  "oidcClientId": "stuff-stash-web",
  "oidcRedirectUri": "https://stuffstash.example.test/callback",
  "mediaUploadPolicy": {
    "supportedContentTypes": ["image/jpeg", "image/png", "image/webp", "application/pdf"],
    "maxBytes": 5242880
  }
}
```

Keep the web upload policy aligned with `STUFF_STASH_MAX_ATTACHMENT_BYTES`; the
API remains authoritative.

## Mobile Runtime Config

The mobile app asks for a Stuff Stash instance URL, reads the API's public
mobile authentication metadata, and signs in with the configured OIDC provider.
These Expo public variables are optional development defaults only.

Mobile booleans accept `1`, `true`, `yes`, `0`, `false`, and `no`.

| Variable | Required | Purpose |
| --- | --- | --- |
| `EXPO_PUBLIC_STUFF_STASH_API_BASE_URL` | no | Optional API base URL seed shown on first launch. |
| `EXPO_PUBLIC_STUFF_STASH_TENANT_ID` | no | Optional initial tenant selection hint. |
| `EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED` | no | Enables mobile voice developer diagnostics. |
| `EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN` | production links | HTTPS web origin whose `/invitations/accept` links open in an installed mobile build. Production EAS builds require it; local builds may omit it and use the custom scheme. |
| `EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP` | no | Explicitly permits a configured loopback or private RFC 1918 HTTP invitation origin in non-production builds for browser acceptance from local devices. |
| `STUFF_STASH_MOBILE_REQUIRE_INVITATION_LINKS` | no | Set to `true` in any non-EAS release pipeline that must include invitation universal/app links. The build fails if the invitation origin is missing. |

The web server publishes the platform association documents at
`/.well-known/apple-app-site-association` and `/.well-known/assetlinks.json`.
Set `STUFF_STASH_MOBILE_IOS_APP_ID` to the built app’s Team ID and bundle ID,
and set `STUFF_STASH_MOBILE_ANDROID_SHA256_CERT_FINGERPRINT` to the SHA-256
fingerprint of the certificate that signs the Android build. Both signing
identities are deployment-specific. When an iOS App ID or Android fingerprint is
not configured, the server publishes an empty relationship for that platform,
so it leaves web links in the browser instead of claiming an identity that cannot
be verified.

The mobile invitation origin must use standard HTTPS on port 443 and must match
the origin used in generated invitation links for verified native app linking.
Associated domains cannot encode a custom port. With the explicit insecure-local
switch, a non-production build may trust loopback or private RFC 1918 HTTP for
browser acceptance, but it emits no iOS associated-domain or Android verified
app-link declaration. The default self-host `:8081` origin is suitable for browser
and local custom-scheme testing; put the web service on standard HTTPS (directly or
through a reverse proxy) before building a mobile app that claims its links.
Apple must also be able to retrieve the association document over publicly
trusted HTTPS. Private hostnames and self-signed certificates are appropriate
for custom-scheme development, not production universal-link verification.

## Docs Build Config

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_DOCS_SITE` | `https://elsell.github.io` | Public site origin for Astro builds. |
| `STUFF_STASH_DOCS_BASE` | `/stuffstash/` | Base path for production or preview docs builds. |

## Local Compose Helpers

These variables are conveniences for local Compose. They are not API runtime
settings unless they map to a `STUFF_STASH_*` variable above.

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_HTTP_PORT` | `8080` | Host port mapped to the API container. |
| `POSTGRES_DB` | `stuffstash` | Local Compose Postgres database name. |
| `POSTGRES_USER` | `stuffstash` | Local Compose Postgres user. |
| `POSTGRES_PASSWORD` | `stuffstash-local` | Local Compose Postgres password. |
| `POSTGRES_PORT` | `5432` | Host port mapped to Postgres. |
| `SPICEDB_GRPC_PORT` | `50051` | Host port mapped to SpiceDB. |
| `DEX_HTTP_PORT` | `5556` | Host port mapped to Dex in the OIDC Compose override. |
| `GARAGE_IMAGE` | pinned digest | Garage image override. Must stay digest-pinned. |
| `STUFF_STASH_API_IMAGE` | pinned digest | Published API image used by self-host Compose. Release automation updates this digest. |
| `STUFF_STASH_WEB_IMAGE` | pinned digest | Published static web image used by self-host Compose. Release automation updates this digest. |
| `GARAGE_S3_PORT` | `3900` | Host port mapped to the Garage S3 API. |
| `GO_BUILDER_IMAGE` | pinned digest | API builder image override. Must stay digest-pinned. |
| `RUNTIME_IMAGE` | pinned digest | API runtime image override. Must stay digest-pinned. |
| `NODE_BUILDER_IMAGE` | pinned digest | Web builder image override. Must stay digest-pinned. |
| `WEB_RUNTIME_IMAGE` | pinned digest | Web runtime image override. Must stay digest-pinned. |
| `PNPM_VERSION` | `11.0.7` | pnpm version used by the web image build. |
