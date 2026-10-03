# Asset Labels And QR Scanning

## Status And Scope

Accepted design direction from the October 3, 2026 discussion; not implemented
or hardware-verified. This spec owns label identity, rendering, QR resolution,
and web/mobile interactions. [Printer integration](printer-integration.spec.md)
owns registered printers and jobs; [CLI](../platform/cli.spec.md) owns the local
consumer. Implementation must follow the test-first repository workflow.

The first release supports item, container, and location assets, individual label
printing, print-after-create, label download/system printing, and in-app QR
scanning. Batch printing, preprinted unassigned labels, public asset sharing,
shared hosted link resolution, and automatic smart-outlet control are deferred.

## Identity And Links

- Labels identify assets; they do not grant access. No bearer token, credential,
  asset name, containment path, or other private content belongs in a QR URL.
- Persist a random, immutable instance identity independent of hostnames, plus a
  random opaque label ID mapped to tenant, inventory, and asset. A label is not a
  capability token. Never reuse or reassign it to another asset.
- Provision the instance identity once through an explicit bootstrap operation;
  replicated API processes must agree on the persisted value.
- One canonical label identity per asset is sufficient initially. Reprinting,
  renaming, editing, moving, or archiving the asset preserves that identity.
- Use a versioned HTTPS link: `{labelBaseURL}/l/v1/{instanceId}/{labelId}`.
  `labelBaseURL` is environment-backed deployment configuration, defaulting to
  the configured public web base URL, never to an untrusted request Host header.
  Respect configured path prefixes and reject credentials/query/fragment in the
  configured base URL. Validate origin and HTTPS at startup.
- A separately operated stable label address is optional. Its operator must keep
  TLS, routing, and any app-association files available. No shared external
  resolver is required by Stuff Stash.
- Backups/restores used for migration preserve instance identity and label
  mappings. A separately operated clone must generate a new instance identity
  and new mappings; it must not silently impersonate the original instance.
- JSON/CSV asset import into a different instance does not preserve scannability
  of old labels in this release. Document the difference from full restore.
- Hard deletion tombstones the label mapping without retaining the asset's
  private contents. Never redirect an old label to a replacement asset.
- A renamed domain cannot be discovered by an ordinary phone camera from a dead
  URL. Browser compatibility requires the printed address to remain reachable
  or a stable label address to redirect it. Do not promise otherwise.

## Resolution And Security

- Web `GET /l/v1/{instanceId}/{labelId}` is a landing route. It preserves the
  intended destination across sign-in and resolves through the authenticated
  API. It must not expose asset metadata before authorization.
- API `GET /labels/v1/{instanceId}/{labelId}` resolves a mapping internally,
  then requires current `inventory.view` through the authorization port before
  returning tenant/inventory/asset navigation data. This is an explicit narrow
  exception to scoped discovery: a label-lookup port accepts both opaque IDs,
  and subsequent asset reads require the returned tenant and inventory scope.
  Missing, foreign-instance, deleted, or unauthorized labels have indistinguishable
  safe responses. Apply rate limiting and safe read audit records on success.
- Archived assets may resolve for authorized readers with their archived state.
  The destination must not present a missing active-list entry as a missing asset.
- `GET /instance` exposes only protocol version and the non-secret persistent
  instance identity before login; it must not expose tenant or inventory metadata.
- An in-app scanner parses the versioned path locally and matches the embedded
  instance identity against the authenticated, configured instance. It can thus
  resolve a migrated label without fetching its obsolete hostname.
- Scanned data must never change the active server, forward credentials to the
  QR origin, trigger arbitrary network fetches, or execute an action. Unknown
  instances require explicit connection/sign-in outside the scanning task.
- Validate URL length, encoding, scheme, version, path structure, and IDs; reject
  malformed data, unrelated QR codes, unsupported versions, and ambiguous paths
  with actionable messages. Do not fall back to arbitrary browser navigation.
