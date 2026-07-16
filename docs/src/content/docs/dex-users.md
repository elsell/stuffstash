---
title: Dex Users And Clients
description: Replace the first-run Dex account before using Stuff Stash for real household data.
---

The bundled Dex service is a first-run identity provider. It lets you test the
same OIDC sign-in boundary that a real deployment uses.

Do not keep the checked-in users or client secrets for household data.

Dex does not provide a built-in user-management UI in this Compose topology.
Manage users by keeping a private Dex config file, restarting the Dex service,
and backing up that private file with your other secrets.

## Create A Private Dex Config

Copy the example config out of the tracked deployment directory:

```sh
mkdir -p .stuffstash/selfhost/dex
cp deploy/selfhost/dex/config.yaml .stuffstash/selfhost/dex/config.yaml
chmod 600 .stuffstash/selfhost/dex/config.yaml
```

Point `.env` at the private copy:

```text
DEX_CONFIG_PATH=.stuffstash/selfhost/dex/config.yaml
```

Keep `.stuffstash/selfhost/dex/config.yaml` out of Git.

## Replace The First-Run Users

Create a bcrypt password hash for each household user:

```sh
docker run --rm dexidp/dex:v2.44.0@sha256:5d0656fce7d453c0e3b2706abf40c0d0ce5b371fb0b73b3cf714d05f35fa5f86 \
  dex hash-password
```

Edit `staticPasswords` in your private Dex config. Use real email addresses,
new `userID` values, and the generated hashes. Remove the example
`owner@example.com` and `viewer@example.com` entries.

## Replace Static Clients

The web client is public and does not need a client secret. Keep its ID and
redirect URI aligned with `.env`:

```yaml
- id: stuff-stash-web-local
  name: Stuff Stash Web
  public: true
  redirectURIs:
    - https://stuffstash.localhost:8081/callback
```

Remove or rotate every checked-in confidential client before real use. The
example Dex config includes `stuff-stash-local` with a known secret for local
testing. If you keep a confidential client, replace `secret` with a new random
value and keep the matching client ID in `.env` only if the API should accept
tokens for that client.

For the bundled web app alone, keep these `.env` values aligned:

```text
STUFF_STASH_WEB_OIDC_CLIENT_ID=stuff-stash-web-local
STUFF_STASH_OIDC_CLIENT_IDS=stuff-stash-web-local
```

Add mobile or other private client IDs to `STUFF_STASH_OIDC_CLIENT_IDS` only
after you add matching private Dex clients with rotated secrets or safe public
client settings.

If you change the hostname, update these values together:

- `STUFF_STASH_WEB_ORIGIN`
- `STUFF_STASH_API_ORIGIN`
- `STUFF_STASH_OIDC_ISSUER`
- `STUFF_STASH_WEB_OIDC_REDIRECT_URI`
- Dex `issuer`
- Dex `web.allowedOrigins`
- Dex client `redirectURIs`

## Restart Dex

After editing the private config, restart the stack:

```sh
docker compose -f compose.selfhost.yaml down
docker compose -f compose.selfhost.yaml up
```

Sign in with the new user before adding household data.

## Back Up Identity Config

Back up these files with your other deployment secrets:

- `.env`
- `.stuffstash/selfhost/dex/config.yaml`
- the Caddy local CA if browsers trust it directly

If you lose the private Dex config, existing Stuff Stash records remain in
Postgres, but users may not be able to sign in with the same OIDC subject.
