# Expiration Notifications

## Scope

Provide a durable personal notification inbox on web and mobile plus mobile push delivery for expiration milestones. Web does not request system notification permission or send browser push. Shared asset dates are defined in `../expiration/expiration-tracking.spec.md`. This is a reusable notification bounded context, with expiration as its first producer.

## Personal Preferences And Inheritance

- Preferences are owned by authenticated principal, tenant and inventory. Never accept a caller-selected recipient identity on personal settings APIs.
- Inventory defaults: expiration reminders enabled, upcoming enabled, expired enabled, advance warning 30 calendar days, and a validated IANA timezone supplied from the user's device on initial settings setup. Server fallback is UTC and must be visible/editable in settings.
- Each available custom asset type inherits defaults unless the user creates an explicit override. An override specifies enabled, upcoming, expired and advance-warning days. Reset to inventory defaults removes the override. Days accept integers 0 through 3650.
- The inventory enabled switch is an inheritable default, not an absolute mute: a type override may enable reminders while the inventory default is disabled. UI must say so. Other members' preferences remain independent.
- Mobile push is an independent personal delivery preference. Denying device permission or disabling push never removes enabled inbox milestones. Upcoming/expired switches control creation of those milestone notifications in both channels.
- Preferences are available to principals with inventory view access; configuration roles are not required for personal choices. Type capability configuration retains its existing stronger permission.

## Milestones And Lifecycle

- One upcoming milestone when the item enters the configured window and one expired milestone after the recorded date ends. Already-expired items generate only the current expired milestone, not a burst of historical upcoming alerts.
- Milestone identity includes principal, tenant, inventory, asset, date value/precision and milestone kind. Durable uniqueness prevents duplicates across workers, retries, restarts, preference toggles and read-state changes.
- Changing or clearing a date, archiving/removing the asset, or disabling its type withdraws obsolete active inbox entries and pending deliveries. Revalidate current state and permission at delivery and read time. Delivered operating-system notifications cannot be recalled reliably; opening one rechecks current item state.
- Restoring/re-enabling the same date does not resend an already delivered milestone. Changing reminder threshold does not resend an existing upcoming milestone for the same date.
- Archived inventories, inactive types and principals without current view permission cannot produce or expose notifications. Membership loss prevents pending delivery, badge counts, inbox reads and deep-link access.
- Disabling personal reminders or a milestone cancels its pending push deliveries but retains previously generated inbox history and read state. It suppresses new matching inbox entries. Re-enabling does not resend previously generated milestones. Type/lifecycle/date withdrawal remains distinct from personal delivery preferences.
- Read status is personal, persisted and synchronized across clients; reading does not change expiration, delete the asset, or resolve its expired status. Inbox supports unread/all views, mark read and mark all read within the authorized inventory.
- Bounded cursor pagination and unread count use the same visibility rules. Loading/error states never masquerade as an empty inbox.

## Push And Background Work

- Notifications must be generated server-side without requiring either client to remain open. A configured bounded worker discovers due expiration milestones through scoped repositories and writes inbox/outbox state transactionally.
- Clock, repositories, authorization, observation and push delivery are injected ports. GORM and provider SDK details remain in adapters. Operational queue discovery may span scopes; every candidate carries tenant, inventory and recipient identity and is reauthorized before publication/delivery.
- Outbox claims are leased and fenced; retries use bounded backoff. Delivery failures preserve the inbox. Permanently invalid device registrations are retired. Do not promise exactly-once OS display after ambiguous network outcomes.
- Device registration is authenticated, bound to the current user/server environment and revocable on sign-out. Re-registration cannot expose another principal's inbox or let an untrusted request choose an arbitrary push URL. Tokens and credentials must never appear in logs, audit metadata, notification URLs or source control.
- Push payloads use opaque scoped notification identifiers and generic lock-screen wording; the authenticated app fetches current content and opens the item. This avoids disclosing medicine names on lock screens by default.
- Provider credentials, project/app identifiers, worker cadence and delivery configuration come from validated environment configuration. Pin any new client dependencies and provider integration versions. Document self-host configuration and ensure release signing supports push before claiming mobile delivery complete.

## Interface

- Place a bell with unread badge in the existing inventory shell on both platforms. Accessible labels announce unread count; color is not the sole indicator.
- Web opens an inbox panel with keyboard/focus management. Mobile opens a native screen/sheet integrated with existing navigation. Both list asset title, expiration status/date and existing location breadcrumb; long breadcrumbs reveal the most specific location by default.
- Selecting an entry marks it read and navigates through existing asset detail controls. Deleted/inaccessible items show a clear unavailable state without leaking details.
- Personal inventory notification settings expose defaults and an asset-type override list, each explicitly labeled Inherited or Custom. Push permission denial links to mobile system settings. Native permission prompts occur only after an explicit enable action.
- Use existing shared controls, loading/error patterns, semantic colors, dynamic type, reduced motion and minimum touch targets. No separate visual theme or new grouping of inventory assets.

## Acceptance

Write failing fake-backed tests before code for inheritance, separate principals/inventories/types, calendar windows, deduplication, concurrent workers, retries/invalid devices, lifecycle cancellation, timezone changes, authorization loss, and pagination/badge consistency. Adversarial HTTP tests cover unauthenticated, expired/malformed token, cross-tenant, cross-inventory, wrong-role configuration, recipient spoofing and legitimate viewer preferences. Browser/mobile tests cover notification navigation, read synchronization, permission denial, inherited overrides and accessible states. Use CI/remote validation and required code critic review. Deploy through GitOps and verify TestFlight upload; separately report any device push delivery evidence that cannot be obtained.

## Personal Settings API And Registration

The authenticated client initializes its personal inventory notification settings when first opening an inventory, using `POST /tenants/{tenantId}/inventories/{inventoryId}/notification-preferences/initialize` with its IANA timezone. Initialization is idempotent and never overwrites existing settings. It registers the recipient for background evaluation even when push is disabled. Clients must not rely on remaining open after registration. `GET` at `/notification-preferences` returns saved settings or visible UTC defaults without mutating state. `PUT` replaces the user's inventory defaults (enabled/upcoming/expired/advanceDays/timezone/pushEnabled). Asset-type overrides use `PUT` and `DELETE` at `/notification-preferences/types/{customAssetTypeId}`; overrides contain only the four expiration settings and inherit timezone and push delivery from the inventory settings. GET includes overrides with their type IDs. All writes are atomic with audit, and unchanged writes avoid duplicate audit. Missing/archived inventory or unavailable/archived override types are rejected. Notifications use the server's authenticated principal identity, never a body/query principal ID. Shared API envelopes and generated client types apply.

