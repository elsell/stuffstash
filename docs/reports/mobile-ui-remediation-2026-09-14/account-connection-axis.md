# Account and connection recovery audit

September 15 source/controlled-render review covers root Settings (R035), Account
(R024) and Connection (R026), with partial comparison of About (R023) and Diagnostics
(R027). Native phone/iPad, accessibility and Android acceptance remain pending.

## Task and pattern fit

Account and server switching are commands with different consequences, not a
single value picker. Each belongs on its own settings destination. Existing native
confirmation alerts explain the retained server on sign-out and the local reset
on changing servers; neither claims to delete inventory data. Cancel is explicit.
Pending guards prevent repeated execution. Keep this separation and confirmation.

Root Settings is navigation. Appearance is already an in-place choice. Grouped
Account/Connection rows lead to tasks, while household/inventory rows depend on
scoped access. About and Connection read local diagnostics. Diagnostics combines
identity and scope and remains dependent on those reads; this is not needed for
sign-out recovery.

## M73 — inventory failure prevents account recovery (P1)

Account reused SettingsModelScreen, which waits for selected-inventory data. A
missing, denied or offline inventory therefore replaced Sign Out with a loading
or error view. Root Settings similarly withheld both Account and Connection until
scope was ready. Two mounted regressions reproduced unavailable recovery links
and blocked sign-out.

Root's loading/error states now retain Account and Stuff Stash server navigation.
Account reads only the principal label, falls back to Current account, and always
exposes confirmation for Sign Out. Principal failure offers retry without removing
the command. Existing authorization for inventory administration is unchanged.
This is task-dependency correction: leaving an account must not require its current
inventory to be readable. It is a project usability inference, not a quoted Apple
requirement about this architecture.

Evidence: 27 remote settings/cache checks, TypeScript and structural checks pass on
paul. Critic found no blocker and requested pending/failed identity cases; both now
verify fallback sign-out. Existing mounted checks retain scoped-denial behavior,
confirmation, pending/error recovery and cache ownership. This is not new proof of
server authorization or native confirmation rendering.

## Other axes and remaining scenarios

Settings rows use shared scalable text, minimum targets and stacked large-text
layout; inspect native narrow/iPad/long identity and URL values before accepting
geometry. Loading/error content scrolls. These screens have no media, search or
notification task; sharing native adapters elsewhere does not make those features
required here. Light/dark, contrast, RTL, VoiceOver/TalkBack order and keyboard
navigation still need runtime coverage. Native screen return after failed teardown
and focus ownership of delayed error notices remain open lifecycle checks.

The account command still delegates to the existing push-session disconnect and
onboarding teardown. Source review found performance disposal precedes asynchronous
profile reset; failure recovery of that resource lifetime needs a separate
engineering review. No implementation change or defect claim is made for it here.

## September15 — About and Diagnostics

Sourceb8d18f5b, R023/R027. Inspected AboutSettingsScreen,
DiagnosticsSettingsScreen, SettingsModelScreen and SettingsQuery. About is a
read-only product/version destination backed by synchronous local diagnostics.
It does not need an inventory query, save action, chooser or modal confirmation.
Keep its scope small; a tutorial or setup flow is not implied by an About screen.

Diagnostics is an inspection task: grouped connection, identity and version rows,
with no inline edits. It intentionally reads principal and selected scope to show
those identifiers. Pending state shows an indicator; initial failure gives a
scrollable error with Retry; a failed background refresh keeps data with a refresh
notice. This differs from account recovery, which M73 decoupled from inventory
availability. Diagnostic unavailability does not justify blocking Sign Out.

SettingsQuery explicitly maps principal ID, selected tenant, URL, auth mode and
version into the view model. No password/token field is rendered by these screens;
this is source-field inspection, not a claim that arbitrary configured URL values
are safe or a substitute for authentication/authorization testing. The screen
uses developer identifiers because its task is diagnostics; ordinary inventory
flows should retain household/inventory language.

Both ready views scroll and use shared value rows; the Diagnostics error also
scrolls. Long URLs/identifiers, copied diagnostic values, VoiceOver grouping,
Dynamic Type, contrast and window adaptation need runtime review. No copy/export
command is currently supplied; assess that as a task enhancement rather than
inventing a compliance requirement. The load Retry still uses a custom command;
include it with the shared native-command follow-through. These screens contain
no search or media editor, but system notification interruption remains possible.
Do not mark global lifecycle axes N/A merely because their content is read-only.


