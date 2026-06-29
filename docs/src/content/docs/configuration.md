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
| `STUFF_STASH_INVITATION_TTL` | `168h` | Default inventory invitation token lifetime. |
| `STUFF_STASH_DEFAULT_PAGE_LIMIT` | `50` | Default API collection page size. |
| `STUFF_STASH_MAX_PAGE_LIMIT` | `100` | Maximum accepted API collection page size. |

## API: Media Storage

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_BLOB_STORAGE_MODE` | `filesystem` | Blob storage adapter. Use `filesystem` or `s3`. |
| `STUFF_STASH_BLOB_STORAGE_PATH` | `.stuffstash/blobs` | Local filesystem blob path. |
| `STUFF_STASH_MAX_ATTACHMENT_BYTES` | `5242880` | Maximum attachment size in bytes. |
| `STUFF_STASH_S3_ENDPOINT` | empty | S3-compatible endpoint host and port, without scheme. |
| `STUFF_STASH_S3_ACCESS_KEY` | empty | S3 access key. |
| `STUFF_STASH_S3_SECRET_KEY` | empty | S3 secret key. |
| `STUFF_STASH_S3_BUCKET` | empty | S3 bucket name. |
| `STUFF_STASH_S3_REGION` | `garage` | S3 region value. |
| `STUFF_STASH_S3_SECURE` | `true` | Uses HTTPS for S3-compatible storage when true. |

Set `STUFF_STASH_S3_SECURE=false` only for trusted local plain-HTTP storage,
such as local Garage verification.

## API: Realtime Voice Providers

| Variable | Default | Purpose |
| --- | --- | --- |
| `STUFF_STASH_VOICE_DEV_FAKE_ENABLED` | `false` | Enables local development fake speech, language, and speech output providers. |
| `STUFF_STASH_VOICE_GOOGLE_ENABLED` | `false` | Enables Google realtime voice provider adapters. |
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
| `STUFF_STASH_PROVIDER_CREDENTIAL_KEY` | empty | Base64-encoded 32-byte AES-GCM key for sealing provider credentials. |

If active provider credentials exist, startup fails closed unless a valid
provider credential sealing key is configured.

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

## Mobile Development Config

These Expo public variables are development defaults for the mobile app. They
are not a production mobile authentication model.

Mobile booleans accept `1`, `true`, `yes`, `0`, `false`, and `no`.

| Variable | Required | Purpose |
| --- | --- | --- |
| `EXPO_PUBLIC_STUFF_STASH_API_BASE_URL` | yes | API base URL used by the mobile app. |
| `EXPO_PUBLIC_STUFF_STASH_TENANT_ID` | yes | Initial tenant ID for local development flows. |
| `EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN` | yes | Local development bearer token value. |
| `EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED` | no | Enables mobile voice developer diagnostics. |

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
| `GO_BUILDER_IMAGE` | pinned digest | API builder image override. Must stay digest-pinned. |
| `RUNTIME_IMAGE` | pinned digest | API runtime image override. Must stay digest-pinned. |
| `NODE_BUILDER_IMAGE` | pinned digest | Web builder image override. Must stay digest-pinned. |
| `WEB_RUNTIME_IMAGE` | pinned digest | Web runtime image override. Must stay digest-pinned. |
| `PNPM_VERSION` | `11.0.7` | pnpm version used by the web image build. |
