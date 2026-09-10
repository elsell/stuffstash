# Expiration Tracking

## Approved Product Scope

Expiration is an optional capability of any user-defined custom asset type. Multiple tenant- or inventory-scoped types may enable it independently. Medicine is an example, never a built-in category. Tags and containment remain unchanged. Each physical package with a different date is a separate asset; this slice introduces no lots, quantity splitting, or product grouping.

The user approved implementation in the existing web and native mobile designs, personal reminder inheritance, PR/CI validation, GitOps deployment and TestFlight release on 2026-09-10. No additional visual approval gate is required for this agreed scope.

## Type Capability And Item Date

- Custom asset types expose `expirationEnabled`, default false, editable under existing type-configuration authorization and audited with type changes.
- Assets expose an optional structured expiration value: `date` in strict `YYYY-MM-DD` or `YYYY-MM` form and `precision` (`day` or `month`). Missing means unknown/unset, never non-expiring proof.
- Both forms validate calendar dates, including leap years, without silently normalizing invalid input. No automatic medicine shelf-life inference is permitted.
- A day value remains valid through that calendar day. A month value remains valid through the final calendar day of that month. The expiration boundary is the following local midnight; reminders use the user's inventory reminder timezone.
- API storage preserves entered precision; a month label must never be presented as a user-entered exact day. Domain logic uses an injected time source. Calendar boundaries and warning starts use the first representable instant of the target local date: choose the first occurrence of a repeated midnight, advance through a skipped midnight, and advance to the next representable date when an entire civil date is skipped. Calculate calendar offsets before resolving local time.
- Date edits require normal asset edit permission, tenant/inventory isolation, atomic persistence and audit. Clearing the date is supported. Normalized no-op edits do not create new date revisions or alerts.
- Setting a nonempty date requires an assigned expiration-enabled custom type. Disabling or archiving that type retains existing dates but stops future reminders. UI shows the retained date with tracking disabled; re-enabling restores tracking without duplicate milestone alerts for the same date.
- Existing untyped assets need a path to assign an existing active custom type. This narrow extension allows initial assignment with validation of all retained custom fields, preserving tags and containment. Replacing an already assigned type remains outside this slice; do not silently discard incompatible fields.
- Date tracking belongs to the expiration bounded context, with project-owned ports and application orchestration. No cross-domain direct imports or framework-dependent domain models.

## User Interface

- Type create/edit settings on web and mobile include an accessible Expiration tracking switch.
- Item create/edit uses one optional Expiration field when enabled by type. Exact-date entry is the default; Month and year only is an explicit mode. Switching precision requires visible confirmation of the resulting value, never silent date invention. Clearing is a labeled action.
- Mobile uses shared native controls; web uses existing shadcn-Svelte form primitives. Calendar entry supports keyboard/text input on web and system date controls on mobile. Month-only entry uses labeled month/year controls. Validation remains inline without dismissing entered content.
- Item detail shows the entered date plus Expiring soon/Expired status when applicable, with text in addition to color. Month-only input explains that tracking runs through the end of the month.
- Date and type assignment integrate with existing save/error/undo behavior; failures preserve edits and never claim success. Existing tags remain visible and unchanged.
- Notification entrypoints, preferences and inbox are specified in `../notifications/expiration-notifications.spec.md`.

## Validation And Release Evidence

Tests precede implementation: precision round-trips, month ends/leap years, invalid dates, timezone/DST boundaries, unset dates, capability disable/re-enable, initial type assignment and field preservation, no-op writes, audit/undo, adversarial authenticated API boundaries and cross-platform entry behavior. Tests/builds run remotely or in CI, never on the user's Mac. Completion requires required checks, code critic review, generated API contracts, healthy GitOps deployment and confirmed TestFlight upload. Record actual evidence after completion; do not equate synthetic tests with physical-device verification.

## Asset Integration Contract

The precision-preserving date is a shared immutable value object in `internal/domainvalue/expirationdate`. Asset and notification application code may use it without importing another domain package. Persistence stores its original date and precision together; application validation requires type capability before setting a date. Asset create accepts optional `expiration: {date, precision}`. Asset PATCH accepts the same object, omission preserves it, and `null` clears it. Responses include the object or null. Expiration changes participate in existing optimistic stale checks, atomic asset edit audit and undo snapshots, including persistence round trips. Initial type assignment is an explicit optional `customAssetTypeId` PATCH value; clearing or replacing an already assigned type is not a general edit operation.

