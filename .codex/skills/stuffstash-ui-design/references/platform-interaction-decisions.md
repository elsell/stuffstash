# Platform interaction decisions

Use this before selecting components. This is a project decision framework informed
by Apple's HIG, not a replacement for its platform-specific guidance. Recheck the
linked guidance for the target OS when behavior changes. Reviewed 2026-09-14.

## Task-to-pattern decisions

| Task | Preferred starting point | When a different pattern earns its place |
| --- | --- | --- |
| Choose one of a few flat values | Labeled pop-up/menu-style picker showing current value | Long descriptions, search, hierarchy, or values that cannot be understood in a menu |
| Toggle a persistent setting | Native switch with immediate, recoverable feedback | Explicit draft form where the whole form commits together |
| Switch between a few peer views | Native segmented control when all choices fit | Tab/navigation structure for distinct destinations; menu when space or labels make segments unsuitable |
| Choose many tags or a parent location | Searchable selection list with selected state; hierarchy for containment | A short local multi-select menu if choices are genuinely bounded and understandable |
| Choose date/time | System date picker; compact or inline according to task | Month-only expiration needs a precision-aware adapter; never silently invent a day |
| Issue contextual commands | Native action/pull-down menu, grouped by task | Prominent button for frequent primary action; alert only for a decision requiring interruption |
| Navigate to an object or substantial settings category | Native navigation destination with normal back behavior | Inline disclosure for a small amount of supplementary content |
| Complete a bounded task | Sheet with clear title, completion, and dismissal | Longer/multistep work may warrant a full-screen task; avoid an app inside a sheet |
| Inspect a photo | Full-screen viewer with familiar zoom/dismiss affordances | System preview/share where supported; custom viewing needs accessibility and gesture parity |
| Search | Platform search entry appropriate to priority and available space | Do not force permanent fields or icon-only search universally; honor accepted app placement |
| Report progress | Local status for local work; pull indicator only for an explicit pull | Blocking progress only when the task truly cannot continue |
| Recover from mistakes | Retain draft, undo, retry, or actionable error at the task | Confirmation where loss is meaningful; avoid confirming every harmless change |