Settings repositories require tenant, inventory and principal scope on every user-facing read/write. Operational background enumeration pages through registered recipients by a stable cursor, with each returned record carrying full scope. Background authorization rechecks current membership, so stale registrations cannot authorize notifications. Settings entries are retained across access loss but inaccessible until authorization returns. Personal settings changes use optimistic revision numbers to reject stale replacements and prevent simultaneous device edits from silently undoing a newer preference. Initialization returns the existing revision; subsequent default/type changes require that revision and advance it once atomically.

The API runtime embeds the Go toolchain's pinned IANA timezone database so personal calendar evaluation does not depend on timezone files installed in the container image.

## Implementation Evidence

The personal settings backend provides scoped persistence, optimistic revision updates, initialization and type overrides through generated REST contracts. Remote tests cover legitimate viewer settings, user and tenant isolation, unavailable types, invalid timezones, stale writes, inventory-default overrides and atomic audit rollback in memory and GORM. This is foundation work; inbox generation, delivery and client surfaces are still required before release.

## Milestone Evaluation Contract

Expiration evaluation receives an immutable candidate snapshot (asset and type IDs, date and precision, and active/expiration-enabled eligibility), personal settings, an injected current instant and the resolved personal timezone. It emits at most the currently due milestone: upcoming while inside the warning window, otherwise expired after the date ends. Disabled preferences produce none. An upcoming entry remains meaningful after expiration, but cannot be newly generated after expiration. Moving a date, removing eligibility or moving an upcoming date outside its current warning window withdraws that entry. Existing history remains visible when only personal enabled/upcoming/expired switches change. The persistent unique identity is scoped to recipient, tenant, inventory, asset, original date/precision and kind; never include mutable threshold or timezone in identity. This separates lifecycle visibility from generation and push eligibility.

## Inbox Persistence Contract

The inbox stores immutable notification ID, full recipient scope, asset ID, expiration value/precision, milestone kind and creation time, plus optional read time. A unique scoped milestone key prevents duplicate insertion even when a retried producer proposes a different notification ID. Inserting a new milestone and its audit is atomic; duplicate milestone insertion returns the original record without changing read state or adding audit history. Read mutations require full scope, are idempotent, and commit audit atomically only when transitioning from unread. An unknown or other user's notification ID is never modified. User-facing repository reads require tenant, inventory and principal. Raw repository pages use a stable descending notification-ID cursor and bounded page size; application reads must revalidate current scope and candidate visibility before returning content, counts or links. Records retain only IDs and expiration metadata, avoiding stale asset titles or paths. Asset deletion may leave a deduplication tombstone; those records must never become visible without a current eligible asset.

## Notification Item Access

Application notification item reads and mark-read commands first authorize the current inventory, then load the notification by the full personal scope. They resolve the current asset and its currently assigned custom type within that tenant/inventory, validate active lifecycle and expiration capability, and apply current personal timezone/window visibility. Missing, foreign and obsolete notification IDs return unavailable. Marking an unavailable entry read is rejected. Successful item reads produce safe read audit; successful first read-state changes produce notification.read audit atomically. Detail returns the current asset projection for existing detail navigation, never a cached title or client-selected asset ID.

## Inbox Pagination

Application inbox pages accept a descending notification-ID cursor, a limit from 1 through 100, and an unread-only filter. Each request scans at most 200 stored candidates, checks current item visibility and returns up to the requested number. The continuation cursor advances over every examined candidate, including withdrawn/read entries. A page may be empty with a continuation cursor when stored history is sparse; clients continue fetching and must not present this as an empty inbox until the cursor is exhausted. Candidate exhaustion ends pagination. Errors other than a genuinely unavailable notification fail the request rather than silently hiding entries. Listing emits one inventory-scoped read audit, not one audit per candidate.

## Recipient Generation Pages

The background application command evaluates one registered recipient and at most 100 active assets per call, using stable ascending asset-ID pagination. It requires current inventory view access and an existing personal registration, reads current settings and the injected clock, resolves each asset's active expiration-enabled type, and inserts only the currently due milestone with audit. It returns created-count and the next asset cursor. Deduplicated milestones do not count as new or emit another creation event. A failed page is safe to replay because milestone identity is durable; completed inserts are retained. The background coordinator must continue all recipient and asset cursors and restart full evaluation on later sweeps. No client connection is needed. Generation errors propagate to the worker for observable retry rather than being treated as successful empty results.

## Background Scheduler

The API starts expiration generation automatically when `STUFF_STASH_NOTIFICATION_WORKER_ENABLED` is true (default true). `STUFF_STASH_NOTIFICATION_POLL_INTERVAL` (default 5s, minimum 100ms), `STUFF_STASH_NOTIFICATION_PAGE_SIZE` (default 100, range 1–100), and `STUFF_STASH_NOTIFICATION_PAGE_TIMEOUT` (default 30s, minimum 1s) bound its cadence and work. Invalid values fail startup. One tick evaluates one recipient asset page; cursors carry the last completed recipient and the active recipient's asset position. Recipient removal/change resets the asset cursor to avoid skipping another recipient's assets. A completed sweep resets its cursor for the next cycle. Denied or missing recipients are skipped; infrastructure errors preserve the cursor and emit a domain worker-failure event for retry. Process restarts begin a fresh sweep; durable uniqueness makes this safe. The worker starts immediately, has no overlapping calls, honors cancellation and joins before repositories close. Push remains a separate delivery worker.

## Inbox REST API

Authenticated `GET /tenants/{tenantId}/inventories/{inventoryId}/notifications` lists personal visible alerts, with `limit` (default 30, maximum 100), optional `cursor` and `unreadOnly`. Shared pagination metadata carries the continuation cursor, including empty continued pages. `GET .../notifications/{notificationId}` returns a currently visible personal alert; `PUT .../notifications/{notificationId}/read` idempotently marks it read. Entries include notification ID, current asset ID/title/parent/type, date/precision, milestone, creation time and optional read time. Selecting an entry uses the current asset ID through normal asset detail navigation. No endpoint accepts a recipient identity. Standard authentication, scoped errors, read audit, and generated contracts apply. Inaccessible or stale item IDs return not found; inventory access denial remains forbidden.


