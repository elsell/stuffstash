# Import And Export Spec

## Purpose

Stuff Stash should make user data portable.

Users should be able to export inventory data and later import data through supported formats without coupling the domain to one file format.

## Scope

This spec covers initial import and export requirements.

This spec covers the first legacy Homebox import workflow and the durable import-job direction for production-scale imports.

This spec does not define the final Stuff Stash-native CSV columns, final Stuff Stash-native JSON schema, backup packaging, media export packaging, newer Homebox entity import behavior, permanent source-link management UI, or all future import conflict resolution modes.

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
- Imports must use a source-neutral planning and execution workflow:
  - Preview builds and validates an import plan without mutating Stuff Stash state.
  - Execution revalidates the reviewed plan and then writes through application services, ports, and adapters.
  - Durable imports start execution through a persisted job and worker, not through a synchronous request-lifetime apply endpoint.
- Import source adapters must convert source-specific data into a Stuff Stash import plan. Source adapters must not directly create Stuff Stash assets, custom fields, custom asset types, attachments, inventories, or audit records.
- Import plans must be source-neutral enough that future import sources can reuse the same validation and apply path.
- Import apply must run under a real authenticated principal and must enforce the same tenant, inventory, authorization, validation, audit, observability, attachment, and containment rules as normal application workflows.
- Import plans must include source references so repeated imports can detect likely duplicates before writing.
- Import preview must report warnings and blocking errors separately.
- Import warnings must be safe to show to users and must not include source passwords, bearer tokens, attachment storage paths, or provider internals.
- Import source validation and connection failures may return a safe, actionable error detail, such as missing URL, unsupported URL scheme, blocked private-network source, TLS trust failure, or Homebox HTTP status, only when the source adapter marks the detail with the import-source user-error contract. Ordinary adapter errors must remain generic at the HTTP boundary. The web UI must prefer typed safe details over a generic invalid-request message.
- CSV upload validation performed before source-adapter execution, such as invalid base64 upload content or files larger than the supported import limit, must use the same safe import-source detail path with user-actionable copy. These details must describe how to fix the upload without echoing file contents, raw payloads, credentials, provider paths, source-specific brands outside the selected adapter, or storage keys.
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

Production-scale live imports use inventory-scoped durable jobs with persisted preview state, progress, cancellation, audit, observability, source-link uniqueness, encrypted temporary credentials, and worker-based execution.

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
Jobs whose discard cleanup failed must remain visible and recoverable until cleanup reaches a terminal cancelled state; users must not be able to remove discard-failed jobs from import history.

Durable import jobs must have explicit permissions:

- Import job view permission.
- Import job create permission.

The first authorization mapping derives these permissions from the existing inventory edit/write permission or relationship. Inventory editors, inventory owners, and tenant owners may view and create import jobs; inventory viewers may not. Application and adapter code must still represent import-job permissions explicitly so future access policy changes do not require changing the import domain model.
Inventory access summaries returned to clients must include the explicit import-job permissions when allowed so clients can render import navigation, history, setup, and unavailable states from the same permission model enforced by the API.

Durable import-job execution must be behind a worker port.
The first runtime may provide an in-process worker adapter hosted by the API process, but route handlers must not directly execute long-running import work.
The worker port must allow a future standalone worker process without changing import application behavior.
Worker adapters must load already-authorized job snapshots through an internal worker application path, not through public user-facing detail queries that re-check current view permission.
Starting a previewed job must use a repository-level conditional transition so concurrent starts cannot both launch workers for the same job.
Temporary source credentials must be stored only by the start request that wins the previewed-to-running transition. A failed concurrent start must not be able to overwrite source credentials or source options for an already-running job.
If startup fails after a job has won the previewed-to-running transition, the application service must terminalize the job as failed, clean any stored source material when possible, and audit the failure when the audit sink is available.
If startup-failure cleanup deletes stored source material, it must write the same safe credential-cleaned audit action as worker and vacuum cleanup.
Restart recovery must claim each running or cancellation-requested job through a repository-level compare-and-set before dispatching it to the worker.
If recovery finds a running job whose source material is missing or unreadable, it must terminalize the job as failed with a safe message and audit record instead of leaving the job running indefinitely.
Jobs already marked cancellation-requested before a crash must be recovered through the worker port and terminalized as partial progress kept, partial progress discarded, or discard failed.
Worker terminalization and cancellation requests must use repository-level conditional transitions so stale workers and late cancellation requests cannot overwrite terminal or cancellation state.
When a running job is cancelled, the cancellation request correlation ID must be retained with the durable job state until terminalization so cancellation-completion audit records can be correlated to the request that cancelled the job rather than the request that started it.

