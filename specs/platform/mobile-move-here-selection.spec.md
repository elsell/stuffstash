# Move Here Selection

## M273: task structure and consistency

Move Here chooses an existing inventory asset for the current containing asset.
Source review confirms it still uses a partial-height iOS sheet, form-style search,
custom selected-row badges, a large arrow preview and footer commands, although
Move now uses a full-height selection task. These peer tasks should share the
same interaction vocabulary. This is a pattern-selection finding, not a claim
that the existing passed recovery checks failed.

Use a full-height iOS sheet with a native title and persistent Cancel/Move here
commands; Android retains stack presentation. Keep the containing destination
visible as context and the chosen item's name visible even when query refinement
hides its row. Reuse native header search on iOS and the existing Android search
adapter, plus checked single-choice rows with kind/path context. A large inventory
justifies searchable selection rather than a flat menu. Explicit Move here commits;
choosing a row, searching, clearing and cancelling do not mutate containment.
Preserve the selected item through search changes and failed submission. Busy,
read-only, focus and retired-row callback protections apply equally to both moves.
Do not change authorization, allowed-candidate filtering or command behavior.

This is a project interaction choice informed by Apple's
[toolbars](https://developer.apple.com/design/human-interface-guidelines/toolbars)
and [sheets](https://developer.apple.com/design/human-interface-guidelines/sheets)
guidance. Apple does not mandate full height for every task; the long searchable
collection and consistency with Move motivate it here.

## Acceptance

Verify a connected normal-text visit on phone/iPad and Android: search, select,
refine/clear while retaining the chosen item, rejected move with retry, then return.
Current native command callbacks must use the committed selection; removed-row
callbacks cannot silently replace it. Existing busy, permission and navigation
ownership tests remain critical regressions. Source tests do not prove native
search attachment, keyboard clearance or command hittability. Keep this follow-up
outside the already merged M260–M264 release.

The grouped move-selection-workflows run covers Move creation/retry and Move Here
lookup/retry/selection/refinement/clear/commit. Use known letter-key readiness,
not literal space (absent in retained iPad evidence). Preserve complete typed-query
and checked-state assertions, exact command fixture inputs and return checks.

Android normal-text installed candidate3f5d6e09 passes lookup retry, checked choice,
search refinement preserving selection, rejected Move with source/destination
retained and successful retry/return. The retained-state screenshot was inspected:
header actions and checked row are visible without overlap. This does not establish
iOS, tablet, physical-device or backend mutation acceptance. After clearing native
search, dismiss the keyboard only if it remains; iPad may already collapse search.
Keep the checked-selection and explicit Move assertions regardless.

## Native35891073312: bounded decision

Move's initial choice exists on both devices as an accessibility Other/radio,
not a Button. Match its semantic label across descendants and retain checked,
hittable and enabled assertions. This is an observation correction.

Move Here completed the entire rejected-command/retry/return path on iPad. Phone
stopped after clearing search because native integrated search remained active
with a visible Close control, so task Move was not yet exposed. Both recovery
checks likewise requested task Cancel while native search was still active.
Keep the user's compact integrated search pattern: use its Close/clear interaction
and then require task commands to return, preserving the chosen item. Do not
replace it with a permanently expanded field merely to satisfy those assertions.
Persistent task commands apply outside the native search submode. Apple's
[search placement API](https://developer.apple.com/documentation/uikit/uinavigationitem/searchbarplacement-swift.enum)
defines integratedButton as the integrated placement whose inactive state is a
button. Captures establish actual control availability; they do not certify the
remaining phone workflow or the Move creation steps that were never reached.

Retain exact query, checked state, failure/retry and successful return assertions.
One corrected grouped acceptance run follows these specific fixes; no provider,
keyboard, timing or production layout changes are justified by these failures.
