# Findings and remediation tracker

| ID | Finding | Status | Evidence / next verification |
| --- | --- | --- | --- |
| M01 | Short filter choices cause unnecessary drilldown | Implemented; runtime pending | Existing audit F01; in-place choices, draft/apply/cancel tests and iOS/Android adapter contracts pass remotely; full suite 1319 plus 3 added adapter tests; critic found no draft/navigation blocker |
| M02 | Custom field short choices use bespoke disclosure radios | Implemented; runtime pending | Shared native Type/Applies to menus, read-only transition guard; 34 focused tests and typecheck/structural; critic found no blockers |
| M03 | Redundant exact-date staging | Implemented; native pending | Compact iOS picker edits parent draft directly; explicit Add initializes absent date; Android cancel remains non-mutating; 5 focused tests/check/structural, critic no blockers |
| M04 | Standard header actions remain custom on some screens | Implemented; native pending | Add, inbox and reminder timing use shared native bar items with disabled guards; 29 focused tests/check/structural; critic copy mismatch corrected |
| M05 | Background queries control pull indicators; some gesture owners retain state across blur | Implemented; runtime pending | Shared focus-aware lifecycle across all refresh owners; real query-cache and inbox blur tests; 1324 remote tests/typecheck/structural green; critic found no blockers |
| M06 | Notice motion/timing/targets need accessibility adaptation | Implemented; runtime pending | Persistent actions/warnings/errors and screen-reader notices; live Reduce Motion, readable labels, 48-point controls and enlarged-text stacking; 8 focused tests/typecheck green; critic requested motion regression, added and passed |
| M07 | Switcher lacks bounded scroll/explicit dismissal and identifies households by name | Implemented; runtime pending | Bounded sheet/native Close; identity collision, safe retry, duplicate and late navigation tests; full1329 tests/check/structural green; critic found no blockers |
| M08 | Adaptive/assistive-tech runtime matrix unverified | Investigating access | F09; phone/iPad/Android runtime needed |
| M09 | Expiration location labels omit ancestry | Implemented; runtime pending | Authorized active-tree path labels, partial paths and duplicate-name fixture; full1331 tests/structural plus typecheck green; critic found no blockers |
| M10 | Expiration tag multi-selection is exposed as radio buttons | Implemented; runtime pending | Checkbox semantics with select2/remove1/apply regression; 4 focused tests/check green; critic found no blockers |
| M11 | Newly selected custom-field applicability targets cannot be removed before saving | Implemented; runtime pending | Saved targets stay immutable; draft checkbox choices can be deselected, including safe unavailable-draft removal; create/edit/scoped-name tests, 37 focused tests/check/structural green; critic found no further blocker |
| M12 | Replacement notices inherit the prior timer/animation lifecycle | Implemented; runtime pending | Monotonic identity, keyed lifecycle and originating-ID dismissal; two rendered regression tests/check green; critic found no blockers |
| M13 | Initial asset/location list load errors have no in-place retry | Implemented; native pending | Scoped Retry in both routed lists and legacy unrouted LocationsScreen; repeat-failure and scope-recovery tests; full1349/check/structural green; critic found no blocker |
| M14 | Native onboarding run shows a shortened typed server address | Named native scenarios pass; broad acceptance pending | Run34920888328 system-address button and Go submission pass on phone/iPad; production phone onboarding passes. Production iPad launch failed before UI assertions. Older input comparison failures remain distinct from the new adapter |
| M15 | Unsaved enum options cannot be removed before saving | Implemented; native rerun pending | Saved/draft distinction with native Remove command; 43 focused tests/check/structural green, critic found no blockers; native removal fixture added |
| M16 | Customization controls remain editable while Save is pending | Implemented; runtime pending | Pending-save inputs stay visible/disabled; open picker guarded; 53 focused tests including all editor kinds and failed-save recovery, check/structural green; critic found no blocker |
| M17 | iPad onboarding stretches the form across the display with excessive separation from its action | Implemented; native rerun pending | Centered 600-point form column and adjacent action; typecheck/structural green, critic found no blockers; iPad landscape fixture added, enlarged text still pending |
| M18 | Native menu pickers omit visible field labels outside a SwiftUI Form | Implemented; native rerun pending | Run34887652455 Browse screenshot; shared LabeledContent wraps menu value; native test requires visible Availability label and in-place selection |
| M19 | Expiration filter sheet loses body/actions during native presentation | Expansion passes; phone keyboard actions unresolved | Run34920888328 direct-root candidate survives expansion on phone and iPad. Search keyboard action reachability passes on iPad but fails on phone. Large-text label finding tracked separately as M53 |
| M20 | Onboarding keyboard does not dismiss with downward content drag | Open | Run34887652455 iPhone preserves full typed URL but fails corrected downward dismissal; investigate actual gesture and scroll bounds before changing behavior |
| M21 | Add fields can change while the submitted item is being saved | Implemented; source tests pass; native pending | Exclusive save/parent/photo operation ownership guards draft edits, duplicate submission and dismissal. Five remote tests cover save failure, parent failure and photo cancellation with draft retention and editing recovery; native verification remains pending |
| M22 | Appearance uses navigation for three flat choices | Implemented; native menu scenario passes both devices | Settings now uses the shared native menu; older route reuses it. Immediate selection and storage-failure rollback are preserved; native menu rendering remains pending |
| M23 | iOS custom tag color requires an extra custom editor around the native picker | Implemented; 15 remote tests/check/structural pass; native pending | Available iOS native picker edits parent draft directly; swatches/Clear, disabled guards and Android/unavailable-native fallback preserved. Critic no blockers; native open/close/clear fixture added |
| M24 | Voice cancellation text is constrained to a 44-point icon box | Implemented; native pending | Single native record/send/cancel command replaces cramped Cancel text and redundant disabled Send; 48-point control, explicit accessible name and disabled guards. Three remote tests, typecheck and structural check pass; critic found no blocker. Native sizing and enlarged text remain pending |
| M25 | Standalone recording meter uses an on-action foreground against the surface | Implemented; check/structural pass; native pending | Both consumers inspected: standalone uses action foreground; inside-button meter explicitly uses onAction. Critic no blocker. Native light/dark contrast remains pending |
| M26 | Voice accessory accessibility label omits its changing status and context | Implemented; check/structural pass; native pending | Accessible name includes complete title and subtitle, preserving clipped status/context. Critic no blocker; screen-reader verification remains pending |