- Universal Links/App Links are supported only for domains declared in the
  distributed app and verified by association files. General distribution cannot
  automatically claim arbitrary self-hosted domains. Extend the existing
  invitation link handling without breaking invitation or OIDC callbacks.
- Unassociated HTTPS domains use the web landing page with an explicit Open in
  Stuff Stash action using `stuffstash://labels/v1/{instanceId}/{labelId}` and validated IDs. Preserve
  a usable web fallback; never assert that app detection or launch is guaranteed.
- Pending navigation is reauthorized after authentication and server changes.
  Custom-scheme handlers apply the same identity and permission checks as scans.

## Templates, Registered Label Size, And Rendering Ownership

- Register the physical printer and its current label size in one task. Each
  registered destination has exactly one configured media setting in the first
  release, for example Garage Brother — 29 x 90 mm. The user selects that size;
  automatic roll detection is not required and never substitutes another size.
- Internally, typed media settings contain physical width/height, printable
  margins, resolution, raster dimensions/orientation, color mode, and cut policy.
  A built-in adapter preset supplies technical values where known. For the first
  QL-800 adapter the user selects 29 x 90 mm; this is the only required printer
  and label-size combination in the first release. Unknown/unsupported media is
  rejected with setup guidance, not silently treated as compatible. Other sizes can be
  supported by adding validated presets without changing the registration model.
- Media settings belong to the printer registration, not an independently managed
  inventory media catalog. Do not require users to create separate size resources,
  assign a list of rolls, or select a size again for each queued print request.
- Users with `inventory.configure` can edit the registered size when changing
  rolls; the printer ID and queue remain stable. Each job captures the original
  effective media values and their fingerprint. Compare effective values rather
  than only revision numbers, so changing back to a previous size permits its
  waiting jobs to print without regenerating them.
- New jobs use the current registered size. Existing jobs with different media
  remain queued with Label size changed; restore the configured size or cancel
  the job. Never resize their images, silently reassign media, or consume failure
  retries. Later compatible jobs may proceed; do not block the entire queue on
  the oldest incompatible job. Configuration edits while printing are rejected
  until the attempt settles or uncertainty is resolved. Start/configuration
  changes race atomically; the worker revalidates before sending device data.
- The configured size is the user's assertion about the loaded roll. When the
  printer cannot identify its stock, trust that assertion; no recurring media
  confirmation or detected-media requirement is part of printing. Explain this
  once at setup and when editing size. Actual hardware-reported errors still
  stop printing normally; never claim to have verified an unobservable roll.
- A template defines appearance: arrangement of QR/title/reference, typography,
  wrapping, and constrained display options. It is independent of the printer's
  registered size. One printer supports multiple compatible templates; the same
  layout rules can render different images for different registered sizes.
- Stuff Stash's API-side rendering service owns composition through a rendering
  port. Its output is used for client previews, downloads, and queued printing.
  The printer adapter owns USB/protocol encoding and status; it cannot redesign
  the label, substitute text/QR content, crop, or scale-to-fit silently.
- Rendering inputs are a typed content snapshot (canonical QR URL, title,
  human-readable reference), template ID/version/options, and the destination's
  configured media snapshot. Never supply arbitrary HTML, scripts, fonts,
  external URLs, or executable printer commands as template options.
- First release ships versioned built-in templates `qr-title` (default, QR beside
  title) and `qr-only` (QR with optional human-readable reference). Both support
  `showReference`, initially true. Layout adapts to the printable bounds rather
  than stretching a fixed image. Arbitrary fonts, colors, coordinates, and a
  visual template editor are deferred.
- Users can select template/options per print/download. Inventory defaults keep
  template/version/options separate from the default printer; selecting a new
  template never requires a new registration or connector restart. No custom
  template CRUD or uploaded template execution is part of the first release.
- Compatibility uses template/options, actual content, configured media, and
  adapter capabilities. Reject a too-dense QR or insufficient printable area with
  guidance to choose a simpler template or configure different supported media;
  do not silently change the destination, template, size, or content.
