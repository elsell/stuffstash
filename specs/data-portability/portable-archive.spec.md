# Portable Inventory Archive And Restore

## Objective and delivery

Implement instance-to-instance inventory transfer as a complete user workflow.
The October 2 user request adds this to the ongoing audit goal; it is not complete
when an encoder, endpoint, or export-only client ships. Delivery requires archive
creation, download, preview, approved restore, both clients, and a cross-instance
round trip including original media bytes. Existing JSON/CSV remain supported.

Three implementation commitments form the batch: A1 durable archive export and
validated packaging; A2 preview and restore into a new inventory; A3 mobile/web
task surfaces, accurate format descriptions and connected acceptance. Foundation
commits may land separately but must not advertise unavailable restore behavior.

## User contract

- CSV is a spreadsheet export, not a backup. JSON is complete inventory metadata,
  including tag/type/field definitions, but excludes binary photos/files. A Stuff
  Stash archive is a ZIP of JSON and selected original media. Only an archive with
  all media included and verified is described as a full inventory backup.
- Inventory settings offers Export archive with Photos and Other files enabled
  by default. Metadata and every referenced definition are always included.
  Disabling a switch explicitly produces a data-only or partial-media archive.
  Archived assets, schema records, retained assignments and attachments are included.
- Export is a persisted background job with queued/running/ready/failed/cancelled/
  expired states, understandable progress, cancellation and retry. Closing the app
  does not cancel work. A restarted worker must recover expired leases safely.
  The job result can be downloaded or shared later until its disclosed expiry.
- Restore accepts a Stuff Stash archive, validates it before approval, and previews
  inventory name, counts, attachment inclusion/omissions and compatibility errors.
  The user supplies the destination inventory name within an authorized household.
  It creates a new inventory; it never merges into or overwrites an existing one.
- Native document selection and system save/share are used on mobile. Web uses
  browser file selection/download. Inclusion controls are native switches on
  mobile and labeled checkboxes on web. Reuse established job progress and task
  presentation patterns; avoid extra screens for simple inclusion choices.

## Archive contract version 1

- A standard ZIP contains exactly `manifest.json`, `inventory.json`, and zero or
  more original media entries named `media/<lowercase SHA-256>`. No directories,
  symlinks, absolute paths, traversal, duplicate names, encrypted or unknown entries.
  Deduplicate identical media content by hash without losing attachment records.
- Manifest identifies `format: stuffstash.inventory`, `archiveVersion: 1`, export
  time, Photos/Other files inclusion, the byte size and SHA-256 of inventory.json,
  and a list of included media content sizes/hashes. Unsupported versions fail
  explicitly. All required fields are validated; unknown manifest fields fail
  closed for this version. A valid ZIP alone never proves a restorable inventory.
- Inventory JSON is separately versioned. It retains all assets and parent links,
  original creation/update times, lifecycle, tags including names/colors, custom
  types/field definitions and applicability, typed field values, expiration precision,
  attachment metadata and current checkout details. Tenant-inherited schema is
  captured and recreated locally in the new inventory, not installed tenant-wide.
- User names/text are preserved verbatim. Source IDs are references within the
  package, never authority to access a destination object. Restore allocates new
  IDs for every resource, remaps all links and validates the complete reference
  graph, kinds, cycles, field/value compatibility, enums and uniqueness first.
- Current checkout state/details survive. Source principal IDs are provenance,
  never destination principals or grants. Restore records its initiating principal
  as the local actor. Source audit history is not replayed as destination history.
- Do not export authentication/provider credentials, memberships, invitations,
  access tokens, signed URLs, blob storage keys or raw audit/undo snapshots.
  This is an inventory backup, not a server or household-security backup.
- Validate declared and actual byte lengths and SHA-256 checksums for JSON and all
  media. Missing/corrupt requested media fails export; it is never silently omitted.
  On restore, omitted media is reported as an explicit source selection, and no
  attachment row pointing to a nonexistent blob is created.

## Inventory metadata version 2

Archive `inventory.json` uses schemaVersion 2, with explicit projections rather
than serialized persistence/domain objects. The document contains source tenant
and inventory IDs, inventory name, export time, assets, tags, custom asset types,
and custom field definitions. All four collections are required, including when
empty. Attachments are nested under their owning asset; current checkout includes
its original creation/update times and actor as provenance. No storage key is
accepted in the input. Unknown or duplicate JSON keys, malformed records and
unsupported versions fail before approval.

Schema keys are unique within each source definition family, including inherited
and archived definitions. Restore preserves source display names and values. If
an existing destination household definition reserves a source key, allocate a
fresh valid local key and remap every custom-field value to it; never overwrite
or reuse destination definitions merely because a name or key matches. Preview
reports these key remappings. Revalidate destination reservations atomically at
publication; a new conflict fails safely rather than changing approved meaning.

## Consistency, resources and failure

- Capture metadata in a coherent repository snapshot through a dedicated port.
  Stream immutable original blobs referenced by that snapshot; if any becomes
  unavailable, fail the job instead of publishing a supposedly complete archive.
- Stream ZIP production and validation through storage ports. Do not load all
  media or a complete archive into memory or mobile server-state caches. Configured
  positive limits bound compressed bytes, total expanded bytes, per-entry bytes,
  inventory JSON bytes, entry count, record count, concurrent jobs and retention.
  Enforce actual streamed byte limits as well as ZIP headers; reject ZIP bombs,
  invalid checksums and trailing/unlisted archive entries. No shell extraction.
