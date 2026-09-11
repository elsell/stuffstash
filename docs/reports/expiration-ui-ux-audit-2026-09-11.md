# Expiration UI and UX audit

September 11, 2026 · iOS screenshot review and mobile/web source inspection

## Assessment

The expiration feature exposes its configuration model too directly. Users have to understand inheritance, device registration, date precision and independent save operations to complete ordinary tasks. Some individual controls are native, but their composition does not behave like a coherent iOS flow. The create form also has a visible layout failure, which takes priority over styling.

This is an audit and proposed design, not an implementation or a claim that the next build is fixed. Evidence comes from the five supplied screenshots and the current mobile and web implementations. No local build or tests were run. Intermittent scrolling, actual touch targets, VoiceOver, Dynamic Type and web rendering still require runtime verification. The screenshots establish overlap; they do not establish its exact layout-engine cause.

The existing domain model remains appropriate: optional dates on items, capability enabled by custom type, and personal reminder defaults per inventory with type overrides. Medicine needs no special treatment. Turning inventory defaults off must continue to permit explicit type overrides, as requested.

## Findings and priorities

| Priority | Finding and evidence | Required correction |
|---|---|---|
| Release blocker | Screenshot 4 shows Description overlapping Clear expiration; screenshot 5 shows Description crowding the date section. The user reports intermittent inability to scroll. | Establish a bounded, keyboard-aware scroll region whose content measures correctly as sections expand. Verify every field and the final action on a physical iPhone. Do not patch this with fixed extra spacer heights. |
| High | Save is below photos, type/date controls, description and the entire tag collection in `AddAssetScreen`. Expanding details can make completion several screens away. | Keep Cancel and Add in the sheet navigation bar, with one scrolling form below. Use a large presentation for this composition task. |
| High | The create type selector renders “Base asset” and every type as plain text Pressables. Selection is indicated visually only by text color, although radio semantics exist for assistive technology. Screenshot 5 makes the choices look like labels or links. | Use a clearly labeled **Item type** selection row, showing its current value and a disclosure affordance. A selection list shows a checkmark, selected accessibility state and an optional “Tracks expiration dates” subtitle. Use “None” instead of “Base asset.” |
| High | Mobile reminder settings render the full policy editor for every type, including disabled inherited values. One type already occupies most of screenshot 1. | Show one summary row per type. Open only the selected type's settings. Replace the inheritance switch plus disabled form with explicit modes and a readable effective-policy summary. |
| High | The page mixes reminder generation, per-inventory push preference, device permission and timezone. “Enable expiration reminders” sounds like a master switch even though type overrides can remain active. | Separate reminder rules from delivery. Label the defaults as defaults; show the number of custom exceptions when defaults are off. |
| High | “Alerts enabled on this device” follows permission/registration setup. It does not verify end-to-end delivery, and APNs server setup remains pending. | Say “Permission allowed on this device” for that fact. Show server delivery readiness only when known from a real capability/status source. Never equate a saved device token with working delivery. |
| High | Both platform reminder editors initialize draft values once. Refreshing parent preferences does not reset the existing default/custom editor state. Mobile and web timezone drafts also initialize only on first load. | Define refresh/conflict reconciliation: update clean drafts from the server; preserve dirty edits and offer a deliberate reload. A success banner must refer to the values actually saved. |
| Medium | Expiration introduces an always-visible precision selector and multiple fields even when no date is set. Month entry requires typing `03` and a year. Exact dates display raw ISO strings in the mobile field. | Start with a compact Expiration row. Use locale-formatted dates and contextual native date controls. Month-only selection must preserve month precision without requiring a numeric month code. |
| Medium | Each precision has its own draft. Switching to an untouched mode produces an empty effective value; clearing only clears the active mode, so switching back can restore an earlier value. This follows the current spec but is easy to misinterpret. | Revise the interaction contract before implementation: make the resulting value explicit, preserve meaningful month/year when changing modes, and make removal clear the expiration as a whole. Never invent an exact day from a month-only label. |
| Medium | “More details” does not control the type/date editor: that editor is rendered outside the conditional details block. All available tags render as a wrapping chip cloud. | Give type and expiration stable, compact positions in the main form. Put description and a selected-tags summary in optional details; open a searchable tag selection view for the full collection. |
| Medium | Inbox and settings repeat “Notifications” beneath the native navigation title. Large Refresh and Settings buttons dominate an empty inbox. | Use one navigation title; name the settings screen “Expiration reminders.” Put settings in a labeled toolbar action, support pull-to-refresh and use a meaningful empty state. |
| Medium | Separate Save reminders buttons and Save timezone coexist with an immediately persisted push toggle. Users cannot infer what is saved. | Use consistent save behavior: immediate persistence for simple settings with visible failure recovery; a focused editor with explicit Done for compound values such as a custom lead time. Avoid several unrelated save forms on one screen. |
| Medium | Raw IANA timezone identifiers and “Saved timezone” are implementation-oriented. | Present a readable Time zone row, such as “New York,” with a searchable selection view and a short explanation of its calendar effect. Preserve the saved inventory preference when traveling; do not silently follow the device. |