## Shared Client Transport

The shared generated-contract client exposes a focused notifications subclient for preference initialization/read/update, setting/removing type overrides, paginated inbox reads, detail and mark-read. It reuses the existing bearer-token, cancellation and safe API error handling. Revision values and explicit false settings must pass through unchanged. Inbox pagination preserves an empty page with a continuation cursor; callers must not infer exhaustion from item count. Platform adapters keep their own domain models above this transport layer.

### Web notification repository boundary

The web frontend owns notification, reminder-policy, and preference domain models. A dedicated inventory-scoped notification repository exposes initialization, preference updates, type override replacement/removal, paginated inbox reads, notification resolution, and read marking. Its API adapter maps transport date/precision fields into the frontend expiration value and preserves cancellation, optimistic revisions, and empty continued pages. Inbox records retain their asset identity for navigation; the server revalidates visibility when a record is resolved.

### Reminder policy editor

Use one reusable web editor for inventory defaults and type overrides. Inventory defaults explain that explicit type overrides can remain enabled. Type editors offer “Use inventory defaults”; inherited controls display the current defaults without allowing edits until an override is selected. Each form saves explicitly, validates whole-number advance days from 0 through 3650, disables controls while saving, and retains the draft on failure (including revision conflicts). A failed save must not announce success. Resetting inheritance submits removal of the override. Remount the editor when its inventory/type identity changes; parent orchestration supplies current policy and revision-aware save callbacks.

### Personal settings application session

Each web settings surface owns an inventory-scoped preference session. Initialization registers the current principal using the device timezone without replacing existing settings. Saves use the last successfully loaded revision and preserve unrelated preferences: changing defaults retains timezone/push, changing timezone retains defaults/push, and type override operations retain other types. Allow only one pending write per session. A failed write leaves its loaded snapshot unchanged; the UI retains its independent draft. Explicit refresh obtains a newer revision after a conflict without silently resubmitting a stale draft. Application observations describe load/save success and failure without preference values or user content.

### Web settings surface

The inventory notification settings surface loads/initializes personal settings with an explicit loading and retry state. It shows inventory defaults, a validated editable IANA timezone, and individual expiration-enabled active type editors. All editing controls are disabled during a save. A “Refresh saved settings” action reloads the revision and current inherited defaults while preserving open policy and timezone drafts; users explicitly save retained drafts afterward. The surface never requests browser notification permission. It explains that mobile push is configured in the mobile app. Mount a new surface when principal/server/tenant/inventory identity changes.

### Web settings navigation and composition

Expose Notifications under inventory settings at `/settings/tenants/{tenantId}/inventories/{inventoryId}/notifications`, available to inventory viewers. It is an inventory-only collection without resource or lifecycle subroutes. The authenticated composition root supplies a dedicated notification repository through workspace context. Key the settings surface by API identity, principal, tenant and inventory so pending drafts never cross those boundaries. Load the complete active type collection, including inherited types, through the inventory customization port; surface failures with retry rather than treating them as an empty list.

### Web inbox application reads

Loading a visible inbox page follows empty continued pages until a visible page or actual exhaustion. Detect repeated/missing continuation cursors and bound traversal to 100 pages; incomplete traversal is an error, never an empty-inbox claim. Pass the unread filter and cancellation signal on every request. Opening an alert resolves it through the current notification detail endpoint, marks that notification read, then returns the resolved asset ID for ordinary navigation. If resolution or marking fails, do not navigate using a cached asset ID. Application observations record operation outcomes without alert titles or dates.

### Web inbox list view

The reusable inbox view presents All and Unread filters, explicit initial/append loading and retry states, and paginated alerts with readable expiration dates and unread text. Month precision is displayed without a day; calendar dates are formatted without shifting across timezones. Refresh/filter changes cancel obsolete loads and discard their results. Appending merges IDs without duplicate rows. Opening an alert disables duplicate activation, uses the verified application open flow, and reports unavailable/failure without navigating. Unmount cancels reads/opening and prevents late navigation. The shell supplies normal asset navigation and the containing accessible panel.

### Bounded unread summary and mark-all work

Unread counts and mark-all-read traverse the same currently visible personal inbox pages as listing. Each service call processes at most one bounded page (100 visible records, at most 200 stored candidates), returning a continuation cursor even for zero visible records. Count results are page contributions; clients accumulate until exhaustion before presenting a complete count. Mark-all clients continue until exhaustion and recheck the count afterward. This is a live traversal, not a transaction-wide snapshot; newly arriving alerts may remain unread and appear on the next refresh. Every read mutation is idempotent and audited atomically. A notification withdrawn between listing and marking is skipped; authorization/infrastructure failures abort the page, whose completed changes may safely be retried. No partial traversal may be presented as completion.

### Inbox batch REST contracts

`GET .../notifications/unread-count` returns `{count}` for the current bounded page. `PUT .../notifications/read-all` marks the current bounded page read and returns `{complete}`. Both accept an optional `cursor` and include shared pagination metadata; complete is true only when no continuation remains. Count is explicitly a page contribution. Both operate on the authenticated principal alone and require current inventory view access. The static paths take precedence over notification ID routes. Clients traverse all continuations; errors preserve prior read changes for safe retry. Read audit and per-notification mutation audit follow the existing application behavior.

### Web batch traversal

The web notification port exposes count and mark-all page operations. Application helpers accumulate count contributions or perform read-page mutations until cursor exhaustion, checking cancellation and rejecting repeated cursors or inconsistent completion flags. Traversal is bounded to 100 pages; reaching that limit is an explicit incomplete-operation error. Never present an incomplete count as zero or partial mark-all as success. Callers refresh the inbox/count after successful mark-all, and may safely retry after failure. Observation records only operation outcomes.

### Web bell and inventory registration

The inventory header mounts a notification bell keyed by API identity, principal, tenant and inventory. It initializes personal reminder registration with the device timezone and then loads a complete unread count. Refresh on opening the panel, window focus, successful read mutations and every 30 seconds while visible; stop timers and abort requests on unmount. Unknown/error counts have an accessible unavailable label instead of a zero badge. The existing task sheet supplies modal focus, Escape dismissal and scrolling; closing restores focus to the bell, while item/settings navigation uses the existing route handler. Inbox errors remain visible and retryable. Web never requests push permission.

