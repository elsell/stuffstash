# Stuff Stash CLI And Print Consumer

## Status And Scope

Accepted design direction, October 3, 2026; not implemented. Ship a project-owned
Go CLI named `stuffstash` under `apps/cli`, separate from the API server binary.
It is an API client and a host for local connector adapters. It never connects
directly to the database, SpiceDB, or another app's internal packages.

This release specifies login/context selection, basic inventory actions, label
render/print commands, printer discovery/registration, and the long-running USB
print consumer. It does not promise full REST parity, interactive terminal UI,
dynamic plugin loading, a general event bus, or service installation automation.

See [printer integration](../printing/printer-integration.spec.md) for security
and job semantics and [asset labels](../printing/asset-labels.spec.md) for QRs.

## Human Commands

The intended command vocabulary is:

```text
stuffstash login --server <https-url>
stuffstash login --server <https-url> --device-code
stuffstash logout
stuffstash inventories list
stuffstash assets list --inventory <id>
stuffstash assets show <asset-id>
stuffstash assets create --kind <item|container|location> --title <title>
stuffstash assets update <asset-id> --title <title>
stuffstash assets move <asset-id> --parent <asset-id|root>
stuffstash assets archive <asset-id>
stuffstash assets restore <asset-id>
stuffstash labels render <asset-id> --format <png|pdf> --output <path>
stuffstash labels print <asset-id> --printer <printer-id>
stuffstash labels resolve <label-url>
stuffstash printers list
stuffstash printers discover
stuffstash printers test <printer-id>
stuffstash connectors print register
stuffstash connectors print run --printer <printer-id>
stuffstash print-jobs list
stuffstash print-jobs show <job-id>
stuffstash print-jobs cancel <job-id>
stuffstash print-jobs reprint <job-id>
```

- Inventory commands expose the same authorized application behavior, typed
  kinds, pagination, lifecycle rules, and audit as the REST clients. They do not
  reimplement domain validation or construct private persistence operations.
- Every scoped command has explicit server, tenant, and inventory context from
  flags or environment-backed configuration. Reject ambiguity; do not silently
  select the first inventory/printer or infer a target from a display name.
- Creation supports explicit `--print-label` and destination/template overrides.
  Omission does not apply the UI's inventory default invisibly to scripts.
- Asset list/search and mutation flags must map to the generated contract. Only
  the commands above are the first supported inventory surface; further REST
  parity must be specified intentionally rather than guessed at runtime.
- `printers list` queries registered inventory destinations, including offline
  ones. `printers discover` examines the local machine without registration or
  printing. Make this distinction clear in help and structured output.
- `labels render`, `labels print`, and creation with `--print-label` accept
  `--template`, `--template-version`, and validated options such as
  `--show-reference`. Resolve omitted templates from inventory defaults. Add
  `stuffstash labels templates` to inspect the authorized versioned catalog.
  Queued printing always uses the destination's registered size; no per-job
  media-size flag or separate media-catalog command is required. Standalone
  download rendering can specify output dimensions without registering a printer.
- `labels print` enqueues an API job and returns its ID; it does not synchronously
  talk to a USB device. `printers test` is an explicit, authorized test-label job
  with the same queue/status semantics; discovery and heartbeat never print.
- `labels resolve` uses the same instance-checked parser and API resolver as the
  client contract, never follows an arbitrary scanned origin with credentials.
- Provide human-readable output and `--json` for finite commands. Paginated
  commands expose cursor/limit and do not imply the first page is the whole set.
  Errors have stable categories and nonzero exit status; never report a queued
  job as printed. Return job IDs even when an optional wait times out.
- Human output is emitted through a presentation/output port. Diagnostics use
  injected observability, keeping machine-readable stdout free of log noise.
- Idempotency keys are generated per logical mutation and reused for transport
  retries. Accept an explicit key for automation; never retry non-idempotent
  creation with a freshly generated key after an ambiguous response.

## Human Authentication And Local Configuration

- Use system-browser OIDC authorization code with PKCE, state, and nonce. Discover
  public CLI client metadata from the API; never collect provider passwords or
  ship a client secret. Add a separate public CLI OIDC client registration and
  audience validation; do not silently repurpose the mobile callback/client ID.
