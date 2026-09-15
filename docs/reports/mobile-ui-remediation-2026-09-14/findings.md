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

M50 Add readiness candidate after run34965113594: the hidden-header sheet remains
loading with zero observers, while the otherwise identical preconfigured-header
sheet reaches the focused form. Production Add now declares its native header and
known title before presentation, retaining the same sheet and screen-owned
commands. No query preloading or connectivity override was introduced. The
configured-header comparison still fails exact typing; current phone/iPad loading,
text retention and rejected-save recovery remain required. This does not close M50.


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
passed on both phone and iPad in run34965113594. The inspected phone capture shows the system picker open. Preserve the failed row result: well activation does not certify row-center activation or VoiceOver activation. M51 remains open.

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

### M54 — Choice adapters forward events while disabled

Source regression tests reproduced disabled callback delivery on iOS, Android
and the generic renderer. Each adapter now rejects events from a disabled render
and resumes valid changes after re-enabling. Android also marks each menu item
disabled, so its open-menu presentation receives the current lock state.

Three failing cases were observed before the fix;22 focused adapter/expiration/
custom-field/reminder tests plus TypeScript and structural checks now pass remotely
(`/tmp/native-choice-lock-green.log`). The shared consumer inventory is the same
as M53. No authorization behavior changed. Critic found no blocker and emphasized
the evidence limit: these checks prove current callback-boundary behavior, not
immediate native handler replacement in an already-open menu. Physical timing and
Android runtime acceptance remain pending.

### M55 — Inbox open completion outlives its navigation intent

Three source tests reproduced opening an asset after blur, after blur/refocus, and
starting a read through an unfocused callback. A focus-session token now gates
open/navigation. Completed reads still reconcile mounted inbox/count state;
unmount cancellation and scoped ownership remain intact. Normal focused opening
retains resolve/read/reconcile/navigate order.

Seventy-two notification tests/typecheck/structural pass remotely. Critic found no
blocker and requested proof that a fresh open works after the old request settles;
the final11 inbox tests include that passing case. Native navigation interruption
remains pending. See `notifications-axis.md` for the broader140-surface ownership
map and explicitly unverified paths.

### M56 — Inbox recovery and paging use custom command controls

Inbox Retry/Load more and route-load Retry now use the existing native command
adapter. This matches their command task without extra navigation. Query scoping,
loading guards, recovery and pagination behavior are unchanged. Eleven existing
inbox behavior tests, TypeScript and structural checks pass remotely. Critic found
no issue. Native control sizing/appearance and paging acceptance remain pending.

### M57 — Device setup feedback outlives the attempt's context

Reminder settings retained a successful permission claim after returning from
OS Settings, and compared display text to choose button behavior. Feedback now
uses a typed outcome, describes a completed setup attempt, and clears on focus
entry or backgrounding. A generation token rejects delayed enabled/denied feedback
after backgrounding or navigation departure without canceling background setup
persistence. Transient inactive permission prompts retain their feedback ownership.

The delayed-background regression failed before the generation guard. Fifteen
remote settings/setup/session tests pass, including background/inactive × granted/
denied results, unchanged inventory preferences, and return/retry behavior.
TypeScript and the mobile structural check pass. Critic's race finding is resolved.
Physical permission prompts, external Settings changes and native lifecycle timing
still need verification. This change follows the PR135 release cut.

### M58 — Customization completions outlive navigation and record ownership

Delayed save and archive completions previously dismissed a newer task. Focus
identity now gates success announcements and navigation; an old confirmation
cannot start a mutation. A separate resource lifetime scopes local completion,
permission failure, refresh and busy state. Each resource gets its own workflow;
old loads retain their original workflow so replacement cannot revive stale loads.

Three deferred focus tests failed baseline. Two resource-replacement tests also
failed before the lifetime guard because the new editor remained locked. The
final45 screen/workflow tests, TypeScript and structural checks pass remotely.
Tests cover delayed granted/denied completion, late confirmation, and fresh save
recovery; critic's resource-lifetime issue is addressed. Native navigation and
accessibility timing remain pending.

Follow-up: the retained editor's pre-existing completed flag suppresses dirty
tracking after subsequent edits; completed create/lifecycle presentation also
needs review before the entire customization lifecycle is considered resolved.
M58 fixes ownership, not that separate completion-state design.

### M59 — Retained completed forms permit misleading repeat editing

After a background completion, the existing completed flag suppressed dirty
tracking but left form controls visible. The retained screen now shows its typed
Saved/Archived/Restored/Deleted result and one native Return to collection command.
Submitted draft and lifecycle controls are removed. Normal focused success still
returns automatically; failures retain editing/recovery. Resource replacement
clears the terminal state. This closes the completion-state follow-up under M58.

Three revised regression cases failed baseline. All45 screen/workflow tests,
TypeScript and structural checks pass remotely; critic found no blocker. This
protects rendered interactions, not arbitrary externally retained callback closures.
Surface S141 adds explicit completion-state coverage. Native result adaptation,
focus and accessibility acceptance remain pending.

### M60 — iPad Return cancellation falls below the visible sheet

Run34923022927 shows Save at y839–887 and Cancel at y903–951 below the
visible sheet. The existing native cancellation test fails hit-testing. The
candidate pairs the native commands in a flexible wrapping row, cancellation
first, inside the directly rooted scrolling form. Each native host receives its
own width-constrained wrapper. It reduces separate-row height without imposing
a fixed form height or removing enlarged-text scrolling.

Twenty-five Home behavior tests, TypeScript and structural checks pass remotely;
critic found no blocker. Existing iPhone/iPad native cancellation and optional
details tests remain unchanged and required. Candidate reachability is unverified
until those tests run; keyboard and enlarged-text acceptance remain open.

### M61 — Checkout history subtitle overlaps native navigation

The iPad screenshot places the asset subtitle in the navigation bar's vertical
space. The sheet previously wrapped a subtitle and nested record ScrollView in
a non-scrolling parent, with no automatic content inset for the subtitle. The
candidate uses one directly rooted ScrollView for subtitle, all states, records
and native retry/paging commands. Content padding remains inside the scroll
container; automatic native insets own navigation/bottom clearance.

Six history behavior/application tests, TypeScript and structural checks pass
remotely. Critic found no blocker. The unchanged native note-hit-testing failure
is not yet proven resolved: iPhone/iPad detents, Close, paging and enlarged text
remain required. This change does not alter query permissions or pagination.

### M62 — Initial Reduce Motion reads override newer preferences

Map started with motion enabled and could overwrite a newer live event with an
older async snapshot; read rejection was unhandled. Voice rails had the same
overwrite race. The shared UI motion-preference hook now starts conservatively,
subscribes before reading, gives live changes precedence, and catches failed reads.
Notices reuse it while preserving independent screen-reader behavior.

Two rendered Map animation-request tests failed baseline. Ten focused checks pass
including pending/late/failure preference reads and later live reenablement, plus
notice behavior and existing voice entity-link coverage. TypeScript and structural
checks pass; critic found no blocker. Device animation, voice rail timing and
system-component adaptation remain unverified. See motion-axis.md for scope.

### M63 — Photo selection unnecessarily requires broad library access

Source-confirmed, P2. The shared Expo photo adapter requested full library access
before opening the system picker and rejected a denied response. Add, asset-detail
attachments, and voice-plan photos all inherit the gate. Their task is choosing
specific photos, for which the platform picker provides scoped access. Pinned
expo-image-picker55.0.20 documents the library prerequisite only for iOS10.