### Native notification repository boundary

Mobile owns its notification and personal-preference models and an inventory-scoped application repository, separate from generated transport types. The API adapter maps date precision into the native expiration value, normalizes absent override collections, and preserves revisions, explicit false values, pagination and cancellation. Convert API failures into safe typed native failures for expired authentication, denied access, unavailable records, revision conflicts, validation and temporary service failure. Never expose raw provider/server error text. Cancellation remains cancellation rather than a visible failure.

### Native inbox queries

Native application queries keep explicit tenant/inventory scope and use the notification repository for sparse inbox traversal, complete unread counts, mark-all pagination and resolving an alert before navigation. Apply the same bounded cursor and incomplete-result rules as web. Use the shared native cancellation helper, including runtimes without `AbortSignal.throwIfAborted`, before and after every request. Typed injected observations record query/mutation outcomes without notification contents. Native screen cache keys and composition must include server/session scope as well as inventory identity.

### Native personal settings session

Create a fresh native preference session per mounted server/principal/inventory settings surface. It serializes initialization, refresh and saves with the last successful revision, preserving unrelated defaults, delivery preference, timezone and type overrides. Failed or cancelled operations retain the previous snapshot and UI draft; explicit refresh obtains current state before retrying an uncertain mutation. Native snapshots are copied using ordinary object/array operations rather than requiring a runtime structured-clone API. A push-preference write is separate from device permission and registration; the UI must complete permission/registration handling before enabling delivery. The composition root provides the session factory and shared injected notification observation.

### Native reminder policy editor

Use a shared native policy editor for inventory defaults and type overrides, with AppSwitchField and AppTextInput, explicit save, and whole-number thresholds from 0 through 3650. Inherited policies display current inventory defaults and disable individual controls until customization is selected. Saving inheritance removes the override. Retain drafts after failures, show safe inline error text, announce saved status, prevent duplicate saves, and ignore late UI updates after unmount. The parent keys the editor by preference scope/type and owns revision-aware save callbacks. No permission prompt occurs in this policy editor.

### Native settings screen

The native settings screen loads a complete active asset-type query and initializes personal preferences before exposing editors. It presents inventory defaults, editable IANA timezone, and expiration-enabled type overrides in a scrollable native surface. Load failures are explicit and retryable. Refresh updates saved revisions and inherited defaults while retaining open drafts. One save disables all editors until completion; a failed save preserves drafts and offers refresh before retry. Unmount cancels pending work. Route composition supplies a fresh session keyed to authenticated service scope, tenant and inventory.

Native inventory settings include Notifications for all members with inventory view access, including viewers; personal reminder settings do not require inventory configuration permission. The route uses the shared selected-scope query and native back navigation. Changing authenticated server scope or inventory remounts the preference session and discards the previous screen draft.

### Native inbox interactions

The scoped native inbox uses the shared native segmented control for All and Unread, renders readable day/month expiration labels without timezone-shifting calendar dates, and supports refresh and cursor pagination. Opening a notification resolves its current visibility and asset before marking it read and navigating; failures keep the inbox open with safe retry guidance. Mark all read processes every bounded server page, then refreshes the list and badge. All operations disable competing controls, cancel on unmount, and ignore late results. Failed loads retain existing rows, and an initial failure must not claim the inbox is empty. The parent keys the screen by authenticated server and inventory and supplies normal asset navigation plus reminder-settings navigation.

### Native inbox navigation and badge

Home includes a native bell beside existing actions, opening the notification screen through the native stack. Its unread count uses a server/principal/tenant/inventory query key, initializes personal preferences with the device timezone once per cache lifetime, polls while foregrounded every thirty seconds, and refreshes on reconnect/focus and notification reads. Initialization failure remains retryable and must not display a false zero. Count failures hide a stale number and expose an accessible unavailable state while retaining inbox access. The inbox route resolves the selected inventory before rendering and invalidates the same count key after reads. Asset taps use the existing asset detail route. Reminder settings use the inventory settings route.

### Push delivery state

Push delivery uses native device tokens behind a provider port (APNs for iOS and FCM for Android), without requiring Expo's hosted push relay. Provider-specific token parsing remains in adapters. The notification domain represents a pending delivery, a leased attempt, provider acceptance, permanent cancellation, or exhausted failure. Provider acceptance means handoff succeeded, not proof the operating system displayed the alert.

Each claim consumes one attempt and carries a nonempty opaque fencing token plus lease deadline. Completion/retry requires the same unexpired lease; stale workers cannot finalize another worker's attempt. A lease may be reclaimed after expiry. Injected worker configuration supplies maximum attempts, initial retry delay and maximum retry delay; configuration must be positive and the cap at least the initial delay. Retry delay doubles by consumed attempt up to the cap without duration overflow. Exhausted retries become terminal failures. Permanent device rejection cancels delivery and retires that registration; authorization, lifecycle, date or preference withdrawal cancels without retiring a valid device. All transition times are supplied by the injected application clock. No token or provider response body belongs in a domain error, audit metadata or logs.

### Device registration ownership

Device registrations are scoped by tenant, inventory and authenticated principal and have an opaque server ID, client installation ID, provider transport, secret native token, active flag and optimistic revision. An installation has at most one registration in that scope. Updating a token, re-enabling or revoking advances the revision; queued deliveries bind to that revision and never silently switch destination after token rotation. Registrations in different inventories may share a token only for the same principal. An active token cannot be claimed by a different principal; sign-out revokes registrations before clearing credentials. Reject conflicts without returning the existing owner's identity. Provider rejection retires matching active registrations; no token values appear in user-facing responses or audit.

Repository saves atomically enforce scope, installation uniqueness, active-token ownership and revision matching with audit. Reads require full scope, and active-device enumeration uses bounded stable ID pagination. Device registration does not backfill old milestones; only future generated notifications create deliveries. Token values are bounded nonempty printable ASCII secrets; provider adapters apply provider-specific validation before registration. Normal string/JSON formatting of the domain token is redacted; persistence and provider adapters use an explicit secret accessor.