- Defaults affect new requests only. Jobs pin normalized options, template
  version, media snapshot, and immutable rendered artifact. Disabled catalog
  versions cannot be selected for new jobs; existing compatible jobs keep their
  artifacts. Unsupported consumers must reject before output, not regenerate.
- Reprint deliberately uses current asset content and the selected printer's
  current settings, creates a new job, and links its predecessor. Show a new
  preview when content/layout/size differs; exact historic reprint is deferred.
- A download/system-print request may specify standalone output dimensions when
  no registered printer is used. This does not create a printer or media catalog
  and does not allow per-job overrides of a registered printer's configured size.

## Label Rendering

- Use a project-owned label rendering port with typed content, template selection,
  and media profile. Vendor commands do not enter its public input model.
- Default layout is black on white, QR beside the asset title. QR identity stays
  fixed; each new job captures the current title and selected layout.
- Render QR modules at integer pixel sizes without interpolation, preserve a
  quiet zone of at least four modules, and enforce profile-specific readability.
  Reject content that cannot fit a scannable QR; never crop it to fit.
- Long titles may wrap and truncate with a visible indication. The initial renderer
  uses Go Regular from the already pinned `golang.org/x/image v0.41.0`, with its
  redistribution license. Normalize text to NFC and support its Latin, Greek,
  and Cyrillic glyph coverage. Reject unsupported glyphs and scripts requiring
  shaping with a recoverable rendering error; never silently print replacement
  boxes or misleading unshaped text. Broader fonts/shaping require a reviewed
  renderer update. Text must not overlap the QR.
- The Go renderer uses `github.com/boombuler/barcode v1.1.0` for QR encoding;
  independent decoding acceptance uses `github.com/makiuchi-d/gozxing v0.1.1`.
  Draw the returned QR matrix at integer scale with an explicit four-module
  quiet zone. Rasterize text with the existing x/image font adapter.
- A fixed image-only PDF adapter wraps the production raster with explicit
  physical page size, printable margins, and disabled image interpolation.
  Validate each printable physical axis against raster dots and declared DPI,
  allowing at most one dot of physical rounding error per axis. Reject inconsistent
  settings rather than stretching square QR modules.
  It performs no separate text/QR composition and includes no active PDF content.
  Keep it deterministic; validate its output using an independent PDF renderer.
- The initial offline `label-catalog` command accepts versioned media descriptors
  exported by printer adapters and emits template metadata and synthetic fixtures
  through the production renderer. It does not register printers or contact the
  API. Candidate fixtures do not establish hardware acceptance.
- The initial Brother profile uses 29 x 90 mm media and the existing script's
  306 x 991 pixel print raster orientation. These are separate physical and
  printable dimensions, not a general pixel-to-mm conversion rule. Verify the
  actual adapter's rotation, margins, resolution, cutting, and QR readability on
  the physical device before accepting this profile.
- Provide downloadable PNG and physically sized PDF, plus browser/system print
  presentation. Document actual-size printing; browser scaling must not silently
  invalidate the media dimensions. No connector is required for this path.
- Download/render access requires `inventory.view`; queued physical printing
  requires `inventory.edit_asset`. Rendering is a safe read, audited accordingly.
  Download or opening a print dialog is not proof of physical completion.
- Persist canonical label identity before issuing an artifact. First-time
  provisioning is an authorized, idempotent command with audit, not a GET side
  effect; viewers may provision a label for an asset they can view, but cannot
  change its mapping or submit a physical print job.

## Inventory Settings And Creation

- Persist inventory label settings: default registered printer (optional),
  independent template/version/options, and `printOnCreateDefault` (initially
  false). Changes require `inventory.configure`; reads require `inventory.view`.
- Enabling the default requires an active configured destination and compatible
  template/media combination, but does not require the printer to be currently online.