The adapter now opens the image-only library picker directly. Cancellation stays
an empty result; camera capture retains its permission gate. Two denied-library
selection/cancellation cases failed before the change; all seven adapter cases,
TypeScript, and mobile structural checks pass on paul. Evidence:
`/tmp/photo-picker-red.log`, `/tmp/photo-picker-green.log` (remote host).
Critic found no blocker; its requested image-only launch assertion was added
and the checks rerun successfully.

Native acceptance remains open: select images with library authorization denied,
cancel without error, and deny/allow camera from Add, attachments and voice on
supported iOS/Android; confirm no unexpected permission prompt and that selected
image content reaches the intended draft. Adapter tests do not prove native prompts.
This finding covers selection permission only, not all photo lifecycle behavior.

### M64 — Asset photo operations outlive their asset context

Source/behavior-confirmed, P2. Asset-detail selection started uploads even after
asset change or route teardown. Duplicate source callbacks could launch duplicate
uploads; upload progress, failure, retry drafts and cleanup lacked the ownership
guard already used by removal. Three rendered regression cases failed before
implementation. Selection, upload, retry and removal now share one asset-owned
pending scope. A stale picker cannot start an upload; already-started commands
finish for their original asset without writing to the replacement view. Failed
photo drafts reset on asset change.

A fourth case exercises an old upload finishing while a new asset upload remains
pending, including progress and final status. Remote test/type/structural evidence
is in `/tmp/photo-ownership-red.log` and `/tmp/photo-ownership-green.log` on paul.
All 65 selected checks, TypeScript and structural checks pass; critic found no
confirmed blockers.
Native chooser interruption/dismissal remains unverified. Route focus without
unmount and other mutation flows are separate lifecycle review work, not cleared
by these checks.

### M19 — Measured keyboard-clearance candidate

The footer now measures an unmoved bottom boundary in the sheet's window and
offsets only the overlap with the native keyboard frame. The direct-root scroll
view remains intact. Pinned React Native0.83 RCTKeyboardObserver converts iOS
keyboard frames from screen to key-window coordinates; no full-screen height or
fixed keyboard offset is assumed. Superseded measurement callbacks and callbacks
after hide/unmount are ignored. Already-resized sheets do not get a second offset.

Seven geometry/hook/filter checks, TypeScript and structural checks pass on paul
(`/tmp/footer-boundary-green.log`). Critic found no confirmed blockers. Existing
native expansion, phone/iPad search-keyboard, rotation and enlarged-text scenarios
remain the acceptance gate. This candidate does not clear M19 until those pass;
keyboard animation and exact native coordinate alignment remain unverified.

### M65 — Return note loses/reorders native text

Runtime-observed, P1. iPad mini native run34927007321 typed `Returned clean` but
the native text view contained `leanR`. Screenshot:
`evidence/ipad-return-note-corruption-34927007321.png`. The same run's controlled
address comparison lost text, while native and uncontrolled comparisons passed.
This supports removing draft-value replay as a candidate, not proof of its cause.

The Return note now has a stable native initial value keyed by return session.
Change events still update the application draft for Save/retry. Twenty-five Home
behavior checks, TypeScript and structural checks pass on paul
(`/tmp/return-native-text-green.log`); critic found no blocker. Removed controlled
value-prop assertions no longer pretend to establish visible text preservation.
The unchanged native full-string typing and rejected-save retry assertions must
pass on phone and iPad before closing this finding. Other controlled fields remain
a broader text-entry audit concern.

### M53 — Reflow after inspecting the enlarged phone screenshot

Run34927007321's explicit accessibility-size screenshot shows the native label
and selected value squeezed into two columns, with the value broken into short
fragments (`evidence/phone-choice-narrow-columns-34927007321.png`). The AX issue
description says the Availability node may clip at larger sizes; normal-size
screenshot alone would not reveal the problem. Passing hit-testing did not prove
readable layout.

The shared iOS picker now uses native VStack label-over-menu at accessibility
font scales, retaining LabeledContent otherwise. The native menu's own label is
hidden in the vertical layout, preserving its explicit accessibility name. Expo
55.0.17 does not expose ViewThatFits; the threshold uses pinned React Native0.83's
default AccessibilityMedium multiplier. This limitation is documented in spec.
All previous shared consumers remain in scope.

One component case failed before the change; eight picker/filter cases, TypeScript
and structural checks pass on paul (`/tmp/choice-reflow-green.log`). The native
large-text scenario now also requires the menu below the label, while retaining
its hit-testing/menu-open checks and the original accessibility audit. All native
reflow, ordinary-size regression, and medium-category clipping outcomes remain
pending; M53 is not cleared. Critic found no confirmed blocker and emphasized
that vertical placement alone does not prove long-value fit.


### M66 — Refinement count badges lose text contrast

The iOS, Android and fallback refinement buttons repeated white badge text on
`palette.accent`. Rendered foreground/background measurements were 3.39:1 in
light and 2.22:1 in dark, below the 4.5:1 target for small text. Users have more
difficulty reading the applied filter count. This is a source/numerical finding,
not a screenshot-derived native geometry finding.

The shared RefinementCountBadge now uses the semantic action/onAction pair.
Eight new rendered checks failed before the change; all 39 selected badge/token
checks, TypeScript and mobile structural checks pass on paul. Critic found no
blocker; its duplicate contrast-helper concern was addressed with a shared test
utility. Android uses the shared component but is not mounted by these tests.
Native badge placement, text growth and material composition remain pending.
This fix is after PR136 and is excluded from the interim release cut.


### M67 — History ignores device date and clock conventions

Activity, checkout history and exact event details forced en-US while neighboring
date surfaces used the device locale. The shared AssetHistoryTimestamp formatter
now uses the runtime locale/local zone and preserves each existing detail level
and invalid-value fallback. Stored timestamps, ordering and authorization are unchanged.

New explicit British/US/precision/invalid checks were written before the helper
(the initial red was a missing-module failure, not a baseline behavioral claim).
Thirteen initial formatter/query checks, eight mounted history checks, TypeScript
and structural checks pass on paul. A fifth formatter case exercises the omitted
locale contract; all five pass with LC_ALL and LANG set to en_GB.UTF-8, with the
actual runtime locale independently verified as en-GB. Critic found no blocker
and requested this default-locale coverage. Native settings changes, clock
overrides, long date layout and time-zone acceptance remain pending.

### M68 — Month-only expiration mislabels alternate-calendar periods

Open. A stored Gregorian month is formatted by converting its first day using
the locale's default calendar; month choices use unrelated synthetic2020 dates.
For en-US-u-ca-hebrew, stored2028-02 is labeled Shevat5788 and January's choice
is named Tevet. These periods do not share boundaries, so choosing a displayed
month can store a different period. Thai year display also differs from the
canonical numeric year field. Source and remote Intl evidence are recorded in
localization-axis.md; native propagation and corrected contract remain pending.


### M19 follow-up — one keyboard-avoidance owner

After native349289 placed the visible Apply button outside its Host bounds,
NativeSheetActions gains a fixed keyboard-avoidance owner. Only the measured
expiration footer selects container ownership, which disables the SwiftUI Host's
keyboard safe area; Browse keeps the native default. Other safe areas remain.