Import jobs must persist normalized plan metadata and source references after preview.
Import jobs must not persist Homebox attachment bytes or other source attachment bytes in the preview plan.
Attachment bytes must be fetched during apply when the source adapter and user options require attachment import.
Source adapters must distinguish preview planning from apply execution at the port boundary.
Preview planning may count and expose safe attachment metadata such as source attachment ID, target source asset ID, filename, content type, and primary-photo status, but it must not download or retain attachment bytes.
Apply execution may fetch attachment bytes only through an explicit worker/apply source request.
Source fingerprints must ignore attachment byte content, byte-size metadata, sniffed attachment content type, generated attachment filenames, and safe warning messages that are only knowable after apply downloads the attachment.
Attachment source identity, target source asset identity, and primary-photo status remain preview-known fingerprint inputs.
Attachment fingerprint input order must be deterministic so provider response ordering changes do not force a new preview when the attachment set is otherwise unchanged.
Plan normalization used for fingerprinting, persisted preview metadata, and safe message handling must not mutate the source adapter's apply plan or remove attachment bytes needed by execution.
If attachment bytes cannot be fetched during execution for an attachment that was visible during preview, the source adapter must keep the attachment's source identity in the execution plan and mark it unavailable instead of returning an empty attachment to the normal upload path.
The application service must count unavailable source attachments as skipped and report a source-download warning.
It must not call the normal attachment creation path for unavailable source bytes, because normal upload validation errors describe malformed user content rather than a source download failure.

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
The first credential vacuum interval is configured by `STUFF_STASH_IMPORT_CREDENTIAL_VACUUM_INTERVAL_SECONDS`, which defaults to `60`.
Successful worker cleanup of stored credentials after terminal execution must write the same safe credential-cleaned audit action as recurring credential vacuum cleanup.
Import job identifiers are globally unique, but credential vacuum persistence must still delete the exact tenant/inventory/job-scoped source material rows selected for cleanup. It must not collapse scoped selection into a job-identifier-only delete.

Import-job credential cleanup must be observable and auditable without exposing credentials, bearer tokens, passwords, or provider internals.
Credential cleanup audit must identify only the import job scope and safe source or job metadata.
Credential cleanup audit must not include encrypted source payloads, ciphertext, nonces, passwords, bearer tokens, raw URLs containing credentials, or provider internals.

Durable import jobs must be observable through domain-oriented probes.
Observability must include job lifecycle, worker claiming, worker retry, worker failure, credential cleanup, source-fingerprint mismatch, source-option mismatch, preview progress, apply progress, cancellation, discard cleanup, and source-link duplicate prevention.
The first durable import probes include previewed, started, progress updated, recovery claimed, source fingerprint mismatch, source option mismatch, cancellation requested, discard cleanup completed or failed, history removed, worker failed, credential vacuumed or failed, and source-link duplicate skipped events.
The in-process worker adapter must record worker execution failures from background jobs because those errors occur after the request that dispatched the job has already returned.
Observability fields must not include source credentials, bearer tokens, raw provider paths, raw storage keys, or raw adapter internals.

Durable import jobs must be fully auditable.
Audit history must include job creation, preview start, preview completion, apply start, apply completion, apply failure, cancellation request, cancellation completion, discard cleanup, credential cleanup, and imported record creation or modification where state changes occur.
Existing asset, custom field, attachment, and other domain audit records must still be written by the existing application services used during import apply.
Import-specific audit metadata must include the import job ID where practical.

