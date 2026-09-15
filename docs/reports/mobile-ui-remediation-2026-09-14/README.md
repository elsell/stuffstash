# Comprehensive mobile UI audit and remediation

Active goal, started 2026-09-14 at source revision 12b5cb5a.

Inventory: **141 route/layout and nested-task surfaces × 24 axes = 3384 review cells**
(retained customization completion added September 15; initial inventory: 132 surfaces; four asset editing/moving and three Home return subtasks added during source inspection).
This is a review worklist, not a count of completed checks. Overlapping shared tasks
are intentional: route coverage and interaction coverage are independent.

- `surfaces.json`: route and nested task enumeration.
- `axes.json`: named review dimensions.
- `matrix.csv`: per-cell state/evidence/finding tracking; initially pending.
- `findings.md`: confirmed findings and remediation evidence.
- `inventory-switcher-axis.md`: hierarchy, context changes, completion ownership and remaining controls.
- `onboarding-axis.md`: prerequisite task fit, editing/recovery and remaining native gates.
- `localization-axis.md`: date conventions, month-calendar semantics and directional-layout work.
- `appearance-axis.md`: shared appearance, materials, contrast evidence and remaining native checks.
- `text-input-sites.csv` and `text-entry-axis.md`: input ownership, external reset paths, and native acceptance work.

Runtime availability: macOS GitHub runners build and launch the genuine application
on iPhone and iPad simulators. The first native run failed; see `native-evidence.md`
for inspected screenshots and the distinction between procedure and app findings.
Android runtime remains unavailable. No local builds/tests, per session constraint.

A source-reviewed cell never implies a runtime pass. Add discovered internal
surfaces during inspection. Record justified N/A per cell; do not default missing
coverage to pass. This effort includes fixing findings and TestFlight release.

## Current checkpoint — September 15

The audit remains incomplete. After the About/Diagnostics source review based onb8d18f5b, the 3384 cells comprise 2798
pending, 540 source-reviewed, 28 finding, 14 runtime-partial and 4 not-applicable.
These are evidence states, not a compliance score. Finding cells can include
implemented corrections whose native acceptance is still pending.

Previous delivered TestFlight checkpoint: **0.24.16 (104.1)** from5775da93.
Apple processing and the exact-build changelog were verified at05:10:25UTC in
[release34930161409](https://github.com/elsell/stuffstash/actions/runs/34930161409).

Current interim release: **0.24.17**, source177c08b6 (PR138). Validation and release
publishing succeeded in
[release34932422663](https://github.com/elsell/stuffstash/actions/runs/34932422663).
Build105.1 uploaded successfully at05:49:31UTC. Apple processing and the exact-build
TestFlight changelog were verified at05:52:22UTC: **0.24.17(105.1) is delivered**. This release contains filter badge contrast,
history locale, month-calendar presentation and keyboard-ownership corrections.
It does not certify the full audit or unresolved native footer behavior.

PR140 merged as `eca1ad7e`. Its interim release is running in
[release34939488611](https://github.com/elsell/stuffstash/actions/runs/34939488611).
This checkpoint improves onboarding, inventory switching and account/invitation
recovery. Upload and Apple processing are not yet verified.

Native run 34937278231 tested merge `6076e824`, whose parents are177c08b6 and
b8d18f5b. The remaining PR140 commit765aa6cd changes only audit documentation.
iPhone onboarding passes one applicable scenario (two iPad-only skips); iPad
onboarding passes all three, including previously failing margin dismissal.
iPhone fixtures pass25/34 and iPad fixtures 28/34. The new switcher recovery
scenario passes on both. These results do not establish a fully passing native
audit. See the evidence log for unresolved failures and screenshot limitations.

Priority open work: Add typing/loading, History query diagnostics and interaction
acceptance, phone clipping and nested-sheet findings, iPad landscape screenshot
validation, Android runtime coverage, and remaining surface/axis reviews.

Earlier delivery evidence remains in [native-evidence.md](native-evidence.md) and
[findings.md](findings.md). No historical release is full audit acceptance.
