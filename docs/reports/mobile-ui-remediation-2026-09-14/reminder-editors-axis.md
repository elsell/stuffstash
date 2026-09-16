# Reminder editors — all 24 source axes

## Inventory overview route R044 follow-up

Reviewed at2edf4dbd across the24 axes below, plus push-permission-axis.md for the
device subsection. R044 resolves the current authorized settings context before
mounting the reminder screen. Initial context loading is named; failure has a
native Retry command. The scoped child key includes service-state scope, tenant
and inventory, so a selection change replaces the session owner. Overview edits
personal inventory defaults, timezone, per-type rules and push preference; it is
not a household-wide policy editor.

Child destinations encode tenant/inventory identities. R043 validates those route
parameters and rejects mismatched currently selected inventory. Changes invalidate
only the scoped inventory query key. Shared screen and editor findings M199/M205
apply here; the earlier24-axis table and push review describe layout, keyboard,
state, selection, accessibility and lifecycle limits. No additional source defect
was established in this route pass. Parser tests and controlled HTTP denied-state
tests do not independently prove backend access enforcement or cold-deep-link Back
behavior; those remain separate acceptance boundaries.

Reviewed R043 and S111–S115 at 0ab593c2, including the route parser, scoped
NotificationSettingsRoute, NotificationSettingsScreen, ExpirationReminderEditor,
ReminderTimingEditor, TimeZonePicker and shared SettingsList rows. This extends
the task review in `notifications-axis.md`; it is not native acceptance.

## Task boundaries

- Defaults (S111) are independent booleans expressed as switches.
- Type policy (S112) is a short in-place Defaults/Custom/Off menu. Custom timing
  opens only when the user needs the separate timing task.
- Preset timing (S113) has Off, seven day choices and Custom. Selecting a preset
  commits immediately and returns after success; this differs intentionally from
  the explicit Save for custom input (S114). Failed selection remains retryable.
- Time zone (S115) is a searchable selection list. Search does not save. A failed
  save retains the previously saved checkmark. Successful selection returns.
- R043 validates view, optional type and required inventory identities. Invalid
  links and changed inventory show recovery instead of a different inventory's
  editable preferences. The root stack supplies the navigation destination.

These are project task-fit judgments using the established platform-interaction
standard. The number and nature of choices justify these separate editors; this
does not establish that every row's implementation looks or behaves natively.

| Axes | Source evidence and outstanding acceptance |
| --- | --- |
| Task, navigation, selection | Boundaries above. Timing has independent expired-reminder behavior. Verify system Back and successful return in the genuine stack. |
| Modality, gestures | Editors are pushed routes, not nested custom sheets. Timing hides Back and disables the interactive gesture while saving. Other editors rely on abort/focus ownership when departed; actual navigation races remain native checks. |
| Layout, adaptation | Parent ScrollView owns content and automatic navigation/keyboard insets. No fixed bottom action overlay. Shared grouped rows do not establish iPad column width, safe-area correctness or native search overlap. |
| Typography | Shared settings text wraps; custom number field has minimum width88 and height44. Long type labels and normal-size keyboard fit still require captures. Enlarged-text testing follows normal-size remediation. |
| Appearance, imagery | Semantic settings palette, switch/menu adapters and a check icon supplement text. No essential photos. Contrast, disabled-state appearance and materials need device inspection. |
| Localization | English labels; timezone identifiers are made readable by replacing underscores and reversing slash components. Search matches readable and IANA forms case-insensitively. This is not translated city naming or RTL certification. Persisted inventory timezone intentionally stays fixed during travel. |
| Targets, accessibility | Choice rows declare radio/checked/disabled and minimum52-point height; switches are named. Native menu and header Save use their shared adapters. Actual hit regions, VoiceOver order, error announcements and focus return remain unverified. |
| Keyboard | Custom input uses number-pad and accepts integer0–3650, rejecting decimal/negative/non-numeric values. Parent permits handled taps, keyboard dismissal and adjusted insets. Verify the native Save is reachable while editing. |
| Motion | No custom editor transition or decorative motion. System navigation/search, progress and Reduce Motion still require runtime checks. |
| Content, search | Timezone results are capped at30 with guidance to search further; saved zone, UTC and device zone are included in candidates. A valid exact identifier can be selected even if absent from Intl's enumerated list. No-match is explicit. This is bounded search, not a fully displayed/paginated directory. |
| Loading, recovery | Parent locks changes during requests. Initial load, retry, missing type and inline save errors have distinct branches. Mode failures offer Retry/Discard; timing retries the retained selection; timezone leaves the saved value intact. The M199 follow-up below corrects denied-refresh/save rendering with cached preferences. |
| Editing | Custom typing is local until Save; Back abandons the unsaved task. Presets commit on selection. Pending refs reject duplicates; failed saves preserve retry state. Retained drafts are in-memory, not process-death recovery. |
| Privacy | Scoped route key includes service-state scope, tenant and inventory; mismatched expected inventory is rejected. M199 verifies UI data retirement through controlled HTTP denial responses. Backend authorization enforcement is not established by these screen tests. |
| Notifications, media | Editors change personal reminder preferences, not physical APNs delivery. Device permission/setup is the separate S116 task. No camera/audio/library interaction originates in these editors. |
| Lifecycle | Child presentation ownership suppresses late errors/completion after focus departure; parent aborts departed requests and reloads on focus. Tests cover returning before settlement and a fresh retry. OS suspension, process termination and external notification interruption remain runtime work. |

Remote verification on paul: **29 tests across7 files**, covering reminder mode,
timing, timezone, retained task presentation, real settings parent/session,
destination validation and preference session. Log: `/tmp/reminder-editor-audit.log`.
No implementation changed in this pass. These tests do not resolve existing
native findings or establish full settings authorization acceptance.

## M199 follow-up — cached settings after access loss

The separate adversarial follow-up reproduced editable defaults remaining after
401/403 refresh or save failures, and a denied supporting-type read. The screen
now removes preferences/types and blocks further saves until an authorized reload.
Ordinary transient errors retain the editor; after denial, transient retry errors
cannot restore it. Retained callbacks cannot write during denied recovery.

Five cases cover actual API client/notification adapter HTTP401/403 reads and
writes, plus a typed supporting-query denial, followed by failed retry and
successful authorized reload/save. The initial six RED cases included a redundant
supporting-query case, consolidated before GREEN. All36 related tests, TypeScript
and structural checks pass remotely on paul (`/tmp/reminder-access-green.log`).
This does not prove server authorization enforcement or native focus/announcement
after the editable content is removed. Those remain separate acceptance work.
