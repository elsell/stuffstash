# Findings and remediation tracker

| ID | Finding | Status | Evidence / next verification |
| --- | --- | --- | --- |
| M01 | Short filter choices cause unnecessary drilldown | Implemented; runtime pending | Existing audit F01; in-place choices, draft/apply/cancel tests and iOS/Android adapter contracts pass remotely; full suite 1319 plus 3 added adapter tests; critic found no draft/navigation blocker |
| M02 | Custom field short choices use bespoke disclosure radios | Implemented; runtime pending | Shared native Type/Applies to menus, read-only transition guard; 34 focused tests and typecheck/structural; critic found no blockers |
| M03 | Redundant exact-date staging | Open | F03; preserve precision and parent draft semantics |
| M04 | Standard header actions remain custom on some screens | Open | F04; verify native adapters and toolbar consumers |
| M05 | Background queries control pull indicators; some gesture owners retain state across blur | Implemented; runtime pending | Shared focus-aware lifecycle across all refresh owners; real query-cache and inbox blur tests; 1324 remote tests/typecheck/structural green; critic found no blockers |
| M06 | Notice motion/timing/targets need accessibility adaptation | Open | F06; persistent recovery and preference-aware feedback |
| M07 | Switcher lacks bounded scroll/explicit dismissal and identifies households by name | Implemented; runtime pending | Bounded sheet/native Close; identity collision, safe retry, duplicate and late navigation tests; full1329 tests/check/structural green; critic found no blockers |
| M08 | Adaptive/assistive-tech runtime matrix unverified | Investigating access | F09; phone/iPad/Android runtime needed |
| M09 | Expiration location labels omit ancestry | Implemented; runtime pending | Authorized active-tree path labels, partial paths and duplicate-name fixture; full1331 tests/structural plus typecheck green; critic found no blockers |

| M10 | Expiration tag multi-selection is exposed as radio buttons | Implemented; runtime pending | Checkbox semantics with select2/remove1/apply regression; 4 focused tests/check green; critic found no blockers |
| M11 | Newly selected custom-field applicability targets cannot be removed before saving | Investigating contract | CustomizationEditorFields only appends targets and displays selected targets as static Existing text, including unsaved additions; distinguish immutable saved scope from editable draft |
| M12 | Replacement notices can inherit the prior notice timer/animation lifecycle | Open | AppFeedback reuses unkeyed AppNotice and callback clears whichever notice is current; old dismissal must not remove newer feedback |
| M13 | Initial asset/location list load errors have no in-place retry | Open | InventoryAssetsRouteScreen/LocationAssetsRouteScreen/LocationsScreen ErrorState renders text only; preserve context and offer retry |

The prior web draft finding is outside this mobile-only task. This list is a seed;
the full surface/axis review must discover and track further findings.