Database registration writes use transaction-scoped locks keyed by a SHA-256 digest of transport and token. Lock rows contain no raw token and remain stable across revocation so concurrent re-registration cannot bypass ownership checks. Rotation acquires old and new digest locks in sorted order; after locking, active ownership checks and the scoped revision update occur in the same transaction as audit. Operational token-ownership checks are an explicit exception to scoped user-facing reads and never return another user's registration. A revision conflict rolls back every change. The active registration listing remains fully scoped. Tokens are persisted only in the device table, and the adapter suppresses ORM query logging for operations that carry those secrets.

### Device commands and audit

Registration commands require inventory view permission and an active inventory, use the authenticated principal, and validate the native token through an injected provider-validation port. Register takes installation ID, transport, token and expected revision (zero for first registration). Retrying the same active registration with revision zero is idempotent; changing an existing registration requires its current revision. Revoke requires device ID and current revision, returns not found outside the caller's full scope, and is idempotent at the current revision when already inactive. Scoped installation lookup supports refreshing registration metadata after conflicts. Read responses omit tokens. Registration changes and revocation record notification_device.updated and notification_device.revoked audit actions against the device ID; lookup records notification_device.viewed. Observation contains only scope/device IDs and outcome, never token or installation ID.

Own-device cleanup is an explicit authorization exception: authenticated users may look up and revoke their own scoped registration after losing inventory membership or after inventory archival. These operations expose only their device metadata and cannot affect another principal's records; they do not require current inventory view permission. Register/re-enable still requires current view permission and an active inventory. This permits sign-out to release token ownership for the next user of the installation. The REST boundary must test authenticated ownership independently from inventory membership, including membership loss followed by revocation and new-user registration.

### Device REST contract

Under `/tenants/{tenantId}/inventories/{inventoryId}/notification-devices`, POST registers with `{installationId, transport, token, revision}` and returns 200 metadata; GET `/by-installation/{installationId}` reads own metadata; DELETE `/{deviceId}?revision=N` revokes and returns 200 metadata. Metadata contains ID, installation ID, transport, revision and active, never token. Authentication is mandatory on all three operations. Unknown request principal fields cannot select another owner. Register requires inventory view; cleanup GET/DELETE require authenticated scoped ownership even after membership loss. APNs tokens must be nonempty even-length hexadecimal strings; FCM tokens remain bounded opaque printable strings, since the provider owns their format. Malformed inputs fail without reflecting secrets. Client types are generated from OpenAPI.

Shared HTTP schema-validation details report the constraint and field location but omit the submitted value. This prevents malformed token types as well as oversized strings from being reflected through validation errors, and applies consistently to other secret-bearing inputs.

APNs token normalization is mandatory before ownership comparison: decode hexadecimal input and encode canonical lowercase hex. The injected token adapter returns the normalized token. Persistence rejects noncanonical APNs values as invalid input, preventing alternate casing from bypassing token ownership even through another adapter. FCM tokens remain case-sensitive and unchanged.

### Device client adapters

The shared authenticated client provides registration, installation lookup and revision-aware revocation from generated OpenAPI types, preserving revision zero and inactive responses. Cancellation propagates through the existing transport. Native device adapters expose separate frontend registration metadata and command models, map safe notification failures, and ignore cancelled responses. Device tokens exist only in registration command input, never in returned frontend metadata. Existing notification and device adapters share one cancellation/error translation helper.

### Native push setup lifecycle

Enable push only after an explicit permission request succeeds, a native token is obtained, and the authenticated device registration completes. Denial changes neither inbox preferences nor registration. Before a registration mutation, persist a token-free cleanup record containing server, principal, tenant and inventory plus a stable installation ID, so an ambiguous response remains discoverable by installation lookup. This is cleanup metadata, not an offline mutation queue. Refresh saved preferences before enabling push and preserve all unrelated values.

Sign-out cleanup enumerates only records for the current authenticated server/principal, refreshes each installation's current registration, revokes active registrations, then removes successfully cleaned metadata. Missing/already inactive registrations count as cleaned. A failed cleanup retains its record and reports failure; callers must finish cleanup before discarding credentials. Revisions are refreshed before each revocation; a conflict retries from current metadata at most once. Setup and cleanup serialize to avoid a late registration after sign-out. Cancellation stops dependent operations and preserves cleanup metadata.

### Secure registration journal

The native journal uses Expo SecureStore with a dedicated keychain service and device-only accessibility. One injected journal instance serializes reads and writes for the app lifetime. A versioned document stores a stable generated installation ID and an exact allowlist of server/principal/tenant/inventory strings; it never serializes arbitrary caller fields or device tokens. Remember is idempotent, forget removes only the exact scope, and successful writes survive a new adapter instance. Storage or validation failures preserve the existing document and surface a safe error; corrupt data is never silently replaced because it may be needed to revoke an active registration. Storage writes complete before setup continues.

Journal validation bounds installation/principal/tenant/inventory identifiers to 128 characters and server identity to 2048 characters, rejects empty identities and unknown document versions, and uses a device-only accessibility option when supplied by the secure-store runtime.

### Native permission adapter

The injected Expo notification adapter supports iOS/APNs and Android/FCM device tokens without a relay. Android creates the expiration notification channel before requesting permission. Existing grants do not prompt again; denied permission returns false. Unsupported platforms, permission API failures, invalid native token shapes and token retrieval failures return safe notification errors without exposing provider data. Tokens must match the current native platform and be nonempty bounded printable strings before reaching registration. Pin SDK-compatible expo-notifications 55.0.23 and expo-crypto 55.0.16 (already present transitively in the native lockfile) for native registration and secure installation ID generation.

### Signed-in push session

Native runtime composes the device adapter, secure journal and authenticated device repository behind a push session. The server identity is the normalized connection URL, never the transient UI cache scope. Every setup/cleanup resolves the current authenticated principal through the principal port; UI inputs cannot select a registration owner. The session supplies the inventory preference session internally for enablement. Cancellation after identity resolution stops subsequent mutations. Explicit sign-out and server changes await registration cleanup before disposing the current services or clearing credentials. Transport, storage and revision cleanup failures leave the signed-in connection intact and are presented by the existing connection-action error surface. Authentication expiry is an explicit exception: preserve the saved server and secure cleanup journal, transition to sign-in through the existing authentication-required flow, and do not complete the requested sign-out/server-reset action. Reauthentication as the same principal permits cleanup to be retried; another principal cannot revoke the retained registrations.

