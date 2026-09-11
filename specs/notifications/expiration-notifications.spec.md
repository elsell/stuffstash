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
