# Import And Export Spec

## Purpose

Stuff Stash should make user data portable.

Users should be able to export inventory data and later import data through supported formats without coupling the domain to one file format.

## Scope

This spec covers initial import and export requirements.

This spec covers the first legacy Homebox import workflow and the durable import-job direction for production-scale imports.

This spec does not define the final Stuff Stash-native CSV columns, final Stuff Stash-native JSON schema, backup packaging, media export packaging, newer Homebox entity import behavior, permanent source-link management UI, exact durable import-job endpoint shapes, or all future import conflict resolution modes.

## Requirements

- Import and export must be behind ports and adapters.
- Export must support JSON.
- Export must support CSV.
- Import should support JSON and CSV once the corresponding import workflows are specified.
- Export must preserve tenant and inventory authorization boundaries.
- Users must only export inventories they are authorized to export.
- Imports must validate tenant, inventory, asset, location, and custom field behavior before applying changes.
- Imports must produce audit records for state-changing operations.
- Import and export adapters must not leak file-format details into domain logic.
- Imports must use a two-step workflow:
  - Preview builds and validates an import plan without mutating Stuff Stash state.
  - Apply revalidates the submitted plan and then writes through application services, ports, and adapters.
- Import source adapters must convert source-specific data into a Stuff Stash import plan. Source adapters must not directly create Stuff Stash assets, custom fields, custom asset types, attachments, inventories, or audit records.
- Import plans must be source-neutral enough that future import sources can reuse the same validation and apply path.
- Import apply must run under a real authenticated principal and must enforce the same tenant, inventory, authorization, validation, audit, observability, attachment, and containment rules as normal application workflows.
- Import plans must include source references so repeated imports can detect likely duplicates before writing.
- Import preview must report warnings and blocking errors separately.
- Import warnings must be safe to show to users and must not include source passwords, bearer tokens, attachment storage paths, or provider internals.
- Import source validation and connection failures may return a safe, actionable error detail, such as missing URL, unsupported URL scheme, blocked private-network source, TLS trust failure, or Homebox HTTP status, only when the source adapter marks the detail with the import-source user-error contract. Ordinary adapter errors must remain generic at the HTTP boundary. The web UI must prefer typed safe details over a generic invalid-request message.
- Import apply may reject stale or tampered preview plans. The client must not be trusted as the source of validation truth.
- Production-scale imports must use durable import jobs rather than request-lifetime HTTP execution.
- Durable import jobs are inventory-scoped. Import jobs must not create inventories in the durable Homebox import slice.
- Durable import jobs must be source-adapter agnostic. Homebox is the first import source adapter, not a special import-job architecture.

## Import Source Ports And Adapters

The import flow uses hexagonal boundaries:

- Source adapters know how to read a foreign source and produce normalized import data.
- The import application service validates normalized import data against Stuff Stash rules.
- The import application service applies validated plans through existing application services and ports.

Initial source adapters:

- `legacy_homebox`: connects to a legacy Homebox instance with URL, username, and password.
- `legacy_homebox_csv`: parses a Homebox legacy CSV export uploaded by the user.

Future adapters may support newer Homebox entity APIs, Stuff Stash JSON, Stuff Stash CSV, folder/photo imports, or other inventory systems without changing the core import apply behavior.

## Durable Import Jobs

Durable import jobs are the production import execution model.

The synchronous Homebox preview and apply endpoints are only the first small import slice. Production-scale live imports must move to inventory-scoped durable jobs with persisted preview state, progress, cancellation, audit, observability, source-link uniqueness, encrypted temporary credentials, and worker-based execution.

Durable import jobs must support:

- In-progress job visibility.
- Past successful job visibility.
- Past failed job visibility.
- Past cancelled job visibility.
- Explicit user removal from import history.
- Inventory-scoped job listing and job detail.

Completed, failed, and cancelled jobs must remain visible until explicitly removed from import history by an authorized user.
Removing a job from import history must not remove audit history.
Removing a job from import history must not remove imported records.
Removing a job from import history must not remove source-link records needed for duplicate prevention unless a future source-link management spec explicitly defines that behavior.