One ownership contract case failed before the change; seven action/filter cases,
TypeScript and structural checks pass on paul. Critic found no source blocker.
These verify configuration and callbacks, not native hit-testing. The hierarchy
also shows a possible62-point boundary/keyboard coordinate discrepancy; that
remains unresolved. The original Apply/Back native hit/navigation assertion stays
the acceptance gate. M19 remains open and no full keyboard fix is claimed.


M68 implemented follow-up: localized Gregorian month choices and month-only
summary formatting share the stored period's calendar, with a clarification for
alternate-calendar locales. Exact-day labels/controls and storage are unchanged.
One baseline label test failed;14 selected checks, TypeScript and structural
checks pass on paul (`/tmp/month-calendar-green.log`). Critic found no implementation
blocker; its domain wording correction was applied. Shared field/status/card,
workspace heading, notification and voice-review consumers were inspected.
Native alternate-calendar settings, digits and larger-text clarification remain
unverified; this is source/test remediation, not complete localization acceptance.

### M69 — Sharing feedback outlives its initiating screen (P2)

Implemented focused-session/scope ownership for create/cancel/copy/share notices.
Three failing rendered departure regressions now pass; normal focused failure
retains the draft and notice. See [Sharing review](sharing-axis.md) for task-fit,
source evidence, privacy boundaries and pending native acceptance. This is the next
batch after PR138 and is excluded from release workflow34932422663.

### M70 — old invitation recovery replaces a newer invitation (P2)

Implemented request-generation ownership for opening/start-over failures and reset
of replacement start-over availability. Two regressions failed before correction;
13 invitation-screen checks, typecheck and structural checks pass remotely. See
[Sharing and invitation review](sharing-axis.md). Native link replacement and
route-side navigation effects remain pending; no broad deep-link acceptance claim.

### M71 — failed system initial-link lookup leaves invitation initialization pending (P2)

Implemented rejection readiness, foreground precedence and disposed-subscription
guards behind the existing native Linking adapter. A mounted fake-source baseline
reproduced the stall and unhandled rejections. Ten remote hook/domain checks plus
typecheck/structural checks pass. See [Sharing/invitation audit](sharing-axis.md).
Native deep-link delivery remains pending; no claim of whole-entrypoint acceptance.

### M72 — stale invitation route completion clears a replacement link (P2)

Implemented focus/reference ownership for clear-and-return navigation after Open
inventory and Start over. Mounted original-callback regressions failed before the
fix. Seven route/progress/selection checks plus typecheck/structural checks pass
remotely. See [invitation audit](sharing-axis.md); native delivery remains pending.

### M73 — unavailable inventory blocks sign-out recovery (P1)

Root Settings loading/error states now keep Account and Connection reachable.
Account reads only identity and permits confirmed sign-out with a fallback label
when identity is pending/failed. Two original regressions failed; 27 remote
settings/cache checks plus typecheck/structural checks pass. Critic requested
identity cases were added. See [Account/connection audit](account-connection-axis.md)
for remaining native layout/lifecycle acceptance and unchanged scope boundaries.

### M74 — Add name loses characters during native typing (P1)

Native-observed on phone run349297 navigation-stack comparison. Implemented an iOS
native-owned Name candidate with explicit restore/reset lifetimes and unchanged
application draft/save guards. Six remote behavior checks plus typecheck/structural
checks pass; native full-string acceptance remains pending. Sheet readiness remains
a separate issue. See [text-entry review](text-entry-axis.md).

### M20 follow-up — preserve readable form and full scroll surface

Repeated iPad evidence distinguishes the failing margin drag from the passing
inside-form drag (run34932076384). Move the centered600-point constraint into a
child form and retain full-width scroll content. No custom gesture or manual
keyboard dismissal is introduced. This is a source candidate, not a verified fix;
existing phone/iPad keyboard and landscape tests remain the native acceptance gate.
Remote validation:4 mounted onboarding/invitation checks, TypeScript and mobile
structural checks pass on paul (/tmp/onboarding-scroll-green.log). Required critic
found no blocker. No local tests/builds were run; native acceptance remains pending.

### M75 — departed onboarding screen sends reset navigation (P2)

Source and mounted-test confirmed: a pending Sign out and start over completes
its authorized reset after unmount, then calls onStartOver and onStateChange from
the departed screen. This can replace the current destination. Connect/Create
already checks the mounted generation; reset now uses the same ownership check.
Teardown continues and the focused success path still returns to connection.

The regression failed with both callbacks observed, then passed with sign-out
completed and no callbacks. Five remote onboarding/invitation checks, TypeScript
and structural checks pass (/tmp/onboarding-reset-green.log). Critic found no
blockers. This proves unmount ownership, not native transition rendering or every
in-place state replacement. Native acceptance and failure recovery remain tracked
separately; no new authentication or teardown behavior is introduced.

### M76 — setup permits progress with missing required values (P2)

Source interaction gap against the entering-data criterion in onboarding-axis.md:
Connect/Create was available with blank required values. Primary actions now reflect
readiness, with a visible explanation naming the missing value. Whitespace stays
incomplete; nonempty invalid URLs still receive application validation. Keyboard
submission uses the same guard, and Start over is independent of incomplete names.

Two mounted regressions failed before the change. Eight remote onboarding/invitation
checks now pass, including both inventory form variants, keyboard bypass prevention,
URL validation and existing recovery. TypeScript/structural checks pass on paul
(/tmp/onboarding-readiness-green.log); critic found no blockers. Native explanation
layout and keyboard timing remain unverified, including M35 compact-phone reachability.

### M77 — partial time-zone identifiers cannot be found (P2)

The native search advertised city/time-zone lookup but matched only the reversed
readable label. `America/New` therefore missed available America/New_York, although
a complete valid identifier could appear through a separate fallback. Search now
matches both label and identifier, preserving case/outer-whitespace handling.
No selection is saved by typing. Existing bounded results and valid-zone fallback
remain unchanged. Mounted regression failed first; eight remote picker/settings
checks, TypeScript and structural checks pass (/tmp/timezone-search-green.log).
Critic found no blockers. Native search integration still requires runtime evidence.

### M78 — inventory switch completion outlives focus (P2)

The switcher canceled pending work on unmount but not blur. A mounted regression
reproduced Back after focus left and returned during selection. Focus cleanup now
aborts the request signal; late navigation/error feedback is suppressed. The
pending guard remains until settlement, after which a fresh focused selection
works. This does not promise reversal of an inventory choice already persisted by
the port. Critic caught the complementary blur → settle → refocus ordering leaving rows
disabled. Its regression failed, then passed after focus entry reconciled busy
state with the actual pending request. Five remote switcher checks, TypeScript and
structural checks pass (/tmp/switcher-focus-final.log). Native interruption
acceptance remains pending.

### M79 — cancellation skips accepted-selection reconciliation (P2)

SelectInventoryCommand checked cancellation between repository success and its
composition-scoped selection observer. If acceptance preceded cancellation, the
chosen inventory could change without notifying the cache to reconcile. The
observer now runs after repository success; a final cancellation check still
rejects obsolete caller success. Initially canceled and rejected selections do
not publish. No authorization rule or API boundary changes.