- Use a loopback callback bound only to 127.0.0.1 on an ephemeral port and a
  one-time path/state. Validate issuer, audience, nonce, redirect, and returned
  state. Require provider support for the registered loopback redirect policy;
  fail with setup guidance if unsupported.
- Support RFC 8628 device authorization for human login through `--device-code`
  in the first release. It requires the configured issuer to advertise a device
  authorization endpoint and the operator to enable the grant for the public CLI
  client. Do not assume every OIDC provider supports it or ship a client secret.
- Show the provider verification URL and user code; the user approves in a browser
  on another device. No callback listener or local browser is required. Keep the
  secret device code out of output/logs and clear pending state on completion,
  denial, expiration, or cancellation. Poll at the provider's interval; honor
  `authorization_pending`, `slow_down`, `access_denied`, and `expired_token`, with
  bounded backoff on transport failures and a hard expiry deadline.
- Device flow uses the configured issuer's discovered HTTPS endpoints, not
  arbitrary endpoints from input. Validate the returned OIDC ID token's signature,
  issuer, CLI audience, and expiry before storing a session. The provider must
  support issuing an ID token for this grant and the requested `openid` scope;
  an OAuth access-token-only response is incompatible with the current API and
  must fail with guidance, not be reinterpreted as an ID token. Apply device-flow
  validation independently; browser-only callback state/PKCE rules do not imply
  that device flow uses a loopback callback.
- Both login methods create the same human principal and obey the same SpiceDB
  permissions. Device authorization does not register a print consumer or mint a
  service-account credential. Connector pairing remains a separate API workflow.