| M27 | Sharing uses custom access choices and editable pending invitation drafts | Implemented; native pending | Shared native Access menu, email/access freeze and stale callback guards. Remote behavior checks cover failure retention and existing scope isolation; native menu rendering remains pending |
| M28 | Invitation acceptance/opening buttons lose their accessible names while busy | Implemented; native pending | Stable Join/Open accessible names, busy/disabled states and visible progress text. Ten focused remote tests, typecheck and structural check pass; critic no blocker. Screen-reader runtime remains pending |

| M29 | Provider detail implies three operations run at once and allows competing editors | Implemented; native pending | Typed active operation identifies only its progress label; all competing actions and editor navigation disabled/guarded. Fifteen remote tests plus typecheck/structural pass; critic no blocker |

| M30 | Provider credential/prompt drafts remain editable during Save | Implemented; native pending | Native editable state and ref guards preserve submitted draft; failure restores editing, successful credential save clears secret. Seventeen remote tests plus check/structural pass; critic no blocker |

| M31 | Voice service choice uses custom action rows and conflates selection with test progress | Implemented; native pending | Shared native service picker retains current choice, excludes other archived services, and ignores same-value selection. Distinct select/test/enable progress with guarded competing navigation. Eighteen remote tests, check/structural pass; critic no blockers |

