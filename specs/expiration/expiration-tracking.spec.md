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


### Web Date Entry

The shared expiration field uses the existing segmented control for Exact date versus Month and year, and a labeled native date/month input through the shared Input component. Separate precision drafts preserve what the user entered when switching modes without inventing a day. The visible selected value is the value submitted on Save. Clearing removes the active value. Invalid or incomplete nonempty input blocks save with inline guidance. In add-item forms, the field appears only for enabled types; changing the selected type resets its date draft to prevent accidental carryover.


### Web Date Editing

The item editor reuses the expiration field for an assigned active, expiration-enabled type. Date-only edits participate in dirty tracking and dismissal protection. Clearing sends explicit null; unrelated edits omit the date. A failed save retains the draft, while reopening initializes from persisted asset data. Retained dates whose type no longer tracks expiration remain visible with a tracking-disabled explanation and can be cleared without re-enabling the type.


### Initial Type Assignment In The Web Editor

An untyped item's editor offers active custom types from its authorized inventory context. Selecting a type reveals its applicable fields and expiration capability, and counts as a dirty change. Save submits initial type assignment and date together. Existing assigned types cannot be replaced here. Changing an unsaved type choice clears its expiration draft but preserves other existing asset fields and tags. Unknown retained fields are not silently discarded; the application service validates compatibility before accepting assignment.

Reselecting the current type in either create or edit is a no-op and must preserve the visible date and its submitted value.


### Native Editor State

Native detail view models retain the original expiration value and assigned type identity. The edit draft distinguishes unchanged/omitted dates from explicit clearing, includes type assignment in dirty detection, and preserves precision through normalization. A date-only change enables Save and protects dismissal; an invalid active date draft blocks Save while retaining edits. The native save command receives the normalized date/type values through the existing application port.


### Native Date Control Dependency

Use `@react-native-community/datetimepicker` pinned to `8.6.0`, the installed Expo SDK 55 compatibility version, for system day selection behind a shared UI component. The installed pinned `@expo/ui` package exposes segmented control but not a date-picker replacement. Month precision uses labeled month/year native text controls and strict calendar validation; no synthetic day is stored. The field preserves separate drafts by precision, supports clearing, and never submits a date merely because a picker opened or was dismissed. UI receives its initial picker date from its caller. Remote tests use a controlled native-picker adapter; device behavior is validated by the CI-built release.


The shared native field stages iOS picker changes until Use date, commits Android's confirmed selection, and leaves the draft unchanged on dismissal. Its date button exposes the current value to assistive technology. Remote controlled-picker tests cover confirmation, cancellation, month/year validation, clearing and precision-draft retention; these do not substitute for native-device verification.


### Native Type Choices

Native asset forms load active tenant and inventory custom types through a dedicated application query using the existing customization ports. The inventory collection already includes inherited tenant types; do not call the tenant configuration endpoint, which requires elevated permissions. The query requires the expected tenant/inventory scope to match the selected context, honors cancellation, and refuses incomplete collections. It filters returned records to that scope and active lifecycle before presenting type choices; missing or failed metadata is not evidence that expiration tracking is disabled.


### Native Edit Integration

The native edit route loads type choices under the current inventory query cache and refreshes them after customization changes. Loading failures show retry without treating tracking as disabled. Untyped assets offer initial type selection; typed assets retain their type. Enabled types show the shared date field; disabled types retain a visible recorded date and allow clearing. Changes merge into the existing text/tag draft. Changing type clears the expiration draft; reselecting the same type preserves it. The picker initial date is captured at the UI boundary, and the field is keyed by asset/type so navigation cannot reuse another item's draft.


### Native Creation Integration

The native add form uses the shared type/expiration editor and scoped type query. Valid chosen dates and initial type IDs join the existing per-principal/inventory draft and create command. Failed saves retain them; successful saves and Clear draft reset them. Invalid dates block submission without clearing other inputs. Date/type controls remain visible outside the optional description/tag section. Partial invalid date text stays in the mounted field; reopening restores the last structured draft and fresh validation state.

### Conversation expiration facts

Authorized asset tool results include a nullable expiration object with original date/precision, calendar status (current/upcoming/expired), trackingEnabled, effective advanceDays, and recipient timezone. Read status uses the personal type window even when notifications are disabled. A retained date on a disabled/archived type still has a factual calendar status, but trackingEnabled is false. Missing dates remain null. Type vocabulary explicitly exposes expirationEnabled so the model can resolve a suitable existing type without inventing capability. Invalid personal calendar configuration fails the read rather than substituting a misleading timezone. Calendar projection belongs to the expiration application package; conversation orchestration only maps its result.

Conversation read-facts implementation evidence: remote targeted tests reproduced missing tool expiration and type capability, then passed with the projections. The remote full Go API suite and structural checks passed; required code critic review found no actionable issues. This covers data supplied to the model and calendar rules, not completion of expiration query tools, action-plan writes, card rendering or live-corpus acceptance.

### Expiration in create action plans

Create-asset and create-location command arguments may include customAssetTypeId and an expiration object with strict date/precision fields. Dates require a nonempty type ID; malformed, null, unknown-field or calendar-invalid date objects fail validation rather than being silently discarded. Omission preserves existing create behavior. Approval execution passes both fields through the ordinary asset application service, including dependent multi-create plans; current type capability and scope are revalidated before the atomic write. Stored command JSON retains the original precision for recovery. Keep create-argument parsing and mapping in their own focused file instead of growing the action-plan execution facade beyond its structural limit. Model schema exposure and review controls must ship together before this is considered a complete voice write flow.

### Voice review expiration presentation

The proposed-command event includes optional expiration {date, precision}, copied from validated create arguments. The mobile adapter validates the calendar value and precision before forwarding a review; malformed values fail the session instead of being hidden. The session retains a copied value through review/history/status transitions. Each affected command displays “Expires <date>” using a shared calendar formatter, with month-only values rendered as month/year. The formatter uses UTC solely to format calendar components and never converts the recorded day to a different local day. The inbox and voice review share this formatter. Dates on unsupported command kinds are rejected until the corresponding edit contract is implemented. Existing commands without dates retain their presentation.

Historical plan summaries also show each command's recorded expiration date beside its title, including cancelled/failed/saved plans. A status transition must not remove the date from the conversation's visible history.