Durable import jobs must have explicit permissions:

- Import job view permission.
- Import job create permission.

The first authorization mapping may derive these permissions from the existing inventory write permission or relationship, but application and adapter code must represent import-job permissions explicitly so future access policy changes do not require changing the import domain model.

Durable import-job execution must be behind a worker port.
The first runtime may provide an in-process worker adapter hosted by the API process, but route handlers must not directly execute long-running import work.
The worker port must allow a future standalone worker process without changing import application behavior.

Import jobs must persist normalized plan metadata and source references after preview.
Import jobs must not persist Homebox attachment bytes or other source attachment bytes in the preview plan.
Attachment bytes must be fetched during apply when the source adapter and user options require attachment import.

Durable import jobs must persist enough source fingerprint information to detect that a reviewed live source changed between preview and apply.
If the source fingerprint changes after preview, apply must require a new preview.
Apply must not silently continue with warning-only skips when the source fingerprint changed.

For live Homebox, the same Homebox instance may contain different records at different times.
Source fingerprints and source links therefore represent repeatable import identity, not full synchronization.
The import model must not imply that Stuff Stash and Homebox are continuously synchronized.
Homebox records added after a preview require a new preview before import.
Homebox records removed or changed after a preview require a new preview before apply when the source fingerprint detects the change.

Durable import jobs must use encrypted stored credentials when the job cannot complete within the request that supplied source credentials.
Credential storage must reuse the repository's existing encryption patterns.
Credentials must never be returned by the API.
Credentials for terminal jobs must be cleaned by a recurring credential vacuum job.
Credentials for non-terminal jobs that run longer than `STUFF_STASH_IMPORT_JOB_TIMEOUT_SECONDS` must also be cleaned by the credential vacuum job.
`STUFF_STASH_IMPORT_JOB_TIMEOUT_SECONDS` defaults to `900`.

Import-job credential cleanup must be observable and auditable without exposing credentials, bearer tokens, passwords, or provider internals.

Durable import jobs must be observable through domain-oriented probes.
Observability must include job lifecycle, worker claiming, worker retry, worker failure, credential cleanup, source-fingerprint mismatch, preview progress, apply progress, cancellation, discard cleanup, and source-link duplicate prevention.
Observability fields must not include source credentials, bearer tokens, raw provider paths, raw storage keys, or raw adapter internals.

Durable import jobs must be fully auditable.
Audit history must include job creation, preview start, preview completion, apply start, apply completion, apply failure, cancellation request, cancellation completion, discard cleanup, credential cleanup, and imported record creation or modification where state changes occur.
Existing asset, custom field, attachment, and other domain audit records must still be written by the existing application services used during import apply.
Import-specific audit metadata must include the import job ID where practical.

## Source Links And Idempotency

Durable import jobs must enforce source-link uniqueness for imported records.

Homebox custom fields such as `homebox-source-id` and `homebox-asset-id` may remain useful for user-facing display, search, and diagnostics, but they are not the correctness mechanism for import idempotency.

The durable import model must include source-link records outside the asset table.
Source-link records must map a source entity to the Stuff Stash resource created or modified from that source entity.
Source links must cover assets and imported images or attachments.
Source links must be source-adapter agnostic.

Source-link uniqueness must be scoped by:

- Tenant.
- Inventory.
- Source type.
- Source instance identity or fingerprint.
- Source entity type.
- Source entity ID.

The implementation must enforce this uniqueness at the persistence boundary, not only through application-layer preflight checks.
Repeated apply, worker retry, route retry, or process restart must not create duplicate Stuff Stash records for the same source entity in the same tenant and inventory.

Durable import apply must also track records created or modified by an import job in separate import-owned tables.
The asset table must not receive a `pending_import`, `import_job_id`, or equivalent import-specific column for this purpose.
If the UI needs to show which assets or images were created or modified by an import job, it must derive that state through import-owned records or query surfaces.

## Durable Conflict Policy

The first durable import conflict policy is conservative:

- If a source link already exists for a source entity, apply must skip that source entity by default.
- If an existing Stuff Stash record appears to match by supported Homebox duplicate fields but lacks a source link, preview must report a duplicate warning and apply must skip by default.
- Import apply must not overwrite existing Stuff Stash records by default.
- Import apply must not delete Stuff Stash records because records disappeared from the source.
- Import apply must not treat repeated imports as continuous synchronization.

Future import modes such as merge, overwrite, fill-empty-fields, replace, delete propagation, or user-assisted conflict resolution must be specified before implementation.

## Durable Cancellation Semantics

Durable import jobs must support two user-visible cancellation choices:

- Cancel and keep partial progress.
- Cancel and discard partial progress.

Cancel and keep partial progress:

- Stops future import work as soon as the worker reaches a safe cancellation point.
- Preserves records and source links already created or modified by the job.
- Preserves all audit history.
- Leaves the job visible in import history as cancelled with partial progress kept.

Cancel and discard partial progress:

- Stops future import work as soon as the worker reaches a safe cancellation point.
- Runs durable discard cleanup for records created or modified by the job up to the cancellation point.
- Deletes or otherwise compensates for all records created or modified by the job, including imported images or attachments, through application services and ports rather than direct persistence mutation.
- Removes source-link records created by the job for records that were discarded.
- Preserves all audit history.
- Leaves the job visible in import history as cancelled with partial progress discarded, unless explicitly removed from import history later.

Discard cleanup must use import-owned mapping records as its source of truth.
Discard cleanup must not depend on import-specific columns on assets.
Discard cleanup must be resumable and idempotent.
If discard cleanup cannot fully complete, the job must remain visible with a safe failure state and safe messages.

The exact terminal status names for cancelled and discard-failed jobs are implementation details, but the user-visible meaning must distinguish partial progress kept from partial progress discarded.

## Legacy Homebox Import

The first Homebox implementation targets Homebox `v0.24.x` style APIs.

Supported live-source endpoints:

- `POST /api/v1/users/login`.
- `GET /api/v1/status`.
- `GET /api/v1/items`.
- `GET /api/v1/items/{id}`.
- `GET /api/v1/locations`.
- `GET /api/v1/locations/tree`.
- `GET /api/v1/tags`.
- `GET /api/v1/items/{itemId}/attachments/{attachmentId}` for image content when the user includes images.

Supported uploaded CSV format:

- The CSV must use Homebox legacy export columns such as `HB.location`, `HB.tags`, `HB.asset_id`, `HB.name`, `HB.quantity`, `HB.description`, `HB.insured`, `HB.notes`, `HB.purchase_price`, `HB.purchase_from`, `HB.purchase_time`, `HB.manufacturer`, `HB.model_number`, `HB.serial_number`, `HB.lifetime_warranty`, `HB.warranty_expires`, `HB.warranty_details`, `HB.sold_to`, `HB.sold_price`, `HB.sold_time`, and `HB.sold_notes`.
- CSV import cannot include binary images in the first slice because the Homebox CSV export does not contain attachment bytes.
- CSV import must still preview image support as unavailable rather than silently claiming images will be imported.

Legacy Homebox mapping:

- Homebox locations become Stuff Stash assets with kind `location`.
- Homebox location hierarchy becomes Stuff Stash asset parent references.
- Homebox items become Stuff Stash assets with kind `item`.
- Homebox item location becomes the item parent asset reference.
- Homebox archived items become archived Stuff Stash assets when the apply path supports doing so without bypassing lifecycle rules; until then, archived source rows must be reported as warnings and skipped or imported as active only when the user explicitly accepts that behavior.
- Homebox `assetId`/`HB.asset_id` becomes a custom field named `homebox-asset-id`.
- Homebox item IDs, location IDs, CSV location references, and import references become a custom field named `homebox-source-id` used for duplicate detection and audit metadata.
- Homebox quantity becomes a number custom field named `homebox-quantity` until a first-class quantity/consumable model is specified.
- Homebox tags become a text custom field named `homebox-tags` containing a stable semicolon-separated tag list until a first-class tag domain is specified.
- Homebox purchase, warranty, sale, manufacturer, model, serial, insured, and notes values become validated custom fields.
- Empty Homebox values must not create noisy custom field values.
- Homebox date values with year `0001`, such as `0001-11-08`, are partial or ambiguous dates for Stuff Stash purposes. They must be imported as text fields with a warning, not as validated date fields.
- Homebox attachments become Stuff Stash asset attachments when importing from a live Homebox source and the user enables image import.
- Imported image attachments must be eligible for the same primary-photo list, search, and detail presentation paths as images uploaded directly in Stuff Stash.