The first durable import audit target type is `import_job`.
The first import-job audit actions are:

- `import_job.previewed`.
- `import_job.started`.
- `import_job.completed`.
- `import_job.failed`.
- `import_job.cancellation_requested`.
- `import_job.cancelled`.
- `import_job.history_removed`.
- `import_job.credential_cleaned`.

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

For live Homebox imports, the first source instance identity is the normalized, credential-free Homebox base URL returned in the source summary.
Live Homebox base URL normalization must preserve an explicit `http://` or `https://` scheme, treating the scheme case-insensitively. If the user enters a schemeless host, the importer may default to `https://` for the first connection attempt.
The source fingerprint must still be used to require a fresh preview when the source contents changed between preview and start, but it must not be the live Homebox source-link instance key because the same Homebox instance can legitimately contain different records over time.

For CSV imports, the first source instance identity is the preview source fingerprint because the uploaded file is the source instance available to the importer.
Future file-import adapters that expose a stronger durable source identity must specify that identity before using it for source links.

The first source entity types are:

- `asset` for imported item, container, and location-like asset records.
- `attachment` for imported image or attachment records.

The first resource types are:

- `asset` for Stuff Stash assets.
- `attachment` for Stuff Stash media attachments.

The implementation must enforce this uniqueness at the persistence boundary, not only through application-layer preflight checks.
Repeated apply, worker retry, route retry, or process restart must not create duplicate Stuff Stash records for the same source entity in the same tenant and inventory.
Imported asset persistence must commit the Stuff Stash asset, audit record, source link, and import job resource record in one persistence unit of work.
Imported attachment persistence must commit the Stuff Stash attachment metadata, audit record, source link, and import job resource record in one persistence unit of work.
Attachment bytes may be written to blob storage before the metadata transaction, but the application must delete the blob when the imported attachment metadata transaction fails.

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
Discard-failed jobs must remain retryable by recovery or worker execution. Retrying discard cleanup must be idempotent: already-deleted records and already-deleted source links must not make the retry fail or duplicate audit history.
Running-job cancellation must preserve the current progress phase, done count, and total count while changing only cancellation status, mode, message, and timestamps.

The exact terminal status names for cancelled and discard-failed jobs are implementation details, but the user-visible meaning must distinguish partial progress kept from partial progress discarded.

## Durable Import UX

Durable imports must be presented as reviewable background jobs, not as modal-only tasks.
The default import surface must be import history, including in-progress jobs and past successful, failed, or cancelled jobs.
The user-facing mental model is: review import history, start a new import when needed, confirm the source connection, preview what will happen, start the durable job, and return to history to monitor progress.
The durable import UI may depart entirely from the first synchronous Homebox import structure.
The durable import UI must optimize for low-friction, low-cognition user workflows, confidence, and transparency rather than mirroring source-adapter names, API endpoints, or backend implementation mechanics.

The durable import UI must live inside the inventory workspace.
The guided new-import flow must show one primary step per screen and make the user's place in the flow visible without requiring them to remember the sequence.
The guided flow must use spatial progress cues for the current step and must not duplicate that cue with redundant visible copy such as `Step 3 of 4` beside the same step indicator.
The guided flow step indicator must be interactive navigation for reachable steps, not static explanatory text.
Users must be able to return to prior source-selection, source-setup, and preview-review steps without discarding the active wizard session's non-secret draft state or existing preview plan.
The step indicator must not allow jumping forward to states that do not yet exist; preview requires an existing preview plan, and run requires a started durable job.
Changing source inputs, source options, selected file metadata, selected file content, image options, security options, or source fingerprint after returning to setup must make the existing preview stale and require a new preview before the import can start.
Import source text inputs must avoid mobile auto-capitalization, autocorrection, and spellcheck where they collect URLs, emails, usernames, passwords, or source identifiers.
The UI must present source snapshots, imported records, and audit links in user-facing language. Internal resource IDs, source fingerprints, and source entity IDs may be available as secondary diagnostic metadata where useful, but they must not be the primary label for records or trust signals.
Removing an import job from visible history must require an explicit confirmation in the UI. The confirmation must state that imported records and audit history remain.
Removing an import job from visible history must use a focused confirmation dialog or overlay. It must not replace the import history or detail surface with a full-page confirmation step.
State-changing import actions must show an in-progress affordance, such as an animated spinner and operation-specific label, while the operation is pending; disabling controls alone is not sufficient feedback.
Import refresh and detail-loading actions must also show a visible loading affordance while the read operation is pending so users can distinguish a delayed response from an inert control.
It must preserve inventory context throughout source setup, preview, running progress, result review, and history.
The import surface must not look or behave like a marketing page or separate administration console.