“Release blocker” means block a redesigned UI release until verified; this audit does not recommend withdrawing the existing build.

## Proposed iOS interaction

### Add and edit an item

Use a native navigation container within a large sheet. The first view has Cancel and Add, a short item form, and stable scroll/keyboard behavior. The existing system switch and segmented-control implementations are useful foundations; replacing them with custom imitations would not solve this problem.

The compact form contains Name, Photos, Location, Item type and, when supported, Expiration. Description and Tags sit under optional details. The type and date values remain visible without expanding details. Expiration reads **Not set**, **Mar 2027**, or **Mar 14, 2027**, formatted for the user's locale.

Tapping Item type opens a selection list within the current navigation flow, rather than stacking another modal sheet. The list includes None, searchable types when needed, and a checkmark on the selected type. Returning preserves the rest of the draft. Changing a draft's type must not silently destroy an entered date: explain incompatible changes before applying them. This does not expand the existing restriction on replacing an already assigned persisted type.

Keep date editing contextual: use the system compact date picker where supported. Reveal precision selection only while editing expiration. Month-only entry uses a month/year picker; iOS's standard date picker does not itself offer a month-only mode, so this requires a suitable native picker composition. Show “Expires at the end of March 2027.” Do not add a fake day to use a date-only control. Removing expiration clears it unambiguously. Cancel restores the prior committed draft value.