- Add forms for every asset kind expose Print label after saving. Use a web
  checkbox and native mobile switch, initialized once from the inventory default
  for each new draft. Refetching settings must not overwrite a user's choice.
- Show destination and any current readiness warning inline. With no configured
  destination, explain setup; allow normal creation with printing off. Do not
  silently turn an explicit print request into an ordinary create.
- The create command carries explicit print intent, destination, expected media
  fingerprint, template/version, normalized options, and copies (one by default). The API resolves media from
  the printer registration; a stale expected fingerprint returns a conflict.
  Persist asset creation, label identity, print job, and audit atomically through
  an application unit-of-work port. A lost response/retry must not create another
  asset or job; use a scoped idempotency key and reject changed-payload reuse.
- Validate authorization, destination ownership, registration, and compatibility
  before commit. Invalid requested printing returns a recoverable validation
  error preserving the draft; the user can choose a destination or turn it off.
- Once committed, printer disconnection, unavailable media, photo-upload failure,
  or print failure must not undo asset creation. Report Saved; label queued and
  link to the job. The user may leave immediately without losing the request.
- Inventory defaults initialize interactive drafts only. Imports, background
  actions, and existing conversational tools do not silently acquire auto-print
  side effects. CLI creation opts in explicitly. A future conversational print
  tool must preserve normal approval and authorization boundaries.

## Client Interactions

- Reuse the asset overflow menu for Print label on item, container, and location
  detail. With a compatible default, this issues one command with one copy and
  shows its job status without a mandatory confirmation step.
- Label options is a bounded task with preview, destination with configured size
  displayed read-only, template/options, copy count, download/system print, and
  explicit Print/Cancel. Use native task presentation on mobile and accessible web controls. No nested modal stacks.
- Changing the template, options, content, or media invalidates the prior preview.
  Show the API-rendered preview of the selected combination; submit its selection
  revision/content fingerprint so a stale preview cannot silently print changed
  asset data. A one-tap default print renders current data without requiring a
  preview step. Browser/system printing uses the same composition, with a
  physically sized PDF wrapper where needed.
- Bound copy count and payload sizes through validated server configuration;
  disclose copy count and keep one copy as the default.
- Printers in inventory settings is a normal navigation destination. Users with
  view access can see registrations, capabilities, current readiness, last seen,
  and job status. Configure actions are permission-gated. Secret credentials and
  local filesystem/device paths never appear in ordinary client responses.
- Show connector online/offline separately from printer ready/unavailable/error/
  unknown. A stale report is unknown, never confidently ready.
- Typical recovery copy: Connector connected; printer unavailable. Check power
  and USB connection. Three labels waiting. Only name a cause such as out of
  labels or cover open when the adapter has evidence for it.
- Queued/waiting jobs show destination and cancellation; uncertain jobs explain
  possible prior printing and offer explicit reprint. Preserve completed/failed
  history independently from current printer availability.
- Add Scan label to Browse/search actions, without a new primary navigation tab.
  Camera scanning is a bounded native task and an HTTPS camera interaction on
  web. Request permission on entry; handle denial, no camera, unreadable code,
  unavailable server, and permission loss. Offer paste-label-link fallback.
- Debounce repeated frames; resolve once, dismiss the scanner, and navigate to
  the existing asset detail route with normal back navigation. Never stack asset
  navigation inside a camera modal or discard an unrelated unsaved draft.
- In focused job views poll the API initially; stop on background/unmount and
  terminal state (including a successful cancel or resolution), back off on failure,
  and refresh on return. Native job reads start at five seconds; consecutive failures
  double the delay up to sixty seconds, and a successful read resets the delay. No mandatory
  WebSocket/SSE delivery or direct CLI-to-client connection is needed.
- Accessible status text must accompany color; announce meaningful transitions
  without announcing every poll. Preserve focus, draft choices, large text,
  keyboard operation on web, and native navigation semantics.

## Interaction Grounding