The import workspace must include:

- A way to start a new import.
- Import detail sections must remain readable on compact mobile widths. If the detail view uses tabs, tab labels must not clip or overlap; the row may use a horizontal rail, concise visible labels, and fuller accessible names where needed.
- A way to view in-progress imports.
- A way to view past successful imports.
- A way to view past failed imports.
- A way to view past cancelled imports.
- A way to open an individual import job detail.
- A way to remove terminal jobs from import history without removing audit history or imported records.

The import history screen must summarize the inventory's import state before the row ledger so users can quickly distinguish active work, previews waiting for confirmation, finished imports, imports with warnings, and imports that require action.
The import history screen must not repeat page-level headings such as `Imports` and `Import history` as competing primary hierarchy. The page title owns the surface; the history content below it should start with controls, status filters, current work, or the ledger.
History status summaries must be aligned, compact controls when they filter or jump the page state. They must expose selected state instead of looking like decorative statistics, and their count/icon alignment must make each summary read as a single clickable control.
History rows must use user-facing status language and must not require users to know durable-job implementation statuses.
History rows should use `Source` as the first ledger column because import sources are pluggable. Source cells should show the source name, instance identity, and source version when available; method details such as live connection, CSV upload, image options, and default network protections must not compete with the source identity unless they require user attention.
History rows should avoid repeating the same warning or result state across multiple columns. A warning-level completed import may show a single amber issue indicator and concise result text; detailed warning counts and explanations belong in job detail.
History row timestamps should favor the user-actionable temporal state, such as completed time for terminal jobs or elapsed/progress state for active jobs. Actor labels are secondary metadata and must not make the timestamp column visually heavy.
Previewed jobs are drafts waiting for confirmation, not completed history; the UI must present them separately from terminal history and make them easy to resume.
If an inventory has active import jobs, the UI must make those jobs easy to resume from the import surface before encouraging a new import.
Active jobs and preview drafts must be grouped as current work ahead of completed history so users can resume, inspect, or cancel ongoing work before starting another import.
Jobs promoted into current-work groups must not be duplicated in the default history ledger immediately below those groups.
Warnings and blocking errors must be visually separated. Warnings are non-blocking review signals and should be shown as amber indicators in the main history ledger. Blocking errors, failed imports, and discard-cleanup failures require action and must use red action-required row emphasis in the main history ledger.
The history page may show a compact action-required alert before the ledger, but it must not render a second row list that pushes the actual history ledger below the first viewport.
In history summaries and filters, `Completed` means a succeeded import without action-required severity. Succeeded imports with warning-only severity may still count as completed; succeeded imports with returned blocking errors or stale error counts must count as action-required instead.
Import detail must preserve the same severity model as history: warning-only runs may show an amber review callout, while blocking errors, failed imports, and discard-cleanup failures must use action-required language and red destructive/error emphasis.
Import detail must avoid repeating the same state in the heading, badge, metric strip, and issue callout. The top summary should make source identity, final state, changed records, and issue severity clear with as few repeated labels as practical.
Import detail must distinguish source identity from import method. For Homebox, the source is Homebox plus the safe instance URL and version when available; live connection or CSV upload are method details.
Default safe options such as normal network protections should not be surfaced as standalone source facts unless the user changed a risky option or needs to understand a failure.
Import detail issue groups must support progressive disclosure at the group level. A dense warning set must let users expand one warning cause without expanding every issue group or turning the page into an unbounded wall of repeated rows.
Import preview plan sections must show compact section counts so users can understand the scale of fields, locations, assets, and photos/files before reading individual samples.
When import detail counts and returned messages disagree, the web UI must render the highest safe severity from terminal job status, count fields, and returned message severities; a returned error message is action-required, and a returned warning message is warning-level unless action-required status or counts are also present.
The history ledger should be compact enough that recent runs remain visible in the first screen on ordinary desktop viewports whenever there is no active blocking work, and terminal rows should avoid wall-of-text descriptions by reserving detailed issue lists for job detail.
Users may explicitly switch the ledger to action-required or warning history when they want the full row-level context for jobs that need review.
If an inventory has no import jobs, the import surface must show an empty state with one clear action to start an import.
The import surface must not show the source setup form by default.

