# Move Destination Selection

Status: M270 creation-name separation, shared choice rows and native search are
implemented and source-verified. Connected Android native acceptance passed;
iPhone/iPad run35888124430 remains pending.
Keep separate from frozen M260–M264 and the M269 native acceptance run.

## Task and pattern

Move chooses a destination and then performs an explicit mutation. Preserve its
native Cancel/Move actions and full-height navigation. Selection alone must not
move the asset. The M262 creation-order correction remains valid.

The current Put in field serves two different tasks: searching existing places
and naming a new destination. Its change handler closes the creation disclosure.
Consequently, correcting a name while creating also hides the Kind/Create controls.
The selected row uses a custom highlighted block and trailing Selected text,
where Add uses the shared checkmarked choice rows and native search. Source and
normal-text iPad capture from run35874141371 confirm the presentation difference;
the creation-edit collapse is source-confirmed, not yet a native reproduction.

Use the same native search and choice-list vocabulary as Add. Follow the existing
platform-interaction decisions and mobile-add-location-selection spec. This is a
project consistency decision based on the same searchable hierarchical task;
small flat choices such as Kind continue to use the in-place native picker.

## Required behavior

- Existing-destination search belongs to NativeNavigationSearch on iOS and the
  existing Android search adapter. Retain the current destination and any selected
  proposal when changing, clearing or closing search. Search starts collapsed on
  iOS and does not consume a permanent form row.
- Use the shared checkmarked selection rows with path/type context. Preserve the
  explicit inventory-root choice, eligibility explanations, pending lock and
  current/proposed location feedback. Keep the Move mutation separate from choice.
- New destination remains an explicit secondary task. Give its name a dedicated
  draft field, initially seeded from the search query. Editing the name keeps the
  creation task and Kind control present. It must not edit the search query or
  silently replace the current selected destination.
- Validate the actual proposed creation name against a current authorized lookup
  before enabling creation; a different search query's results are insufficient.
  Keep loading, error, retry and duplicate-name behavior distinct.
- Cancel new destination returns to the existing choices with search and selection
  unchanged. Preserve the existing Kind preference; reopen with a fresh proposed
  name seeded from the current search. Creation failure retains name/kind for retry;
  success selects the created destination, shows its title in search, and closes
  the creation task. Retained Create events use committed current draft and lookup
  availability; canceled, pending, disabled and completed creation reject them.
- Reuse the existing command port, independent creation boundary, operation lock,
  committed-current callbacks, and stable native removal behavior. Scope/permission
  changes or a departed owner reject late interaction and presentation callbacks.

## Acceptance

First verify the connected normal-text workflow: search, select, open creation,
correct its name without losing Kind/Create, cancel, reopen, reject creation,
retry, reject Move, retry and return. Check that name/query/selection semantics are
preserved and that no mutation occurs merely by searching or selecting.

Use focused mounted tests through controlled ports, representative Add regression
coverage for reused controls, code review, then phone/iPad/Android native acceptance.
Retain existing typed-input and rejected-command evidence; do not reopen keyboard
provider experiments unless a new failure distinguishes a specific product cause.
Judge task continuity and action reachability before detailed edge cases.

## Current evidence

The mounted name-edit scenario first failed without a dedicated name field. The
retained Create regression then reproduced submission of the previous name while
the current name lookup was pending. Both pass after separate draft ownership and
the shared focused action guard. All 51 focused Move/Edit/Add selection tests, TypeScript
and mobile structural checks pass on the remote Linux validation host. Code critic
cleared the source corrections. Native acceptance uses the shared radio checked
accessibility value and intended-key keyboard readiness, retaining exact name and
command payload assertions. Shared choice acceptance verifies single selection, no
mutation on selection, retired callbacks after a row disappears, explicit Move,
and retained selection after rejection. This does not establish native presentation
or complete M270.

## Visual redesign after user rejection

Functional native run35895924050 does not establish design acceptance. The user
rejected the flat text hierarchy, unexplained row indentation and floating text
creation command. Current MoveSelectionRow wraps a custom SettingsChoiceRow inside
formScrollContent: both own horizontal inset. Replace this composition.

For iOS Move and Move Here, use the pinned Expo UI SwiftUI List and Section
primitives as the scrolling body, with one owner of native row insets/separators.
Show a compact subject/current-location section and a separately labelled choices
section. Candidate rows use an appropriate SF Symbol, primary name, secondary
location path, and trailing selection checkmark. Selecting proposes the move;
the persistent native confirmation command commits it. Keep an existing selected
candidate understandable through search refinement without redundant free-floating
Selected paragraphs. Preserve disabled candidates and their reasons.