| M32 | Some settings errors and permission-denied states cannot scroll | Implemented; native pending | Scoped settings, shared customization denial and sharing error content now use ScrollView with growing content; focus/retry preserved. Initial 58 checks plus follow-through 61 existing tests cover collection/editor, notification and filter recovery. Check/structural pass; critic no blockers. Seven additional fallback consumers now scroll. Large-text runtime remains pending |

| M33 | Shared settings values do not shrink within horizontal rows | Implemented; native pending | Read-only and trailing navigation values can shrink/wrap within available width; stacked large-text layout retained. Both style consumers inspected, check/structural pass; critic no blockers. Native long-string verification remains pending |

| M34 | Time-zone search uses a plain custom field instead of native navigation search | Implemented; native pending | Shared native search preserves readable-city/IANA matching and cancellation without save. Two focused remote tests, check/structural pass; critic no blockers. Verify title integration and route cleanup natively |

| M35 | Onboarding Connect is largely covered by the phone keyboard | Open; native-confirmed in submission fixture | Run34906713382 retains full URL but Connect tap does not submit; inspected screenshot shows only a thin portion above the keyboard. Explicit dismissal can unblock command testing, but does not resolve keyboard layout |

| M36 | Type reminder mode uses custom choice rows for three flat values | Implemented; native pending | Shared native Reminders picker preserves inheritance, failed selection and Discard. Six focused remote tests, typecheck/structural pass; critic no blockers. Native Custom/defaults scenario added |

| M37 | Asset Move allows destination changes and Cancel during submission | Implemented; native pending | Shared synchronous operation guard freezes draft/Cancel/selection and blocks system Back; successful return explicitly allowed, unavoidable teardown suppresses late navigation. Full1372 remote tests (245 files), check/structural pass; critic removal finding fixed and rereviewed. Native dismissal verification remains pending |

| M38 | Move destination kind uses custom tab semantics for a form value | Implemented; native pending | Shared native Kind picker replaces tab semantics; create command receives selected kind. 52 remote tests/check/structural pass; critic no blockers. Native menu rendering remains pending |

| M39 | Asset Edit and Move text fields lack explicit accessible names | Implemented; native pending | Stable accessible names added to native text fields; route behavior tests now locate fields by those names. Twenty focused remote tests/check/structural pass; critic no blockers. VoiceOver/TalkBack verification pending |

| M40 | Tag and voice photo actions have undersized touch targets | Implemented; native pending | Shared48point minimum for tag choices/fields; native Add tag, photo Add/numberedRemove and Retry commands replace small targets. Photo previews and separate commands scroll in a rail. Twenty remote tests/check/structural pass; critic no blocker. Enlarged text and hit areas pending native verification |

| M41 | Edit tag resolution overwrites selected IDs with a second draft update | Implemented; native pending | One atomic tag callback updates selected IDs and pending definitions together. Regression reproduced normalized existing-tag selection disappearing; fix preserves selection through Save and unrelated description. Focused remote tests/check/structural pass; critic no blockers |

| M42 | Photo viewer permits another removal while deletion is pending | Implemented on continuation branch; native pending | Scoped synchronous deletion guard, disabled Remove/progress and selected-ID preservation. Duplicate-confirmation regression reproduced; full1377 remote tests/check/structural pass before final asset-switch case, then six focused route cases/check/structural pass. Covers changed photo index after reconciliation, retry, teardown, stale confirmation and old completion during new asset operation. Critic gaps addressed. Excluded from interim0.24.11 |

| M43 | Gallery Add photos uses a custom styled command despite an available native adapter | Implemented on continuation branch; native pending | Shared native command preserves permission/callback gating and separate placement below imagery. Eight remote gallery/route tests, typecheck and structural check pass; critic no blockers. Legacy mocked style snapshots replaced with mounted behavior checks; paging and appearance runtime remain pending. Not included in interim0.24.11 |

