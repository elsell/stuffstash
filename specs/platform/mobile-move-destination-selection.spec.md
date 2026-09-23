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