Initial custom field definitions created for Homebox import are inventory-scoped and apply to all assets unless a future spec defines Homebox-specific custom asset types.

Initial Homebox custom fields:

- `homebox-asset-id`: text.
- `homebox-source-id`: text.
- `homebox-tags`: text.
- `homebox-quantity`: number.
- `homebox-insured`: boolean.
- `homebox-notes`: text.
- `homebox-purchase-price`: number.
- `homebox-purchase-from`: text.
- `homebox-purchase-time`: text.
- `homebox-manufacturer`: text.
- `homebox-model-number`: text.
- `homebox-serial-number`: text.
- `homebox-lifetime-warranty`: boolean.
- `homebox-warranty-expires`: text.
- `homebox-warranty-details`: text.
- `homebox-sold-to`: text.
- `homebox-sold-price`: number.
- `homebox-sold-time`: text.
- `homebox-sold-notes`: text.

Homebox image import requirements:

- Image import is available only for live Homebox source imports in the first slice.
- The importer must download attachment bytes through the Homebox attachment endpoint. It must not use Homebox internal attachment storage paths.
- The importer must sniff content signatures and use the actual supported MIME type rather than trusting Homebox attachment metadata.
- Unsupported attachment content types must be skipped with warnings.
- Oversized attachments must be skipped with warnings unless the current Stuff Stash attachment limit allows them.
- Imported attachment file names must pass Stuff Stash file-name validation. Unsafe or missing names must be replaced with safe generated names.
- Attachment import must use the same attachment application path as normal uploads so content type validation, hashing, blob storage, authorization, observability, and audit behavior remain intact.
- Live Homebox source URLs and redirects must reject loopback, private, link-local, multicast, unspecified, and other internal-only address ranges by default. Private or local network addresses require explicit user opt-in for the source request.
- Attachment downloads must be bounded by the configured Stuff Stash attachment size limit.
- Uploaded CSV content must be bounded before and after base64 decoding.

## Import Preview API

Initial endpoint:

- `POST /tenants/{tenantId}/inventories/{inventoryId}/imports/legacy-homebox/preview`

The request supports two mutually exclusive source shapes:

- Live Homebox source:
  - `sourceType`: `legacy_homebox`.
  - `baseUrl`.
  - `username`.
  - `password`.
  - `includeImages`.
  - `allowInsecureTLS`, default `false`. This option is only for self-hosted Homebox sources with self-signed or otherwise locally trusted certificates. The UI must make the insecure TLS state explicit, and the importer must not silently disable certificate verification.
  - `allowPrivateNetwork`, default `false`. This option is only for Homebox sources that intentionally resolve to private or local network addresses. The UI must make the private-network state explicit.
- Uploaded Homebox CSV source:
  - `sourceType`: `legacy_homebox_csv`.
  - `fileName`.
  - `contentBase64`.

Preview requires authentication and `inventory.configure`.

Preview response must include:

- Source summary.
- Counts of source locations, assets, tags, custom field definitions, attachments, warnings, and blocking errors.
- A bounded sample of planned locations and assets.
- A bounded sample of planned images when available.
- Safe warnings and blocking errors.

Preview must not persist source passwords or Homebox bearer tokens.
The first synchronous Homebox slice does not persist preview plans. Apply re-reads and revalidates the submitted source input immediately before writing.
Production-scale imports must follow the durable import-job, source-link, conflict-policy, and cancellation semantics defined above.

## Import Apply API

Initial endpoint:

- `POST /tenants/{tenantId}/inventories/{inventoryId}/imports/legacy-homebox/apply`