- If device flow is unavailable, fail explicitly and explain browser login with
  PKCE where usable; never start a browser silently or fall back to password
  collection. Test the bundled pinned Dex setup and document provider prerequisites
  before claiming headless human login works. Standard reference:
  [RFC 8628](https://www.rfc-editor.org/rfc/rfc8628).
- Public CLI auth discovery is `GET /auth/cli/config`, containing issuer, public
  client ID, scopes, allowed loopback redirect policy, and enabled login methods.
  Device-flow availability must reflect issuer discovery and operator policy. It is additive to
  mobile auth discovery and must not expose secrets or change its callback rules.
- Match the existing API's verified OIDC ID-token bearer convention using the
  explicitly allowed CLI audience. Refresh through the configured issuer when
  supported; refresh failure requires login. No acceptance of arbitrary audiences.
- Human auth is needed for inventory commands. An unattended print worker uses
  the browser-approved, restricted connector credential instead; pairing may be
  approved from another device, so a headless USB host needs no local browser.
  This credential authenticates a service account; the API authorizes printer
  access with SpiceDB on every operation. It does not encode a trusted local
  permission list, and the CLI never connects directly to SpiceDB.
- Permission changes take effect without replacing connector credentials. Honor
  authorization-pending state during pairing and fail-closed denial/outage during
  work; credential rotation retains the same service-account identity.
- Store secrets behind a credential-store port using an OS store where usable;
  on headless Linux support an explicitly configured owner-only credential file,
  refuse unsafe permissions, and write atomically. Do not copy arbitrary shell
  environment or use a logged-in human's token as the worker's credential.
- Operational settings come from validated environment-backed configuration;
  flags may override non-secret settings and select context. Provisioned IDs,
  credential references, and recovery journals are persistent local state, not
  hard-coded deployment configuration. Never require secrets on command lines.
- HTTPS is required except explicitly configured loopback development service
  access. Certificate validation stays enabled; no inherited insecure behavior
  from the historical Homebox script.

## Built-In Printer Adapter

- Ship only the required Brother QL-800 adapter and 29 x 90 mm profile initially.
  Bundle them with the CLI distribution; other models/media are future extensions.
  No separate user-installed Stuff Stash plugin or Python/uv setup is part of
  the intended experience. OS USB permissions or a packaged native USB runtime
  may require documented installation setup.
- Keep command parsing, consumer orchestration, credential storage,
  device discovery, local locking/journaling, and printer I/O separate. Label
  composition runs in the API-side renderer. The CLI submits template selections
  and consumes immutable artifacts; it does not bundle a second layout engine.
  Thin Go entrypoints assemble injected ports in focused bootstrap packages.
- The printer port supports discovery, capabilities, readiness, submit, and
  completion observation, with typed outcomes distinguishing definite no-output
  failure from uncertain/partial output. Wake is optional and unsupported until
  verified by an adapter. Generic orchestration contains no Brother protocol.
- First release acceptance platform is Linux with a USB QL-800. Other CLI builds
  may support normal API commands without supporting local USB printing; report
  this capability honestly. Windows/macOS USB adapters are follow-up work.
- Adapt the Brother raster/status protocol through a focused adapter. Before
  implementation, record exact reviewed Go/USB dependencies or protocol-source
  versions in the tooling spec. Do not adopt floating brother_ql dependencies
  or assume the historical Python package is the chosen production dependency.
- Hardware spike must determine status/finish reporting, device identity,
  optional media detection, USB claim behavior, and physical rendering. If a
  packaged helper proves necessary, specify its pinned version, packaging, and
  license first; preserve the one-install user experience.
- Acquire an OS-level lock by physical device identity. A second worker reports
  Device already in use and does not claim work. Permissions use a narrow Linux
  udev/group setup; do not run the entire network-connected CLI as root.
- QL-800 auto power-off is supported as an ordinary unavailability case. Do not
  promise USB wake. Brother's documented Auto Power On concerns AC connection;
  disabling Auto Power Off is optional operator guidance, not an automatic
  configuration change. Smart-outlet adapters are deferred.

## Consumer Runtime

- Registration selects device and label size together, for example QL-800 with
  29 x 90 mm stock. The built-in preset supplies margins/resolution/orientation;
  size selection is an explicit user action, not inferred from USB availability.
  Registration pairs and persists the approved printer/media settings and
  connector credential with the server. A restart reuses this registration.
- Add `stuffstash printers configure <printer-id> --label-size <supported-size>`
  as a human-authorized, revision-checked update of the same registration.
  The label-size argument is an exact preset ID from the authenticated profile
  catalog for that printer's adapter. Fetch the current registration, resolve one
  supported preset version, then PATCH only revision, preset ID, and preset version.
  Preserve name, retirement state, adapter, and connector assignments. Unsupported
  or ambiguous presets fail without mutation; a stale revision fails without an
  automatic overwrite/retry. Offline readiness does not block configuration.
  Initial support remains only `brother-ql800-29x90` version 1.
  Web/mobile printer settings expose the same edit. Worker credentials cannot
  change media settings; they retrieve and apply the server's configured snapshot.
  No automatic roll detection, multiple saved-roll UI, or recurrent confirmation
  is required. Trust the configured size; report actual hardware errors normally.
- Changing size never resizes a queued artifact. Only claim jobs compatible with
  current configuration, and refresh/recheck it before start. A different size
  leaves the original job waiting; changing back makes matching jobs eligible.
- `connectors print run` stays in the foreground and uses outbound HTTPS to
  heartbeat, report printer state, claim work, renew leases, fetch authorized
  artifacts, and report outcomes. It opens no network listener.
- Consume the versioned job/device interface in
  [printer integration](../printing/printer-integration.spec.md#print-job-and-device-interface).
  Multiple templates may target one printer. Printer adapters preserve the
  API-rendered composition, including orientation, instead of interpreting a
  template or silently fitting the image to different stock.
- Fetch artifacts only from the configured authenticated API path. Validate
  content type, digest, size, and profile. Never execute a job-supplied command,
  fetch arbitrary URLs, or render arbitrary HTML on the printer computer.
- Journal attempts atomically on local persistent storage. Recover/reconcile
  before claiming more work. If state is missing or corrupt, do not assume a
  previous physical submission was absent and replay it.
- A credential may identify an installation across restarts; each process has a
  separate session ID. Sharing credentials does not authorize claiming another
  session's active attempt. Duplicate configured workers obey API reservations
  as well as local locks.
- Reconnect with bounded jittered backoff; keep valid jobs on the server while
  offline. Never start an unacknowledged/expired claim or hide revocation behind
  an infinite retry loop. Terminal credential failure reports re-pair guidance.
- Graceful shutdown stops claims, releases unstarted work when safe, and drains
  or records uncertainty for submitted work. Forced shutdown is covered by
  server lease expiry and reconciliation, not by assuming a clean exit.
- Support running the same foreground command under systemd or another service
  manager with a dedicated user and persistent state directory. Future docs must
  include verified installation, pairing, USB permission setup, status checks,
  restart/recovery, revocation, and uninstall instructions. Automatic service
  installation is not required for the first release.

## Tests, Packaging, And Evidence

- Add `apps/cli` to the Go workspace and root test/format/structural-hook coverage
  when implementation begins. No CLI entrypoint, adapter, or source file is
  exempt from project architecture, observability, or size rules.
- Every Stuff Stash API call, including auth configuration, pairing, printer
  status, claim/recovery, and artifact endpoints, uses a Go SDK automatically
  generated from the same Huma-produced OpenAPI artifact used by web/mobile:
  `packages/api-client/openapi.json`. The existing client there is TypeScript;
  the Go SDK now uses that same contract. No separately hand-maintained Go API schema or
  endpoint/DTO layer is permitted.
- Generate the Go SDK into `apps/cli/internal/adapters/httpapi/generated` with a
  reviewed pinned generator recorded in the tooling spec before implementation.
  Keep generated transport behind a project-owned CLI API port; map DTOs there
  and never import the API module's `internal` packages or edit generated files.
- Add root `make cli-client-generate` to regenerate the canonical OpenAPI artifact
  first and then the Go SDK reproducibly. Add `make cli-client-check-generated`
  to regenerate in isolation and fail on missing, stale, or unexpected output.
  CI and applicable pre-commit checks must catch drift from API route/DTO,
  OpenAPI artifact, generator configuration/version, or SDK changes. Retain the
  existing TypeScript generation and drift checks from the same source.
- SDK authentication hooks, streaming, timeouts, and error mapping are transport
  concerns behind the adapter. No hand-written REST escape hatch for printer
  operations. Provider OIDC discovery/token/device endpoints are external
  protocols behind the auth adapter, not Stuff Stash OpenAPI endpoints; use a
  reviewed pinned standards implementation for those integrations.
- Unit/application tests use behavioral in-memory fakes for API, clock, printer,
  credential storage, journal, and output. Adapter integration tests verify actual
  parsing/transport and atomic persistent recovery, not mock call sequences.
- End-to-end auth tests cover callback forgery/replay, wrong issuer/audience,
  expired tokens, cross-tenant scope, wrong role, pairing/revocation, and allowed
  human versus connector operations. CLI subprocess tests cover stable JSON,
  exit statuses, secrets absent from logs, cancellation, and restart recovery.
- Device-flow tests cover supported/unsupported discovery and operator policy,
  successful headless login, wrong issuer/audience/signature, access-token-only
  response rejection, expiry/denial/cancellation, polling slowdown, transport
  recovery, and no device-code/token leakage. Include a real configured OIDC
  provider integration; fakes alone do not establish Dex compatibility.
- Package pinned dependencies and required notices; verify checksums/provenance
  and a clean Linux installation that does not require the old Python project.
- Physical printing and scan checks remain explicitly unverified until performed.
  Collect user-device checks in `docs/reports/user-testing-checklist.md` during
  implementation; do not repeatedly request them or block unrelated release work.
- The connector and Linux USB adapter are implemented. Stateful protocol,
  journal, and worker tests verify recovery without duplicate submission. Real
  Dex browser PKCE and device-code login have been exercised against the API;
  these checks do not establish physical printer or scan behavior. Read-only
  discovery on Paul found no currently attached Brother device. Physical
  printing and scanning remain on the user-testing checklist.

## Generated First-Party Printer Documentation

Built-in adapter registrations expose offline, deterministic public descriptors
for [generated printing docs](printing-catalog-docs.spec.md). Registered runtime
adapters, supported platforms/transports, and media presets drive the website's
support catalog. Verify export/runtime parity; no docs-only list or USB access is
required for export. Candidate support and physical verification remain distinct.

## GitHub Release Binaries And Download Documentation

- Once the CLI ships, every normal project release that is cut must build and
  publish the CLI from that exact release commit under the same project tag.
  Do not create an unrelated CLI version stream or publish an unversioned build
  from a moving branch. Documentation-only changes retain existing no-release
  behavior; they do not manufacture tags just to refresh download instructions.
- Embed the exact release tag/version and source commit in the binary using
  build-time values. `stuffstash version` and `stuffstash version --json` expose
  both plus OS/architecture and enabled connector capabilities. Development
  builds clearly identify themselves as development; a release build must fail
  if required version inputs are absent or inconsistent with its release plan.
- Attach per-platform archives, checksum manifest, and available build provenance
  to the GitHub release. Each archive includes the executable and required
  licenses/notices. Archive names include tag, OS, and architecture, with a stable
  naming convention documented before publishing the first CLI release.
- Initial required downloadable target is Linux amd64, covering the user's USB
  QL-800 host; additional API-only platforms may be shipped when verified, but
  must not imply USB printing support on unimplemented platforms. Decide and
  record the supported release matrix before wiring builds. Pin toolchain,
  build actions, system/native USB dependencies, and generator versions.
- Build and smoke-check `version --json` against the planned tag/commit before
  publication. Verify archives can be unpacked and the documented binary runs
  on supported clean hosts. Checksum/provenance generation covers final bytes,
  including any required native runtime, not an earlier unsigned artifact.
- Publish release assets before advertising download availability. Partial
  failures must not advance the documented current CLI version; never overwrite
  existing versioned assets with different bytes. A retry verifies existing
  assets and completes missing publication idempotently.
- The docs install section shows a copyable `curl --fail --location` command
  using the exact newest successfully published stable CLI tag and a compatible
  platform archive, followed by checksum verification, extraction, and a version
  check. No `curl | sh`, floating binary URL in the generated command, or promise
  that a missing platform asset exists. A release-index link may be provided too.
- On successful release publication, automatically refresh version/asset/hash
  metadata through the existing release-maintenance update mechanism and trigger
  the documentation build. Generate the install snippet from this metadata;
  never manually replace versions across prose. The update must not trigger an
  endless release loop. Track docs-update failures separately and retry safely.
- Metadata contains only verified published assets and records the release
  commit. PR/offline docs builds use checked-in metadata without querying GitHub
  or guessing from tags; release automation performs online asset verification.
  Prerelease/draft/failed releases do not replace the stable install snippet.
- Before the first CLI release, docs say it is not yet available rather than
  emitting a broken download command. Generated supported-printer pages link to
  installation instructions, keeping binary download facts in one place.

## Delivery And Critical-Test Budget

The user explicitly prioritizes full delivery in small, coherent PRs and asks
for coordinated parallel agents. Parallelize renderer/registry, API/security/job
services, CLI/SDK/release, and client/docs work at stable interfaces with explicit
file ownership and integration review. Do not ship disconnected scaffolding as
completed functionality or mistake a passing isolated module for the full feature.

Write only critical tests: authentication/authorization and tenant isolation,
atomic claims/idempotency and duplicate-output risks, crash/recovery paths,
render/scan integrity, essential connected user flows, and release/generated
artifact correctness. Existing acceptance lists name risks to verify, not a
mandate for one test per bullet or exhaustive combinations. Do not write tests
that repeat generated code, constants, types, or implementation statements already
codified by the code. Preserve required checks and review, use behavioral fakes,
and resolve integration defects rather than expanding low-value test coverage.

### First API Authentication Prerequisite Slice

Public CLI discovery uses `STUFF_STASH_OIDC_CLI_CLIENT_ID`,
`STUFF_STASH_OIDC_CLI_SCOPES` (default `openid,email,profile,offline_access`), and
`STUFF_STASH_OIDC_CLI_DEVICE_AUTH_ENABLED` (default false). A configured CLI client
is added explicitly to accepted OIDC audiences. Missing client or non-OIDC mode
returns unavailable metadata, without affecting existing mobile sign-in.

`GET /auth/cli/config` returns the standard envelope with `issuer`, `clientId`,
`scopes`, `loginMethods` (`authorization_code`, optionally `device_code`), and
`loopbackRedirect: {host: "127.0.0.1", pathPrefix: "/callback/", ephemeralPort: true}`.
Device login is advertised only when operator-enabled and discovered from the
configured issuer. Discovery failure is a startup/configuration error when CLI
OIDC is enabled; never advertise guessed provider capabilities. Bundled Dex uses
a separate public CLI client whose empty redirect list permits Dex's supported
loopback/device authorization policy. These values are public, never secrets.
## Initial Human CLI Delivery

The first executable slice provides login, logout, version, inventories list, and
asset list/show/create/update/move/archive/restore. Printer commands arrive with
their API contracts. No unavailable command is presented as operational.
`STUFF_STASH_CLI_SERVER`, `STUFF_STASH_CLI_TENANT`, and
`STUFF_STASH_CLI_INVENTORY` provide explicit context; flags override them.
`STUFF_STASH_CLI_CREDENTIAL_FILE` explicitly selects the headless file store.
`STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP=true` permits HTTP only for loopback
development API/issuer URLs. The normal default credential store is the OS
keyring, keyed by canonical API origin. Login never silently switches servers.
Finite commands accept context/output flags before or after positional arguments.
Human inventory output includes IDs and pagination continuation; JSON preserves
client-owned result models and pagination, without token or provider error bodies.
Initial portable release targets are Linux/macOS amd64 and arm64, plus Windows
amd64; API commands use pure Go builds. USB capability is separately advertised
from the registered adapter platforms; Linux includes the Brother adapter.

## Initial Linux Brother Transport

The first adapter uses Linux's bidirectional `usblp` character device through
standard nonblocking file I/O and poll. Operators enable the kernel module,
disable Editor Lite using the hardware control where needed, and grant a narrow
printer group access. Discovery reads USB/sysfs metadata without opening the
printer or transmitting data. No automatic kernel-driver detach, Editor Lite or USB-interface
configuration changes, privileged helper, Python runtime, or libusb installation
is performed. Missing device nodes, unsupported unidirectional interfaces, or
permissions are actionable unavailable states. Serial identity is preferred;
serial-less devices use an explicitly reported physical-port identity.

The printer port submits one physical copy at a time. The worker owns bounded
copy sequencing, lease renewal, and durable completed-copy progress under its
exclusive physical-device lock. Submit requires an immutable validated PNG and
exact registered media profile. The adapter does not resize, rotate, dither, or
compose content. It packs the feed-oriented monochrome raster into Brother's
720-pin rows with the documented offset/bit order.

Completion requires both a spontaneous printing-completed status and the following
waiting-to-receive phase from the same open connection. Successful USB writes
are only submission evidence. Observation never sends status requests during
printing. A partial write, disconnect, malformed status, or timeout after possible
output is uncertain, including a partial set of copies; never infer no output
from a missing final print command. No new submission can follow uncertainty
on the same connection. Reopening is not recovery evidence.

`stuffstash printers discover` performs local read-only discovery without login.
`stuffstash printers catalog --json` exports deterministic built-in descriptors
and media presets without probing devices. Catalog availability is distinct from
physical verification. QL-800 with 29 x 90 mm stock is the only initial profile.
Physical verification is pending: read-only inspection on Paul found no attached
Brother printer or usblp node on October 3, 2026.

The built-in preset records the manufacturer's effective height 89.8 mm while
its user-facing stock name remains 29 x 90 mm. Physical margins are 18 dots
(1524 micrometers) across and 35 dots (2963 micrometers) along the feed. Raster
size remains 306 x 991 at 300 DPI; rounded physical dimensions are subject to the
renderer one-dot tolerance. Brother's printhead offsets (6 right, 408 left) are
transport mechanics, distinct from these physical margins. Sysfs and device-root
defaults can be configured by `STUFF_STASH_CLI_SYSFS_USB_ROOT` and
`STUFF_STASH_CLI_USB_DEVICE_ROOT`; job data cannot override local device paths.

### Retryable release publication

The release workflow builds the five portable CLI targets and self-host bundle
from the validated release commit before creating a tag. It retains those exact
bytes, release notes, image references, and a SHA-256 publication manifest as a
GitHub Actions artifact. A release starts as a draft; publication uploads missing
assets, verifies every downloaded asset against the manifest, and only then makes
it stable. Existing assets with different bytes are an error, never overwritten.

An explicit `repair_run_id` workflow input resumes the retained artifact of a
completed Release run in this repository on main. The workflow verifies the
original workflow identity, run commit, manifest commit, and tag target before
writing release state. It does not rebuild from a moving branch or guess an old
release's contents. Missing/expired artifacts require investigation, not silent
replacement. Ordinary release planning is unchanged and repair does not cut a new
tag or trigger TestFlight again.

Only a fully verified, published, non-prerelease release can update checked-in CLI
download metadata. The maintenance PR regenerates exact versioned URLs and hashes;
its `chore(release)` commit does not manufacture another release. Older repair runs
cannot replace a newer stable download. Production docs are dispatched against
main after that PR is observed merged; requesting auto-merge alone is insufficient.
A failed docs refresh is reported separately and can be retried by dispatching
Docs Pages against main. Failed Release runs never trigger a production docs build.

Release publication retains the release ID returned by draft creation and refreshes
that resource directly while uploading and verifying assets. It does not depend
on a newly created draft immediately appearing in the release collection. Remote
read failures still fail closed; no unverified asset or release is advertised.
### Worker journal and local device reservation

The Linux print worker holds a nonblocking OS lock on the actual resolved USB
character-device inode before claiming jobs. It retains that device connection
for its active worker lifetime, so separate users or journal directories cannot
bypass physical exclusion. Device identity is revalidated through trusted local
discovery before opening; unsupported locking fails closed. A separate journal
file lock serializes recovery state keyed by the discovered physical identity. Its
private state directory and files are owned by the current user and inaccessible
to other users. Lock release follows process exit; a second worker fails clearly
instead of sharing an active printer connection. Unsupported platforms reject USB
worker startup without affecting ordinary CLI commands.

Each device journal is a versioned, checksummed envelope bound to that device.
Missing files and invalid/corrupt state are distinct from an initialized idle
journal. A write atomically replaces a same-directory temporary file, fsyncs the
file before replacement and the directory afterwards, and acknowledges persistence
only when both succeed. Completed-state clearing writes a durable idle envelope;
it does not erase evidence by deleting the journal. Claim tokens remain owner-only
local secrets and are never included in diagnostics.

The worker records each copy's submission intent durably before sending printer
bytes, then records positively observed completion before attempting another
copy. An API start acknowledgement and current lease are additional gates before
submission. Recovery inspects API state before new claims; an earlier process's
submission intent can produce uncertainty/reconciliation, never an automatic
physical replay. Journal or lock failures are fail-closed printer conditions.

### Local Connector Registration State

- `connectors print register --name NAME` discovers candidate devices without
  opening them, starts key-bound browser pairing, and displays the short code
  and approval URL. The approving user selects the inventory, printer, and
  explicit media profile in the browser. The CLI never borrows a human token.
- Save the exchanged connector credential, canonical API server, tenant,
  inventory, connector ID, and expiry before sending the activation heartbeat.
  A failed save must not activate the credential. A lost activation response can
  be retried with the saved credential. Pairing private keys and poll secrets
  remain temporary process state and never appear in output.
- Connector credentials use a separate OS-keyring namespace. Headless operators
  may explicitly set `STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE` to an owner-only
  file. This file stores one connector; replacing a different registration needs
  an explicit separate path. It must never overwrite human login credentials. Cross-process locking
  serializes identity checks and atomic file replacement so competing
  registrations cannot both activate after overwriting each other.
- `--connector ID` or `STUFF_STASH_CLI_CONNECTOR_ID` selects the saved connector
  for foreground operation. Live authorized device bindings and media come from
  the API. `STUFF_STASH_CLI_PRINT_STATE_DIRECTORY` or `--journal-dir` selects
  persistent recovery state separately from credential storage.

The foreground worker consumes all assigned printers by default. One connector
heartbeat loop refreshes assignments; each distinct physical device has its own
serial worker holding its device connection and journal reservation. An offline
printer does not block another printer. Removed or changed physical/media assignments cancel and join their old worker
before a replacement can open that device. Retirement stops new claims while
allowing an already-started attempt to finish and durable evidence to reconcile;
recovery for a retired printer does not require the hardware to be online.
Duplicate physical bindings fail closed instead of racing two queues. Terminal
credential failures cancel all workers and return re-pair guidance; transient
failures use bounded jittered backoff. Runtime timing and backoff are injected
through ports, with environment-backed defaults selected in bootstrap.

Foreground runtime defaults are a 5-second connector heartbeat/assignment refresh,
2-second idle job poll, 5-second readiness probe, 1-second completion observation
interval, 5-second lease safety margin, and 16 MiB artifact limit. Operators may
set `STUFF_STASH_CLI_PRINT_HEARTBEAT_INTERVAL`, `STUFF_STASH_CLI_PRINT_POLL_INTERVAL`,
`STUFF_STASH_CLI_PRINT_READINESS_TIMEOUT`, `STUFF_STASH_CLI_PRINT_OBSERVE_INTERVAL`,
`STUFF_STASH_CLI_PRINT_LEASE_SAFETY`, and `STUFF_STASH_CLI_PRINT_MAX_ARTIFACT_BYTES`.
Reconnect backoff uses equal jitter from 1 to 30 seconds, configurable through
`STUFF_STASH_CLI_PRINT_BACKOFF_MIN` and `STUFF_STASH_CLI_PRINT_BACKOFF_MAX`.
Durations and limits are validated before device access. The default persistent
journal directory is `stuffstash/print-state` under the OS user configuration
directory, with owner-only permissions; `--journal-dir` overrides it. Recovery
runs before hardware access, including when an active printer is powered off.

### Human queue-command delivery

Human printing commands resolve omitted template/options and destination from
inventory print settings, then fetch that destination's current registered media
fingerprint. Explicit flags override the corresponding setting. An offline
registered destination remains selectable; a missing destination is a usage
error, never an implicit first-printer choice. Scripts only print during asset
creation when `--print-label` is present, regardless of the inventory UI default.
Every enqueue/create-with-print command reports its logical request key to stderr
before sending; an ambiguous failure can be retried with that exact key and
selection. The CLI never automatically retries with a new key. Cancellation reads
the current revision and relies on the API's compare-and-swap fence. Human output
identifies queued jobs as queued and includes the creation's print-job ID.

### Existing connector credential rotation

`stuffstash connectors print rotate --connector ID` loads the existing connector's
local identity, creates a fresh key-bound pairing, and directs the user to browser
approval for that exact tenant, inventory, and connector. The verification URL
carries those public identifiers as `tenantId`, `inventoryId`, and `connectorId`
query parameters; these express intent and never authorize rotation. The browser
requires authenticated configuration permission and explicit replacement approval
through the existing credential-rotation command, rather than new registration.

The CLI rejects an exchanged credential for any different server, tenant,
inventory, or connector before saving or activating it. Expired local credentials
may identify the target: only the human approval grants a replacement. A successful
exchange persists the replacement before its activation heartbeat; failed storage
leaves the old credential untouched, and failed activation retains the replacement
for the worker's existing activation recovery. No credential or polling secret is
included in the URL or output. Run the normal worker command after activation.

Rotation starts an explicitly marked `rotation: true` pairing with no discovered
printer candidates; it works while the printer is disconnected or powered off.
Ordinary registration still requires candidates. Rotation-only pairings cannot
use the new-registration approval command, and review exposes their purpose.
### Standalone Label Commands

`labels templates` lists the authorized versioned catalog. `labels resolve URL`
needs a configured authenticated server but no selected inventory: parse locally,
check `/instance`, then use only the generated configured-server resolver endpoint.
HTTPS links may retain obsolete hosts/path prefixes; no request reaches that host.
Apply the shared label protocol restrictions, including canonical opaque IDs,
version, path, whitespace/encoding, and rejection of credentials/query/fragment.

`labels render ASSET --format png|pdf --output PATH` uses explicit scoped context
and inventory template defaults, with the existing template/version/reference
flags. Media comes from `--printer` (or the configured default printer), an
explicit `--media-preset` from the authenticated registry, or paired positive
`--width-mm` and `--height-mm` values selecting an exact supported catalog size.
Dimensions never synthesize unreviewed printer geometry; initial standalone stock
remains 29 × 90 mm. Conflicting media selectors fail before rendering. Rendering
without a registered printer does not create one, enqueue work, or access USB.

Provision label identity, request the immutable artifact, then download only its
scoped generated content endpoint. Ignore returned content URLs. Verify the expected
PNG/PDF content type, a bounded 16 MiB download, and SHA-256 before publishing.
Write a private file atomically without replacing any existing path or symlink;
failed or canceled downloads leave no output file. Output reports the path, format
and digest through the presentation port, never binary bytes mixed with JSON.

Private label-file publication initially supports Linux and macOS. Windows must
fail closed without creating a file until a Windows adapter establishes and verifies
an owner-only DACL; Unix mode 0600 alone is not evidence of Windows privacy.
Other Windows label and queue commands remain supported.
### Release capability metadata

The Linux Brother adapter is now integrated. Linux amd64 and arm64 binaries
advertise `usbPrinting: true`; macOS and Windows API binaries advertise false.
This states compiled adapter availability, not device readiness or physical
verification. The `version` command derives support from registered printer
adapter descriptors for its target OS without touching USB devices. Release
manifest capabilities derive from the same exported registry, and the release
build compares native binary output with that registry. Release builds keep
`CGO_ENABLED=0`: the Linux usblp transport is pure Go and must remain included.