A port-level regression failed before the correction and now verifies acceptance,
observer notification and canceled caller outcome. Negative cases cover rejection
and initial cancellation. Twenty-eight remote command, switcher and inventory
adapter checks plus TypeScript/structural checks pass on paul
(/tmp/selection-cache-green.log). Critic found no blockers. Native interrupted
selection and cache-driven screen transition still require runtime evidence.

### M80 — switcher commands bypass the native adapter (P2)

Switch household/Back and load Retry used custom styled Pressables despite the
existing NativeCommandButton adapter. They now use that adapter. Its full-width
Host sits below the wrapping household heading, avoiding a competing horizontal
width constraint. Loading copy uses inventories rather than internal tenant
terminology. Selection rows and ownership behavior are unchanged. Five existing
remote switcher checks, TypeScript and structural checks pass
(/tmp/switcher-native-commands.log); critic found no blockers. Native narrow,
large-text and sheet-layout acceptance remains pending.

### M81 — Settings retry commands bypass the native adapter (P2)

Root Settings and Diagnostics load failures, plus the shared refresh notice used
by root Settings, Account and Diagnostics, used custom Pressables for commands
already supported by NativeCommandButton. They now use that adapter in vertical
content. Retry callbacks, retained values, error copy and account recovery links
are preserved. Other users of the shared retry styles are unchanged and remain
part of the audit. Twenty-seven existing remote Settings behavior tests,
TypeScript and structural checks pass (/tmp/settings-native-retry.log on paul).
This presentation change adds no prop-mirroring tests. Native error-state layout,
large text and VoiceOver acceptance remain pending.


M81 follow-up extends the native commands to scoped Settings, customization
collections and editors (including Refresh access), provider state and voice
setup. Shared refresh-notice consumers also include Sharing, scoped Settings,
customization and provider/voice editors; their vertical composition was reviewed.
Remaining route, navigation-guard and Sharing-specific retry controls are outside
this pass, and their shared styles remain. The first validation caught a duplicate
import and premature style removal; both were corrected before committing.
All 115 tests across five Settings, customization, provider and Sharing suites
pass on paul, followed by TypeScript and structural checks
(`/tmp/settings-retry-consumers.log`, `/tmp/settings-retry-consumers-check.log`).
The critic found no remaining blockers. Native layout acceptance remains pending.

### M82 — Add save error is behind the native sheet (P1)

Run 34939793483, actual checkout 17c9a1c94fa38092ac965c9eccfb9b33051c0c1f,
iPhone configured-header comparison retained and submitted Native draft name.
The final hierarchy B690DF46-1CAF-4374-AD62-85028EC65E52.txt contains the root
notice, but inspected screenshot 3CBE115A-6EBB-46A7-8737-CACC78FE931F.png shows
no error in the presented sheet. The test also queried StaticText while the old
notice grouped its message; existence alone would not prove visibility.

Save failures now persist inside the Add form, scroll into view on layout, and
announce through iOS accessibility or Android live region. Draft edits and the
next save clear stale failure state. The draft and retry/close lifecycle remain.
An ownership regression asserts that the error is inside the form scroll view;
it fails against HEAD and passes with the correction. Sixteen remote Add checks,
TypeScript and structural validation pass on paul (/tmp/add-inline-error-green.log);
critic found no blockers. Native inset/keyboard visibility and announcement
acceptance remain pending. The separate Add typing and loading failures are open.

M82 consumer follow-up: Add parent creation, library and camera failures used
the same root notice. They now share the form-owned error path with accurate
headings. Starting a new operation clears stale errors; cancellation stays silent.
Three operation error ownership regressions fail before the correction. Eighteen
remote Add checks, TypeScript and structural checks pass afterward
(/tmp/add-operation-error-green.log on paul); critic found no blockers.
Native camera/library return and announcement checks remain pending.

M81 remaining shared-style consumers: expiration filter Retry/Cancel, notification
settings Retry, Sharing/Voice guard recovery, invitation Retry and pagination now
use native commands. Authorization decisions and request callbacks are unchanged.
The shared custom retry styles are now unused and removed. Twenty-six remote
guard, Sharing and notification checks plus TypeScript/structural checks pass
(/tmp/settings-last-retry.log). Critic caught missing pagination progress copy;
an adjacent Loading older invitations status now preserves feedback while the
native command retains its stable label. The nine Sharing checks and static
checks passed again after that correction. Native recovery layout remains pending.

### M83 — photo-removal recovery belongs above the viewer (P1)

Source inspection finds AssetPhotoViewerSheet retains its overFullScreen viewer
after failed removal, while AssetDetailRouteScreen sends failure to the root
notice. The Add runtime evidence established that such notices can remain behind
a native modal. Photo-specific visual failure is not yet captured.

The rejected destructive operation now uses the existing native dialog adapter
with one OK acknowledgment. It preserves the photo and retry state, and suppresses
late failures after the operation owner leaves. The failure/teardown regression
fails before correction and checks acknowledgment does not retry. Sixty-one
remote asset/photo checks, TypeScript and structural checks pass
(/tmp/photo-removal-alert-green.log); critic found no blockers. Native viewer/alert
layering, VoiceOver focus return and retry remain required acceptance evidence.

M83 native scenario added: the real photo viewer and native feedback adapter are
composed with a synthetic failure. XCTest checks confirmation, reachable failure
alert/OK, two attempts, preserved viewer controls and closing to the retained
photo count. Captures must be inspected after execution. This verifies modal
presentation, not production deletion or authorization. Remote structural and
two fixture-preparation checks pass; critic found no blockers. Native pending.

### M84 — photo viewer chrome ignores Reduce Motion (P2)

Pinned image-viewing0.2.2 uses200ms Animated.timing translations to ±300points
for zoom-triggered chrome changes, without reading Reduce Motion. The wrapper's
fade prop does not control this path. Source confirmed; repair and native preference
testing pending. See photo-viewer-axis.md.

### M85 — photo load failure has no viewer recovery (P1)

Pinned image-viewing0.2.2 resolves failed dimensions to0×0 and has no image onError
handler to leave loading or offer retry. Both platform image components retain
loading until their success path. The wrapper exposes no load-error callback.
Source confirmed; failed-media runtime reproduction and dependency repair pending.
Close remains an escape, but does not explain or retry the failure.

M84 candidate repair: pnpm applies a content-hashed patch to0.2.2; the version
and other package resolutions remain unchanged. The chrome hook keeps stable
animation values, starts with motion suppressed, handles live preference changes
and stale/failed initial reads, and stops animation when reduction is enabled.
Tests import the installed dependency hook rather than the viewer test double.
The initial regression fails against upstream;15 focused remote checks plus
TypeScript/structural checks pass against the patch. A clean web-container-shaped
frozen-lockfile install also passes. Dockerfile.web now copies patches before
install. Native zoom/preference-change verification remains pending; M85 is open.

Critic requested stronger motion evidence: the controlled animation fake now
tracks pending children and stop operations; the regression verifies settled
positions, stable values across rerenders and cleanup on unmount. Twenty-five
focused viewer, route, Map and feedback checks pass, along with TypeScript and
structural checks. This still does not establish native timing or zoom behavior.

M85 candidate repair now handles both dimensions and native decode errors, presents
Photo unavailable with the existing native Retry command, and remounts the image
for a fresh attempt. Late events cannot replace the current attempt; only the
active photo restores viewer chrome on failure. Both caller projections are stable
across unrelated updates. Eleven focused checks, eighteen Add checks, TypeScript,
structural validation, and an iOS Metro export passed remotely. This is not native
visual acceptance; M85 remains open pending the runner scenario and inspection.

