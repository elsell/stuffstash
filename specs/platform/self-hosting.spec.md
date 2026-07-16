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
- The release workflow must create or refresh a branch-protected pull request
  that updates the tracked self-host release env artifact with the newly
  published API and web image digests before creating the release.
- The release workflow must explicitly dispatch and require success from the
  required CI workflow for that digest-update branch when the automation token
  would otherwise suppress normal pull request checks.
- The release workflow must merge the validated digest-update pull request
  before creating the GitHub release, while keeping the release tag on the
  source commit that produced the published image attestations.
- Each GitHub release must attach a checksum-protected self-host bundle that
  contains the Compose file, digest-pinned environment example, mounted
  configuration, and operator scripts needed to start that release. Public
  quick-start documentation must use this bundle instead of cloning a moving
  branch.
- A release bundle must embed the API and web digest references produced by
  that release, even when the tracked env example is updated later by PR.
- Production documentation deployment must wait for the main-branch Release
  workflow to finish successfully so newly documented release assets exist
  before the public quick start points to them.
- The public quick start must configure direct LAN access by default. Published
  ports may bind to all host interfaces so a browser on another household
  device can open the app without SSH forwarding.
- Compose must declare a stable project name so release-bundle directory names
  do not change the names of persistent volumes during upgrades.
- The shared HTTPS host may be a valid DNS name or strict IPv4 literal. It must
  not be a URL, include a port or userinfo, or contain malformed DNS labels or
  IPv4 octets. IPv6 literal support is deferred until URL rendering, Caddy, and
  client behavior are specified and tested together.
- Caddy must issue a local certificate for the configured host and provide that
  host as its default SNI value. IP-literal TLS clients omit hostname SNI, so
  the default is required for the web app, API, Dex, and Garage to remain
  reachable at an IPv4 address.
- The release bundle must include one setup command that creates `.env`, detects
  or accepts the server LAN IPv4 address, keeps every public origin and OIDC
  callback consistent, creates the private Dex config, and prints the final web
  URL. DNS configuration must remain optional.
- A private Dex config must remain private on the host. Compose must stage it
  into a named volume with ownership and mode readable by the non-root Dex
  process rather than requiring a world-readable host file.
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
- Strict preflight must reject the bundled Dex emails, client secret, user IDs,
  and known password hash. Renaming a public example account must not make its
  public password acceptable for household use.
- Dex operator guidance must be honest that Dex has no built-in user-management
  UI in this topology. If persistent Dex storage is introduced later, the spec
  must define its storage mode, backup requirements, and user-management
  workflow before docs present it as an operator UI.
- Public docs must name the persistent volumes and explain which hold Stuff
  Stash metadata, SpiceDB authorization data, Garage metadata, Garage object
  data, and Caddy certificates.
- The public happy path must run an operator preflight before Compose. It must
  check Docker Compose availability, host and origin consistency, port
  availability, bind-address intent, and required files. The default check may
  warn about bundled example users and secrets so a newcomer can start without
  a separate evaluation mode. An explicit `--strict` check must reject them for
  operators who want a hardened configuration.
- Public documentation must keep the first-run path concise and sequential.
  Certificate trust, LAN exposure, secrets, backup, and upgrade detail must be
  visually separated or linked from the quick start.
- The quick start must follow one top-to-bottom LAN path, end at the configured
  IPv4 URL, explain how the browser device receives the Caddy root, and state
  plainly that anyone who can reach the server can use the bundled example
  credentials. SSH forwarding, optional DNS, private users, backups, and wider
  network exposure belong in visually separate operator guidance.

## Verification

- `scripts/check-selfhost-happy-path.sh` must reject self-host Compose drift that
  can be checked mechanically, including source-build defaults, missing
  published image digest references, missing Postgres `PGDATA` under mounted
  volumes, missing Garage CORS setup, unsafe default port binding, bypassed Dex
  config staging, missing release-bundle assets, and docs that still tell
  operators to clone a moving branch or run source builds by default.
- CI must configure the release bundle with the runner's LAN IPv4 address,
  start that topology with a mode-`0600` private Dex config, verify the public
  health and OIDC discovery endpoints through the IP-literal HTTPS edge, and
  tear down the isolated project and volumes afterward.
- Browser-level self-host audits must verify Dex sign-in, first inventory
  creation, item creation, Garage-backed image upload, reload, and restart
  durability.
