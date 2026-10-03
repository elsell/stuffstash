# Registered Printers, Connectors, And Print Jobs

## Status And Ownership

Accepted design direction, October 3, 2026. Not implemented. Printing is a new
bounded context: it owns labels, printer registrations, connector credentials,
capability reports, jobs, and attempts. Assets retain asset lifecycle ownership;
inventories retain inventory configuration; identity/access supplies human
permissions and a restricted connector authentication boundary.

The API is the durable source of truth. Web, mobile, and CLI consume its contract;
connectors never publish status directly to other clients. See
[asset labels](asset-labels.spec.md) and [CLI](../platform/cli.spec.md).

## Resource Model

- A registered printer is an inventory-owned logical destination with immutable
  ID, display name, enabled/retired state, one user-configured label size with
  adapter-supplied technical media settings, and queue. Registration requires the
  size; editing it later keeps the same printer identity. Templates are independent
  catalog entries; see [label rendering](asset-labels.spec.md#templates-registered-label-size-and-rendering-ownership).
  Its identity survives connector restart/replacement and must not be a USB path.
- First release registers a physical printer in one inventory. Cross-inventory
  physical-printer sharing and scheduling are deferred. Reject known duplicate
  device bindings during registration; a connector must not bind a device twice.
- A connector is an explicitly registered CLI installation within that inventory,
  with ID, name, enabled/revoked state, service-principal ID, approved binding
  records with authorization-sync state, credential version,
  version/capability report, and server-observed heartbeat timestamp.
- One connector may serve multiple explicitly approved printers. Several approved
  connectors may be eligible for a logical printer, but one local process per
  physical USB device is the initial recommended setup.
- Discovery is local and read-only. Device discovery does not publish every USB
  device or grant access automatically. Registration requires an administrator
  to select and approve printer bindings; reconnecting an existing device must
  not create duplicate registrations.
- A printer report contains typed readiness (`ready`, `unavailable`, `error`,
  `unknown`), safe reason code, media information when observable, capabilities,
  and server receipt time. Device identifiers stay in protected connector
  configuration; client views expose only useful model/profile information.
- Heartbeat freshness determines connector availability independently of printer
  readiness. Once stale, readiness becomes unknown. Multiple reports retain their
  source connector; a fresh eligible ready connector can make a destination
  available, but stale reports and conflicts must not masquerade as certainty.
- A print job has kind `asset_label` or `printer_test`. Asset jobs require an
  asset/label reference; test jobs use fixed diagnostic content with no asset QR
  and are explicitly requested by an editor. Tests consume media and are never
  emitted by discovery, registration, or heartbeat.
- A print job captures tenant/inventory/printer, asset/label reference when
  applicable, immutable API-rendered artifact and content snapshot,
  template ID/version/options,
  effective registered-media snapshot/fingerprint and preset version, copies,
  requesting user, timestamps, idempotency key, status, and attempts. Never silently rerender an
  existing job with later asset edits or changed printer defaults.
- Represent identities, profiles, statuses, outcomes, and reasons as typed domain
  concepts. Repository reads carry explicit tenant/inventory scope.

## Print Job And Device Interface

The consumer receives an immutable rendering, not a template to execute. The
API creates a raster suitable for the selected media/device profile; the CLI
converts it to the device protocol without changing composition. PDF download
is a client/system-print output; the initial QL-800 consumer accepts PNG only.

Human job requests contain `printerId`, `expectedMediaFingerprint`, `templateId`,
`templateVersion`, validated `templateOptions`, `copies`, and an idempotency key.
The API resolves media settings from that printer's registration, rejecting a
stale expected fingerprint; requests cannot override a registered printer's size.
The asset ID is the scoped route target. An optional preview
fingerprint must match the current render inputs; a mismatch returns a conflict
and requires a refreshed preview. Create-with-print carries the same selection.

A successful claim returns this versioned contract through the generated SDK:

| Field | Meaning |
| --- | --- |
| contractVersion | Consumer protocol version; initially 1 |
| jobId, attemptId, printerId | Scoped identities; no local USB path |
| claimToken, revision, leaseExpiresAt | Attempt control, separate from identity and SpiceDB authorization |
| template | Immutable ID/version/options, informational for the consumer |
| media | Immutable configured-media snapshot/fingerprint and preset ID/version, with physical micrometers, printable rectangle, raster dimensions, DPI, orientation, and color mode |
| artifact | API-relative authenticated content path, MIME type, byte length, SHA-256 digest |
| output | Positive bounded copies and typed cut policy compatible with the profile |

- Profile integer physical dimensions and printable bounds are distinct from
  raster dimensions. The payload specifies whether the raster is already in
  device-feed orientation; initial QL-800 output is. The adapter must not rotate
  twice or infer scale from PNG display metadata. Capabilities explicitly name
  supported contract versions, profile versions, format/color mode, and cut
  policies; unknown required values are rejected before any device output.
- Claim eligibility requires the job media fingerprint to match the printer's
  effective configured settings. Recheck atomically at start, and require the
  worker to apply that configured snapshot locally before output. Configuration
  changes before start release a mismatching claim to waiting without output.
  Reject media edits while printing or uncertain; physical roll changes cannot
  be locked by software, and remain the operator's responsibility.
- The server validates and renders before a job becomes claimable. Persist an
  immutable artifact reference/digest with the job; a staged blob is not exposed
  until the asset/label/job transaction commits. Clean up abandoned staging
  through existing blob lifecycle ports. Render failure before creation commit
  preserves the draft and returns a recoverable error; it is not a printer error.
- Expired/missing/corrupt artifacts detected before any possible device
  submission fail with definite no-output status. Artifact loss during recovery
  cannot prove nothing printed: preserve confirmed completion evidence or mark
  uncertainty, and never replay. Do not rerender from live asset/template data
  or keep reclaiming a broken job.
  Retention must not evict queued, active, or unresolved-attempt content; explicit
  cancellation/resolution makes it eligible for terminal retention policy.
- Template metadata is for diagnostics. The consumer does not need the template
  implementation or fonts. It verifies artifact digest, type, dimensions,
  decoded-pixel limits, and device compatibility before requesting start.
- API start acknowledgment precedes device submission. Pass only validated local
  content and typed device instructions to the printer port, never claim tokens,
  API credentials, template code, or arbitrary command strings.

The project-owned printer port has these conceptual operations (exact Go types
are finalized in implementation without changing the behavior):

| Operation | Input | Result |
| --- | --- | --- |
| Discover | Local discovery context | Device references and model information |
| Capabilities | Selected local device | Supported media/profile/format/protocol versions and status/wake capabilities |
| Readiness | Selected local device | Typed readiness, observed media if known, and safe reason |
| Submit | Locked device, attempt ID, validated artifact, media profile, copies/cut policy | Submission reference or definite no-output/uncertain error |
| Observe | Device and submission reference | Pending, confirmed completion, definite no-output failure, or uncertain/partial result |

The worker owns API claims, local journaling, locks, retries, and reporting. The
adapter owns protocol conversion, device communication, and hardware evidence.
Submission success is not completion. An adapter incapable of observing completion
must report that capability honestly. Optional wake remains outside the initial
QL-800 implementation. Third-party connectors can consume the same versioned
contract; none may substitute their own layout for a job's immutable artifact.

## Registration And Credentials

- Inventory viewers can list safe printer/connector/job information. Configuring
  printers, approving/revoking connectors, binding devices, and changing defaults
  require current `inventory.configure`. Queueing, canceling, and reprinting jobs
  require `inventory.edit_asset`; viewers can render/download labels separately.
- Bootstrap connector registration through a browser-approved pairing request.
  CLI creates an ephemeral pairing key and receives a secret polling token plus
  short user code/verification URL. The unauthenticated request grants no access
  to inventory names, printers, or jobs. Rate-limit creation and guessing.
- In an authenticated web/mobile approval surface, the user enters the code,
  reviews connector name and public-key fingerprint, selects inventory/printers,
  and explicitly approves with `inventory.configure`. Pending discovery data is
  untrusted descriptive data. A code alone cannot retrieve a credential.
- CLI exchanges the secret polling token with proof of possession of the bound
  key for an opaque connector credential. Approval and consumption are atomic,
  expire after a bounded configured interval, and cannot be replayed. Lost
  exchanges restart pairing; issued-but-unconfirmed credentials expire unless
  activated by a first authenticated heartbeat within the pairing activation
  window. No permanent unowned credential may remain active.
- Store connector credentials hashed server-side and securely locally. Bind them
  to the connector service principal, registration, expiry, and version.
  Credentials establish identity, not a self-contained printer permission list.
  Rotation invalidates the old version; revocation applies on the next request.
  Never put secrets in URLs, command arguments, logs, job artifacts, or lists.
- Authenticated connectors authorized by SpiceDB may heartbeat, report assigned printer status, claim
  eligible jobs, retrieve only their claimed artifacts, and report attempts.
  They cannot browse assets, create inventory content, register arbitrary
  printers, alter permissions, or reuse human endpoints as a human principal.
- Implement this machine-credential authentication behind dedicated ports and
  route guards; it is a deliberate extension to the human OIDC-only model.
  Do not weaken ordinary OIDC verification or authorize by possession of a
  connector ID. Reject stale/revoked credentials even with a valid claim token.
- Pairing/lease/freshness/credential lifetimes and rate limits are validated
  environment-backed policy supplied through injected configuration and clocks.

## Connector Authorization Through SpiceDB

- Authenticate the credential as a distinct `service_account` principal. It is
  neither the approving human nor a `user` principal. Approval does not copy the
  human's inventory role to the connector.
- Use the authorization port and SpiceDB for connector/printer access, as defined
  in [SpiceDB schema](../identity-access/spicedb-schema.spec.md#planned-print-connector-authorization).
  The API, not the CLI, calls SpiceDB. Database binding records capture approved
  relationship intent and lifecycle; they are not an independent allowlist that
  can grant access when SpiceDB denies or is unavailable.
- A `print_connector#agent` relationship grants the service principal permission
  to report its own connector heartbeat. A `printer#consumer` relationship grants
  access to that printer's safe configuration/status, reporting, and work queue.
  Both resources link explicitly to their inventory. No human inventory `view`,
  `edit_asset`, or `configure` permission is granted to service accounts.
- Every consumer request checks `print_connector.report` on its authenticated
  registration. Each printer operation additionally checks the corresponding
  `printer.view_consumer`, `printer.report`, or `printer.consume` permission.
  For artifact reads, attempt recovery, renew/start/outcome/reconciliation, derive
  the printer from the scoped persisted job and check `printer.consume`. Claim
  ownership/session/token rules remain additional requirements, never substitutes
  for authorization. Batch reports and discovery responses authorize each printer.
- Resource tenancy, inventory equality, active registration, credential validity,
  binding revocation/sync state, and operation-specific printer lifecycle rules
  are independent deny-only checks. A mistaken cross-inventory relationship
  cannot bypass these boundaries. Retirement blocks new claim/start/artifact
  delivery, but an otherwise authorized owner may read, report the outcome of,
  or reconcile an existing attempt. Explicit connector/binding revocation still
  denies access; retirement alone must not strand completion evidence.
- Approval/binding changes persist relationship intent, audit, and authorization
  outbox events atomically. New grants remain pending until relationship delivery
  succeeds; API responses expose pending state and CLI retry guidance. Do not
  activate the pairing heartbeat deadline until its required grants are ready.
- Binding removal immediately denies that binding while its relationship removal
  is pending; full connector revocation also invalidates its credential. These
  guards prevent stale SpiceDB relationships from granting revoked access during
  outbox failure. Regrant requires a new binding generation and completed sync.
- Outbox reconciliation must serialize by resource/binding generation and apply
  the latest desired relationship state; delayed grants must not resurrect a
  revoked relationship. Retries and repeated removals are idempotent.
- Use fully consistent SpiceDB checks for consumer operations initially; do not
  cache allows across requests. Fail closed on timeout/error/indeterminate result.
  An operation already authorized/in flight may finish; already submitted device
  output cannot be revoked. Loss of authorization before the next API operation
  stops further output and uses the existing uncertain-outcome recovery policy.
- The credential answers who, SpiceDB answers which printer/operation, and the
  claim token answers which current attempt. Rotating credentials preserves the
  service-principal identity and relationships; changing permissions requires
  authorized relationship changes, not minting a more privileged credential.

## Atomic Claims And Attempts

Job states: `queued`, `claimed`, `printing`, `completed`, `failed`, `uncertain`,
`canceled`. Waiting for printer is a queued job's derived presentation/reason,
not failure and not proof a claim exists.

| From | Trigger | To / rule |
| --- | --- | --- |
| queued | Eligible ready connector atomically claims | claimed |
| claimed | API accepts start with current lease/token | printing |
| claimed | Explicit no-output release or lease expires | queued |
| claimed | Validation/permanent no-output error | failed |
| printing | Confirmed hardware completion | completed |
| printing | Authoritative evidence of zero output | queued for transient error, failed for permanent error |
| printing | Partial output, ambiguous disconnect, or lease expires | uncertain |
| queued/claimed | Authorized cancellation before start | canceled |
| printing | Adapter accepted submission without completion evidence | Remain printing while observed; eventually uncertain if unconfirmable |

- Claims use an atomic repository operation and a unique attempt/fencing token,
  connector session ID, expiry, and monotonic job revision. First eligible claim
  wins. Claims are scoped to the requested registered printer and credentials.
- Claim at most one job per printer; enforce printer-level reservation as well
  as job-level exclusivity. Two consumers must not claim two different jobs for
  the same printer concurrently. Prefer oldest eligible job with ID tie-breaker.
- A connector acquires an OS-level exclusive lock for the physical device before
  claiming work. Persist stable local device identification; VID/PID alone cannot
  distinguish two QL-800s. Ambiguity requires explicit selection, never guessing.
- Renew the lease while preparing/printing. The API must acknowledge the transition
  to printing before the connector writes any bytes that can cause output.
  If acknowledgment is lost, read current state; do not print speculatively.
- Cancellation and start compete atomically. Cancellation that wins prevents
  start; after printing starts return a conflict rather than promising cancellation.
- Lease expiry before start permits retry. Lease expiry after start creates
  uncertainty, not automatic replay. Never assume physical output can be undone.
- Persist local attempt phases and any spool identifier durably before device
  submission; report status idempotently. Local history supports recovery but
  cannot eliminate the crash gap between physical output and durable recording.
- Expired tokens cannot make normal status transitions. A reconnecting connector
  may submit evidence for its original attempt through a reconciliation command.
  Verified completion of that attempt may resolve uncertainty; this must be
  audited and must not alter a later reprint job or resurrect canceled work.
- While an attempt is uncertain, pause that printer's queue until reconciliation
  or an authorized user acknowledges uncertainty and the serving connector has
  confirmed no active device submission. No silent concurrent writes during
  takeover. A stopped/stale connector must not initiate additional USB writes.
- For unavoidable hardware activity already submitted, fencing only protects
  server state: document this limit; do not claim exactly-once physical printing.
- Reprint creates a new job linked to its predecessor with explicit user intent
  and a new idempotency key. Completed/failed/canceled/uncertain records remain
  available; do not erase attempts by resetting a terminal job to queued.
- Success removes a job from eligible work permanently. Failure does not blindly
  release it for replay; retry depends on evidence of zero physical output.
- UI shows Sent to printer as a phase/detail when submission alone is known;
  Completed/Printed requires hardware completion evidence. Unknown results stay
  honest even if a driver process exited successfully.

## Readiness And Recovery

- A healthy connector reports printer absence without failing the connector.
  Queued jobs wait, and all clients show the reason and last status time.
- Probe status at a bounded interval with backoff; a ready transition resumes
  eligible queued work automatically. Do not continually claim/release jobs for
  a printer known to be unavailable or consume failure retries while powered off.
- Report paper/media mismatch, cover open, cutter error, USB access failure, and
  unsupported device only when observable; otherwise report unavailable/unknown.
- Printer adapters may expose wake capability. No USB wake is promised for the
  QL-800. Smart outlets, power cycling, and wake integrations are outside the
  first release; no arbitrary shell hook or network webhook is required.
- Retiring a printer blocks new jobs and safely cancels unstarted work; starting
  versus retirement is atomic. In-flight work follows normal reconciliation.
  Clear affected defaults and disable print-on-create with visible feedback.
- Connector revocation blocks claims/status access immediately. Unstarted claims
  recover after cancellation/expiry; started work becomes uncertain if not
  already settled. Inventory/tenant archive blocks dispatch; deletion cancels
  pending work and revokes bindings without pretending to retract printed output.
- Before start, recheck the initiating user's current print permission and asset
  existence/lifecycle for asset jobs. Test jobs require current requester print
  permission and an active inventory. Permission loss, archive, or deletion cancels unstarted
  jobs. Already submitted output cannot be recalled. Archived assets remain
  scannable but cannot start new physical jobs in this first release.

## API Contract

All inventory routes are under `/tenants/{tenantId}/inventories/{inventoryId}`
and use standard envelopes/errors, scoped cursor pagination, and generated
OpenAPI contracts. Mutations require idempotency or revision preconditions as
appropriate; conflicts distinguish stale state without leaking other scopes.

- `GET/POST /printers`, `GET/PATCH /printers/{printerId}`: list/register/read,
  rename, configured label size/media settings, or retirement with a revision
  precondition. POST requires a user-selected size or supported preset; the server
  validates adapter-supplied technical values. PATCH size changes preserve printer
  identity and follow queued-job/start race rules. No separate media CRUD exists.
- `POST /printers/{printerId}/test-jobs`: explicit one-copy diagnostic job.
- `GET /print-connectors`, `GET /print-connectors/{connectorId}` and
  `PATCH /print-connectors/{connectorId}`: safe lists/detail, approved binding
  changes and revocation. Credentials are never included.
- `POST /print-connectors/{connectorId}/credential-rotation`: an administrator
  initiates fresh key-bound pairing for the same connector; activating the new
  credential atomically invalidates the old version. Return a pairing reference,
  never expose the credential through the human endpoint.
- `GET /label-templates`: inventory-view-authorized versioned built-in templates
  with supported options and compatibility constraints. Printer setup displays
  supported adapter presets for selection, without a separate media management
  destination. Consumers obtain current configured media and its fingerprint for
  SpiceDB-authorized printers through `GET /print-consumer/printers`; this grants
  no access to human inventory/catalog endpoints. Observed roll metadata is
  optional telemetry, not the source of configuration or a requirement to claim.
- `GET/PATCH /label-settings`: inventory defaults and revision-based updates.
- `POST /assets/{assetId}/label`: idempotently provision the canonical label.
- `GET /assets/{assetId}/label`: existing label identity and link metadata.
- `POST /assets/{assetId}/label-renders`: validate template/media/content selection and return an
  authenticated artifact reference plus the selection/content fingerprint; `GET /label-renders/{renderId}/content`
  streams the PNG/PDF with private/no-store caching and scoped authorization.
  Artifact lifetime is bounded and cannot become a public bearer link.
- `POST /assets/{assetId}/print-jobs`, `GET /print-jobs`,
  `GET /print-jobs/{jobId}`: create/list/detail including attempts and revision.
- `POST /print-jobs/{jobId}/cancellation`, `POST /print-jobs/{jobId}/reprints`,
  `POST /print-jobs/{jobId}/resolution`: explicit cancellation, reprint, or human
  acknowledgement of uncertainty. Acknowledgement records reported outcome and
  queue-unblocking intent; it must not fabricate hardware-confirmed completion.
- Asset create accepts the explicit label-print request described in the labels
  spec and returns the resulting job reference in its usual response envelope.

Pairing uses `/print-connector-pairings` (create),
`/{pairingId}/approval` (authenticated human approve), and
`/{pairingId}/credential` (secret/key-bound exchange); all are POST. Pairing `GET /print-connector-pairings/{pairingId}` status polling
uses the secret polling token in a request header rather than in the URL.

Connector operations use `/print-consumer`; credentials authenticate the service
principal, while scoped persistence and SpiceDB checks determine permitted access:
`POST /heartbeat`, `POST /printer-reports`, `POST /claims`,
`POST /claims/{attemptId}/renewal`, `POST /claims/{attemptId}/start`,
`POST /claims/{attemptId}/outcome`, `POST /attempts/{attemptId}/reconciliation`,
and `GET /claims/{attemptId}/content`. Claim token and revision are required on
claim mutations/artifact access. `GET /attempts?status=unsettled` and
`GET /attempts/{attemptId}` provide recovery reads restricted to attempts owned
by the authenticated connector and authorized through current SpiceDB
`printer.consume` checks plus active scoped bindings and the operation-specific
lifecycle rules above (including recovery reads for retired printers).
Return job/attempt phase, lease validity, session owner, current revision, and safe
outcome evidence; never unrelated jobs or asset content. A restarted session can
inspect an earlier session's attempt and reconcile its evidence, but cannot use
that read to take over its lease or submit more device output. Test missing start
responses and restart discovery through these restricted endpoints.
Heartbeats carry a process session identity;
restarting or duplicating a process cannot silently take over another attempt.

Consumer polling/long-polling over outbound HTTPS is sufficient. No inbound port,
message broker, public webhook receiver, or persistent event stream is required.

## Ports, Persistence, Audit

- API printing application services live under `internal/app/printing`, with
  domain types and focused repository, rendering, machine-authentication,
  authorization, clock, artifact, and observer ports. Cross-context creation uses
  application orchestration, never domain-to-domain imports.
- GORM persistence must enforce unique canonical label mapping, scoped
  idempotency keys, current printer reservation, attempt identity, and revision
  transitions transactionally. Use a shared unit-of-work port for asset/create
  plus job/audit persistence. Do not run printer I/O inside database transactions.
- No broker dependency is needed; durable database jobs are the queue. Worker
  lease recovery uses an injected clock and cannot duplicate active claims.
- Recheck authorization before artifact delivery/start; minimize snapshot content
  to label needs. Artifact retention/size and terminal job retention have bounded
  environment-backed policies; retain safe audit history after content expiry.
- State changes produce domain audit events: label provisioned; printer
  registered/configured/retired; connector approved/rotated/revoked; print job
  queued/started/completed/failed/uncertain/canceled/reprinted/resolved. Record
  human initiator and connector actor separately. Never record credentials.
- Reads follow the REST safe-read audit contract. Heartbeats/lease renewals are
  operational telemetry, not individual domain-history entries; connectivity
  transitions and consequential job changes are observable through injected
  ports. This is an explicit exception to auditing every mechanical state write.

## Required Tests And Acceptance Evidence

- Use faithful, stateful printer fakes behind the production ports, never mocks
  or scripted method-call expectations. Model readiness, accepted output,
  completion evidence, partial output, disconnects, and reconnect/restart state.
  Drive real application and worker transitions with those fakes. Hardware and
  production-database acceptance evidence remain separate requirements.

- Real-SpiceDB tests must prove grant/revoke, wrong-service-principal denial,
  no human-role inheritance, cross-inventory edge rejection, permission loss
  between claim and start, credential rotation without privilege change,
  pending outbox grants/removals, reordered grant/revoke events, and fail-closed
  SpiceDB outages, and retirement between start and outcome/reconciliation.
  No database-only authorization fallback is permitted.
- Real HTTP adversarial tests before endpoints: anonymous, wrong-role,
  cross-tenant/inventory/printer, forged IDs, expired/revoked credentials, pairing
  guessing/replay/approval races, wrong key, stolen code without polling secret,
  stale claim/session/version, privilege escalation, and artifact leakage.
- Authorized viewer discovery, editor submit/cancel/reprint, administrator
  registration/defaults/revocation, and restricted connector work must succeed.
- Verify immutable API rendering and template/media version selection, stale
  preview rejection, required manual size registration/editing, waiting jobs after
  size changes and recovery when switching back, no mandatory roll detection,
  artifact loss both before submission and during recovery, content-size/digest/decoded-pixel
  bounds, unsupported protocol/profile rejection, and no adapter layout changes.
- Concurrent consumers against production PostgreSQL prove one claim per job
  and one active reservation per printer. Fakes must model these guarantees;
  SQLite-only/unit tests do not establish production concurrency correctness.
- Test lost create/start/outcome responses, duplicate status reports, connector
  restarts, stale leases, network partitions, local journal loss, cancellation
  races, revocation, archive/delete, printer retirement, and permission changes.
- Test failures before output versus uncertain/partial output, multiple copies,
  dead connector versus disconnected printer, stale capability reports, queue
  pause/recovery, and explicit reprint after uncertainty.
- Physical QL-800 evidence must establish media handling, status observability,
  USB disconnect/power-off recovery, no duplicate print on acknowledgment retry,
  and whether completion evidence is actually available. Until established,
  report the hardware limitation and do not certify successful printing.

User-device checks belong in `docs/reports/user-testing-checklist.md` during
implementation. Pending checks do not block independent implementation or release;
keep hardware acceptance explicitly unverified and do not represent it as passing.

## Registry Implementation Contracts

- A shared plain-data Go package under `packages/printingprofiles` owns built-in
  printer/media descriptors. API and CLI adapters map those descriptors to their
  local models; neither maintains another independently editable QL profile.
- Registry persistence uses `printers` and the connector/binding tables. Queue
  transitions lock connector, binding, printer, then job, in that order. Media
  changes lock the printer and reject printing/uncertain reservations. Retirement
  preserves the active job reference and completion/recovery evidence.
- Authorization checks return an internal consumer authority carrying the
  credential version and binding generation. Stateful command transactions
  recheck those deny fences under row locks after the SpiceDB check. Session
  ownership remains a queue concern and is not implied by connector identity.
- Readiness is per connector/printer binding with a freshness deadline, never a
  last-writer-wins global report. A claim uses the requesting connector's report.
- Pairing uses Ed25519 public keys. The credential exchange signs the UTF-8
  domain-separated message `stuffstash-print-pairing-v1\n<pairing-id>\n<poll-token>`.
  The polling token and credential are cryptographically random and stored only
  as SHA-256 digests. Polling/exchange headers and request bodies are never audit
  metadata. Exchange consumes the approved request atomically once.
- Relationship reconciliation serializes by connector generation while applying
  the latest desired bindings. Persistence locks remain held through the bounded
  authorization write, so a delayed grant cannot overtake a committed removal.
  Partial external delivery leaves the generation pending and therefore denied.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/printer-profiles` lists the
  finite built-in adapter catalog and complete versioned media snapshots for an
  active inventory with `inventory.view`. It requires no registered printer and
  permits selecting media for a download before pairing hardware. It uses the
  standard response envelope, contains no device identifiers, and does not create
  audit history because it reads static shipped capability metadata, not stored
  inventory content. Profile responses identify supported platforms and whether
  the profile has been physically verified.

## Queue Implementation Contract

The queue aggregate owns transitions and attempt history. It receives time from
its caller (the application injects its clock) and never reads wall time itself.
An attempt records a connector, process session, constant-time-verifiable claim
secret digest, lease expiry, start time, and outcome evidence. Mutations compare
an expected monotonic job revision as well as ownership and lease. Repeating an
already recorded identical outcome is safe without changing revision/history.
An expired pre-start attempt requeues; an expired started attempt becomes uncertain
and retains the printer reservation. Reconciliation may resolve only the owning
connector's original uncertain attempt, after current authorization, without
resuming device output. Explicit no-output evidence may requeue a transient
failure; partial output or ambiguous evidence never does. Physical evidence is monotonic:
a known completed copy or partial output cannot later become zero output.
Repeating an identical recorded uncertain report is also an idempotent readback,
including after lease expiry; it cannot resume output or change history.

A mutex-backed repository fake must reproduce atomic printer reservation, job
selection and compare-and-set transitions, not scripted responses. Production
PostgreSQL transactions lock the printer row before selecting or mutating its
active job. All repositories require tenant and inventory scope. The application
rechecks current permission, binding lifecycle, media and asset state before start.

Durable storage keeps indexed scope, printer, queue state and idempotency columns
alongside the versioned rendering snapshot and attempt evidence. A unique scoped
actor/request key prevents duplicate jobs even when concurrent requests select
different printers. An attempt index retains its connector and job association
for authenticated lost-response recovery, including after settlement. Job state,
printer reservation, attempt index and audit writes commit or roll back together.
Registration denial fences use connector → binding → printer → job lock order;
no caller treats a reservation or database registration as an authorization grant.
- Registry services live in `internal/app/printregistry`; rendering and labels
  remain in `internal/app/printing`. A human creates the logical printer with a
  name, adapter, and media preset. Pairing creation supplies local discovered
  candidates with an opaque candidate ID, display name, adapter ID, and protected
  device identity. Approval maps candidate IDs to existing printer IDs; the
  server requires matching adapters and persists the device binding. Human DTOs
  never return the protected device identity, credentials, or local paths.
- Heartbeats include a process session ID for observability but do not replace an
  existing attempt owner. Connector last-seen time advances only on a heartbeat;
  readiness reports have separate timestamps. Fresh heartbeat and readiness are
  dispatch requirements, not prerequisites for reporting an earlier outcome.
- Printer registration requires `Idempotency-Key`; its scope includes the human
  actor and inventory. An identical replay returns the existing printer; reusing
  the key with different registration content conflicts. Persistence stores only
  a hash of the scoped request key and a canonical request fingerprint.
- The initial printer CRUD slice creates logical registrations before connector
  authorization. Pairing reconciliation must sync each printer's inventory
  relationship before granting/synchronizing connector bindings; pending sync
  remains denied by the persistence fence. A human registration alone grants no
  service principal access.

Lease renewal uses a dedicated repository command, under the same connector,
binding, printer and job locks, that can change only the current owned attempt
lease and job revision/timestamp. It does not emit a history record and cannot
be used as a general mutation or audit bypass.

The initial queue stores bounded rendered PNG bytes with the job in the same
transaction, so an acknowledged queued job never refers to an uncommitted or
mutable render. This is separate from short-lived download previews. Runtime
configuration controls maximum copies (default 20), artifact bytes (default 1 MiB),
artifact lifetime (default 7 days), terminal history lifetime (default 30 days),
claim lease (default 60 seconds), and readiness freshness (default 90 seconds).
Artifact expiry makes content unavailable for new output; terminal cleanup removes
its private bytes while preserving safe audit history. An
uncertain reservation is never removed by retention cleanup.
### Durable Connector Reconciliation

The connector's desired `generation` and acknowledged `synced_generation` form
its durable authorization outbox record. Approval and binding/revocation changes
persist the new generation together with audit history in one transaction. The
worker scans mismatched generations and synchronizes the latest complete desired
state under the connector lock; it never replays a stale captured grant payload.
The acknowledged generation advances only after all required printer inventory
relationships and connector relationships have been delivered. A failed external
write or failed database commit leaves the outbox pending and consumers denied.

Consumer readiness report reasons are finite safe codes: empty for ready,
`device_unavailable`, `device_busy`, `paper_empty`, `cover_open`, `hardware_error`,
or `unknown`. Device error strings and USB paths must not become report reasons.

Connector runtime policy is configured with `STUFF_STASH_PRINT_CONNECTOR_`
variables: `PAIRING_LIFETIME` (10m), `CREDENTIAL_LIFETIME` (720h),
`ACTIVATION_LIFETIME` (5m), `AUTHORIZATION_TIMEOUT` (5s), `POLL_INTERVAL` (5s),
`REPORT_MAX_AGE` (1m), and `BATCH_SIZE` (100). Invalid bounds fail startup.
The existing environment-configured per-client-IP HTTP rate limiter covers
pairing creation, status, proof exchange and approval guessing. Request bodies
and polling-token headers are never included in observability or audit fields.

`POST /print-connector-pairings/{pairingId}/review` is the human approval preview:
it requires current `inventory.configure` for the selected tenant/inventory and
the short code in its body. It returns the connector name, public-key fingerprint
and candidate ID/name/adapter only; protected device identities remain hidden.
Approval submits the same code with the explicit candidate-to-printer selection.

Pairing creation returns `verificationUrl` at the configured HTTPS
`STUFF_STASH_PUBLIC_WEB_BASE_URL` plus `/print-connectors/pair/{pairingId}`.
The pairing ID is a public reference, not authorization; neither short code nor
polling token appears in this URL. Without a configured public web base, new
pairing creation is unavailable (503); already configured consumers remain
operational. Existing label rendering uses the same public web base setting.

### Credential Rotation Protocol

The CLI starts a fresh key-bound pairing. An inventory administrator submits its
pairing ID and short code to the existing connector's `credential-rotation`
endpoint with the connector generation precondition. This approves the new key
for the same connector and service principal without changing any printer/device
binding. The human response contains only the pairing reference and expiry.

Exchange stores a pending credential with the next credential version and a
bounded activation deadline. The old credential remains valid until the first
heartbeat authenticated by the pending credential atomically activates it. That
heartbeat clears the pending fields and invalidates the previous version. Pending
credentials may only activate through heartbeat; they cannot claim or fetch work.
Expired pending credentials never become active, and a subsequent fresh pairing
may replace them. Replayed pairing exchanges never return either credential.
Approval, issuance, and activation write atomic, safe lifecycle audit records;
ordinary heartbeat timestamps and readiness reports remain operational telemetry
and do not create one audit entry per polling request.

Rotation approval captures the current credential version. Exchange rejects a
stale approval if another rotation has already activated, and allows only one
unexpired pending credential at a time. A delayed exchange therefore cannot
replace a newer activated credential or overwrite another still-live pending key.
### Consumer Claim Transport And Recovery

The worker durably creates a random attempt ID, process session ID, and canonical
base64url 32-byte claim secret before POST `/print-consumer/claims`. Its body
contains these values and `printerId`. Only the secret's SHA-256 digest is stored.
An identical claim retry reads that same attempt; changing printer/session/secret
for an existing attempt conflicts and never creates another output attempt.
Attempt and session identifiers are bounded opaque ASCII values (16–100 characters).

Start, renewal and outcome bodies contain `sessionId`, `claimToken`, and expected
`revision`; outcome adds finite kind/reason, completed copies and retryability.
Content reads use `X-Print-Session-ID`, `X-Print-Claim-Token`, and
`X-Print-Revision` headers. No secret appears in a URL. Every operation checks the
current credential and fully consistent printer authorization. Content additionally
requires ownership of the current unexpired attempt and an unexpired artifact.
Recovery reads never return the claim secret, label title, QR URL, or PNG bytes.
Unsettled discovery is filtered to the authenticated connector, supports an optional
printer filter, and uses scoped cursor pagination. A recovered record is evidence,
never permission to resume earlier-session physical output.

### Scheduled Job Maintenance

The API runs bounded job maintenance at the configured cleanup interval using its
injected clock. Expired unstarted leases return to queued with no-output evidence;
expired started leases become uncertain and retain their printer reservation.
Artifact expiry before a job starts makes that job definitively failed; it must
not be reclaimed or rerendered. Expiry never changes a printing or uncertain job
into a safe-to-retry result.

Private artifact bytes are removed only from terminal completed, failed, or
canceled jobs after artifact expiry. Queued, claimed, printing, and uncertain
content stays protected until the job transitions to a terminal state. Terminal
job and attempt metadata may be removed after the terminal history lifetime,
measured from its last state transition; audit history remains. The request
idempotency window ends when that terminal job is removed. Cleanup uses the same
printer lock as claims and state changes and rechecks eligibility under the lock.

### Dispatch readiness fence

The start transition rechecks the serving connector heartbeat and its readiness
report under the same transaction and locks as the printer reservation. A lease
renewal does not refresh either health signal. Stale, missing, future-dated, or
unavailable health rejects a new start without consuming the claim or emitting
output. An identical already-started retry remains readable/idempotent. Outcome
reporting and reconciliation remain usable while the printer is unavailable.