Apple specifically describes flat exclusive choices in [pop-up buttons](https://developer.apple.com/design/human-interface-guidelines/pop-up-buttons)
and context-preserving selection in [pickers](https://developer.apple.com/design/human-interface-guidelines/pickers).
Short-choice defaults are not a ban on selection screens. Read the iOS/iPadOS
sections; a watchOS navigation picker is not an iPhone design precedent.

## Review every applicable design dimension

| Dimension | Decision to establish before implementation | Evidence to seek |
| --- | --- | --- |
| Purpose, familiarity, inclusion | Does the flow match a household task and familiar concepts? | Task walkthrough, terminology and alternatives |
| Navigation and hierarchy | Which actions navigate, select, or mutate? Are back and tab return predictable? | Entry/deep link/back traces; no duplicate chrome |
| Menus and controls | Current value, exclusivity, enabled/selected/destructive states | Open/select/dismiss using touch and assistive tech |
| Modality | Why interrupt? What commits, cancels, or preserves work? | Button, swipe, system-back, outside-tap and async completion |
| Layout and safe areas | Which system owns each inset/scroll edge? | Top, middle, bottom, keyboard, sheet detents, rotation |
| Adaptation | Does content reflow across phone/tablet/window widths and large text? | Narrow/wide layouts, split window, longest labels, accessibility sizes |
| Typography, language, RTL | Are scale, wrapping, reading order, units and dates appropriate? | Dynamic Type, expanded strings, RTL, locale and timezone fixtures |
| Color, materials, appearance | Are semantics legible in light/dark and accessibility appearance modes? | Contrast measurement and native render; no stacked custom glass over native material |
| Icons and imagery | Familiar symbols, understandable action labels, useful photos? | Accessible names, missing media, crop and thumbnail behavior |
| Touch, gestures, pointer | Can tasks be completed without hidden gestures? | Target bounds, gesture competition, explicit alternatives, pointer affordances |
| Keyboard, focus, data entry | Appropriate input, submit/dismiss, focus restoration? | Software/hardware keyboard, Tab/Escape, dictation, validation recovery |
| Accessibility | Name, role, state, value, reading order, focus and nonvisual feedback? | VoiceOver/TalkBack traversal; web semantics; not just role props |
| Motion and haptics | Is motion informative and optional? | Reduce Motion, interruption, no essential haptic-only meaning |
| Content and collections | Scan, selection, empty/error states, long lists, hierarchy? | Dense and sparse fixtures, loading more, scroll restoration |
| Search and refinement | Scope and active filters clear? Draft versus applied state clear? | Type, clear, cancel, no results, scope change, return from detail |
| Loading and feedback | Does status describe the user's operation? Can they continue/retry? | Slow/failure/stale data, background refetch, announcements and timing |
| Editing and undo | What constitutes saved work? Can users recover from dismissal? | Dirty draft, partial save, duplicate submission, undo after notice disappears |
| Launch and onboarding | Can users reach the task with minimal required setup? | Cold/warm start, interrupted sign-in, returning user, no inventory |
| Accounts, privacy, permissions | Does the user understand scope, data destination and permission request? | First use, denied/revoked permission, sign-out/switch; do not invent deletion promises |
| Notifications | Relevant, actionable, controlled by user? | Device receipt, cold/warm tap, unavailable item, read state, settings |
| Sharing and files | Use platform acquisition/share where appropriate; show progress and recovery | Camera/library/files, cancel, limited access, upload failure, export/share |
| Audio and generative AI | Clear listening/processing state, user control, understandable data use? | Interrupt/stop, no microphone permission, typed fallback, plan review/cancel |
| Help and writing | User language, concise recovery, no developer narration? | Labels, errors, empty states, advanced settings separated from normal tasks |
| System integrations | Does a shipped integration follow its own interaction contract? | Widget/shortcut/link/system entry traces; mark unimplemented technologies N/A |

## Authoritative entry points

- [Design principles](https://developer.apple.com/design/human-interface-guidelines/design-principles)
- [Foundations](https://developer.apple.com/design/human-interface-guidelines/foundations): accessibility, inclusion, privacy, layout, type, color, materials, motion, imagery, writing, RTL.
- [Patterns](https://developer.apple.com/design/human-interface-guidelines/patterns): task flows, lifecycle, search, sharing, settings, feedback and recovery.
- [Components](https://developer.apple.com/design/human-interface-guidelines/components): content, organization, actions, navigation, presentation, input and status.
- [Inputs](https://developer.apple.com/design/human-interface-guidelines/inputs): touch, keyboard, focus, pointer and device-specific input.
- [Technologies](https://developer.apple.com/design/human-interface-guidelines/technologies): integration-specific guidance, including AI and VoiceOver.
- [iOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-ios) and [iPadOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-ipados).
- [Android design](https://developer.android.com/design/ui/mobile) and [web interaction patterns](https://www.w3.org/WAI/ARIA/apg/patterns/).

## Existing Stuff Stash adapters

Inspect before creating a substitute: `NativeChoicePicker`, `NativeActionMenu`,
`NativeSegmentedControl`, `NativeHeaderActions`, `NativeSheetActions`,
`NativeNavigationSearch`, `NativeTagColorPicker`, system date picker,
`AppSwitchField`/`SettingsSwitchRow`, native tabs and native-stack navigation.
Android implementations may differ or fall back to React Native: inspect the actual
platform-resolved file. Web primitives live under `apps/web/src/lib/components/ui`.

`SettingsNavigationRow` means navigation; it is not the default value picker.
A text footer saying how a custom interaction works does not justify that interaction.
A shared primitive is reusable only when its task semantics match its consumers.

## Deliberate departures

Record: task; preferred platform pattern; concrete unmet need or verified adapter
limitation; substitute; accessibility/keyboard/dismissal behavior; consumer list;
verification evidence; condition for revisiting. “Easier in React Native,” matching
an old screen, or copying another app's screenshot is not sufficient justification.
Do not require a new dependency merely to remove every custom visual element.

## Command appearance preference

Stuff Stash defaults commands to bounded native buttons. Use the existing secondary
native style for ordinary commands and primary emphasis for the task's main action.
Borderless commands require a contextual reason, such as established navigation-bar
placement; do not use blue text as a blanket default. Navigation rows and selection
rows retain their platform conventions. This is an explicit product preference,
not a claim that Apple's HIG forbids borderless controls.