These are proposed extensions of the existing product, not observed native UI.
Use `AssetOverflowMenu`/`NativeActionMenu` for secondary commands, existing Add
forms for an inline choice, and the Settings grouped-row hierarchy for printer
management. Inspect all consumers before extending shared components. The task
fits those existing hosts; it does not justify a new primary tab or extra modal.
The repository's reference-layout reset (Files secondary actions, Settings choice
rows, and focused commit/cancel forms) governs placement and emphasis. Printer
selection with descriptive readiness belongs in a selection task when it exceeds
a short native menu. Native visual acceptance must validate the resulting flow.
All new user-facing strings, plural counts, errors, and accessibility labels use
the existing localization infrastructure rather than hard-coded English.

## Acceptance Evidence Required

- Tests first for stable identity, idempotent provision/create/print, all three
  asset kinds, renamed/moved/archived/deleted assets, restores and foreign clones.
- Adversarial boundary tests for unauthenticated and unauthorized scans/renders,
  cross-tenant/inventory IDs, malicious QR origins, redirects, callback replay,
  expired sessions, and permission loss while resolving or submitting a job.
- Verify successful authorized viewer scans/downloads, editor printing, and
  administrator configuration through real HTTP and client boundaries.
- Verify print default initialization/override, offline queueing, missing/default
  printer changes, draft preservation, duplicate submissions, and photo failure.
- Test one printer with both built-in templates, one template across compatible
  registered sizes, independent defaults, version pinning, stale-preview rejection,
  manual media setup without detection, size-change/start races, switching back
  to an old size, queued mismatch recovery, and adapter preservation of layout.
- Test QR decoding of rendered artifacts, long Unicode titles, physical sizing,
  orientation, and profile limits. These do not replace physical print/scan tests.
- Verify web camera and paste fallback; named iOS/Android builds on native
  runtimes for permission, scan, cancellation, navigation, and sign-in return.
  Verify physical QL-800 labels using both in-app and ordinary phone cameras,
  including domain migration and app-installed/uninstalled cases.

## Research And Limits