The new-import flow must follow the durable import job lifecycle:

- Source selection.
- Source setup and connection confirmation.
- Preview review.
- Job execution.
- Return to import history for progress and result review.

The new-import flow must be step-by-step, with one primary task per screen:

- Step 1 selects how to bring Homebox data in.
- Step 2 configures and confirms the selected source connection or file.
- Step 3 previews the normalized import plan and asks for explicit import confirmation.
- Step 4 starts the durable job, shows a brief handoff confirmation that the import is running in the background, and gives the user a direct path back to the import-history screen showing the in-progress job.

On mobile viewports, fixed browser or application chrome must not cover import setup summaries, source options, or primary step actions. The source setup screen must reserve enough bottom clearance for the confirm and back actions to remain fully visible and tappable.

Source selection must present live Homebox connection and Homebox CSV upload as two large, clearly differentiated choices with enough explanatory detail for the user to decide without opening either option.
The source-selection screen must frame the choice as an import method choice, not as an adapter or endpoint choice.

Source setup must:

- Let the user choose the import source adapter.
- Present source choices in user-understandable import terms before exposing technical adapter details.
- Show only the fields and options required for the chosen source adapter.
- Keep advanced or risky source options visually subordinate to common options, preferably collapsed or grouped away from the primary source fields until the user needs them.
- Make image-import choices explicit when the source adapter can import images.
- Avoid showing or retaining source passwords, tokens, Homebox internal storage paths, or raw attachment bytes in visible UI state.

Preview review must:

- Clearly state that nothing has been saved yet.
- Show source identity and source version when available.
- Show planned counts for locations, assets, images or attachments, custom fields, duplicates, warnings, and blocking errors when available.
- Group warnings and blocking errors by user-understandable cause.
- Collapse exact duplicate safe warning or blocking messages in visible issue summaries so repeated provider messages do not create wall-of-text detail pages. Distinct affected source records must remain distinguishable when their safe source name or source ID differs. When reported warning or blocking counts differ from the deduplicated visible messages, summary counts must use the reported counts while the issue-list affected-record stat must describe the distinct visible records.
- Put readiness, blocking status, and re-preview requirements ahead of sample rows.
- Show bounded samples of planned records and state when only a subset is shown.
- Disable start-import actions when the preview has blocking errors.
- Require a new preview when the source input, source options, selected file, file content, image option, security options, or source fingerprint changes.
- Treat source-fingerprint changes as blocking re-preview requirements, not warning-only states.

Job execution must:

- Show that the import is running as a durable background job.
- Let the user leave and return without implying that the browser tab owns the job.
- Show the current phase in user-facing language.
- Show progress counts when a phase has a known total.
- Avoid misleading exact percentages when the current phase total is unknown.
- Show safe warnings and failures as they are available.
- Show terminal results in terms of created, reused, skipped, discarded, and warning counts before exposing detailed source references.
- Provide a cancellation action while the job is cancellable.

