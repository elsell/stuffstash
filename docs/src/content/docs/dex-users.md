---
title: Dex Users And Clients
description: Replace the bundled Dex users and clients.
---

Dex has no user-management screen in the bundled setup. Users live in a
private config file that you edit and back up.

## 1. Open The Private Config

The setup command creates the file named by `DEX_CONFIG_PATH` in `.env`. Edit
that file in place so its generated IP address or DNS name stays intact. Keep
it at mode `600`, out of Git, and in your backups.

## 2. Replace Users

Create a bcrypt hash for each password:

```sh
docker run --rm dexidp/dex:v2.44.0@sha256:5d0656fce7d453c0e3b2706abf40c0d0ce5b371fb0b73b3cf714d05f35fa5f86 \
  dex hash-password
```

In `staticPasswords`, replace both example users. Give each person a real email,
a unique `userID`, and a generated hash.

## 3. Check Clients

The bundled web client is public and needs no secret. Keep its generated
redirect URI aligned with `STUFF_STASH_WEB_OIDC_REDIRECT_URI` in `.env`:

```yaml
- id: stuff-stash-web-local
  public: true
  redirectURIs:
    - https://<server-address>:8081/callback
```

```text
STUFF_STASH_OIDC_CLIENT_ID=stuff-stash-web-local
STUFF_STASH_WEB_OIDC_CLIENT_ID=stuff-stash-web-local
STUFF_STASH_OIDC_CLIENT_IDS=stuff-stash-web-local,stuff-stash-mobile-local
```

Remove the example confidential client, or replace its known secret. Add mobile
or other client IDs only when matching clients exist in Dex. Keep
`stuff-stash-mobile-local` when the bundled mobile client remains enabled.

## 4. Apply And Test

Return to [Replace The Example
Credentials](../self-host-operations/#replace-the-example-credentials) to
replace the remaining example secrets. To apply a Dex-only change on a private
installation:

```sh
docker compose -f compose.selfhost.yaml down
./scripts/selfhost-preflight.sh --strict
docker compose -f compose.selfhost.yaml up -d
```

Sign in with the new user before adding household data. Back up `.env` and the
private Dex config together; losing the config can change OIDC identities and
prevent existing users from signing in.