- [Homebox label implementation](https://github.com/sysadminsmedia/homebox/blob/main/backend/pkgs/labelmaker/labelmaker.go)
  provides browser/download/server-command precedent; no implementation is copied.
- [Apple domain associations](https://developer.apple.com/documentation/xcode/supporting-associated-domains)
  and [Android App Links](https://developer.android.com/training/app-links/about)
  constrain automatic app opening.
- Existing script inspected read-only on Paul:
  `~/code/homebox-label-printer/brother-ql800/auto-label/main.py` and `README.md`.
  It uses brother_ql/PyUSB, QL-800, USB vendor/product 04f9:209b, and 29x90 media.
  Script settings are evidence of prior configuration, not current roll detection
  or physical acceptance. Do not copy embedded credentials or disabled TLS checks.

## Generated Template Documentation

Template registrations and rendering behavior feed the public catalog and PNG
examples defined by [generated printing docs](../platform/printing-catalog-docs.spec.md).
Use the actual renderer with fixed synthetic fixtures for each supported
template/media combination; no independent documentation layout implementation.
Changes to templates, presets, or renderer behavior update generated docs/images
in the same PR and must pass generation/drift checks.

## Initial Label API Delivery Contract

- `stuff-stash labels bootstrap-instance` explicitly initializes the persisted
  singleton identity after database migration. Repeating it returns the same ID;
  competing processes cannot replace it. Before bootstrap, `/instance` and label
  operations return service-unavailable without creating identity as a read side
  effect. The operation is a local deployment command, not a public HTTP mutation.
- `STUFF_STASH_LABEL_BASE_URL` defaults to `STUFF_STASH_PUBLIC_WEB_BASE_URL`.
  Empty configuration disables label operations; configured values must be valid
  HTTPS bases with no credentials, query, fragment, ambiguous escaped separators,
  or dot segments. Preserve a configured path prefix. Never derive it from Host.
- The initial download slice accepts a standalone media snapshot in
  `POST /assets/{assetId}/label-renders`, plus explicit template/version/options
  and PNG/PDF format. It requires a previously provisioned canonical label.
  Printer-backed requests are added with registered printer resources; they may
  not accept a caller override of registered dimensions.
- Responses expose a scoped render ID, content fingerprint, expiry, media
  fingerprint, content type, digest, and authenticated relative content path.
  Content reads live under the tenant/inventory path and reauthorize the current
  principal, active tenant/inventory, asset existence, and expiry on every read.
- Short-lived preview/download artifacts use a dedicated repository port storing
  immutable bytes and metadata in the database, separately from future durable
  job artifacts. `STUFF_STASH_LABEL_RENDER_TTL` and
  `STUFF_STASH_LABEL_RENDER_MAX_BYTES` bound lifetime and artifact size; periodic
  cleanup removes expired previews. No public blob URL or bearer download token
  is issued. Asset deletion atomically tombstones identity and removes previews.
- Canonical provisioning is naturally idempotent per tenant/inventory/asset and
  records `label.provisioned` only on first creation. Label metadata/resolve,
  template listing, rendering, and content access have safe audit actions against
  the existing asset or inventory target; content, QR URLs, and titles never enter
  audit metadata. Resolver failures for absent, foreign, tombstoned, and forbidden
  identities return the same safe 404. Existing HTTP rate limiting applies.

### Inventory Print Settings Contract

- `GET` and full-replacement `PUT /tenants/{tenantId}/inventories/{inventoryId}/print-settings`
  return the shared success envelope with `revision`, nullable `defaultPrinterId`,
  `template: {id, version, options: {showReference}}`, and `printOnCreateDefault`.
  A missing row is a virtual revision-0 default: no printer, `qr-title` version 1,
  reference shown, and automatic printing off. GET does not create settings.
- PUT supplies the revision it read; initial creation requires 0 and subsequent
  changes require the current revision. A conflict never overwrites another
  administrator's change. Settings and their audit record commit atomically.
- Inventory viewers may read; only inventory configurators may write. Explicit
  tenant/inventory scope applies to settings and printer lookup. Connector
  credentials cannot read or change these human settings.
- A selected printer must belong to this inventory and be active. Automatic
  printing requires a selected printer; clearing the printer requires automatic
  printing off. Disabling automatic printing preserves independent template
  selection, including when no destination is selected. Online/readiness state
  does not constrain settings changes.
- Template ID/version must exist in the built-in catalog. For a selected printer,
  the real renderer validates compatibility using a synthetic title/reference and
  a realistic instance/label URL under the configured public base URL. This does
  not create label identity or artifacts and never prints. Actual asset content
  is still validated independently when a job is requested.
- The selected printer's active state and media/revision snapshot are rechecked
  atomically when settings are saved. A concurrent retirement or media change
  produces conflict. Later printer edits do not rewrite settings or queued jobs;
  each new request revalidates its current destination.
- A dedicated inventory print-settings repository owns persistence; this is not
  mutable state inside the printer aggregate. Safe read history uses
  `print_settings.viewed`; successful replacement uses `print_settings.updated`.

### Atomic Create-And-Print Boundary

The existing asset-create endpoint accepts an optional explicit print selection
and a scoped idempotency key. A coordinator prepares asset validation, optional
parent promotion and tags, label identity, immutable rendered content, and their
audit records before invoking a dedicated unit-of-work port. Neither rendering
nor preparation writes the new asset, label, or job.

The persistence adapter locks the selected registered printer, rechecks the
registration revision/media and scoped request key, and commits the prepared
asset, parent promotion, tag assignments, label, job, undo operation, and audit
records together. A failure rolls back every change. Concurrent equivalent
requests return the original asset/job; changed payloads conflict. The stored
job retains the asset-creation operation reference for equivalent response
recovery. Request fingerprints cover the full asset draft and print selection,
not generated identifiers or request-correlation IDs. A retry remains subject
to current inventory authorization and cannot recreate a deleted result.
## Client Download And Scan Delivery Slice

- Before printer-job UI ships, asset overflow exposes Label options for authorized
  viewers on all three asset kinds. The bounded web task sheet and native task
  route show the current API-rendered PNG, built-in template choice, reference
  toggle, configured catalog size, PNG/PDF export, and system printing. No
  physical completion is inferred from downloading or presenting a print dialog.
- Standalone download media comes from the authenticated printer-profile catalog,
  never a second hard-coded client copy of physical profile values. The initial
  single 29 x 90 mm option is displayed read-only. Templates remain independent.
- Preview requests are cancelled/superseded when selection or asset scope changes;
  stale responses cannot enable export. Binary content uses generated authenticated
  transport with redirect rejection and an explicitly scoped render identifier.
- Web scanning uses HTTPS camera capture with a pinned `jsqr` 1.4.0 decoding
  adapter, with manual pasted-link fallback when camera access fails. Decode is
  local, bounded to camera frames, and never downloads scanned origins.
- Native camera scanning and system printing use Expo SDK 55 bundled versions
  `expo-camera` 55.0.18 and `expo-print` 55.0.15 behind adapters. Reuse existing
  file-system/sharing for temporary label documents. Camera permission is requested
  when entering scanning; denial retains the paste fallback.
- A shared project-owned label protocol parser validates HTTPS/custom-scheme URLs
  locally, supports deployment path prefixes, rejects escaped/ambiguous paths,
  credentials/query/fragment, and extracts only opaque instance and label IDs.
  The generated API client always resolves against its configured API destination.
- Keep the pending label reference outside native authentication/server composition;
  re-check instance identity and authorization after either changes. Web sign-in
  retains the local label landing route through the existing validated return path.
  Existing invitation and OIDC links remain independent.
- Browse/search owns the scan entry, without another primary tab. A successful
  scan closes its bounded task before navigating to existing scoped asset detail;
  archived detail navigation uses the resolver state instead of active-list lookup.
  Extend native retained-action callback tests only where shared header semantics
  change. Add adversarial parser/transport and navigation-race tests; runtime camera,
  native sharing/printing, and physical QR checks remain explicit device checklist
  items until actually observed.

### Native Atomic Creation Delivery

The Add item draft initializes its print switch once from scoped inventory defaults
and persists the user's choice through navigation and metadata refresh. Loading
label settings does not overwrite an explicit choice. If loading fails, show Retry
and an explicit Add without a label choice; never silently drop a requested label.

When enabled, creation captures the destination's media fingerprint, independent
template, one copy, current inventory and a random request key. The mobile draft
retains that complete create input before submission and freezes edits while its
outcome is ambiguous. Retry Save sends that same intent, including previously
prepared tag IDs, rather than creating tags/assets/jobs again. An initial definite
rejection unlocks correction; a definite rejection after any ambiguous attempt
does not erase the original intent. Scope changes cannot submit the captured
request against a different selected inventory. Success clears the request and
exposes View print job; parent creation never inherits the item print switch.

The API client accepts `createAsset(tenantId, inventoryId, input, idempotencyKey?)`;
`input.printLabel` comes from the generated atomic selection contract and the
mapped result retains optional `printJobId`. No extra enqueue call is made.

Every pending-tag write that prepares a create-and-print request carries its
captured tenant/inventory scope. The mobile adapter checks that scope against the
current selection before each write and uses the captured IDs in the API request.
Changing inventory while an earlier tag write is in flight may finish that write
in its original inventory, but no later tag or asset may target the replacement
inventory. The final atomic create retains its own scope fence.

### Atomic creation before label bootstrap

Create-with-print must return HTTP 503 when instance identity has not been
bootstrapped, with no asset or job committed. Authentication and inventory
creation authorization still run first: anonymous and unrelated principals must
not learn instance readiness by bypassing access checks. After the operator
bootstraps the identity, the same authorized create command can succeed.