M82 header-reveal follow-up: the iPad screenshot in run349441 shows the error
heading under the navigation bar, despite a readable message. Add now requests
the measured negative iOS header offset when revealing an error; RN's bounded
programmatic-overflow option permits the automatically inset position. Android
keeps zero. The existing draft-recovery regression failed at the old zero offset
and now passes for two header measurements. The native scenario additionally
requires the entire heading below the navigation bar. Native acceptance is pending.
The broader remote check of the changed shared ScrollView fake passed all1495
mobile tests across257 files; TypeScript and structural checks also passed.
This validates source behavior, not iOS geometry.

### M86 — asset command callbacks lack completion ownership (P1)

Checkout, return and lifecycle callbacks could submit twice before the busy render
and issue UI effects after route teardown/replacement. Four regression cases
reproduced duplicate command calls. The candidate extends the existing photo
operation owner to these commands; tests cover late deletion navigation, failure
feedback, stale confirmations and replacement-asset busy state. Authorization and
domain command behavior are unchanged. Native focus/blur and interruption coverage
remain pending. See asset-actions-axis.md.

### M87 — asset sheet recovery bypasses native commands (P2)

Edit and Move query retries used unstyled Pressables. Edit also rendered each
metadata failure in an expanding error panel, with indistinguishable Try again
labels. The candidate reuses NativeCommandButton for asset, placement, suggestions,
types and tags. Supplementary errors are compact inline messages. A regression
verifies independent type/tag retries retain a dirty name;12 asset-sheet behavior
checks, TypeScript and structural checks pass remotely. The new isolated native
Edit scenario checks simultaneous failures at largest text size, including Cancel.
That scenario has not run; combined-height and native reachability remain open.


### M88 — contained workspace retains custom search and commands (P2)

AssetContainedWorkspace uses an AppTextInput and separate Clear control, plus
custom spatial and maintenance buttons. The inline search follows an older spec,
so this is design-contract drift rather than failure to follow that contract.
Update the contract to scoped native search on demand, and use native command
controls while preserving Add prominence. Inspect both regular detail and map
sheet consumers. Source-confirmed; implementation and native acceptance pending.
See contained-items-axis.md for the complete24-axis review and acceptance.

### M89 — unknown contents are presented alongside empty-state claims (P2)

AssetDetailView builds empty section rows while contents are loading or unavailable.
AssetDetailRouteScreen reports query failure through a root notice suggesting a
pull gesture, without persistent region-level retry. In a map detail sheet the
root notice may be obscured; that occlusion is a source risk, not a new screenshot
observation. Retain available content and provide explicit independent native
contents/photo retries; do not claim an unknown collection is empty. Source
confirmed; reproduction tests, implementation and runtime acceptance are pending.


M89 candidate: contents and photos now expose independent native retry commands
inside their owning detail screen/sheet. Unknown contents and photos no longer
render empty claims; initial retry returns to the loading indicator, while cached
content remains visible. A real query regression reproduced false empty copy
before implementation; independent failure/retry and cached-refresh retention
checks pass.35 remote detail tests, TypeScript and mobile structural checks passed.
Critic found no blocker. Native sheet placement and announcements remain pending;
M89 is not closed by this source result. This change is after the PR142 release.


M89 route correction: current Map info pushes assetDetailHref rather than a
sheet. The earlier sheet-occlusion rationale does not apply to that current path;
false empty claims and lack of persistent local retry are still source-confirmed.
A runner fixture now exercises the actual shared detail route at largest text,
with independent contents/photo recovery and Back. Native execution is pending.


M88 search candidate: location contents now use NativeNavigationSearch with the
existing20-row threshold. The inline field is removed and no-match Clear search
uses NativeCommandButton. Route-owned search resets on asset/eligibility changes;
owner guards and keyed adapter lifetime reject obsolete callbacks. A regression
reproduced stale callbacks clearing a newer asset query before the guard. Shared
adapter events after unmount are also ignored. Spatial and maintenance controls
remain open under M88; search header/keyboard behavior is not native-verified yet.

M88 validation:41 final remote adapter/consumer/route checks passed, plus18
detail-presentation checks before the ownership follow-up, TypeScript and
structural checks. Critic confirmed the stale-event fix; native acceptance pending.


M88 command candidate: spatial, availability and maintenance actions now use the
shared native command adapter. Add item here and direct item availability retain
primary prominence; contained availability and maintenance use standard commands.
Authorization-derived visibility, missing-handler/pending guards and ordering are
preserved. Removed custom icon/button styling. The native adapter regression
reproduced missing primary emphasis, then passed with37 detail/native checks,
TypeScript and structural validation. Native width, multiline labels and wrapped
maintenance rows at large text remain unverified; M88 remains open for acceptance.

The broader shared-adapter run passed all1505 mobile tests across258 files on
paul. Code critic found no confirmed blocker. This does not establish native
button geometry or Android runtime behavior.


M87 native follow-up: phone run349502 at97edb367 confirms Cancel outside the
visible sheet at largest text. Metadata errors are siblings above the entire edit
form rather than part of its scrollable content. The retry labels also visibly
overlap adjacent messages; their reported native button frame is46.1points despite
a two-line large label. Do not treat hittability as proof of label layout. Repair
scroll ownership and investigate hosted-label measurement before closing M87;
keep existing native assertions. See phone-edit-errors-large-text-349502.png.

M87 scroll-ownership candidate: Edit metadata recovery now renders inside its
form ScrollView rather than above the form. Cancel/Save retain their existing
fixed action region. A regression first failed on the old layout and now passes
while preserving independent retries and dirty names (12 action-sheet tests,
TypeScript and structural checks on paul). Native acceptance scrolls each retry
fully into view and checks Cancel throughout; native execution remains pending.
The hosted native-label overlap remains under investigation, so M87 stays open.
Adjacent Move/Move here candidate status is still outside their forms and needs
its own large-text review; this Edit-only change does not certify those layouts.

### M90 — Move here reports unknown suggestions as empty

Source-confirmed, recovery priority P2, S136 loading/recovery. A failed current
query previously rendered No movable matches beside its retry. The intended
pattern distinguishes unavailable results from known empty results. The candidate
fix gates empty copy on current-query data and places status/retry in the results
scroll region; existing cached candidates and draft query remain. Regression
failed before the change and all13 asset action-sheet checks, TypeScript and
structural checks passed on paul. Critic found no confirmed issue. Native large
text, keyboard and selection-retention acceptance remain pending. This does not
resolve the other fixed-content layout risks in Move or Move here.

### M91 — asset form completion commands remain custom

Source-confirmed platform-pattern gap, P2. Shared SheetActions in
AssetDetailSheets.tsx renders custom Pressable Cancel/Save/Move commands for Edit,
Move and Move here. Native command adapters already exist; no concrete platform
limitation is documented for this substitute. Choose the adapter while preserving
busy semantics and one keyboard owner, then verify narrow/large-text sheet
geometry and dismissal. Implementation remains pending. See move-axis.md.

