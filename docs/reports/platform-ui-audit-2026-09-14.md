# Platform interaction audit — 14 September 2026

The recurring pattern suggests **implementation-led interaction design**.
This is an inference about the process, not a claim about an author’s motivation.
Using a native container did not ensure that the task belonged in that container.
The strongest current example is a three-option filter taking over a sheet instead
of opening an in-place menu, despite an existing native picker adapter.

This audit updates the UI design skill and surveys the product against Apple's HIG
coverage framework. It does not change production UI. The separate, previously
authorized Home action reorder is being released independently.

## Scope and evidence

Reviewed source revision: `8392eeb4cedbd7e327c344f4e6b5a78a2fff4851`.
Coverage includes mobile route entrypoints and shared iOS/Android implementations,
web workspace modes and task surfaces, onboarding and system entry points, and the
documentation shell. iPad matters: the app declares `supportsTablet: true`.

- [Surface inventory and family assessments](platform-ui-audit-2026-09-14/surfaces.md)
- [Apple topic applicability ledger](platform-ui-audit-2026-09-14/apple-topics.md)
- [Updated skill](../../.codex/skills/stuffstash-ui-design/SKILL.md)
- [Review policy](../../specs/platform/platform-interaction-review.spec.md)

**Evidence limit:** this is a whole-product source/design audit, not a completed
on-device usability certification. Live automation returned `CUA_REPL_ENABLED_SURFACES
is required`; no running native UI or browser was available through that tool.
No native builds/tests were run locally. Historical user screenshots establish the
reported regressions, but do not prove that those defects persist in today's build.
Source inventories are exhaustive for the declared roots; detailed findings are
based on inspected implementations. Device-only questions remain explicitly open.

The HIG ledger contains 170 topic/category documents discovered through Apple's
Getting Started, Foundations, Patterns, Components, Inputs, and Technologies
hierarchy. This is coverage of the retrieved index, not a claim to have audited
all linked developer API documentation or every future HIG revision. Apple DocC
JSON supplied the page contents where HTML retrieval returned a JavaScript shell.

## Contributing factors inferred from the workflow and source

1. **The skill did not cover mobile.** Its discovery description, required artifact,
   and workflow were web/SvelteKit-specific. Mobile lacked an equally explicit
   task-to-platform-pattern review.
2. **Shared navigation components became defaults for value selection.**
   `SettingsNavigationRow` supplies a chevron and “Opens a settings screen” hint;
   filter callers used it for small enumerations. That made navigation easy to
   implement without establishing whether navigation helped the task.
3. **We evaluated construction more than appropriateness.** Structural checks can
   ban raw React Native modals or enforce adapters. They cannot determine whether
   a short choice warrants another page. Renderer tests can pass either choice.
4. **Fixes were too local.** Home's refresh presentation was corrected, but other
   lists retain the same coupling to background query activity. Shared-pattern
   searches must accompany local fixes.
5. **Release evidence was overstated.** Tests and successful native archives do not
   establish toolbar fit, focus, gestures, accessibility, or scroll restoration.

Apple's [design principles](https://developer.apple.com/design/human-interface-guidelines/design-principles)
provide the foundation: use familiar concepts, preserve agency, reduce unnecessary
work, and adapt to varied contexts. The new project rule is to record the task and
appropriate interaction **before** choosing a component or adding a route.

## Findings

Priorities: P1 = potential loss of work or inability to finish; P2 = repeated friction,
accessibility or interaction inconsistency; P3 = refinement. A priority on an
unverified risk indicates verification urgency, not a confirmed runtime failure.

### F01 · P2 · Short filters introduce unnecessary navigation

**Source-confirmed choice; design recommendation.** Browse Type (4), Status (3),
Availability (3), and Sort (2) replace the overview with a page of choices. Expiration
Kind and Availability do the same. These are internal page-state transitions, not
separate native stack routes, although they feel like a drilldown to the user.

Evidence: [BrowseFiltersScreen](../../apps/mobile/src/ui/screens/BrowseFiltersScreen.tsx)
and [ExpirationFiltersScreen](../../apps/mobile/src/ui/expiration/ExpirationFiltersScreen.tsx).

