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
