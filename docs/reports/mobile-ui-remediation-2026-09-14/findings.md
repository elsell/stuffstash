# Findings and remediation tracker

| ID | Finding | Status | Evidence / next verification |
| --- | --- | --- | --- |
| M01 | Short filter choices cause unnecessary drilldown | Implemented; runtime pending | Existing audit F01; in-place choices, draft/apply/cancel tests and iOS/Android adapter contracts pass remotely; full suite 1319 plus 3 added adapter tests; critic found no draft/navigation blocker |
| M02 | Custom field short choices use bespoke disclosure radios | Implemented; runtime pending | Shared native Type/Applies to menus, read-only transition guard; 34 focused tests and typecheck/structural; critic found no blockers |
| M03 | Redundant exact-date staging | Open | F03; preserve precision and parent draft semantics |
| M04 | Standard header actions remain custom on some screens | Open | F04; verify native adapters and toolbar consumers |
| M05 | Background queries control pull indicators; some gesture owners retain state across blur | Implemented; runtime pending | Shared focus-aware lifecycle across all refresh owners; real query-cache and inbox blur tests; 1324 remote tests/typecheck/structural green; critic found no blockers |
| M06 | Notice motion/timing/targets need accessibility adaptation | Implemented; runtime pending | Persistent actions/warnings/errors and screen-reader notices; live Reduce Motion, readable labels, 48-point controls and enlarged-text stacking; 8 focused tests/typecheck green; critic requested motion regression, added and passed |
| M07 | Switcher lacks bounded scroll/explicit dismissal and identifies households by name | Implemented; runtime pending | Bounded sheet/native Close; identity collision, safe retry, duplicate and late navigation tests; full1329 tests/check/structural green; critic found no blockers |
| M08 | Adaptive/assistive-tech runtime matrix unverified | Investigating access | F09; phone/iPad/Android runtime needed |
| M09 | Expiration location labels omit ancestry | Implemented; runtime pending | Authorized active-tree path labels, partial paths and duplicate-name fixture; full1331 tests/structural plus typecheck green; critic found no blockers |

| M10 | Expiration tag multi-selection is exposed as radio buttons | Implemented; runtime pending | Checkbox semantics with select2/remove1/apply regression; 4 focused tests/check green; critic found no blockers |
| M11 | Newly selected custom-field applicability targets cannot be removed before saving | Implemented; runtime pending | Saved targets stay immutable; draft checkbox choices can be deselected, including safe unavailable-draft removal; create/edit/scoped-name tests, 37 focused tests/check/structural green; critic found no further blocker |
| M12 | Replacement notices inherit the prior timer/animation lifecycle | Implemented; runtime pending | Monotonic identity, keyed lifecycle and originating-ID dismissal; two rendered regression tests/check green; critic found no blockers |
| M13 | Initial asset/location list load errors have no in-place retry | Implemented; native pending | Scoped Retry in both routed lists and legacy unrouted LocationsScreen; repeat-failure and scope-recovery tests; full1349/check/structural green; critic found no blocker |

| M14 | Native onboarding run shows a shortened typed server address | Investigating runtime evidence | Run34882515267 iPhone screenshot shows h.invalid after typing https://example.invalid; add explicit value assertion and reproduce before classifying simulator input vs app loss |

| M15 | Unsaved enum options cannot be removed before saving | Implemented; native rerun pending | Saved/draft distinction with native Remove command; 43 focused tests/check/structural green, critic found no blockers; native removal fixture added |
| M16 | Customization controls remain editable while Save is pending | Implemented; runtime pending | Pending-save inputs stay visible/disabled; open picker guarded; 53 focused tests including all editor kinds and failed-save recovery, check/structural green; critic found no blocker |

The prior web draft finding is outside this mobile-only task. This list is a seed;
the full surface/axis review must discover and track further findings.

| M17 | iPad onboarding stretches the form across the display with excessive separation from its action | Implemented; native rerun pending | Centered 600-point form column and adjacent action; typecheck/structural green, critic found no blockers; iPad landscape fixture added, enlarged text still pending |
| M18 | Native menu pickers omit visible field labels outside a SwiftUI Form | Implemented; native rerun pending | Run34887652455 Browse screenshot; shared LabeledContent wraps menu value; native test requires visible Availability label and in-place selection |
| M19 | Expiration filter sheet renders no body content | Open; phone failure persists | Run34887652455 iPhone screenshot and accessibility hierarchy contain footer only; run34897215957 disproves KAV cause; iPad passes, phone remains blank; detent comparison pending |
| M20 | Onboarding keyboard does not dismiss with downward content drag | Open | Run34887652455 iPhone preserves full typed URL but fails corrected downward dismissal; investigate actual gesture and scroll bounds before changing behavior |

Run34887652455 also passed the iPhone persistent actionable-feedback scenario.
This establishes retention and action reachability for that fixture, not full
assistive-technology or enlarged-text verification. Its iPad onboarding entry
still lost characters (`h//example.invalid`); M14 remains unresolved.