- Uploads and generated archives are private leased objects. Authorize creation,
  polling, cancellation, preview, approval and every download by caller and scope.
  Export requires current inventory view; restore requires destination inventory
  creation authority. Recheck permissions before publication and final restore.
  Never fetch network URLs supplied by an archive. Cleanup handles abandonment,
  expired jobs, interrupted uploads and restart without deleting live user blobs.
- Preview is bound to immutable uploaded bytes and checksum. Approval cannot swap
  the package or destination. Retries are idempotent and cannot duplicate an
  inventory or attachments. Persist progress and bounded safe failure categories.
- Stage and validate restore bytes before creating visible inventory content.
  Final metadata publication is atomic through a restore unit-of-work, including
  required domain validation/audit and reliable authorization synchronization.
  Failed/cancelled jobs must not expose a partial inventory as a successful restore.
  Orphan staged blobs are cleaned by retryable lifecycle work.
- Jobs and repositories preserve tenant ownership; operational lease claiming is
  the only unscoped worker read. Use injected clocks, identifiers, authorization,
  storage and domain observers. ZIP/JSON implementations remain adapters.

## Durable job invariants

Export and restore use a dedicated archive-job aggregate; the Homebox import
record cannot express atomic new-inventory publication or immutable archive
approval. Each job has a tenant, requesting principal, kind, immutable source
inventory or uploaded artifact/checksum, creation/expiry times and monotonic
revision. Export options are immutable. Restore starts with validation work,
then awaits approval; approval binds the destination name and a newly allocated
destination ID before restore execution. No worker can skip that approval.

Claims carry a unique fencing token and an expiry, with repository compare-and-swap
on the revision. Only the current unexpired claim may heartbeat or complete.
Expired claims may be reclaimed with a new token; stale workers cannot publish.
Cancellation and expiry invalidate any claim immediately. Publication and its job
transition must use the same atomic unit-of-work; a cancellation/reclaim race must
fail publication. Retry from failure retains the immutable source and approved
destination identity. Validation failures return to validation; approved restore
failures retry execution without silently choosing a different inventory.
Job updates increment revision rather than relying on timestamp precision.
- Job timestamps use UTC microsecond precision so PostgreSQL indexed timestamps and the serialized aggregate agree. Repository updates must validate a complete domain transition against the persisted predecessor under revision fencing, including immutable request, creation/expiry, and approved destination fields.


## Critical acceptance and release

1. Export on instance A, restore on independently initialized instance B, then
   compare metadata and attachment hashes with new resource IDs. Cover nested
   containment, active/archived records, inherited schema, every supported field
   type, current checkout, Unicode and original photo/file bytes.
2. Verify each media inclusion combination and empty inventories. Missing files,
   bad hashes, duplicate IDs/ZIP entries, traversal, unsupported versions, dangling
   references, cycles and size/count overruns must fail before publication.
3. Adversarial real HTTP boundary tests cover unauthenticated/wrong-role/cross-tenant
   callers, job/upload guessing, changed permissions and approval replay. Authorized
   callers must complete the same flows. Test restart, cancellation, expiry and
   retry around staging/publication; no credentials or user content in telemetry.
4. Exercise real authenticated web and native export/preview/restore with reachable
   controls and persisted progress. Preserve explicit limits of native fixtures;
   physical save/pick evidence is separate. Do not silently replace missing runtime
   evidence with source assertions.
5. Ship backend, migrations and clients together. Updating production GitOps in
   `paul:~/code/local-k8s/infra` and verifying rollout is part of this delivery,
   alongside TestFlight upload and changelog. Tag publication alone is not deployment.

### Snapshot adapter boundary

Archive collection runs within a read-only repeatable-read transaction on PostgreSQL
(or the equivalent SQLite snapshot). A snapshot port supplies repositories bound
to that transaction to the existing bounded export collector. Scope IDs remain
explicit on every collection call. Authorization is checked by the job service
before collection and before publication; the snapshot adapter does not authorize
users. No transaction stays open while streaming blob data or waiting for approval.

The required PostgreSQL CI job verifies a concurrent mutation stays outside the
archive snapshot and exercises migration-backed job timestamp/revision fencing.

### Restore plan

After validating source bytes, build a destination plan with fresh identifiers for
assets, tags, custom types, fields, checkouts and included attachments. Retain all
creation/update times and expiration precision. Parent links, tag assignments,
field applicability, custom types and field-value keys are remapped together.
The plan drops intentionally excluded attachment rows and reports their count.
Checkout source actors are retained as restore provenance; the local initiating
principal owns the recreated checkout action. Persist the exact plan privately
before approval so retries reuse it; approval cannot regenerate identifiers or
silently change key remappings. Blob keys are derived from the new destination
scope and attachment identifiers, never taken from archive input.

### Streaming storage

Archive workers use a bounded private temporary file while producing or receiving
an archive, then stream the known size to object storage. Closing the workspace
removes it, including on cancellation/failure. This avoids holding the ZIP in
memory and permits bounded random-access ZIP verification. S3 archive publication
uses a single bounded PUT (maximum 5 GiB), avoiding orphan multipart uploads on
cancellation. Archive limits are separate from ordinary photo upload limits.
Original blob reads and archive reads expose streams/random access rather than
byte slices. S3 reads bind to the observed ETag so a changed object fails instead
of mixing bytes from different versions. Filesystem writes stage privately and
rename only after a complete, size-checked copy.