M91 candidate: Edit/Move/Move-here SheetActions now delegates to NativeSheetActions
with container-owned keyboard avoidance. Optional secondaryDisabled preserves
Cancel locking during mutation; default false leaves filter dismissal available.
iOS, Android and preview callbacks honor their disabled states. The busy-action
regression failed before the adapter extension;25 focused tests including filter
consumers, TypeScript and structural checks passed remotely. Critic found no
confirmed blocker. Native stacked footer is taller than the prior custom row;
large-text form space and Cancel reachability remain unverified. M91 stays open.

The full mobile suite also passed:1507 tests across258 files on paul. This is
behavioral coverage, not native geometry evidence.

### M92 — Move offers creation while suggestions are unknown

Source-confirmed P2, S134/S135 loading and recovery. Current-query suggestions
were converted to an empty array, so the existing same-kind/title/parent check
offered Create during debounce or failed lookup. The candidate now requires known
results before offering creation and locates retry/loading in the results scroll.
Cached results still support the existing check; this is not global uniqueness.
The regression failed before the change;14 action-sheet tests, TypeScript and
structural checks pass remotely. Native layout and creation recovery remain
unverified. Query and selected destination are not reset by retry.

Move layout follow-up to M91/M92: both forms now scroll their title, help,
query, previews and results together, with only completion controls fixed. The
280-point result cap is removed. Two regression assertions failed before the
change because query entry was outside the scrolling region;14 action-sheet
tests, TypeScript and structural checks passed remotely. Critic review requires
full query visibility before native typing, now reflected in the journey. Native
footer/keyboard reachability remains pending; no visual closure is claimed.

Add follow-up: S086/S087 had the same unknown-result creation offer. Add now
shares one eligibility decision between the offer and command, requiring known
current-query suggestions while retaining the existing name-match heuristic.
The regression covers debounce, failed lookup, retry, known empty results and
retained query; 10 Add tests, TypeScript and structural checks passed on paul.
The render harness needs a second settle after debounce to observe the query
subscription; the corrected test fails against the original creation offer.
Critic review found no confirmed issue. Native suggestion controls, keyboard
behavior and creation recovery remain pending. This follow-up is excluded from
the interim release of PR148.

### M93 — Edit tag-name rejection has no explanation

Source-confirmed P2, S133 recovery. Names over the resolver's limit disabled
Add tag without feedback. The candidate adds Use a shorter tag name beside the
entry controls using the existing resolver status; it preserves the typed value
and draft and clears after correction. Native color validation remains separate.
The regression failed before the change;24 resolver/action-sheet tests, TypeScript
and structural checks passed on paul. VoiceOver announcement and large-text
placement remain unverified. Long selected-tag truncation is a separate pending
review, not fixed by this validation message.

### M94 — Edit ignores large-tag-set disclosure

**P2, source-confirmed contract drift.** EditTagPicker in AssetDetailSheets.tsx
uses tags.map with no bounded initial choices or show/hide control. The asset-tags
spec requires twelve naturally sorted initial options, a retained selected-tag
summary and explicit disclosure for larger sets. This is a project requirement,
not a numeric Apple guideline. Large inventories crowd the form and move inline
creation farther down. No runtime clipping is asserted.

Acceptance: with more than twelve tags, initially show the specified ordered subset
and retain all selected tags in the summary; expand/collapse without changing the
draft; assign an initially hidden tag, collapse and save without losing it. Verify
large text, keyboard, native disclosure actions and accessibility state on device.
The correction is not in PR146 or its interim release. See edit-tags-axis.md.

M94 correction candidate: Edit now sorts naturally and initially shows twelve
options plus selected extras. Pending definitions remain visible. Native Show all
tags / Show fewer tags commands only change disclosure. A real-route regression
failed before the fix and verifies ordering, hidden-tag selection, collapse and
Save retention. Sixteen action-sheet tests, TypeScript and structural checks run
remotely; native reachability and enlarged text remain pending.

### M95 — Unstaged Edit tag input could be discarded silently

**P1, source-confirmed draft loss.** EditTagPicker held name/color locally, outside
the route dirty check. Typing a new tag then Cancel returned without confirmation;
saving another field could omit that entry. The failing real-route regression
confirmed missing discard feedback. The candidate moves the entry into EditDraft,
counts nonblank name or color as dirty, and disables Save with a nearby instruction
until Add tag stages it or the entry is cleared. Staging and entry clearing are
atomic; normalized command data excludes the unfinished entry.

Sixty-seven remote action-sheet/edit/expiration tests, TypeScript and structural
checks pass; critic found no confirmed issue. Native input, color-picker callbacks,
message placement and discard interaction remain pending. This correction is after
PR146 and excluded from 0.24.21.

### M96 — Add drops unfinished tag input when details closes

**P1, source-confirmed draft loss.** AssetTagPicker owns newTagName/newTagColor
locally and is conditionally mounted by showDetails in AddAssetScreen. Collapsing
More details discards that entry. AddAssetDraftStore persists selected IDs and
staged definitions but has no unfinished-entry field; Save also omits it.

Correction must route-own and persist unfinished entry within the existing scoped
draft store, retain it across details collapse and close/resume, and prevent silent
omission on Save. Clear draft and successful staging must clear the entry
intentionally. Test name-only/color-only input, unrelated draft changes, and scope
isolation; verify keyboard, native color callbacks and feedback on device.
Source review only so far; no implementation or runtime claim. See add-tags-axis.md.

M96 correction candidate: unfinished entry is now route-owned and included in the
existing scoped Add draft. Disclosure and remount preserve it; native Save and
its command guard reject omission. Collapsed details has an adjacent reopen
instruction. Add tag updates selections, staged definitions and cleared entry in
one callback; Clear draft and successful Save reset it. The regression failed on
collapse before the fix, then19 remote Add tests, TypeScript and structural checks
passed. It covers scoped restoration, color storage, clear and Save retention.
Critic review found no confirmed issue. Native typing/color/layout acceptance is
still pending; other Add search/validation findings remain separate.

M87 follow-up after iPad349548 inspection: the Edit title moves into the form
scroll with metadata and fields, retaining separate completion actions. Its
containment regression failed before the change. Actual enlarged-text scrolling,
button measurement and footer reachability remain pending on the corrected build.
See native-evidence.md for the old119-point viewport and search-selector findings.

M93 shared-consumer follow-up: Add had the same unexplained overlong-name
rejection as Edit. It now resolves once per render for both eligibility and
staging and displays Use a shorter tag name beside the field. The regression
failed before the change, then18 Add/resolver checks, TypeScript and structural
checks passed remotely. Critic found no confirmed issue. Native feedback and
announcement remain pending; existing color validation is separate.

M94 shared-consumer follow-up: Add previously hid all unselected tags until a
query was entered, contrary to the initial-choice disclosure contract. Add/Edit
now share naturally ordered choice presentation with twelve initial matches and
retained selected extras. Add trims search, retains selected choices across search
and shows No matching tags when appropriate. Native disclosure actions remain
in-place. The Add regression failed before the fix;26 Add/Edit tests, TypeScript
and structural checks passed remotely. Native discovery/large-text acceptance
remains pending. This is a project contract, not an Apple numeric requirement.

### M97 — Add parent commands use custom controls and lose their pending label

Source-confirmed P2, S086/S087 task, targets and accessibility. Retry suggestions
was a bare text Pressable without a minimum target; quick creation used a custom
bordered button that replaced its label with a spinner while pending. These are
in-place commands, so the existing NativeCommandButton is the selected platform
adapter. Searchable parent selection remains in the form because the options are
hierarchical and query-driven; changing these commands needs no new destination.

