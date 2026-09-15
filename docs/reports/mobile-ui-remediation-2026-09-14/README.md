# Comprehensive mobile UI audit and remediation

Active goal, started 2026-09-14 at source revision 12b5cb5a.

Inventory: **136 route/layout and nested-task surfaces × 24 axes = 3264 review cells**
(initial inventory: 132 surfaces; four asset editing/moving subtasks added during source inspection).
This is a review worklist, not a count of completed checks. Overlapping shared tasks
are intentional: route coverage and interaction coverage are independent.

- `surfaces.json`: route and nested task enumeration.
- `axes.json`: named review dimensions.
- `matrix.csv`: per-cell state/evidence/finding tracking; initially pending.
- `findings.md`: confirmed findings and remediation evidence.

Runtime availability: macOS GitHub runners build and launch the genuine application
on iPhone and iPad simulators. The first native run failed; see `native-evidence.md`
for inspected screenshots and the distinction between procedure and app findings.
Android runtime remains unavailable. No local builds/tests, per session constraint.

A source-reviewed cell never implies a runtime pass. Add discovered internal
surfaces during inspection. Record justified N/A per cell; do not default missing
coverage to pass. This effort includes fixing findings and TestFlight release.

September 15 release checkpoint: the user requested an interim TestFlight build
with the current fixes, followed by continued audit work. Native sheet expansion,
Add launch diagnosis, keyboard scenarios, Android runtime coverage and remaining
review cells stay open; this release is not full audit acceptance.
