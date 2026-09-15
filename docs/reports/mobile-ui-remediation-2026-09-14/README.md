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
- `checkout-history-axis.md`: all24 axes, independent name recovery and access-retry evidence limits.
- `provider-editors-axis.md`: all24 axes for credential/prompt editing, native commands and draft protection.
- `add-tags-axis.md`: tag discovery, scoped draft preservation and all24 review axes.
- `edit-tags-axis.md`: all 24 axes for tag selection, draft creation and remaining native acceptance.
- `contained-items-axis.md`: scoped search, shared detail controls, unknown-data states and all24 review axes.
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

The latest batch includes focused-screen notices, reminder and asset-command visit ownership, independent checkout-history name recovery, and measured filter action clearance. All 1,624 mobile tests (265 files), TypeScript and structural checks passed remotely against source 79412f8d; native geometry and lifecycle acceptance remain open.

The audit remains incomplete. After the Expiration filter source review, the 3384 cells
comprise 2198 pending, 911 source-reviewed, 234 finding,
19 runtime-partial and 22 not-applicable.
The [global-notice review](global-notice-axis.md) covers all 24 axes and links the
36-call-site consumer inventory. These are evidence states, not a compliance score. Finding cells can include
implemented corrections whose native acceptance is still pending.

Previous delivered TestFlight checkpoint: **0.24.16 (104.1)** from5775da93.
Apple processing and the exact-build changelog were verified at05:10:25UTC in
[release34930161409](https://github.com/elsell/stuffstash/actions/runs/34930161409).

Historical interim release: **0.24.17**, source177c08b6 (PR138). Validation and release
publishing succeeded in
[release34932422663](https://github.com/elsell/stuffstash/actions/runs/34932422663).
Build105.1 uploaded successfully at05:49:31UTC. Apple processing and the exact-build
TestFlight changelog were verified at05:52:22UTC: **0.24.17(105.1) is delivered**. This release contains filter badge contrast,
history locale, month-calendar presentation and keyboard-ownership corrections.
It does not certify the full audit or unresolved native footer behavior.

PR140 merged as `eca1ad7e`. Its interim release completed in
[release34939488611](https://github.com/elsell/stuffstash/actions/runs/34939488611).
This checkpoint improves onboarding, inventory switching and account/invitation
recovery. **0.24.18 (106.1) is delivered.** Upload succeeded at07:37:33 UTC; Apple
processing and the exact-build TestFlight changelog were verified at07:39:56 UTC
on September15. The later PR142 Settings changes are not included.

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

Asset command review: [overflow, checkout/return and lifecycle actions](asset-actions-axis.md).

Historical delivered checkpoint: **0.24.20 (108.2)**, PR144 at aecaeedc. Upload
succeeded at10:16:30UTC and exact TestFlight changelog verification completed
at10:18:59UTC on September15 in
[release34954415338](https://github.com/elsell/stuffstash/actions/runs/34954415338),
attempt2. The first attempt hit a GitHub tag-push server error before upload.
PR146's asset-form and suggestion-recovery changes are not in this release and
still require native acceptance. Earlier0.24.19 delivery is in the evidence log.

Move source review: [destination selection, creation and Move here](move-axis.md).

PR146 merged as `de5d87b0`. Its interim release completed in
[release34958198826](https://github.com/elsell/stuffstash/actions/runs/34958198826).
Required checks and final code review passed. This checkpoint includes native form
actions, Move reflow, suggestion recovery and Edit tag-name feedback. **0.24.21 (109.1) is delivered.** Upload succeeded at11:00:34 UTC; Apple
processing and exact-build changelog verification succeeded at11:02:58 UTC. All later PR148 audit corrections are excluded.

PR148 validation at a761b0d8: the complete mobile suite passed **1,512 tests across
258 files** remotely on paul. Changed mobile/script files match the checked
workspace by SHA-256. This expands the focused behavior evidence; native runs
remain separate and do not yet cover the latest Add/Edit corrections. No local
tests or builds were run.

Latest delivered checkpoint: **0.24.22 (110.1)** from PR148 at `4b5f7f89`.
Upload succeeded at 11:47:59 UTC and the exact TestFlight changelog was verified at
11:50:22 UTC. This contains Add/Edit tag drafts/discovery and Edit title scrolling.
The larger PR150 batch remains unreleased; native acceptance and the comprehensive
audit remain open. See [release evidence](native-evidence.md).

Tab and nested stack source review: [all 24 shell axes](tab-shell-axis.md).

Voice entry/status control: [all 24 accessory axes](voice-accessory-axis.md).

Browse filter overview, tags and expiration handoff: [all 24 axes](browse-filters-axis.md).

The [Expiration filter review](expiration-filters-axis.md) covers R017 and S075–S079 across all24 axes. M117 corrects the persistent-search inconsistency in source; native acceptance remains pending.

[Expiration results](expiration-results-axis.md) now has source review across all24 axes, with M119 recovery findings still open.

[Containment Map and path search](map-axis.md) now have all24 source review axes; M125–M126 remain open normal-size findings.

[Appearance settings interaction](appearance-settings-axis.md) covers the detail route across all24 axes, with shared inline-picker findings and explicit native gaps.