The candidate gives Retry the shared native target and draft-busy guard and keeps
Creating place… as a disabled, named command. Existing command ownership and
failure recovery remain intact. Ten Add behavior tests, TypeScript and structural
checks passed remotely after two failing command-label regressions. No shared
adapter changed. Native multiline measurement, VoiceOver announcement and keyboard
reachability still require verification; M97 remains open. Apple button guidance
is the relevant topic, but the current documentation page returned a JavaScript
shell during this pass; this is a project adapter decision, not a newly verified
quotation or claim about an Apple requirement.

### M98 — General TestFlight builds reject created invitation links

P1, R048 privacy/recovery. The user supplied
[evidence](evidence/user-invitation-link-error.jpg); exact installed build is unknown.
Source confirms the release sets the invitation origin to empty, while the creation
parser required a configured origin for HTTPS. Every normal creation response in
that configuration failed validation. The error also conflates several rejection
causes; the screenshot alone cannot establish the actual server mutation outcome.

The spec now distinguishes outgoing authenticated creation responses from incoming
links. The server that mints the token is the authority for its browser acceptance
URL. The candidate permits a canonical HTTPS creation response in the general
build, retaining explicit origin pins and all path, credential, token, field and
identity checks. Incoming trust and verified app-link declarations do not change.
There is no automatic navigation or credential request to the returned URL.

Two regressions failed before the fix. Eighty-five mobile invitation/sharing tests,
TypeScript and structural checks passed remotely. The generated HTTP client with a
controlled transport verifies the selected API/auth header, valid response,
401/403 rejection and malformed/cross-scope responses. Existing real API invitation
create/accept/revoke, malformed-token and expiration endpoint tests also passed
on paul. Critic review found no confirmed security blocker. Device creation and
copy/share acceptance remain pending; this fix is outside the running PR148 release.

### M99 — Sharing failure feedback hides the header and obscures mutation state

P2, R048 layout/recovery. The same user screenshot shows a banner covering the
navigation area. Source also refreshes invitations only after the creation response
passes link validation, so a rejected link can leave the list stale after server
creation. Preserve the email draft and validation boundary, refresh safe metadata
when mutation succeeded, and explain an unavailable link separately from failed
creation. The candidate moves creation failure feedback into the current form, outside the
overlay system. A typed link-unavailable outcome is emitted only after a successful
response matches tenant/inventory metadata. It invalidates safe list data and
explains that the invitation was created but its link cannot be used; the user can
cancel it before retrying. Rejected URLs never enter the error. Email/access are
retained, and feedback is cleared on focus/scope change or another attempt. Late
failures still require the captured focused scope.

The new route regression failed before implementation and now verifies list refresh,
inline scroll ancestry, draft retention and absence of a usable link. It covers
success A followed by unavailable link B, with distinct B metadata proving refresh.
A review-discovered stale A link is cleared when a new attempt starts; the link
lifetime explanation now says this explicitly. Thirty-four
sharing tests, TypeScript and structural checks passed remotely. Native inline
feedback visibility/announcement remains pending. Other overlay uses, including
copy/share/cancel feedback, were reviewed in a follow-up: link status and failures
now appear beside the one-time link, and cancellation failures beside their row.
Retries clear their old message; focus/scope changes clear task feedback. A link
operation generation also rejects delayed results after another action or new
creation. Five added regression cases cover local placement/retry and old-link
success/failure after replacement. The earlier creation ancestry assertion now
checks the actual ScrollView type, rather than accepting a null ancestor. Native
visibility and announcements still need verification. Other screens using global
banners remain outside this fix; it does not certify global banner layout.

### M100 — Move has unreadable disabled action and excessive summary chrome

P2, S134 appearance/layout/task. User
[evidence](evidence/user-move-disabled-contrast.jpg) shows dark text on a black
primary action, a large wrapped title and identical From/To blocks above the
picker. Exact build is unknown; the current native footer must be tested for
appearance inheritance, tint and disabled contrast before choosing a repair.
Simplify the context and distinguish the current parent from an actual destination
change. Keep Move disabled until a valid change is selected and Cancel reachable.
Native verification must include dark/light, long titles and large text.

Context candidate: Move now has a short task heading and a separate wrapping
asset name, with no oversized name heading or instructional subtitle. Current
location appears once in quiet form context; Move to appears only for a changed
destination. The colored padded summary panel is removed. Seventeen action-sheet
behavior tests, TypeScript and structural checks passed remotely after the changed
context regression failed first. It preserves valid-change gating and mutation
locks. Critic review found no blocker. Native long-title layout and disabled-action
contrast remain unverified; M100 stays open. The appearance provider already calls
Appearance.setColorScheme, so a missing explicit SwiftUI Host scheme alone is not
evidence of the contrast cause.


### M101 — Sharing commands lack native treatment and link-operation progress

P2, R048 task/loading, source-confirmed. Create used a custom filled button and
Copy/Share custom outlined controls despite an existing native command adapter.
Copy and Share also accepted duplicate/overlapping operations without pending
feedback. Two controlled regressions failed before the correction.

Create now uses native primary emphasis; Copy and Share use ordinary native text
commands. Their shared lock prevents overlapping activation, pending text names
the operation, and failure restores both commands while retaining the link.
Replacement/focus generation protects newer work from older completion; regression
cases cover old success and failure while the replacement operation stays locked.
No shared adapter behavior changed. Cancellation's custom row control was
addressed in the subsequent M102 candidate; its native confirmation remains appropriate.

