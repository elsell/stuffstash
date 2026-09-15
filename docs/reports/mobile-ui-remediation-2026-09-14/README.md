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

The audit remains incomplete. After the switcher source review based on064fbaf9, the 3384 cells comprise 2808
pending, 530 source-reviewed, 28 finding, 14 runtime-partial and 4 not-applicable.
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

Next remediation batch: [PR140](https://github.com/elsell/stuffstash/pull/140), branch
`codex/mobile-audit-after-138`. M69–M73 cover Sharing feedback ownership, invitation
replacement/startup/navigation recovery and account recovery without inventory data.
See [Sharing/invitations](sharing-axis.md) and
[Account/connection](account-connection-axis.md). These changes are excluded from0.24.17.

Latest completed native slice: iPhone17 onboarding in
[run34929746647](https://github.com/elsell/stuffstash/actions/runs/34929746647),
actual sourcebabf6765 (parents50b598ae and904684a1). One applicable test passed;
two iPad-only tests skipped. The phone fixture job then completed with22 passes
and9 failures; iPad fixtures completed27 passes and4 failures. iPad onboarding
remains active. The run predates PR138 and PR140; it cannot verify their changes.

Priority open work: native Add loading diagnosis, filter keyboard hit targets,
large-text choice reflow, history accessibility reachability, Android runtime
coverage and the remaining surface/axis reviews. Preserve running native jobs;
collect their screenshots and hierarchies before selecting further fixes.

Earlier delivery evidence remains in [native-evidence.md](native-evidence.md) and
[findings.md](findings.md). No historical release is full audit acceptance.