Use labeled native menu-style selection in the overview, showing the current value
and selected option. Keep the existing filter draft and explicit apply behavior.
Tags and large type/location collections can retain searchable selection views.
Do not collapse every selection into a menu.

Apple describes this distinction in [pop-up buttons](https://developer.apple.com/design/human-interface-guidelines/pop-up-buttons)
and [pickers](https://developer.apple.com/design/human-interface-guidelines/pickers).
**Acceptance:** change Availability without leaving Filters; cancel preserves
applied filters; apply commits the draft; long-label and VoiceOver checks pass.

### F02 · P2 · Generic value selection is implemented several different ways

**Source-confirmed choice.** `CustomizationEditorFields.SingleChoicePicker` creates
an expanding group of custom radio rows for Type and Applies to. The existing iOS
`NativeChoicePicker` already uses a SwiftUI menu picker. Android falls back to a
custom expanding `SelectionRow`, despite native Compose adapters elsewhere.

Evidence: [CustomizationEditorFields](../../apps/mobile/src/ui/components/CustomizationEditorFields.tsx),
[NativeChoicePicker.ios](../../apps/mobile/src/ui/components/NativeChoicePicker.ios.tsx),
[fallback picker](../../apps/mobile/src/ui/components/NativeChoicePicker.tsx).

Consolidate small exclusive values behind a task-appropriate platform adapter;
preserve search/list selection for descriptive or large collections. Add an
Android-native equivalent only after checking the pinned adapter capability.
**Acceptance:** equivalent value fields use consistent selection semantics,
selection is announced, and dismissal does not accidentally commit another value.

### F03 · P2 · Date entry has extra staging layers

**Source-confirmed choice; recommendation.** Exact expiration entry opens a
SelectionRow, then a Choose date button, then a wheel picker with separate Cancel
and Use date buttons, inside a parent editor that has its own save action.

Evidence: [ExpirationField](../../apps/mobile/src/ui/components/ExpirationField.tsx).
Prefer a compact or inline system date picker bound to the parent draft; the parent
editor owns save/cancel. Month/year precision remains explicit: do not substitute
a full-date picker that silently invents a day. Apple offers multiple valid date
picker styles; the issue is redundant staging, not that wheels are forbidden.
**Acceptance:** edit a date in context, cancel the parent, and retain the saved date;
verify locale, leap day, timezone, and month-only semantics.
[Apple pickers](https://developer.apple.com/design/human-interface-guidelines/pickers).

### F04 · P2 · Native toolbar adoption is incomplete

**Source-confirmed inconsistency; runtime fit unverified.** Home/Browse use native
header action adapters, while Add draws its own Cancel/title/Add row with the native
header hidden. Notification Inbox puts React Native Pressables into `headerRight`;
custom reminder timing also uses a custom Done control.

Evidence: [AddAssetScreen](../../apps/mobile/src/ui/screens/AddAssetScreen.tsx),
[NotificationInboxScreen](../../apps/mobile/src/ui/screens/NotificationInboxScreen.tsx),
[ReminderTimingEditor](../../apps/mobile/src/ui/components/ReminderTimingEditor.tsx).

Use system toolbar items for standard navigation/task actions where the existing
adapter supports them. Do not rebuild a navigation bar in content merely because
the parent presentation is a native sheet. Preserve task-specific content buttons.
**Acceptance:** compare Home, Browse, Add, Inbox, and reminder editor on the same
OS at ordinary/enlarged text; all actions are reachable and named.
[Apple toolbars](https://developer.apple.com/design/human-interface-guidelines/toolbars).

### F05 · P2 · Background loading still controls pull indicators elsewhere

**Source-confirmed coupling; stuck-spinner recurrence not reproduced.** Assets,
Location Assets, Sharing, and Expiration connect pull indicators to query refetch
state. `LocationsScreen` also does, but its current reachability needs confirmation.

Evidence: [InventoryAssetsRouteScreen](../../apps/mobile/src/ui/screens/InventoryAssetsRouteScreen.tsx),
[LocationAssetsRouteScreen](../../apps/mobile/src/ui/screens/LocationAssetsRouteScreen.tsx),
[InventorySharingScreen](../../apps/mobile/src/ui/screens/InventorySharingScreen.tsx),
[Expiration route](../../apps/mobile/src/app/expiration.tsx).

Apply the gesture-owned refresh contract across reachable consumers. Keep normal
background reads and local progress independent. **Acceptance:** background refetch
and return from detail do not activate the pull control; explicit pull does; blur,
failure, and late completion restore the correct state.
[Apple loading](https://developer.apple.com/design/human-interface-guidelines/loading).

### F06 · P2 · Global feedback does not accommodate motion/timing preferences

**Source-confirmed implementation gap; assistive-technology behavior unverified.**
The shared notice always springs into view and automatically disappears after 4.2
seconds, or 6.5 seconds with an action. It has no Reduce Motion branch or
assistive-technology timeout adaptation. The action area has a 38-point minimum
height. This differs from Map and voice-card code that already reads Reduce Motion.

Evidence: [AppFeedback](../../apps/mobile/src/ui/feedback/AppFeedback.tsx) and
[AppFeedbackPresentation](../../apps/mobile/src/ui/feedback/AppFeedbackPresentation.ts).

Respect motion preferences; keep essential recovery available outside a timed
notice; verify announcements and action targets. Native alerts need not replace
all nonblocking feedback. **Acceptance:** reduced-motion notice has no spring;
VoiceOver can hear and reach the action; recovery remains possible after timeout.
[Apple accessibility](https://developer.apple.com/design/human-interface-guidelines/accessibility),
[motion](https://developer.apple.com/design/human-interface-guidelines/motion),
[feedback](https://developer.apple.com/design/human-interface-guidelines/feedback).

### F07 · P1 verification · Inventory switching has unbounded content and no explicit Close

**Source-confirmed structure; large-inventory/device failure unverified.** The
switcher renders household/inventory rows in Views, with no scrolling container.
Its native sheet uses fit-to-content sizing and hides the header. “Back” changes the
internal household-selection mode; there is no explicit close/cancel action.

Evidence: [TenantSwitcherSheetScreen](../../apps/mobile/src/ui/screens/TenantSwitcherSheetScreen.tsx)
and [root layout](../../apps/mobile/src/app/_layout.tsx).

Keep the sheet for multi-household context if needed, but bound and scroll the list,
provide an explicit dismissal, and label the selected context clearly. A small
single-household case may fit a menu. **Acceptance:** 30 inventories, long names,
large text, and VoiceOver allow selecting the last entry or dismissing without a
change. [Apple sheets](https://developer.apple.com/design/human-interface-guidelines/sheets).

### F08 · P1 · Web Add discards a draft on ordinary dismissal

**Source-confirmed loss path; browser reproduction pending.** AddAssetTray resets
its draft on each opening, forwards closing directly to `onClose`, and revokes
photo previews after closing. The workspace close handler closes/navigates without
a dirty check. This differs from field settings, which already has discard handling.

Evidence: [AddAssetTray](../../apps/web/src/lib/components/workspace/AddAssetTray.svelte)
and [InventoryWorkspaceApp.closeAdd](../../apps/web/src/lib/components/workspace/InventoryWorkspaceApp.svelte).

Retain the draft or confirm meaningful discard, consistently across Close, Escape,
outside interaction and browser navigation. Do not add confirmation to every empty
form. **Acceptance:** type a title and choose photos, dismiss accidentally, reopen,
and recover the draft or receive a clear discard decision before loss.
[Apple modality](https://developer.apple.com/design/human-interface-guidelines/modality)
and [web dialog behavior](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/).

### F09 · P2 · Platform adaptation lacks acceptance evidence

**Unverified risk, not a confirmed broken layout.** iPad is declared supported, but
this session has no native evidence for split-window layouts, enlarged text,
keyboard, VoiceOver, reduced transparency, or RTL. Several shared controls still
use fixed widths/heights and physical chevrons. The fact that the app has some
responsive branches and accessibility props does not prove these combinations.

Evidence: [app configuration](../../apps/mobile/app.config.js),
[SettingsList](../../apps/mobile/src/ui/screens/SettingsList.tsx),
[NativeSegmentedControl](../../apps/mobile/src/ui/components/NativeSegmentedControl.tsx).

Run the acceptance matrix in the skill before declaring broad native quality.
Do not automatically redesign every iPad view into a sidebar or flag every fixed
icon size. **Acceptance:** all audited tasks remain usable across supported width,
text, appearance and input configurations; record actual build/device evidence.
[Apple layout](https://developer.apple.com/design/human-interface-guidelines/layout),
[iPadOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-ipados).

### F10 · P2 · Location selection loses distinguishing context

**Source-confirmed limitation; duplicate-name fixture not exercised.** Expiration's
filter route maps locations to `{id, label: title}`. The selection view displays
only that label. Two locations with the same title cannot be distinguished by
ancestor context, despite having different IDs.

Evidence: [Expiration filter route](../../apps/mobile/src/app/expiration-filters.tsx)
and [ExpirationFiltersScreen](../../apps/mobile/src/ui/expiration/ExpirationFiltersScreen.tsx).

Keep a searchable selection view here, but include an ancestor path or another
meaningful distinguishing label. Do not substitute a large flat menu. Assess
virtualization if the supported collection size requires it. **Acceptance:** select
the intended “Shelf” from two different rooms with touch and VoiceOver; selected
context remains clear after returning to the filter overview.
[Apple lists and tables](https://developer.apple.com/design/human-interface-guidelines/lists-and-tables).

## Patterns worth retaining and deliberate decisions

- Native tabs, navigation, Home/Browse toolbar adapters, system photo-source choice,
  and native switches are useful foundations. Their existence is not blanket QA.
- Time zones, large tag/type sets, hierarchical parent destinations and provider
  setup have legitimate search, description or configuration needs. Separate
  selection/configuration views are reasonable there.
- The web expiration dialog already uses in-place Select controls and labeled date
  inputs. Do not replace accessible web controls with an iOS visual imitation.
- Full-screen photo viewing is an appropriate task; its custom toolbar needs
  accessibility/runtime review, not automatic removal.
- Mobile Add has draft-store logic; mobile edit has a dirty-close prompt and disables
  swipe dismissal; customization editing has a navigation guard. The web Add finding
  must not be generalized into “all editors lose work.”
- Bottom filter Apply/Cancel actions reflect earlier explicit user direction. Apple's
  iOS sheet guidance normally places Cancel and Done in the top toolbar. Record this
  tension as an intentional project choice to revisit, not an accidental violation
  or permission to silently reverse the user's decision.
- The containment Map is not geographical. Apple Maps integration guidance does not
  require adding cartography. A calendar is likewise not mandatory for expiration.

## Remediation order and new procedure

First address draft loss and verify switcher reachability. Next replace the short
filter/customization choices and propagate the refresh contract. Then consolidate
standard toolbars, date entry, and accessible feedback. Run the cross-platform
acceptance matrix alongside each affected family rather than waiting for another
whole-app redesign.

The updated skill now requires a task/pattern decision, current-value and
commit/cancel semantics, an adapter/consumer inventory, and separate source/runtime
evidence. A whole-product audit must account for every entrypoint and each HIG
family, including justified N/A topics. Add narrow mechanical checks only where
behavior can actually be enforced; do not use a regex ban on navigation or custom
components as a substitute for design judgment.

The policy and skill are applied in this report. Remediation is proposed, not
implemented by this audit. Native/browser execution is the outstanding evidence
layer; it is not represented as complete.

## Validation of the updated review workflow

The revised skill was independently applied to four contrasting cases: a short
availability filter, 500 hierarchical locations, reminder presets/custom input,
and web Add dismissal. The review preserved justified selection destinations and
confirmed the web draft-loss path. Its finding about duplicate location labels is
included as F10. The final critic found no blockers; causal claims were softened
to clearly labeled inference.

Skill metadata validation passed on `paul`. All 425 local Markdown links across the
skill/report references resolved. No production code changed, so no new app test
suite or native build was needed for these documentation changes. Runtime audit
coverage remains as stated above; these validations do not substitute for it.