The signed-in session serializes enablement and disconnection through completion and rejects further setup after a successful disconnect. When the secure journal has no registration for this server, disconnect proceeds without network identity lookup or creating an installation ID. This preserves ordinary offline sign-out for users who never enabled push.

The iOS lockfile includes the pinned notification module and shared date picker using the CI CocoaPods resolver. Preserve the previously locked libwebp version during this feature; image-codec upgrades require their own review. CI must confirm that the resulting native dependency lock remains reproducible before release.

### Mobile push preference control

The mobile notification settings screen uses the shared native switch for mobile push delivery. The switch reflects the saved personal inventory preference and explains that it controls mobile alerts for this inventory while inbox reminders remain available. Turning it on invokes the signed-in push session, requests device permission and registers this installation before saving the preference. Denial leaves the preference unchanged and explains how to allow notifications in device settings. A busy indicator covers permission/registration/save work and disables competing settings mutations. Turning it off saves only pushEnabled=false, preserving reminder policies, timezone and inbox history. Refresh the screen's preference session after successful enablement so later edits use the latest revision. Unmount cancellation prevents later UI updates and dependent writes.

When the saved inventory push preference is already enabled, expose a separate “Set up alerts on this device” action. It performs permission and registration without toggling the account preference off first. This supports second devices and registration after signing back in; an enabled inventory preference alone must not be presented as proof that the current installation is registered.

### iOS push signing capability

The generated iOS entitlements include aps-environment. Debug/default development builds use development; Release and the production build environment use production. The Expo notification config plugin mirrors the production-build choice, sets the expiration Android channel, and does not enable background silent notifications. Visible expiration alerts need no background execution permission. TestFlight signing validation requires a production APNs entitlement in the distribution provisioning profile, and the archived app's signed entitlements are checked before upload. Missing capability is a release failure, not a silent fallback to an app that cannot register native push tokens.

### Atomic delivery queue contract

A notification publication can atomically insert its inbox record, creation audit and one pending delivery per eligible active device. Each delivery stores its own ID, the full recipient scope, notification ID, device ID and device revision, creation time, and delivery state; it stores no token or copied message content. Reject invalid, duplicate-device, cross-scope, inactive-device or stale-revision destinations before any write. Repeating an existing milestone returns the original notification without creating new deliveries, including when a later device is registered. Queue claiming is an operational cross-recipient stream, ordered by ID and bounded to 100 records. Claims atomically advance the domain retry/lease state with a caller-generated unique fence. Settlement requires the current unexpired fence and uses typed accepted/cancelled/retry outcomes; exhausted attempts become terminal. Queued work is revalidated against current authorization, lifecycle, preference and device revision by the application before provider delivery.

### PostgreSQL delivery persistence

Migration 57 adds a token-free notification_deliveries table with recipient scope, inbox/device references, device revision, creation timestamp and lease/retry columns, plus uniqueness per notification/device and a due-work index. Publication shares one database transaction with inbox/audit insertion and locks destination device rows in stable ID order before committing deliveries. Workers claim due pending or expired leased rows using FOR UPDATE SKIP LOCKED, then persist domain transitions within that transaction. Settlement locks the delivery row and validates its current fence/expiry before applying an outcome. SQLite repository tests use the same mapper and application contract; real PostgreSQL tests verify concurrency.

The PostgreSQL CI service runs notification device-ownership and delivery publication/claim concurrency tests as a separate required step, rather than leaving them skipped in the default SQLite suite.

### Publishing mobile deliveries

Recipient generation uses the atomic inbox/delivery publication port. For a due milestone with personal push enabled, enumerate all active registrations for that exact recipient scope in stable bounded pages and capture their current revisions. With push disabled, publish the inbox record with no deliveries. Recheck inventory access after device enumeration and before publication. A device-revision race fails the whole publication so a later generation pass can retry; never publish only the inbox and silently lose its mobile deliveries. Repeated milestones never backfill alerts onto devices registered afterward. Runtime composition must supply the delivery repository for both memory and PostgreSQL deployments.

### Delivery execution

The application claims a bounded delivery page and resolves each job through the current inbox, asset/type lifecycle, personal policy and active device revision before sending. Re-evaluate the milestone using current time/timezone: an obsolete upcoming job is cancelled once expired. The push port receives only a device token/transport, opaque notification scope/ID, delivery idempotency ID and generic title/body; asset names and dates never enter lock-screen text. Provider outcomes are accepted, retryable or invalid-device. Invalid-device retires only the still-matching registration revision with audit; changed registrations are not retired. Dependency/provider failures retry with bounded backoff. Context cancellation leaves leases for recovery. Recheck access immediately before the provider call and bound its context to the remaining lease time. Observations record only delivery ID and outcome, never provider errors or token data.

### APNs provider authentication

The APNs adapter uses Apple's HTTP/2 token-authenticated API. Its injected credential signer accepts an Apple key ID, team ID and PKCS#8 P-256 private key, signs ES256 JWTs with the team issuer and injected-clock issued-at time, and caches tokens for 50 minutes. Concurrent requests share the cache; a backwards clock invalidates the cached token. Configuration and signing failures return fixed safe errors without key material. Key/team identifiers must be ten uppercase alphanumeric characters. The private key remains inside the adapter and no JWT or key is emitted to observations. Source: Apple, “Communicating with APNs” and “Sending notification requests to APNs.”

### APNs request adapter

APNs requests target the configured production or sandbox Apple endpoint over HTTP/2 and use the configured bundle topic. The adapter disables redirects, propagates cancellation, bounds response reads, sends alert priority 10 with expiry zero, and derives a stable collapse ID from the delivery ID. The generic alert plus scoped notification identifiers and configured API-server identity form the payload; no asset content is included. A 200 acknowledges provider acceptance. A 400/BadDeviceToken identifies an invalid token. A 410/Unregistered includes its provider invalidation time; retirement requires a usable timestamp at or after the current registration time. Missing or malformed timestamps remain retryable; other statuses, malformed responses and transport failures remain retryable without revealing provider response text. Redirects never forward credentials. Collapse IDs reduce duplicate pending notifications but do not establish exactly-once delivery.

Registration with the current nonzero revision records a fresh registration, even for the same token, advancing revision and UpdatedAt. Revision-zero duplicate creation remains idempotent and does not represent a fresh registration. Provider invalidation timestamps travel as typed times through the push result port. The application never retires a registration newer than that timestamp, and revision fencing protects a registration refreshed during the provider request.