Undo/redo that changes type assignment must validate the target type is currently active; restoring a snapshot that retains the same archived type remains allowed. This must behave consistently across memory, SQLite and PostgreSQL adapters.

### Foundation evidence (2026-09-10)

The type capability and asset date API/persistence foundation is implemented. Remote Go tests passed for the full API tree; subsequent focused regressions passed for archived-type redo and response-schema nullability. Memory-backed HTTP tests cover date create/edit/clear, initial type assignment, invalid calendar values, missing/disabled types and cross-tenant rejection. SQLite tests cover date persistence and undo snapshot restoration. The shared calendar tests cover leap years and clock transitions at midnight, including a date that briefly repeats after rollback. Generated API types explicitly represent absent expiration as null. Remote mobile type checking and web checking passed (one existing web CSS warning); required code critic findings were fixed and re-reviewed. PostgreSQL migration CI, the cross-platform feature UI and notification delivery remain pending; this is not release-completion evidence.

## Client Type Settings

Existing asset-type create/edit forms expose a shared labeled checkbox or native switch, “Track expiration dates,” defaulting off for new types. Supporting text explains that individual assets receive their own optional date. Detail views show whether tracking is on. The capability participates in dirty-form detection and is preserved through API and in-memory adapters. Client models accept an absent capability from older cached records as false; explicit save sends the chosen boolean. Disabling the capability preserves recorded dates while suppressing expiration tracking, as defined above.

Closing or discarding the type editor clears its draft initialization state, so reopening the same type restores the saved capability instead of reviving an abandoned toggle.

The native type editor uses a reusable `AppSwitchField` backed by the platform switch, with an accessible label and explanation. It is disabled for inherited/read-only types and during save or lifecycle changes. Editor snapshots include expiration capability so toggling it alone counts as an unsaved change; creation and update both send the selected boolean through the customization application service.

## Client Date Transport

Client asset models preserve expiration as an optional `{ date, precision }` value; absent/null server values represent no date. Create requests may omit expiration. Updates distinguish omission (keep the stored date), a date object (replace), and explicit null (clear). Existing untyped assets may send their initial custom type assignment along with a date, subject to server validation. Client adapters must not convert month precision to an invented day or drop explicit null during request construction.


## Conversational Expiration Support

The approved release includes expiration handling through voice and typed conversation on supported clients, using the same application services as ordinary asset forms. This is part of the feature acceptance criteria, not a later enhancement.

- Queries such as “what medicine expires soon” filter authorized inventory assets by the resolved custom type or tag and their actual expiration status. Medicine remains user vocabulary, never a hard-coded type. Support upcoming, expired, specific-date/range queries and follow-up context. Unknown dates must not be described as safe or non-expiring.
- “Soon” uses the requesting user's effective inventory/type advance-days setting and reminder timezone. Notification delivery switches do not hide matching assets from an explicit query. Explicit user-specified time ranges override the default query window. Upcoming excludes already expired assets; report expired separately when relevant.
- Add requests such as “add Tylenol that expires in February 2028” preserve month precision; exact dates preserve day precision. Edit and clear requests use the same structured approval flow. Relative dates resolve using the injected clock and requesting user's timezone and appear as the resolved date in review. Ambiguous numeric dates, missing years or unclear phrases require clarification instead of guessing.
- A type must support expiration before saving a date. Resolve an existing enabled type from authorized vocabulary and context; if not clear, ask which type to use. Do not silently enable a type, replace an existing assigned type, or discard a requested date. Initial type assignment follows the asset integration contract above.
- Review widgets show each proposed expiration value and its precision, alongside the item and destination, before approval. Approval, recovery, retries, audit and undo retain that date. A rejected or interrupted request cannot claim the item or date was saved.
- Relevant location responses include expiration context from current authorized data: for example, “I found your Tylenol in bin 8 in the hall closet. It expires soon, on February 12, 2028.” Month-only values are spoken as a month/year, never an invented exact day. Expired items are identified as expired. Result cards show the same date/status and preserve asset and location navigation.
- Tool contracts expose typed expiration values, resolved status and query filters; model output cannot bypass validation, type capability, tenant/inventory scope or approval. Typed input retains the existing no-speech behavior.
- Acceptance coverage includes day/month add, date edit/clear, ambiguous-date clarification, type/tag query resolution, personal thresholds/timezones, expired versus upcoming, missing dates, disabled types, location-answer enrichment, follow-up queries, cross-tenant denial and interrupted approval/retry. Remote realistic voice-corpus traces must be reviewed using the voice-evaluation workflow, in addition to deterministic tests.