| M44 | History reversal completion can navigate after leaving its detail | Implemented on continuation branch; native pending | Deferred repository tests reproduced extra Back after blur/refocus, stale confirmation execution, and inherited activity busy state. Activity/focus ownership now gates presentation; completion still invalidates cache. Applied outcome prevents resubmission. Ten focused History tests/check/structural pass; critic finding addressed. Native Back/gesture acceptance pending; excluded from0.24.11 |

| M45 | History commands retain custom controls despite available native adapters | Implemented on continuation branch; native pending | Shared native Retry/pagination/reversal buttons and checkout-history native title/Close replace custom controls. Twelve remote History tests/check/structural pass; critic no product blocker. Runner-only production checkout-history fixture adds medium/expanded/older-page/Close acceptance; fixture typecheck/structural and two installer tests pass. Actual native layout and enlarged text still pending; excluded from0.24.11 |

| M46 | Add crashes in a screen-options update loop on native launch | Implemented on continuation branch; native pending | Bounded navigation-feedback fake reproduces nonsettling header updates. Shared stable-presentation hook retains committed handlers and current disabled/removal guards; Add composes memoized options. Seven focused tests/check/structural pass, including latest draft, rejected-save recovery, stale actions and teardown. Critic no blocker; native Add launch/type/save-failure scenario must still pass |

The prior web draft finding is outside this mobile-only task. This list is a seed;
the full surface/axis review must discover and track further findings.
Run34887652455 also passed the iPhone persistent actionable-feedback scenario.
This establishes retention and action reachability for that fixture, not full
assistive-technology or enlarged-text verification. Its iPad onboarding entry
still lost characters (`h//example.invalid`); M14 remains unresolved.

### M46 shared-consumer follow-through

The Add navigation-feedback finding prompted preventive stabilization of Home,
Browse, notification inbox, inventory switcher, checkout-history dismissal and
reminder timing headers. Home/Browse feedback regressions failed before the fix;
current commands, inventory labels, badges and permission changes remain live.
This does not imply native crashes were observed in all six consumers.

Home's legacy hook mocks and direct component invocation were replaced with
mounted React components, real application queries/commands and repository fakes.
Coverage retains item/location/section navigation, compact tag suppression,
initial recovery, pending Return and background reconciliation without pull
indicators. Header sizing/order has separate coverage. The full remote suite
passed 1,387 tests across 246 files, TypeScript and structural checks; critic
coverage feedback was addressed. Native verification remains pending.

These preventive changes follow the PR129 release snapshot and are not included
in that snapshot.

### M47 — Home return details uses an inline panel instead of its specified sheet

