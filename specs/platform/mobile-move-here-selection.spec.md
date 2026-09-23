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
