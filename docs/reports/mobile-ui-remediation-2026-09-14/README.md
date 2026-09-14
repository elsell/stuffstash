# Comprehensive mobile UI audit and remediation

Active goal, started 2026-09-14 at source revision 12b5cb5a.

Inventory: **132 route/layout and nested-task surfaces × 24 axes = 3168 initial review cells**.
This is a review worklist, not a count of completed checks. Overlapping shared tasks
are intentional: route coverage and interaction coverage are independent.

- `surfaces.json`: route and nested task enumeration.
- `axes.json`: named review dimensions.
- `matrix.csv`: per-cell state/evidence/finding tracking; initially pending.
- `findings.md`: confirmed findings and remediation evidence.

Runtime availability: validation host paul is Linux with adb installed; attached
Android devices and Mac/simulator access are being investigated. Prior CUA attempt
had no enabled surfaces. No local builds/tests, per session constraint.

A source-reviewed cell never implies a runtime pass. Add discovered internal
surfaces during inspection. Record justified N/A per cell; do not default missing
coverage to pass. This effort includes fixing findings and TestFlight release.
