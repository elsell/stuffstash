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
  ID, display name, enabled/retired state, compatible profiles, and queue.
  Its identity survives connector restart/replacement and must not be a USB path.
- First release registers a physical printer in one inventory. Cross-inventory
  physical-printer sharing and scheduling are deferred. Reject known duplicate
  device bindings during registration; a connector must not bind a device twice.
- A connector is an explicitly registered CLI installation within that inventory,
  with ID, name, enabled/revoked state, authorized printer IDs, credential version,
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
- A print job captures tenant/inventory/printer, asset/label reference when applicable, immutable
  rendered content or verified render snapshot, profile, copies, requesting user,
  timestamps, idempotency key, status, and attempts. Never silently rerender an
  existing job with later asset edits or changed printer defaults.
- Represent identities, profiles, statuses, outcomes, and reasons as typed domain
  concepts. Repository reads carry explicit tenant/inventory scope.

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
  to connector, tenant, inventory, authorized printers, expiry, and version.
  Rotation invalidates the old version; revocation applies on the next request.
  Never put secrets in URLs, command arguments, logs, job artifacts, or lists.
- Connector credentials may heartbeat, report assigned printer status, claim
  eligible jobs, retrieve only their claimed artifacts, and report attempts.
  They cannot browse assets, create inventory content, register arbitrary
  printers, alter permissions, or reuse human endpoints as a human principal.
- Implement this machine-credential authentication behind dedicated ports and
  route guards; it is a deliberate extension to the human OIDC-only model.
  Do not weaken ordinary OIDC verification or authorize by possession of a
  connector ID. Reject stale/revoked credentials even with a valid claim token.
- Pairing/lease/freshness/credential lifetimes and rate limits are validated
  environment-backed policy supplied through injected configuration and clocks.

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
  rename, profile configuration, or retirement with a revision precondition.
- `POST /printers/{printerId}/test-jobs`: explicit one-copy diagnostic job.
- `GET /print-connectors`, `GET /print-connectors/{connectorId}` and
  `PATCH /print-connectors/{connectorId}`: safe lists/detail, approved binding
  changes and revocation. Credentials are never included.
- `POST /print-connectors/{connectorId}/credential-rotation`: an administrator
  initiates fresh key-bound pairing for the same connector; activating the new
  credential atomically invalidates the old version. Return a pairing reference,
  never expose the credential through the human endpoint.
- `GET/PATCH /label-settings`: inventory defaults and revision-based updates.
- `POST /assets/{assetId}/label`: idempotently provision the canonical label.
- `GET /assets/{assetId}/label`: existing label identity and link metadata.
- `POST /assets/{assetId}/label-renders`: validate render options and return an
  authenticated artifact reference; `GET /label-renders/{renderId}/content`
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

Connector operations use `/print-consumer` and derive scope from credentials:
`POST /heartbeat`, `POST /printer-reports`, `POST /claims`,
`POST /claims/{attemptId}/renewal`, `POST /claims/{attemptId}/start`,
`POST /claims/{attemptId}/outcome`, `POST /attempts/{attemptId}/reconciliation`,
and `GET /claims/{attemptId}/content`. Claim token and revision are required on
claim mutations/artifact access. `GET /attempts?status=unsettled` and
`GET /attempts/{attemptId}` provide recovery reads restricted to attempts owned
by the authenticated connector and its currently authorized printer bindings.
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

- Real HTTP adversarial tests before endpoints: anonymous, wrong-role,
  cross-tenant/inventory/printer, forged IDs, expired/revoked credentials, pairing
  guessing/replay/approval races, wrong key, stolen code without polling secret,
  stale claim/session/version, privilege escalation, and artifact leakage.
- Authorized viewer discovery, editor submit/cancel/reprint, administrator
  registration/defaults/revocation, and restricted connector work must succeed.
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
