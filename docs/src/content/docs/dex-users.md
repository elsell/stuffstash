---
title: Dex Users And Clients
description: Replace the first-run Dex identity before using household data.
---

Dex has no user-management screen in the bundled setup. Users live in a
private config file that you edit and back up.

## 1. Make The Config Private

```sh
mkdir -p .stuffstash/selfhost/dex
cp deploy/selfhost/dex/config.yaml .stuffstash/selfhost/dex/config.yaml
chmod 600 .stuffstash/selfhost/dex/config.yaml
```

Set its path in `.env`:

```text
DEX_CONFIG_PATH=.stuffstash/selfhost/dex/config.yaml
```

Compose securely stages this file for the non-root Dex container. Keep the host
copy at mode `600` and out of Git.

## 2. Replace Users

Create a bcrypt hash for each password:

```sh
docker run --rm dexidp/dex:v2.44.0@sha256:5d0656fce7d453c0e3b2706abf40c0d0ce5b371fb0b73b3cf714d05f35fa5f86 \
  dex hash-password
```

In `staticPasswords`, replace both example users. Give each person a real email,
a unique `userID`, and a generated hash.

## 3. Check Clients

The bundled web client is public and needs no secret. Keep these values aligned:

```yaml
- id: stuff-stash-web-local
  public: true
  redirectURIs:
    - https://stuffstash.localhost:8081/callback
```

```text
STUFF_STASH_WEB_OIDC_CLIENT_ID=stuff-stash-web-local
STUFF_STASH_OIDC_CLIENT_IDS=stuff-stash-web-local
```

Remove the example confidential client, or replace its known secret. Add mobile
or other client IDs only when matching clients exist in Dex.

For LAN hostname changes, follow [Use Stuff Stash On Your
LAN](../self-host-operations/#use-stuff-stash-on-your-lan).

## 4. Apply And Test

```sh
./scripts/selfhost-preflight.sh
docker compose -f compose.selfhost.yaml up -d
```

Sign in with the new user before adding household data. Back up `.env` and the
private Dex config together; losing the config can change OIDC identities and
prevent existing users from signing in.
