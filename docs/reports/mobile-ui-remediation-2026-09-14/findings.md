# Findings and remediation tracker

| ID | Finding | Status | Evidence / next verification |
| --- | --- | --- | --- |
| M01 | Short filter choices cause unnecessary drilldown | Implemented; runtime pending | Existing audit F01; in-place choices, draft/apply/cancel tests and iOS/Android adapter contracts pass remotely; full suite 1319 plus 3 added adapter tests; critic found no draft/navigation blocker |
| M02 | Inconsistent custom/native choice controls | Open | F02; platform adapter and consumers |
| M03 | Redundant exact-date staging | Open | F03; preserve precision and parent draft semantics |
| M04 | Standard header actions remain custom on some screens | Open | F04; verify native adapters and toolbar consumers |
| M05 | Background queries control pull indicators outside Home | Open | F05; scan all RefreshControl/FlatList consumers |
| M06 | Notice motion/timing/targets need accessibility adaptation | Open | F06; persistent recovery and preference-aware feedback |
| M07 | Switcher lacks bounded scroll/explicit dismissal | Open | F07; also investigate selection failure and duplicate household names |
| M08 | Adaptive/assistive-tech runtime matrix unverified | Investigating access | F09; phone/iPad/Android runtime needed |
| M09 | Expiration location labels omit ancestry | Open | F10; duplicate-name selection fixture |

The prior web draft finding is outside this mobile-only task. This list is a seed;
the full surface/axis review must discover and track further findings.