## Follow-up review at 570b804c

Current follow-up at8599d4b3 covers R024/R026 across every axis below. Re-read both
route wrappers, AccountSettingsScreen/ConnectionSettingsScreen and confirmation
ownership helpers. The historical custom-command observations below are superseded:
Sign Out and Change Server now use NativeCommandButton. Native route Back remains
available; neither task needs a value-selection menu or a draft editor. Both
commands explain their effect in a native confirmation with Cancel, reject reused
or departed acceptance, and prevent duplicate execution while working. Failure
restores command availability and only reports in its originating visit.

Account continues to offer sign-out independently of inventory availability;
Connection reads the injected server diagnostics without a network prerequisite.
The displayed values remain read-only and selectable. These routes add no media,
notification, search or editable-selection task. Current source establishes no
new defect. The table's native layout, long-value, focus, contrast and lifecycle
gaps remain; the full1,849-test checkpoint is not native session-transition proof.


R024/R026, source570b804c plus M141. Inspected route wiring, both screens,
SettingsQuery, SettingsList, SettingsRefreshNotice and shared styles. This is a
source review across all24 axes; runtime-sensitive cells remain pending.

| Axis | Source result and remaining evidence |
| --- | --- |
| Task | Account identifies the principal and signs out; Connection identifies the server and changes it. These are commands, not value selectors. Custom action-row rendering still needs native-adapter review (M142). |
| Navigation | Thin routes inject service actions; native stack owns Back. Confirmations retain their originating visit (M140). Actual Back/return remains pending. |
| Selection | No selectable setting or choice list. Displayed email and server address are read-only values. |
| Modality | Explicit native alerts explain sign-out/server-change consequences and provide Cancel. M140 rejects obsolete/reused acceptance. Native dismissal pending. |
| Layout | ScrollView contains grouped rows and explanatory footers; no fixed footer. Actual safe-area, bottom reachability and iPad widths pending. |
| Adaptation | Shared rows switch layout using font scale; native narrow/large-text/tablet evidence pending. Normal-size checks first. |
| Typography | Values shrink and are selectable. Long email/URL wrapping needs runtime verification; source rules alone do not close M33. |
| Appearance | Palette colors and shared disabled opacity. Native contrast/material appearance pending. |
| Localization | English prose; serverHostname derives display host while full URL remains available. Long strings, RTL and localization coverage pending. |
| Imagery | No screen-specific media or imagery; native navigation icons remain shared-shell work. |
| Targets | Action-row minimum52×44 is declared. Native bounds, focus and touch activation remain pending. |
| Gestures | No required hidden gesture. Explicit command and native Cancel/Back; interactive dismissal remains pending. |
| Keyboard | No editable field. Selectable values support copying; hardware focus and navigation remain pending. |
| Accessibility | Action rows declare button, busy and disabled; value rows have combined labels. Actual traversal, duplication and announcements pending. |
| Motion | No local animation; shared alert/notice transitions require platform checks. |
| Content | Small static groups. Account reads principal separately from inventory; Connection uses injected diagnostics. No pagination. |
| Search | Neither screen owns search or filtering. |
| Loading | Sign-out remains available when identity is pending/failed. Pending command prevents duplicate execution. Loading account identity falls back to Current account. |
| Recovery | M140 suppresses departed action errors and preserves retry. M141 corrects initial identity error copy; retained-data refresh copy stays unchanged. |
| Editing | No field draft. Single-use confirmations and current command eligibility are mounted-tested; successful session transition remains an integration gate. |
| Privacy | Service actions own session changes. Confirmation states what is kept/forgotten; no data-deletion promise beyond existing contract. No authorization rule is changed by M140/M141. |
| Notifications | No notification preference or push handler on these screens; interruption is shared lifecycle work. |
| Media | No acquisition, upload, playback or export action. |
| Lifecycle | Query/identity and focused visit own confirmation/error feedback. Cold start, system background and session transition still require native/integration evidence. |

M141 mounted recovery: initial principal read fails, accurate unavailable-details
copy is shown, sign-out remains present, retry loads principal and removes error.
61 settings tests, TypeScript and structural checks pass remotely. Native acceptance
is not inferred from these tests. The shared refresh notice's other callers keep
the same default message and behavior.