Apple recommends focused sheets, conventional completion/dismissal controls, and considering a larger presentation for extended tasks. Its picker guidance favors contextual editing and compact date pickers when space is limited. These inform the proposal; the precise field order is a Stuff Stash design decision. [Apple: Sheets](https://developer.apple.com/design/human-interface-guidelines/sheets), [Apple: Pickers](https://developer.apple.com/design/human-interface-guidelines/pickers).

### Expiration reminders

Use grouped settings sections and disclosure rows. A brief subtitle establishes scope: “Your reminders for Main Inventory.”

| Section | Example rows |
|---|---|
| Default reminders | Reminders: On; Before expiration: 30 days; When expired: On |
| By item type | Medicine: Uses defaults; Food: Custom; Batteries: Off |
| Delivery | Push notifications: On; This iPhone: Permission allowed |
| Calendar | Time zone: New York |

For a type, offer **Use defaults / Custom / Off**. Use defaults shows one sentence: “30 days before expiration and when expired.” Custom exposes only that type's controls. Off retains its explicit override. Defaults off shows “Default reminders are off. 2 types use custom reminders” when applicable; it is not mislabeled as a universal kill switch.

“Before expiration” combines the upcoming switch and number field into one value row: Off, common intervals, or Custom. Preserve the currently supported custom range rather than restricting users to presets. Keep “When expired” separate. Collapse irrelevant controls when reminders are off, retaining their values for re-enabling. Provide a concise effective summary instead of requiring users to mentally evaluate multiple toggles.

Apple recommends minimizing settings, choosing useful defaults and avoiding setup questions that can be answered automatically. Progressive disclosure is our application of that guidance. [Apple: Settings](https://developer.apple.com/design/human-interface-guidelines/settings).

Delivery must distinguish an inventory's push preference, this device's OS permission and server readiness. Show Enable only when permission has not been requested, Open Settings when denied, and a factual status when allowed. In-app reminders remain available independently of push delivery. Ordinary expiration reminders should respect Focus and system notification preferences; an expiration date alone is not justification for Critical or Time Sensitive delivery. [Apple: Managing notifications](https://developer.apple.com/design/human-interface-guidelines/managing-notifications).

### Inbox, item details and conversation

**Updated unread-indicator finding:** The user confirmed that the numeric unread badge does appear and is acceptable; replacing it with a dot is no longer required. Investigate freshness instead. The mobile count query waits for preference initialization and polls every 30 seconds; the web bell also refreshes on a 30-second interval. These are plausible contributors to delayed appearance, not proof of the exact delay observed. Account for notification generation and network latency separately. Keep one shared bell per platform. Refresh after read actions, app resume and inventory changes, and invalidate the count when this client learns that reminders changed. The badge clears when no unread reminders remain, not merely when the inbox opens. Never carry another inventory's badge across a scope switch. A failed refresh must not falsely report zero unread; retain the last known indicator for the same scope and expose refresh failure in the inbox. Verify all entry points and announce unread state accessibly.

Keep All/Unread and replace utility buttons with native toolbar actions and pull-to-refresh. The empty state should say “No expiration reminders yet,” with a short explanation and a settings link. Distinguish no reminders from no unread reminders, loading, offline and failed loading.

Each reminder should lead with the item and event: “Tylenol expires in 7 days,” followed by the recorded date and location. Use an unread marker with an accessible label, not a repeated “Unread” paragraph. Opening goes to the item; marking read must not imply the expiration is resolved. Deleted or inaccessible items need a recoverable destination state.

Item details, search previews, notification rows and conversation cards should use the same date vocabulary and preserve month precision. An unknown date is “Not set,” not “Does not expire.” Retained dates on disabled types need a clear tracking-disabled explanation. Conversation reviews must show proposed dates before approval, and historical responses must remain recognizably historical. These are acceptance checks beyond what the supplied screenshots can verify.

### Web

Preserve the same meaning with web-native components: compact form fields, date/month inputs, searchable type/tag controls and type override disclosures. Do not copy iOS navigation chrome into the browser. The existing web type settings already collapse individual policies, which is better than mobile, but share the save, draft-refresh and date-mode ambiguities. Verify keyboard focus, browser month-input fallback, dialog overflow and responsive layouts before calling the web surfaces reviewed visually.

## Populated tray follow-up: September 11, 07:22 screenshot

The additional screenshot confirms the inbox problems with real content, not only an empty state. Four equally prominent outlined rectangles compete for attention: Refresh, Reminder settings, the notification and Mark all read. The item itself is visually no more important than maintenance actions. “Unread” is a text line, while the row otherwise uses the same styling as a read notification. The repeated heading consumes space before any content appears.

Required redesign:

- Keep the native page title only, followed by All/Unread and the list. Move reminder settings to a labeled toolbar action and Mark all read to the toolbar's actions menu, enabled only when useful.
- Replace boxed notification cards with list rows separated by subtle dividers. Lead with the item name, then the expiration event and date, with location as secondary context. A thumbnail is useful when present but must not create empty image-shaped whitespace when absent.
- Unread rows have a leading unread dot and semibold title; a subtle background treatment may reinforce this. Read rows lose the dot and use regular title weight. Keep both fully legible; do not communicate read state by fading the entire row. Include read/unread in the accessibility announcement.
- Make the primary row target open the item. Provide an accessible Mark read action without requiring navigation. Do not promise Mark unread until its command contract is specified and supported. Keep location links separate from the primary target.
- Use native pull-to-refresh, including in the empty state. Preserve content during refresh and show a contextual retry only after a failure. Do not replace the large Refresh button with another permanently prominent utility control. On desktop web, retain a compact accessible refresh action because pull gestures are not a sufficient desktop affordance.
- Reading a reminder changes its inbox state only. It must not erase the item's expiration warning or imply the item has been replaced, discarded or made current.

## Expiration visibility throughout the inventory

Expiration belongs on the things users browse, not solely in the notification inbox. Source inspection confirms that the shared mobile AssetCard currently prints a neutral “Expiration: <date>” line. Its summary mapping carries the date but does not carry the detail model's expiration context into that card. This makes adding color locally insufficient: browse surfaces need an authoritative, scoped status as well as a date.

Use a shared expiration-status presentation on mobile and web:

| Item state | Visual treatment |
|---|---|
| Approaching expiration | Small warning icon and **Expires soon · Sep 12** text, with an accessible amber semantic accent. |
| Expired | Distinct status icon and **Expired · Sep 10** text, with an accessible stronger warning accent. |
| Date outside the upcoming window | Neutral date metadata where useful; no warning decoration. |
| No date | No warning in compact cards; show Not set in details/editing. Never imply the item is safe or non-expiring. |
| Type tracking disabled with retained date | Preserve the date and a tracking-disabled explanation; do not claim active tracking. Explicit expiration queries still retain their existing ability to find these dates. |

Keep item photos, names and locations recognizable. Do not tint the entire card, add another outline, or turn every item into an alert panel. Read/unread dots belong to notifications; expiration icons and words belong to assets. These are separate meanings and must not share an ambiguous indicator.

Apply the same treatment to inventory list and grid views, Home/recent items, search results, container/location contents, item details and conversational result previews. Compact selection rows should show the status when choosing an existing item, without distracting from the selection task. A location or container must not inherit an expired label simply because something inside it has expired; descendant counts would be a separate, explicitly labeled aggregation feature.

“Soon” uses the viewing user's inventory/type advance window and calendar timezone, consistently with explicit expiration queries. Push permission, delivery preferences and notification read state must not suppress item status. Preserve month-only labels as month/year; never render an invented recorded day. Recompute or refresh status at the relevant calendar boundary, on foregrounding, after date/type/preference edits and when switching inventories. Use shared application rules and scoped data, not independent card-level clock arithmetic or an API call per card.

Acceptance now includes mixed current/upcoming/expired/unknown items across every surface, two users with different thresholds, read notifications whose items remain expired, midnight transitions, month-only labels, disabled tracking, dark/light appearances and large text. Verify status wording plus icon shape, not color alone, following [Apple's accessibility guidance](https://developer.apple.com/design/human-interface-guidelines/accessibility).

## Layout investigation and implementation boundaries

The Add route uses `formSheet` with `sheetAllowedDetents: 'fitToContents'`, while the screen uses a flex-filled shell and a growing ScrollView containing native controls. This is a concrete investigation target for the sizing/scroll issue, not a proven diagnosis. The form already requests keyboard inset adjustment, so simply adding that prop is not a fix. Also inspect native picker/segment intrinsic measurement, content size after expansion, sheet gesture ownership and keyboard dismissal overlays.

Key source locations:

- `apps/mobile/src/app/_layout.tsx`: Add sheet presentation.
- `apps/mobile/src/ui/screens/AddAssetScreen.tsx`: form composition, scrolling, completion action and tag collection.
- `apps/mobile/src/ui/components/AssetExpirationEditor.tsx`: type choices and date composition.
- `apps/mobile/src/ui/components/ExpirationField.tsx`: native date interaction and precision drafts.
- `apps/mobile/src/ui/components/ExpirationReminderEditor.tsx`: inheritance, local draft and save behavior.
- `apps/mobile/src/ui/screens/NotificationSettingsScreen.tsx` and `NotificationInboxScreen.tsx`: navigation and delivery presentation.
- `apps/web/src/lib/components/workspace/settings/NotificationSettings.svelte`, `ExpirationReminderEditor.svelte`, and `../ExpirationField.svelte`: corresponding web state and controls.

Build shared mobile selection rows, grouped setting sections and form navigation behavior rather than adding more handwritten controls to the large Add screen. Keep preference resolution and date validation in application/domain layers; views present effective values and dispatch changes through existing ports. Update the expiration and client UX specs before changing interaction semantics. This audit adds no dependency or new domain concept.

## Verification required before the redesign ships

| Scenario | Acceptance evidence |
|---|---|
| Small and large iPhones; light/dark; largest accessibility text | No overlaps or clipping; every field and action reachable; readable labels and values. |
| Keyboard open; expand/collapse details; open/close date editor; 50 tags and 20 types | Predictable scroll ownership; focused field visible; completion and dismissal remain reachable; no loss of draft. |
| Type selection by sight, VoiceOver and keyboard | Current value, selected checkmark, accessible selection state and discoverable activation. |
| Exact date, month-only, leap day, mode switch, removal, cancel and reopen | Displayed value equals saved value; no fabricated day, silent clearing or revived removed value. |
| Inventory defaults on/off crossed with type defaults/custom/off | Effective summaries and actual generated reminders agree for each user and inventory. |
| Refresh from another device with clean/dirty drafts; save failure and conflict | Clean values refresh, dirty changes are protected, no false saved state or stale overwrite. |
| Push allowed/denied/not requested; preference off; server unavailable | Accurate separate statuses, useful recovery, no claim of guaranteed delivery. |
| Empty/unread/error/paginated inbox; deleted or inaccessible item | Clear states, accessible read actions, correct navigation and no scope leakage. |
| Every bell; new unread reminder; individual/all read; app resume; inventory switch; refresh failure | Existing badge reflects unread state consistently, survives simply opening the inbox, clears only after all are read, and never shows another inventory's state. Measure and explain refresh latency. |
| Web narrow/wide viewports, keyboard-only, browser date/month behavior | Usable dialog scrolling, focus restoration, correct semantics and parity with mobile. |

Use Apple accessibility guidance to verify Dynamic Type, sufficient contrast in both appearances, non-color selection cues and familiar interactions. These cannot be certified from screenshots or component tests alone. [Apple: Accessibility](https://developer.apple.com/design/human-interface-guidelines/accessibility).

The next implementation should start with form containment and shared selection controls, then simplify reminder settings and delivery status, then polish the inbox. CI behavior tests must be supplemented with recorded simulator/device interaction covering the cases above. Passing API or component tests is insufficient evidence that the native flow is usable.

## Implementation follow-through

The approved revision adds mark-unread, compact type/date controls, stable Add completion, searchable tags, personal reminder modes, native mobile pull-to-refresh and shared item expiration warnings. Regression coverage includes PostgreSQL audit atomicity, authorization, date precision, dirty drafts, external read-state refresh and calendar-boundary retry. Native simulator/device layout acceptance remains separate from component-test evidence; do not treat a successful TestFlight upload as proof of those visual checks.

## Follow-up from device screenshots: reminder settings

The first revision still did not use the existing grouped native SettingsList. The days disclosure exposed a second toggle, a full-width unstyled input and two vertically stacked actions. Type and timezone disclosures could lengthen the same page indefinitely. `1 days` was grammatically wrong. The inbox ScrollView did not explicitly fill its viewport/content, leaving blank-space pull gestures outside the intended scroll surface; its refresh indicator also represented unrelated operations.

Fix: reuse the existing settings groups and native switches, use native stack screens for type rules, timing and timezone, retain a compact overview, use checked timing choices with a focused custom entry and navigation Done action, and fill the scroll viewport/content for short and empty states. Preserve inheritance, failed drafts and timezone semantics. Review against Apple Settings, Lists and Tables, Pickers, and Toggles guidance; test scope and save behavior remotely and build via CI. Native touch geometry cannot be inferred from renderer tests alone.
