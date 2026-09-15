# Comprehensive mobile UI audit and remediation

Active goal, started 2026-09-14 at source revision 12b5cb5a.

Inventory: **139 route/layout and nested-task surfaces × 24 axes = 3336 review cells**
(initial inventory: 132 surfaces; four asset editing/moving and three Home return subtasks added during source inspection).
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


Interim release delivered: **0.24.11 (99.1)**, processed with TestFlight changelog
verified on September 15 at 00:53 UTC. [Release workflow](https://github.com/elsell/stuffstash/actions/runs/34913014534).
The comprehensive audit remains active. Photo-removal and gallery follow-up fixes
were merged in PR129 for a second interim release, v0.24.12. Build **100.1** uploaded at01:55:59UTC; Apple processing and the exact-build
TestFlight changelog were verified at01:58:20UTC in
[release34917914704](https://github.com/elsell/stuffstash/actions/runs/34917914704). Shared-header and Home
return ownership/permission fixes continue separately in draft PR131.