Native search uses a deliberate stacked navigation placement for these selection
tasks on both phone and iPad. Browse retains its existing compact search icon;
this is task-specific, not a global search redesign. The pinned screens adapter
supports stacked placement and disables toolbar integration for it. Compare
matching idle/search/selected states across devices.

Move creation is a secondary native toolbar action, not a centered body text
button. Entering creation should present one coherent bounded form with native
Cancel/Create commands; do not leave competing Move controls active or scatter
creation below the destination list. Preserve independent creation name, kind,
placement explanation, validation, rejection/retry and return with the created
destination selected. No new dependency is required: pinned Expo UI supports
List/Section/HStack/VStack/Image/Text/Button and RNHostView for existing status
content where necessary. Validate navigation safe areas and bridged content on a
native runtime before adopting it; source support is not rendering evidence.

Android follows its platform list/search/action conventions and the same information
hierarchy; do not inject SwiftUI or copy iOS appearance into Android. Acceptance
requires retained functional scenarios plus matching normal-text screenshots of
entry, search, selected destination, creation and recovery judged as a whole.

Android subject summary uses the asset name above its secondary current-location
text, aligned with the choice text. It is descriptive context, not a setting's
label/value row; long names and paths must wrap independently without competing
for horizontal space. Android runtime review of bca78c94 exposed this distinction.

### Native list inset and status ownership

Native run35901635030 exposed a second inset from the old editor wrapper,
zero-gap icon/text composition, and a clipped RN status view inside a SwiftUI
list cell. The selection body must occupy the full available sheet width; the
system List alone owns grouped-list margins. Keep form padding only for creation
and the Android editor. Specify a readable icon/text gap and a smaller title/path
gap rather than relying on the bridge's zero-spacing defaults.

Render typed recovery content inside the native list using native text/buttons.
Do not use a React Native sibling above the list or an intrinsically measured
RNHostView inside its cells: the former falls behind native headers and the latter
can exceed the cell width. Preserve retry focus guards and command behavior.
Capture phone and iPad recovery as well as idle selection before acceptance.

The entire candidate row, including empty space between its text and checkmark,
must select the candidate. The pinned SwiftUI plain button needs an explicit
rectangular content shape on its label stack. The iPad center-row taps in the
existing connected native scenarios are the regression check; do not move those
taps onto the text to hide a deficient touch target.

## Status ownership correction

Run35907123046 phone captures confirm the React Native status sibling renders
under the native header; Retry suggestions has no hittable point. Replace the
arbitrary ReactNode status slot with typed message/retry data rendered inside
the platform list. On iOS use SwiftUI text and buttons so List owns width, safe
area, scrolling and hit testing. Android retains inline recovery in its scroll
body. Preserve current retry handlers, lock/focus guards and read-only/busy/empty
messages. Do not add header-height padding or another unconstrained RNHostView.
The connected move/detail/reopen case passes; creation search shows exact Audit
in the failure capture, so its five-second observation timeout is separate from
this demonstrated status-placement defect.

Reuse the existing bounded 30-second exact-text observation, with elapsed timing,
for the demonstrated Move creation-search observation timeout. Retain exact Audit
and all subsequent creation/retry assertions; do not retype or change providers.

The same run's iPad creation case stops before typing: its sole readiness check
took4.23 seconds and returned false before the five-second deadline. Recorded
post-failure state confirms finite, hittable keyboard and t key. Permit the same
bounded30-second readiness observation for this case, retaining all geometry,
hittability and exact-text checks. Leave other readiness deadlines unchanged.


## Creation form header ownership

Run35911803207 confirms the iPhone new-destination field starts at y99 while the
native header ends at y116. The focused React Native creation form must reserve
the current measured native header height on iOS, with automatic scroll content
insets disabled so there is one inset owner. Keep SwiftUI selection-list inset
ownership unchanged. The form must remain reachable after choosing Kind and
returning from cancellation, with keyboard avoidance and retained draft intact.
No hard-coded device/header height is allowed. Existing native creation checks
must reach and edit the name, create after a rejected attempt, and return to Move.
For stacked Move Here search, test cleanup clears the search field directly;
it must not use an unscoped Cancel lookup that can dismiss the task itself.

Move Here keyboard cleanup observes either native dismissal after clearing or a
hittable dismissal command, then still requires the keyboard to disappear. The
iPad failure in run35911803207 captured no keyboard immediately after an earlier
existence snapshot; this is a transition race, not evidence of a missing command.
