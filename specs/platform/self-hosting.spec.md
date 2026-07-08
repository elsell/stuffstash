# Self-Hosting Spec

## Purpose

Stuff Stash needs a public self-host path that a technically capable household
operator can run without reading source code or depending on contributor-only
development loops.

## Requirements

- The public self-host path must use Docker Compose with Caddy HTTPS, bundled Dex
  OIDC, Postgres metadata persistence, datastore-backed SpiceDB, Garage media
  storage, and the static web container.
- `compose.selfhost.yaml` must use published Stuff Stash API and web images by
  default. Building from source is a contributor option, not the public happy
  path.
- Published API and web image references used by the self-host env example must
  be immutable digest references.
- The release workflow must update a tracked self-host release env artifact with
  the newly published API and web image digests before creating the release.
- Postgres services in self-host Compose must set `PGDATA` inside the mounted
  data volume so database rows survive `docker compose down` and a later
  `docker compose up`.
- Self-host restart verification must preserve the same Dex user, tenant,
  inventory, assets, authorization state, and Garage objects without relying on
  container-local state.
- Garage direct browser upload must work in the default self-host topology.
  Falling back to the API JSON upload route is acceptable only for explicit
  local-development sentinel targets or non-Garage client limitations; it must
  not hide Garage self-host misconfiguration.
- The bundled Dex config is a first-run example only. Public docs must link to
  an operator-safe recipe for replacing static users and clients before
  household use.
- Dex operator guidance must be honest that Dex has no built-in user-management
  UI in this topology. If persistent Dex storage is introduced later, the spec
  must define its storage mode, backup requirements, and user-management
  workflow before docs present it as an operator UI.
- Public docs must name the persistent volumes and explain which hold Stuff
  Stash metadata, SpiceDB authorization data, Garage metadata, Garage object
  data, and Caddy certificates.

## Verification

- `scripts/check-selfhost-happy-path.sh` must reject self-host Compose drift that
  can be checked mechanically, including source-build defaults, missing
  published image digest references, missing Postgres `PGDATA` under mounted
  volumes, missing Garage CORS setup, and docs that still tell operators to run
  source builds by default.
- Browser-level self-host audits must verify Dex sign-in, first inventory
  creation, item creation, Garage-backed image upload, reload, and restart
  durability.