### FCM HTTP v1 request adapter

Android delivery uses FCM HTTP v1 at the fixed Google endpoint and a configured Firebase project ID, behind the push sender port. Authentication is an injected context-aware OAuth access-token source. Requests contain a generic visible notification and string-valued scoped navigation data, the configured Android notification channel, explicit HIGH priority for these visible alerts, zero TTL and a stable delivery-derived notification tag. Redirects are disabled; requests honor cancellation; provider bodies and credential errors never escape into application errors or observations. Payload and response reads are bounded. Only a well-formed 200 message name is acceptance; only a 404 with typed FCM UNREGISTERED detail invalidates a device. Authentication, project mismatch, invalid arguments, malformed/oversize bodies and transient errors remain retryable. Retry-After and quota retry timing must be honored by the delivery scheduler before enabling this adapter in runtime. Sources: [FCM v1](https://firebase.google.com/docs/cloud-messaging/send/v1-api), [FCM errors](https://firebase.google.com/docs/cloud-messaging/error-codes).

### Provider-directed retry deadlines

A push result may carry a `RetryNotBefore` instant. The application propagates it only for a retry outcome; settling a leased attempt retains the later of the provider deadline and the ordinary exponential-backoff deadline. The backoff cap cannot shorten a provider deadline. Missing or past deadlines preserve ordinary backoff, and exhaustion still terminates the attempt. Lease ownership is evaluated at the actual injected current time, never at the future retry instant. Both repositories persist the resulting next-attempt instant atomically with the lease fence. FCM reads Retry-After as nonnegative decimal seconds or an HTTP date, using an injected clock, and uses at least one minute for 429 quota failures. Invalid headers fall back to ordinary retry policy; numeric overflow is rejected. The delay survives malformed provider response bodies.

### FCM credential adapter

The first FCM credential adapter accepts an explicitly supplied Google service-account JSON document at the infrastructure boundary, rather than discovering arbitrary credentials from application code. It uses the pinned Go OAuth library with the firebase.messaging scope and Google's fixed OAuth token endpoint. Credential-supplied alternate token endpoints are rejected before secrets can be sent. RSA keys must be at least 2048 bits and parse at construction. The library owns OAuth JWT signing; cache freshness uses the injected clock. Refresh uses the current request context and a redirect-disabled HTTP client. Concurrent requests share a token until one minute before expiry; waiting for refresh is cancellable, and a backwards clock invalidates the cache. Empty, expired, malformed or failed credential responses return a fixed safe error. No private key, JWT, access token, or provider response appears in application errors or observations. Other Google credential formats require a separately specified adapter.

### Push runtime activation

Native push providers are independently enabled through environment configuration and default off until credentials are supplied. APNs requires key ID, team ID, bundle topic and a mounted private-key file; production is the default with an explicit sandbox override. FCM requires project ID, a mounted service-account JSON file and the native expiration channel. Both require a configured public API-server identity URL. Enabled but missing/invalid credentials fail startup with fixed safe errors. Credential files are read only during infrastructure construction and bounded to 64 KiB. Sender routing uses only the registered device transport; a missing provider never falls back to another transport. The sender is injected into the notification service. A cancellable delivery worker runs when at least one provider is enabled, with configurable page size, poll interval, page timeout, lease and retry bounds. Defaults are 10 jobs, five-second polls, 30-second page timeout, one-minute leases and six attempts with one-minute initial/fifteen-minute maximum backoff. Page timeout must be shorter than the lease. Shutdown cancels and joins the worker; failed pages emit a safe domain worker-failure observation. In-app notification generation remains independent of push activation.

Delivery pages claim one job immediately before processing it, up to the configured page size. A stalled send or cancelled page must not consume retry attempts for jobs the page has not started.

### Mobile registration reconciliation

On authenticated startup, return to the foreground, or native push token change,
reconcile only journaled registrations for the current server and principal.
This background operation checks permission without prompting. If permission is
revoked, revoke the device registration while retaining its journal record for a
later permission restoration. Do not change personal inventory preferences.
With permission, read the current native token and register idempotently with
revision zero first; only on revision conflict fetch the current revision and
retry. Unchanged registrations must not rotate revisions or invalidate queued
notifications merely because the app returns to the foreground. Account/server
cleanup and reconciliation share the session operation guard. Abort outstanding
work when its authenticated composition unmounts. Coalesce repeated triggers and
retry failed reconciliation on the next foreground or token event; report safe
notification observability without including tokens.

Native token events carry the new token into a validated in-memory adapter cache. Reconciliation caused by that event uses the supplied token rather than requesting another native token, preventing Expo token-fetch event feedback. Foreground transitions invalidate the cache before reconciliation; tokens are never persisted in the registration journal.
A native token read that finishes after a newer token event or foreground invalidation must not overwrite that newer state.

### Native notification taps

Direct APNs compatibility (2026-09-12): the pinned Expo Notifications 55.0.23
iOS serializer exposes direct APNs custom fields in the push trigger's `payload`;
`content.data` reads the Expo-specific nested `body` dictionary and may be null.
The native adapter must use the raw payload only for an explicitly identified
`push` trigger when content data is absent. Do not merge routing fields across
payload sources or fall back from present malformed content data. Both launch
and live responses must enter the same application validation and authorized
notification lookup. Cover real direct-APNs response shapes, ordinary content
data, non-push triggers, wrong server/account, malformed fields, cancellation,
and unavailable notifications. Invalid routing hints must give notification-open
guidance, not tell the user to change reminder settings.

Handle the launch notification and later default-action taps through the native
notification adapter. Deduplicate a launch response also delivered by the live
listener. Treat payloads as untrusted routing hints: require bounded nonempty
server, principal, tenant, inventory, and notification identifiers. Match the
configured server and authenticated principal before any notification request.
Never use an asset identifier or URL supplied by the payload. Resolve the current
notification through the authorized inbox API, mark it read, select its inventory
through the existing inventory selection command, and open the returned item in
its normal native detail screen. Inventory selection must honor cancellation
before changing the composition-local selection. Cancel superseded taps and
pending work on composition unmount; cancelled work must not navigate or show
feedback. Unavailable, expired, or inaccessible notifications show a native
message, without changing account or server automatically.

Suppress duplicate delivery while a tap is being handled and suppress previously consumed launch responses. A later explicit tap on the same notification is a new attempt and must remain usable after a transient failure.

### Expiration visibility in inventory browsing

Web and mobile item details and shared inventory cards show a compact expiration
label whenever a stored date is present. Preserve day versus month precision in
localized formatting; month-only labels must not invent a day. Use calendar-date
formatting independent of the viewer's UTC offset. Undated items add no empty
row. The date label remains visible when type tracking is disabled, so stored
information is not hidden. These date labels do not imply reminder delivery or
an upcoming/expired state computed from a different user's preferences.

### Current notification placement

Inbox list and detail responses include `parentTrail`, ordered from the outermost
known ancestor to the immediate parent, with each ancestor's asset ID, title, and
kind. Resolve this at read time within the recipient's authorized tenant and
inventory; do not persist location snapshots in notifications. Bound traversal to
128 ancestors, detect cycles, and never include foreign, missing, or archived
ancestors. Return `parentTrailIncomplete` when the full chain cannot be resolved,
so clients can indicate a partial path. Reuse ancestor reads within one inbox
page. Placement enrichment applies to displayed list/detail results, not delivery,
unread counts, or marking read. Existing notification read audit covers this
response metadata.

### Inbox placement controls

Both inboxes render the authorized parent trail beside each notification using
horizontal breadcrumb controls, initially revealing the immediate parent. Each
ancestor opens that asset through normal inventory navigation. Keep ancestor
buttons separate from the button that opens and marks the notification read;
opening an ancestor does not mark the notification read. Disable ancestor actions
while inbox mutations or an item-opening operation are pending. Partial trails
show an accessible incomplete-location indication. Legacy servers without trail
fields render no invented breadcrumbs. Empty complete trails add no row.

## Approved native usability revision (2026-09-11)

The user approved implementing all findings in `docs/reports/expiration-ui-ux-audit-2026-09-11.md`, including populated-inbox hierarchy, reliable existing numeric badges, and explicit mark-unread support.

- Inbox uses one navigation title, All/Unread filtering, separated rows instead of outlined cards, native pull-to-refresh on mobile and a compact accessible refresh action on web. Settings and mark-all actions live in navigation/toolbars. Unread entries use a dot and stronger title weight; read entries remain legible. Individual read/unread actions are accessible without opening the item.
- `PUT .../notifications/{notificationId}/read` preserves the existing mark-read contract. `DELETE` at the same read resource marks unread, idempotently, after current notification visibility and recipient/inventory authorization checks. The state transition and `notification.unread` audit record are atomic. It does not recreate a milestone or push delivery. Cross-principal, cross-tenant, inaccessible/withdrawn and unauthenticated requests must fail at the REST boundary.
- Preferences use concise grouped default rules and one row per type. Type policy modes are Use defaults, Custom, Off. Inventory defaults remain inheritable, never a global override of type choices. Before expiration combines off/preset/custom-day selection; expired is independent. Hide inactive subordinate editors while preserving saved choices.
- Clean editor drafts follow refreshed server state; dirty drafts are never silently overwritten. Explicit conflict reload reconciles the displayed values. Simple controls save consistently; focused compound edits have explicit commit/cancel. Readable timezone selection preserves the stored timezone across travel.
- Push preference, OS permission and server readiness are separate facts. Permission/registration success must not claim verified delivery. Native system settings is the recovery for denied permission.
- Badge refresh follows mutation, app foreground and inventory scope changes. Failed refresh retains known same-scope counts while announcing unavailable freshness, not false zero. Opening the inbox alone does not mark all read.

Migration 58 expands the PostgreSQL audit action constraint for notification.unread. Its down migration intentionally retains this additive action allowance so historical audit records remain valid; no historical action is deleted or rewritten.

CI records desktop and phone-width browser evidence for personal reminder editing and the read → unread inbox journey. Toolbar and row utility actions use compact, labeled icon buttons. Native device layout verification remains separate.

### Native settings audit follow-up (2026-09-11)

- Inbox and reminder screens fill the available navigation viewport. Their scroll content grows to fill it, and vertical bounce/refresh works when empty or shorter than the screen, including drags beginning in blank space. Refresh indicators represent list refresh, not unrelated saves, opening or read-state mutations.
- Reuse the existing grouped SettingsSection, row typography, insets, separators and Dynamic Type layout. Switches remain platform switches. Never expand an entire rule form inside an overview disclosure.
- The overview shows personal inventory rules, push preferences, a timezone row and a row for each expiration-enabled type. Type customization and timezone selection use ordinary native stack navigation with back behavior. A type screen offers Use defaults, Custom and Off with a visible selection; inherited rules show a concise summary.
- Before expiration opens a focused timing screen with Off, same day, one day, one week, two weeks, one month (30 days), two months (60 days), three months (90 days), and custom whole days (0–3650). Labels explicitly use days rather than implying calendar-month arithmetic. Presets save on selection; custom days commit from the navigation Done action and back/cancel discards the unsaved field. Failed saves keep edits and expose retry without silently saving them through another control.
- Saved values use correct singular/plural labels (1 day). Expired reminders remain independent. Type overrides still take precedence over the user's inventory defaults.
- A focused edit loads the same authorized personal inventory preferences through existing ports. Returning to the overview refreshes current settings; dirty focused fields are not overwritten by refresh. No draft or preference state crosses tenant/inventory/account scopes.
- Timezone selection uses a searchable, checked list with clear empty and failed-save states. It preserves the stored timezone across travel. Device permission status stays distinct from server delivery readiness; use concise status and supporting section footers rather than repeated setup prose in the primary form.
- Tests cover empty/short refresh, preset and custom timing, cancellation, save failures, inheritance, singular labels and retained personal scope. Device geometry/gestures require native CI evidence separately from renderer assertions.

### Direct APNs tap repair evidence (2026-09-12)

Native response regressions reproduced the reported settings error for both cold
launch and live direct APNs taps. The adapter now extracts absent content data
from explicit push-trigger payloads without merging sources, then preserves the
existing server/account/authorized inbox checks. Malformed routing has dedicated
notification-open guidance. Remote mobile typecheck, 1,261 tests and mobile
structural checks passed; required critic found no substantive issues. Actual
phone tap acceptance follows the signed release; receipt was already user-verified.
