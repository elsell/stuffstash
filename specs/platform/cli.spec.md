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
- Creation supports explicit `--print-label` and destination/profile overrides.
  Omission does not apply the UI's inventory default invisibly to scripts.
- Asset list/search and mutation flags must map to the generated contract. Only
  the commands above are the first supported inventory surface; further REST
  parity must be specified intentionally rather than guessed at runtime.
- `printers list` queries registered inventory destinations, including offline
  ones. `printers discover` examines the local machine without registration or
  printing. Make this distinction clear in help and structured output.
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
  fail with setup guidance if unsupported. Device-code login is deferred rather
  than assumed to exist on every OIDC provider.
- Public CLI auth discovery is `GET /auth/cli/config`, containing issuer, public
  client ID, scopes, and allowed loopback redirect policy only. It is additive to
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

- Ship the Brother QL-800 adapter and 29 x 90 mm profile with the CLI distribution.
  No separate user-installed Stuff Stash plugin or Python/uv setup is part of
  the intended experience. OS USB permissions or a packaged native USB runtime
  may require documented installation setup.
- Keep command parsing, consumer orchestration, rendering, credential storage,
  device discovery, local locking/journaling, and printer I/O separate. Thin Go
  entrypoints assemble injected ports in focused bootstrap packages.
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
  installed media detection, USB claim behavior, and physical rendering. If a
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

- Registration discovers/selects device and profile, initiates pairing, and
  persists the approved connector/printer mapping and restricted credential.
  Approval is required for new bindings; a restart reuses existing registration.
- `connectors print run` stays in the foreground and uses outbound HTTPS to
  heartbeat, report printer state, claim work, renew leases, fetch authorized
  artifacts, and report outcomes. It opens no network listener.
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
- API clients are generated from the pinned Huma OpenAPI contract with a reviewed
  pinned Go generator recorded before use. Generated transport stays behind a
  CLI API port; do not import the API module's `internal` packages.
- Unit/application tests use behavioral in-memory fakes for API, clock, printer,
  credential storage, journal, and output. Adapter integration tests verify actual
  parsing/transport and atomic persistent recovery, not mock call sequences.
- End-to-end auth tests cover callback forgery/replay, wrong issuer/audience,
  expired tokens, cross-tenant scope, wrong role, pairing/revocation, and allowed
  human versus connector operations. CLI subprocess tests cover stable JSON,
  exit statuses, secrets absent from logs, cancellation, and restart recovery.
- Package pinned dependencies and required notices; verify checksums/provenance
  and a clean Linux installation that does not require the old Python project.
- Physical printing and scan checks remain explicitly unverified until performed.
  Collect user-device checks in `docs/reports/user-testing-checklist.md` during
  implementation; do not repeatedly request them or block unrelated release work.
- Current evidence is read-only inspection of the old script on Paul, not a new
  print, current-device discovery, or completed connector implementation.