The application service must persist progress updates at safe checkpoints inside durable execution, not only when a phase starts or ends.
For phases with known totals, the job's `done` count must advance after each processed field definition, location-like asset, item asset, and attachment, whether the individual record was created or skipped.
Phase totals must be scoped to the phase currently displayed:

- Custom field creation progress uses the number of planned custom fields.
- Location creation progress uses the number of planned location-like assets.
- Asset creation progress uses the number of planned non-location assets.
- Attachment import progress uses the number of planned attachments.

If a phase has no work, execution may skip directly to the next phase without persisting a misleading zero-total progress state.
Cancellation checks must run at the same safe checkpoints used for progress updates so that a cancellation request is reflected before more records are created.
Terminal progress must preserve a truthful final phase message and must not imply that an unknown-total phase reached a precise percentage.

Cancellation UI must present both supported choices:

- Cancel and keep imported items.
- Cancel and discard imported items.

The cancellation confirmation must explain the consequence of each choice in plain language.
The discard choice must make clear that audit history remains even when imported records are removed from the inventory.

Result review must:

- Show final status.
- Show created, modified, skipped, warned, failed, and discarded counts when applicable.
- Distinguish import failure from non-fatal refresh or navigation failures.
- Provide a path to inspect imported records when records remain in the inventory.
- Label imported records with safe user-facing names, such as the asset title or attachment file name, when the records still exist. Internal source entity IDs and resource IDs may remain available as secondary diagnostic metadata, but they must not be the primary imported-record label.
- Jobs that completed cancellation with partial progress discarded must not expose discarded resource summaries as openable imported records, even though audit history and internal cleanup tracking remain.
- Provide a path to relevant audit history.
- Preserve safe failure detail without exposing credentials, bearer tokens, provider internals, raw source paths, or raw storage keys.

Import history rows must summarize:

- Source adapter.
- Status.
- Started time.
- Completed time when available.
- Actor when available.
- Counts or progress summary when available.
- Whether partial progress was kept or discarded for cancelled jobs.

History rows must read like a compact job ledger: status, source, actor, important times, progress or result counts, and available actions must be visually distinct enough to scan without opening each job.

Import job detail must include:

- Summary.
- Source and options summary without secrets.
- Preview plan summary.
- Progress and phase history.
- Warnings and failures.
- Deduplicated, grouped warning and failure summaries for exact duplicate safe messages.
- Created or modified records when records remain.
- Cancellation and discard outcome when applicable.
- Link or navigation path to audit history.
- Live navigation for audit history and imported-record actions when the web app can handle the target in the current workspace, with normal links retained as reload and deep-link fallback.
- URL-addressable detail state for the selected import job and selected detail tab so users can reload or share the current import detail context without losing their place. The canonical web route for an import job detail must identify the job in the path and may identify the selected detail tab with a query parameter. If no tab is selected in the URL, the detail view may choose the most helpful default tab for the job, such as `Issues` when warnings or errors are present.

When a job completed with partial progress discarded, job detail may acknowledge that records were created and discarded, but it must not render those discarded resources as normal openable inventory records.

The import job API response may include a bounded safe imported-resource summary derived from import-owned resource records.
Imported-resource summaries must include only tenant/inventory/job-scoped resource identifiers, resource type, resource owner identifier when applicable, optional safe display name, source entity type, source entity ID, and creation time.
The optional display name must be derived from the current Stuff Stash record when that record still exists, such as an asset title or attachment file name, and must not be accepted from arbitrary provider metadata.
Imported-resource summaries must not include source credentials, bearer tokens, provider paths, storage keys, attachment bytes, encrypted payloads, or arbitrary provider internals.
The first resource summary limit is implementation-defined and must be enforced at the import-owned repository/query boundary before application mapping or JSON response construction.
The web UI must keep the resource summary visually bounded when shown in job detail. Large imported-record summaries should use pagination or an equivalent bounded table pattern rather than an expanding wall of rows.
The import job API response must include the safe actor principal identifier when the job has an actor so import history can identify who started or prepared the job without exposing provider credentials or source internals.
When the API can resolve the actor through the existing user profile repository, it must also include a safe actor summary with the principal ID and email so import history can show a user-facing label.
The web UI must prefer a known user-facing actor label from the API actor summary or the signed-in user's email when the actor matches the current principal, before falling back to a compact opaque principal identifier. Compact history rows may omit actor attribution when it would compete with source, status, record, and completion-time scanning; job detail, current work, or confirmation surfaces should still expose the actor when available.