This choice follows the project's native-control preference and Apple's
[button guidance](https://developer.apple.com/design/human-interface-guidelines/buttons),
which discusses appropriate button styling and communicating pending activity.
The specific adapter and lock are engineering choices. Eighteen selected route
and native-adapter tests, TypeScript and structural checks passed remotely.
Current-build normal-text light/dark appearance, keyboard reachability and system
share return remain pending; this is not native visual certification.


### M102 — Invitation cancellation loses pending ownership and uses an ambiguous X

P2, R048 task/loading/lifecycle, source and controlled-render findings. One
`cancellingId` represented all rows, and the confirmation callback had no duplicate
or focused-session guard. A second cancellation made the first row appear idle;
replayed confirmations submitted again, including after leaving the route.
Three failing regressions established these defects.

The native contextual menu now names Cancel invitation as a destructive action.
This adds discovery before an infrequent irreversible operation; retaining the
recipient-naming confirmation is a deliberate project choice. Each scope and
invitation has its own pending key and Cancelling… status. Confirmation callbacks
are single-use and bound to their focused session; authorized in-flight commands
still finish and update scoped cache. A follow-up regression also reproduced an
old confirmation submitting again after its failed operation finished; a one-shot
confirmation guard fixes that case.

The normal-text native Sharing walkthrough now opens the menu, confirms, observes
failure, and retries. Current-build menu presentation, progress visibility and
return behavior remain pending. No shared native adapter or API permission rule
changed.

Combined validation after M102: all 1,535 mobile tests in 258 files passed remotely,
followed by mobile TypeScript and structural checks. Critic review found no
remaining blocker after the one-shot confirmation regression was added. These
checks include the current branch's Sharing, parent-command, Move context and
native-command changes; they do not establish native runtime appearance.

### M103 — Global notices reserve no space for native navigation

P2, S128 navigation/layout, source-confirmed with historical user evidence in M99.
AppNotice is a root absolute layer at safe-area top plus small spacing, with no
active-header geometry. Sharing's local-feedback fix does not repair the other
38 call sites. See [global notice review](global-notice-axis.md) and its complete
call-site inventory. A correction must preserve cross-navigation View/Undo while
keeping notice actions and native chrome reachable; a guessed header offset or
blanket alert replacement is insufficient. Current-build native placement remains
unverified. A screen-scoped presentation candidate is now implemented; see the global notice review for source evidence and remaining native gates.

### M104 — Global notice content/actions survive service transitions

P1, S128 privacy/lifecycle, source-confirmed. AppFeedbackProvider wraps the inner
services gate and retains ActiveNotice through sign-out, session expiry and server
change. Those transitions do not clear the notice or invalidate its action closure.
Old item/profile text and actions can remain on onboarding or a replacement session.
This does not prove an API authorization bypass. Add controlled transition/action
regressions before implementing a service-context ownership boundary; preserve
same-session completion handoffs. See [global notice review](global-notice-axis.md).

M104 candidate: services state now supplies the provider's notice scope. Owner
cleanup rejects old publishers and action callbacks across transitions/unmount,
while same-context View/Undo survives. The new-owner cleanup race is covered by a
layout-effect publisher test. All 1,538 mobile tests, TypeScript and structural
checks passed remotely. Five additional mounted cases exercise the production gate
with the real OnboardingCommand and controlled ports: sign-out, server change,
expiry, reconnection and rejected push cleanup. Startup/composition counts remain
stable across notice changes; successful transitions invalidate old actions, while
failed cleanup preserves the current session. The targeted 18 tests, TypeScript
and structural checks pass remotely. Critic found no extraction regression. Native
transition visibility remains pending; M104 stays open. See the global-notice
appendix for evidence limits.

### M105 — Departed provider tasks still navigate or present completion

P1, provider creation/detail and credential/prompt editors; navigation/lifecycle.
Source inspection and five initially failing mounted regressions show that saves
and profile actions can publish notices or navigate after leaving and returning.
A retained archive confirmation can also start its command from the earlier visit.
The service-wide notice boundary does not invalidate same-session navigation tasks.

Creation, credential, prompt and detail actions now capture a provider task focus
session keyed by command/resource identity. Late outcomes cannot publish notices,
navigate or explicitly refresh the new screen. Existing mutation observers still
invalidate the original tenant's cache; authorized requests finish and synchronous
pending guards remain held until settlement. Archive confirmation uses the visit
that opened it. This changes presentation ownership, not API authorization.

Tests cover success/failure after blur/return, retained confirmation, successful
credential cleanup and replacement-profile input. Review caught the first patch
skipping secret cleanup after blur; a failing regression was added and the keyed
form now clears its own submitted secret even when navigation has changed. Failed
credential replacement retains its draft. Native leave/return, keyboard, notice
placement and current-profile interaction remain pending; M105 stays open.

M105 validation: 46 settings/query/mutation-observer tests, mobile TypeScript and
structural checks passed remotely; the final replacement-profile regression also
passed in the 32-test Settings suite with TypeScript. Critic re-review found no
remaining blocker after the credential cleanup correction.

M105 neighboring-stage follow-through: VoiceCapabilityScreen had the same late
feedback/reload path in service selection, test and enable. Six departed
success/failure regressions failed before reusing the visit guard, keyed by query
scope and capability. Six matching focused cases preserve normal success/error
feedback. The 53-test settings/query/observer set passed with TypeScript and
structural checks; the final 44-test Settings suite and TypeScript also pass.
Native return and notice placement remain pending. In-place picker semantics and
mutation observers are unchanged.

Combined PR150 checkpoint after stage ownership: all 1,565 mobile tests in 259
files, TypeScript and the mobile structural check pass remotely. This includes the
services gate, provider/editor lifecycle and earlier Sharing changes. Critic found
no remaining source blocker for this pass. This is source/runtime-harness evidence,
not native acceptance or complete audit coverage.

### M106 — Provider editor commands and recovery bypass native form patterns

P2, credential/prompt editors. At da14195c, both forms use bespoke Pressable
Cancel/Save controls despite existing native adapters. Save is disabled only while
saving: blank/whitespace credential or prompt is still offered, then application
validation rejects it into a global notice. API failure also uses the global
notice rather than the field context. This is source-confirmed; no current native
editor placement is claimed. Use native Save, field readiness, local recovery and
retained drafts; preserve empty-input server ADC. See
[all-axis editor review](provider-editors-axis.md). Native Save, required-input readiness and field-local errors are implemented as a candidate; native runtime acceptance remains pending.

### M107 — Provider editor navigation can discard an unsaved replacement

P1, credential/prompt editors. Local values live in keyed forms; route Cancel and
success both go Back, and there is no usePreventRemove or equivalent dirty-draft
contract. Back/Cancel can remove entered replacement text; native Back is also not
covered by the local buttons' saving state. Add a task-owned dirty/pending removal
policy with a native discard decision and single authorized successful exit.
Do not persist secrets to solve accidental navigation. M105 fixes departed
completion ownership, not draft protection. A focused-visit removal guard, native discard decision and authorized successful exit are now implemented. Native gesture/removal acceptance remains open; see the editor review.

### M108 — Reminder edits retain outcomes from a departed visit

P2, defaults/type mode, timing and timezone selection. Deferred saves can publish
child errors after blur/refocus; the timing component can also call its completion
callback after departure. The production parent already invalidates successful
save/navigation on blur, but did not prevent the retained child error. Initial
controlled component tests reproduced four failures before the fix.

The shared focused-visit presentation helper now serves reminder and provider
tasks. Pending guards stay locked until the actual request settles. Departed
outcomes cannot publish local errors or navigate; fresh actions work afterward.
Mode and timing drafts reconcile with the latest saved policy on departed
settlement, including an unchanged-policy refresh. This closes an optimistic-state
regression caught by review: suppressing an error alone could leave an unsaved
value looking saved. Current-visit failures still retain the draft for retry.

Six deferred component cases cover success/failure across mode, timing and timezone;
a seventh uses the real settings screen, preference session and HTTP repository
with a controlled failed PUT and unchanged refresh. The strengthened reconciliation
checks failed before correction. All 73 focused tests, TypeScript and structural
checks pass remotely; critic re-review found no remaining blocker. Native
Back/swipe/refocus and keyboard acceptance remain pending. M108 is a source-tested
candidate, not a native completion claim.

Combined M108 checkpoint: all 1,580 mobile tests in 261 files, TypeScript and
structural checks passed on paul. This is controlled-source evidence only.

### M109 — Name-read failure hides successfully loaded checkout history

P2, R008. AssetCheckoutHistoryScreen treated every failed core/name read as access
denial, including a transient500. A mounted regression reproduced ready history
being replaced by “Could not load checkout history.” The candidate preserves
independently loaded records with a separate native name retry and local error.
Actual401/403/404 still hide records. Review additionally found retry clears the
query error before success; three deferred regressions reproduced premature
redisplay. The candidate retains denial in the current scoped asset owner through
pending/repeated failure until a successful core read. Name retry does not reload
history pages. Sixteen remote history/query tests, TypeScript and structural
checks pass. Native recovery reachability and remount-during-retry remain pending;
see checkout-history-axis.md for the complete24-axis source review and limits.