Source-confirmed, open. `HomeScreen.tsx` renders `ReturnDetailsSheet` as a `View`
at the end of dashboard content, with bespoke buttons and a placeholder-only
input. The asset checkout spec explicitly calls for a native sheet. The follow-up
may be offscreen after Return, has no modal focus boundary, and lacks a persistent
input label. A bounded optional-details task fits a native sheet with clear Save
and Cancel return semantics; this is consistent with [Apple's sheets guidance](https://developer.apple.com/design/human-interface-guidelines/sheets).
Acceptance must include actual sheet presentation, keyboard, long content, error,
dismissal/undo, and phone/iPad adaptation. This finding is not closed by replacing
buttons alone.

### M48 — Home return operations have no workflow owner

Implemented; native acceptance pending. `DashboardHeader` has no synchronous duplicate guard,
focus ownership or tenant/inventory reset. Only the currently returning card is
disabled. A second Return can replace the first optional-details editor, and a
late completion can open that editor after leaving Home. Save/undo callbacks also
accept repeat invocations while their rendered disabled state catches up.
Acceptance: deferred operations reject stale/repeated commands, retain failure
recovery, suppress new presentation after blur/refocus or scope change, reconcile
successful operations, and prevent repeat returns from stale cards.

### M49 — Home exposes Return without a mutation-permission projection

Implemented; native permission presentation pending. `HomeDashboardViewModel` carries only `canAdd` for toolbar
creation. Checked-out Home cards always create a Return footer whenever a
checkout command is present, including viewer inventories. The API remains the
authorization boundary, but this violates the mobile requirement that viewers
never see checkout/return actions. Add a correctly scoped permission projection
and real viewer/editor boundary tests before changing this interaction; do not
substitute create permission for edit/return permission.

M48 validation: four failing regressions reproduced duplicate submissions, late
presentation, late refresh feedback, and a newer checkout incorrectly disabled.
The scope-owned hook now serializes Return/Save/undo, uses current checkout
identity, preserves failure drafts and gates reconciliation feedback. Home's
dashboard subtree is keyed by tenant/inventory. Twenty focused query/interaction
tests, TypeScript and structural checks pass; full remote suite passes 1,394 tests
in 246 files. Critic findings were fixed and re-reviewed with no further blockers.
Native return/keyboard/dismissal acceptance is still pending, including open M47.

M49 implementation: Home projects `canReturn` from selected-inventory
`edit_asset`; create permission stays independent. Viewer and create-only cards
hide the command while preserving checkout status/navigation. A committed
permission ref rejects stale Return callbacks after revocation. Three regressions
failed before the fix; all 22 Home cases pass, including edit-only permission.
TypeScript and mobile/Go structural checks pass. The API checkout boundary suite
passes with pinned Go1.25.8 and verifies rejected mutations leave the open checkout
intact, alongside existing editor success and adversarial cases. Critic found no
blocker; its edit-only coverage suggestion was added. Permission changes while
the optional-details task is already open remain an acceptance case for M47.

### M50 — Native Add and checkout-history fixtures remain loading

Runtime observed, investigating. Run34917318548 iPad final screenshots show stable
Add chrome with Loading inventory, and checkout history with Loading checkout
history. Synthetic repositories are expected to resolve immediately, but native
readiness assertions fail on both devices. The former Add render-loop exception
is absent from the inspected final hierarchy. Do not certify M46/M45 or attribute
this to production connectivity without query-state evidence. Runner-only query
metadata diagnostics are the next discrimination step; no cache pre-seeding or
connectivity override is an acceptable substitute for the acceptance scenario.

### M19 direct-root candidate after native comparisons

ExpirationFiltersScreen now exposes its ScrollView directly to the native sheet;
the bottom native-action footer is a sibling with measured space reserved in the
content and scrollbar. This follows the three direct-root variants that passed
on both devices in run34917318548. Header search presentation is memoized across
unrelated draft edits. Production route and isolated fixture consumers were
inspected; neither adds an outer ready-state host container.

Four existing selection/date/menu/tag behavior tests, TypeScript, structural checks
and two fixture-installer tests pass remotely. The failing native expansion and
keyboard scenarios remain unchanged, and an overview accessibility audit adds
hit-region, description, traits, Dynamic Type and clipping checks without suppressions.
Unsupported pre-iOS17 audit runtimes explicitly skip. Critic found no source blocker.
This is a new candidate, not a declaration that expansion or keyboard behavior is
fixed; those native results remain required.

### M14 system address-field candidate

The iOS onboarding address now uses the pinned SwiftUI TextField path that preserved
both native and callback values on phone and tablet in run34917318548. Other form
fields and Android retain their existing input implementation. The shared iOS
keyboard accessory now asks the native keyboard controller to resign the current
responder; React Native's focused-input registry does not include the SwiftUI field.
This follows a confirmed critic finding before native acceptance.

The existing full-address typing, explicit accessory dismissal and command
submission assertions remain intact. Added native scenarios cover Go submission
and draft preservation when help opens/closes. These changes are candidates,
not a verified resolution of M14; native keyboard/layout/adaptation remain pending.
They are excluded from the interim release built from merged PR131.

Remote validation: 12 existing onboarding/accessory/invitation behavior tests,
TypeScript, mobile structural checks and two fixture-installer safety tests passed
(`/tmp/native-address-final.log` on paul). The accessory test first failed against
the old RN dismissal path. Critic re-review found no remaining confirmed blocker.
These tests cover source behavior and wiring; the SwiftUI adapter still needs its
native run and must not inherit the generic renderer's test coverage claim.

### M47 native-route implementation candidate

The inline editor has been replaced with the existing native-stack form-sheet
presentation. An app-tree UI context carries the editor into its route while the
inventory-keyed Home hook retains command ownership. A direct React Native Modal
candidate was rejected by the structural check and removed before finalization.
The native route includes its title, persistent input label, platform command
buttons, local failure feedback and permission-loss recovery. It retains the
native keyboard dismissal path and scrollable content.

Critic findings corrected before native testing: explicitly show the native header;
defer dismissal while another route is above this sheet; bind callbacks to their
originating editor session. Tests exercise Back/undo, failed-save retry, permission
revocation, a second return editor, focused/background completion and missing-task
entry. The navigation fake now cleans up removed guards. Obsolete inline styles
were removed. Native full-Home fixtures cover cancel restoration and failed-save
recovery with keyboard entry; actual native acceptance is still pending.

Final remote source validation passed 1,404 tests across 247 files, TypeScript and
mobile structural checks (`/tmp/home-return-full-final.log` on paul), plus both
fixture-installer safety tests. Native fixture source is not native execution;
large text, long details, modal focus and dismissal remain acceptance work.

### M51 — Native color-row activation is not reliable in the audit

Run34919776387 failed opening the system color picker on phone and iPad, after an
earlier pass. The inspected iPad screenshot shows no presented picker. This is an
observed acceptance failure, not yet a proven implementation defect: the native
accessible row spans its label and trailing well, so a separate well-target probe
is pending. Preserve both results and do not certify row activation from a well tap.

### M45 initial native header configuration

The same run loaded iPad checkout records but left the native header absent. The
shared sheet initially hid that header while the mounted screen requested it.
The candidate now shows the title/header from the initial sheet configuration,
matching the screen. Existing Close and expansion assertions remain the native
acceptance gate; M45 is not closed by this source change.

### M52 — Field-editor commands retain radio/custom-button presentation

Source-confirmed task/pattern mismatch in `CustomizationEditorFields.tsx`:
`Expand to all assets` was an always-unchecked radio although it changes the draft
through a command. `Add option` used a custom inline button despite the shared
native command adapter. Both now use that adapter, with Add below its input to
avoid squeezing input beside a full-width native host. Type and initial Applies
to remain pickers, and saved options/targets retain their existing protection.

Two new tests reproduce the missing command semantics and verify expansion,
normalized option addition, duplicate handling and draft clearing. Eight shared
control and37 consumer behavior tests, TypeScript and mobile structural checks
pass remotely. Household/inventory create/edit share this consumer. Native layout,
keyboard, large text and assistive-technology acceptance remain pending.

### M53 — Native choice label can clip at accessibility text sizes

Run34920888328's phone accessibility audit flags the Availability label in the
Expiration overview. The source uses a string-label LabeledContent; the candidate
now uses its supported native Text label slot with unconstrained vertical size.
This preserves the menu/value interaction and allows the host to grow with text.

Shared consumers inspected: Browse/Expiration filters; Appearance; custom field
Type/Applies to; expiration Month; type reminder mode; voice Service; invitation
Access; and Move destination Kind. Selection values, disabled guards and callbacks
are unchanged. No Android presentation change.

Nine remote picker/filter behavior tests, TypeScript and structural checks pass;
two native-fixture installer tests pass. Existing failing AX scenario is preserved
and a separate largest-accessibility-text scenario opens the native menu after
locating the visible label. Critic found no blocker but correctly notes that its
height check is only a coarse enlarged-text check: screenshots and the AX audit
remain required to establish no clipping. Native outcome remains pending.