The import job API response must include a safe progress history for job detail.
Progress history records must include only the phase, done count, total count, safe user-facing message, and update time.
Progress history is a phase timeline, not an unbounded per-record log: it may collapse repeated updates within the same phase and message while the current progress snapshot remains authoritative for exact current `done` and `total` counts.
Progress history must not include source credentials, bearer tokens, provider paths, storage keys, attachment bytes, encrypted payloads, or arbitrary provider internals.

The durable import UI must follow Stuff Stash brand guidance:

- Clean, calm, task-focused layout.
- Compact workspace-native information density.
- System-like primary actions.
- Semantic status colors only where they communicate real status.
- No decorative import illustrations, marketing hero treatment, or enterprise data-loader styling.
- Direct, specific, actionable copy.
- Familiar background-job and upload-history interaction patterns where they reduce cognitive load.

The web API adapter must record durable import workflow observability events at the frontend boundary for history loading, preview, start, cancellation, and history removal. Frontend import observability attributes must be safe: they may include source type, job ID, cancellation mode, and counts, but they must not include Homebox passwords, bearer tokens, raw CSV content, source payloads, or provider internals.
The web API adapter must attach safe request correlation IDs to import preview, start, cancellation, and history-removal mutations so API audit records can be correlated to browser-originated actions without exposing source credentials or payloads.

Import web tests must mirror the workflow boundaries:

- Workspace app tests may cover route integration, canonical navigation, and handoff into the import workspace.
- Import-workspace behavior tests must mount the import workspace or an import-specific harness directly.
- Import-specific fakes, fixtures, and helpers must not accumulate in broad workspace app tests once the import workspace owns the behavior being verified.
- Durable import API adapter tests must stay at the adapter boundary and must not duplicate component interaction tests.

Import web implementation must keep the durable workflow split by UI responsibility as it grows. The workspace shell may own inventory scope, route handoff, polling, and command orchestration, but history rows, current-work presentation, preview/detail presentation helpers, source setup controls, and destructive confirmations should move into focused helpers or components before the workspace file becomes a catch-all for every import screen state.

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
- Attachments that cannot be downloaded during execution must be skipped with source-download warnings, not normal attachment validation warnings.
- Imported attachment file names must pass Stuff Stash file-name validation. Unsafe or missing names must be replaced with safe generated names.
- Attachment import must use the same attachment application path as normal uploads so content type validation, hashing, blob storage, authorization, observability, and audit behavior remain intact.
- Live Homebox source URLs and redirects must reject loopback, private, link-local, multicast, unspecified, and other internal-only address ranges by default. Private or local network addresses require explicit user opt-in for the source request.
- Attachment downloads must be bounded by the configured Stuff Stash attachment size limit.
- Uploaded CSV content must be bounded before and after base64 decoding.

## Durable Import Job API

Canonical import endpoints:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/imports/jobs`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/preview`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/{jobId}`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/{jobId}/start`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/{jobId}/cancel`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/{jobId}`

Import job preview creates a durable reviewed plan record.
Import job start revalidates the source fingerprint before worker execution.
The synchronous legacy Homebox preview/apply endpoints must not be kept as a parallel user-facing import API once durable jobs are available.

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

Uploaded CSV bytes are valid only for uploaded CSV source requests. Mixed source requests, such as a live Homebox request that also includes uploaded CSV bytes, must fail with a safe user-actionable import-source detail instead of being coerced into another source shape.

