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

## Design acceptance hold

User review rejects the Move visual experience despite functional native passes.
Run35895924050 establishes operation and recovery only; it does not close visual
or whole-task acceptance. Hold this batch until the Move/Move Here layout has a
coherent hierarchy and normal-text phone/iPad review. Do not weaken acceptance or
ship the rejected design on the strength of green automation.

Review priorities: compact subject/current-location context, clearly grouped
and aligned destination list, identifiable places/containers, discoverable native
search, distinct selection/commit semantics, and a deliberate secondary creation
action. Avoid stacked ungrouped status paragraphs, unexplained cumulative insets,
and floating centered text commands. Native search is currently collapsed behind
an iPhone toolbar icon and expanded in the reviewed iPad capture; compare matching
states before claiming platform consistency. Reuse native list/toolbar patterns;
SettingsChoiceRow is a custom React Native row, not a system-native list.

### Creation form coherence

Run35917033325 iPad passes all four Move workflows, and reviewed selection/recovery
captures have aligned native sections and visible search. Its creation capture
still uses a blue custom panel, inconsistently inset explanatory text and very
heavy labels. Keep the functional result separate from this visual finding.
Replace the creation panel with the shared grouped form: a Name section, an
in-place Kind picker section with a single explanatory footer, and native
Cancel/Create. Use the same neutral background and row alignment as Add place.
Preserve independent naming, placement policy, pending guards and retry. Verify
name/header clearance and nonoverlapping Kind on phone/iPad, and inspect the whole
form; no standalone custom panel or new action navigation is needed.
