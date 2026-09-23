# Mobile Selection Batch

## Frozen scope

This follow-up batch contains M265–M273: clearer root Map and empty-photo detail
hierarchy; consistent optional Add tag creation; searchable Add/Edit tag selection;
draft-safe Add destination selection; independent Move search and creation drafts;
native draft reappearance retention; visible selection above modal owners; and
Move Here using the same native selection structure. Do not add unrelated audit
findings while acceptance runs.

M260–M264 already merged as690ee8e4 (PR171) and is releasing separately. This branch
integrates that base's accepted Browse outer-card measurements and gutter gesture.
The only production change brought forward during integration is the AssetCard
layout test identifier; follow-up production behavior remains the reviewed4317a261.

## Current acceptance

Full mobile integration:1968 of1970 tests passed; the two failures enforced removed
inline Add picker structure and hidden Move Here header. Updated route/native-task
expectations pass with54 other tests in those files. TypeScript, mobile structural
checks and10 fixture-preparation tests pass. Critic found no merge-resolution
blocker. Source/mounted evidence is not native runtime acceptance.

M265/M266 have scoped phone/iPad acceptance. Android connected Add destination,
Move creation and Move Here retry/return checks passed; retained screenshots were
inspected. M271/M272 grouped Add/Edit selection native run is pending, as is the
corrected Move/Move Here native run. Their current specs retain exact acceptance
requirements and known diagnoses. Do not merge this batch until its changed native
workflows and required CI pass. Real backend, physical-device and Android-tablet
evidence remains distinct from fixture/simulator checks.