Preview, start, cancel, and remove-from-history require authentication and import job create permission.
List and detail require authentication and import job view permission.
Remove-from-history must return `204 No Content` after the history entry is hidden. It must not return the removed job as a live resource representation.
Preview, start, cancel, and remove-from-history must preserve the caller's request correlation ID in import-job audit records when the request supplies one.
Import-job request correlation IDs must be trimmed and limited to 128 characters before any durable job mutation is written.

Preview response must include:

- Source summary.
- Counts of source locations, assets, tags, custom field definitions, attachments, warnings, and blocking errors.
- A bounded sample of planned locations.
- A bounded sample of planned assets, excluding location records.
- A bounded sample of planned images when available.
- Safe warnings and blocking errors.

Preview must not persist source passwords or Homebox bearer tokens.
Durable preview persists normalized plan metadata and source references, but not attachment bytes.
Start request includes the same source input and user choices such as whether to import images so the source fingerprint can be revalidated before worker execution.
Start and worker execution must store and use a source request that explicitly permits attachment byte fetching only after the previewed-to-running transition wins.

Worker apply behavior:

- Revalidate the plan before writing.
- Create missing custom field definitions.
- Create location assets before item assets.
- Create item assets after parent location IDs are known.
- Upload supported image attachments after their target assets exist.
- Produce audit records for state-changing operations.
- Use audit source `import`.
- Return counts for created, skipped, warned, and failed records.
- Return per-record safe failures when partial failure is allowed.

The first worker apply implementation may use best-effort partial failure for attachments after core asset import succeeds. Core location, custom field, and item import failures must stop before later dependent writes when continuing would create misleading containment.

## Duplicate And Conflict Handling

The first durable Homebox slice uses conservative duplicate handling:

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
- The import surface must present the workflow as source setup, preview review, durable job start, and job history/result review so users can tell whether anything has been saved and whether work is still running in the background.
- The import surface must show active operation state for preview, job start, background progress, cancellation, and result review separately. The UI must make the current operation and disabled controls clear without implying that the browser tab owns durable execution.
- Preview shows source version when available, counts, warnings, planned field definitions, locations, assets, and image readiness.
- The web preview badge counts total planned records: planned field definitions, locations, assets, and attachments.
- Preview gives users enough detail to understand tag, quantity, date, duplicate, and image warnings before starting the import job.
- Preview samples must be visually bounded and must state when only a subset of fields, assets, images, or messages is shown.
- Start import must only be enabled for the current source input that was previewed. Changing the source URL, credentials, security options, image option, selected file, file content, or source fingerprint after preview must require a new preview before the durable job can start.
- Preview failures must be labeled as preview or source-connection failures, not as applied import failures.
- If a browser or transport failure prevents the preview request from completing, the UI must replace raw fetch errors such as `Load failed` with actionable copy that explains the preview request could not complete and points users toward API reachability, Homebox reachability, and the explicit private-network option for local Homebox sources.
- Browser CSV upload must enforce the same 10 MiB decoded CSV limit before reading file bytes and must not use the full base64 payload as a repeated reactive preview key.
- The web preview freshness key must not retain the live Homebox password in plaintext; it may retain a one-way comparison marker sufficient to detect source-input changes.
- Durable job history and detail show progress states and final results, including created field definitions, locations, assets, attachments, reused field definitions, skipped assets, and skipped attachments when those counts are nonzero.
- After a durable job reports success, the web UI must present the import as completed even if a follow-up workspace refresh fails. Refresh failures may be shown as a non-fatal warning, but they must not overwrite the successful job result with an import-failed state.
- The UI must avoid showing passwords, tokens, Homebox internal storage paths, or raw attachment bytes.
- The UI must fit the existing Stuff Stash web design language and use the real inventory workspace rather than a separate marketing-style page.
- As the import surface moves from the synchronous Homebox slice to durable import jobs, the web UI must follow the durable import UX requirements above.

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