Apply requires authentication and `inventory.configure`.

Apply request includes the same source input and user choices such as whether to import images.

Apply behavior:

- Revalidate the plan before writing.
- Create missing custom field definitions.
- Create location assets before item assets.
- Create item assets after parent location IDs are known.
- Upload supported image attachments after their target assets exist.
- Produce audit records for state-changing operations.
- Use audit source `import`.
- Return counts for created, skipped, warned, and failed records.
- Return per-record safe failures when partial failure is allowed.

The first apply implementation may use best-effort partial failure for attachments after core asset import succeeds. Core location, custom field, and item import failures must stop before later dependent writes when continuing would create misleading containment.

## Duplicate And Conflict Handling

The first slice uses conservative duplicate handling:

- The importer must search existing active and archived assets in the target inventory for matching `homebox-asset-id` values.
- When a source item or location likely already exists, preview must mark it as a duplicate warning.
- Apply must skip duplicates by default.
- Overwrite, merge, replace, delete, and source-link editing are out of scope for the first slice.

## Web UX

The web application must provide a polished import workflow for the first Homebox slice:

- Users can choose live Homebox connection or Homebox CSV upload.
- Live Homebox connection collects base URL, username, password, and image-import preference.
- Live Homebox connection may offer an explicit self-signed certificate option for local or personal Homebox instances. It must be off by default.
- Live Homebox connection may offer an explicit private-network address option for local or personal Homebox instances. It must be off by default.
- CSV upload accepts a Homebox CSV export and clearly states that images are not included.
- The import surface must present the workflow as source setup, preview review, and apply result so users can tell whether anything has been saved.
- The import surface must show active operation state for preview and apply separately. Long-running preview or apply requests may remain synchronous in this slice, but the UI must make the current operation and disabled controls clear.
- Preview shows source version when available, counts, warnings, planned field definitions, locations, assets, and image readiness.
- The web preview badge counts total planned records: planned field definitions, locations, assets, and attachments.
- Preview gives users enough detail to understand tag, quantity, date, duplicate, and image warnings before applying.
- Preview samples must be visually bounded and must state when only a subset of fields, assets, images, or messages is shown.
- Apply must only be enabled for the current source input that was previewed. Changing the source URL, credentials, security options, image option, selected file, or file content after preview must require a new preview before apply.
- Preview failures must be labeled as preview or source-connection failures, not as applied import failures.
- If a browser or transport failure prevents the preview request from completing, the UI must replace raw fetch errors such as `Load failed` with actionable copy that explains the preview request could not complete and points users toward API reachability, Homebox reachability, and the explicit private-network option for local Homebox sources.
- Browser CSV upload must enforce the same 10 MiB decoded CSV limit before reading file bytes and must not use the full base64 payload as a repeated reactive preview key.
- Apply shows progress states and final results, including created field definitions, locations, assets, attachments, reused field definitions, skipped assets, and skipped attachments when those counts are nonzero.
- After apply returns successfully, the web UI must present the import as applied even if a follow-up workspace refresh fails. Refresh failures may be shown as a non-fatal warning, but they must not overwrite the successful apply result with an import-failed state.
- The UI must avoid showing passwords, tokens, Homebox internal storage paths, or raw attachment bytes.
- The UI must fit the existing Stuff Stash web design language and use the real inventory workspace rather than a separate marketing-style page.

## Media And Backups

- Photo and file export should be supported eventually.
- Initial export may exclude binary attachment content if the export clearly states that limitation.
- Tenant-level backups should be modeled as exports unless a future backup spec defines a separate mechanism.
- Export packaging for photos and files must be specified before binary media export is implemented.

## Testing

- Tests must verify JSON export, CSV export, authorization, tenant isolation, inventory isolation, custom field handling, and audit behavior.
- Tests must use fakes for storage and repositories where appropriate.
- Import tests must verify validation failures and partial-failure behavior before import is exposed to users.

## Open Questions

- What exact JSON export schema should be used first?
- What exact CSV columns should be used first?
- Should exports include audit history?
- Should exports include attachment metadata before binary file export is supported?
