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

### M110 — Asset command completion revives after leaving and returning

P1, S096/S097/S098. Checkout/return/archive/restore/delete used the mounted asset
owner from M86, so blur/refocus retained permission to show status or errors and
delete could navigate the later visit. A retained lifecycle confirmation could
also start after its originating visit ended. Thirteen controlled cases reproduced
these failures before correction.

The commands now capture the focused visit and scoped core-resource identity.
Late outcomes cannot publish status/notices, start explicit screen refresh or
navigate a later visit. The synchronous operation lock remains until settlement;
a fresh command works afterward. Lifecycle confirmation is single-use and rejects
callbacks from an earlier visit. Authorized requests and mutation observers still
finish. Ninety-one focused detail/query-observer tests plus TypeScript/structural
checks pass remotely; three additional confirmation cases verify repeat callbacks
after successful completion are ignored. This does not cover photos, edit undo,
backgrounding without blur or every mutation path. Native menu/alert/Back and
blur/refocus acceptance remain open.

M110 combined checkpoint: all1,600 mobile tests in261 files, TypeScript and
structural checks pass on paul. Critic found no remaining source blocker.

### M111 — Browse tag rows paint behind persistent actions

P2, R016/S071–S073. The user's normal-size Tags screenshot shows rows continuing
behind Show results and Back; its build number is not established. Current Browse
source uses a transparent sibling footer without measuring its height or reserving
scroll clearance. Expiration already has an opaque measured footer and direct
scroll body; that existing pattern is now shared as NativeFilterSheet.

Browse and Expiration pages retain native actions/search and draft semantics. The
footer uses an opaque theme surface, reserves its measured height in content and
scroll indicators, and shares the existing sheet-boundary keyboard handling. The
native body remains direct to avoid the separately observed nested-sheet failure.
Two mounted consumer tests cover resizing and last-tag selection through Back and
Apply: Browse failed before correction while Expiration passed. Thirteen focused
checks plus TypeScript/structural checks pass remotely.

A native long-tag journey now checks the last row above the entire visible footer,
fully contained action bounds, selection, Back and applied IDs. Its execution,
search/keyboard, light/dark and supported-device acceptance remain pending; callback
and style tests do not prove pixel separation. The footer extraction does not claim
all other sheet-layout findings fixed.

## M112 — Android tab icon sources are missing

P2, source-confirmed at e3367204. `(tabs)/_layout.tsx` supplies only `sf` icons
for Home/Browse. Installed Expo NativeTabTrigger selects Android sources from
`drawable`, `md` or `src`; its Android icon converter cannot use SF symbols.
The tab labels remain, but the Android icon configuration is absent. Add native
Android equivalents while preserving the iOS symbols, then verify both tab states
on Android. No Android screenshot or runtime pass is claimed. R003 imagery;
see [tab-shell-axis.md](tab-shell-axis.md). The candidate now supplies Material
`home` and `grid_view` through the existing adapter; native acceptance remains open.

## M113 — Voice entry depends on an iOS26-only accessory

P1, source-confirmed at b74f28a1. R003 mounts its only fresh-conversation control
inside NativeTabs.BottomAccessory. The installed react-native-screens
`src/components/tabs/TabsHost.tsx` renders that subtree only for iOS with version
at least26. The committed iOS project declares15.1 deployment target, and Android
is an explicit product target. This is not merely a style difference: those
platforms have no initial voice entry from Home/Browse in the inspected shell.

Counterevidence checked: VoiceConversationReturn navigates to `/voice`, but only
on asset/location paths and only when a ready context has realtime state or
history. Settings routes configure voice providers; they do not start a
conversation. A manually entered deep link is not an in-app entry alternative.
No older-iOS or Android runtime capture is claimed.

Provide a persistent, safe-area-aware entry on platforms without native accessory
support. Reuse the voice state/actions and preserve Home's requested Add,
Notifications, Profile ordering; do not add a third tab or force voice through
Settings. Keep iOS26's native accessory. The unsupported native extension is a
concrete reason for a fallback, not permission to replace native tab navigation.
Acceptance: fresh session entry, loading/error entry, listening/send and review
return on Android and older iOS, both tabs, keyboard visibility and tab changes;
verify no duplicate accessory on iOS26. Candidate implementation uses
VoiceTabContent inside both tab destinations, with a reserved sibling action area
and direct stack rendering on iOS26+. VoiceAccessoryContent shares the existing
presentation/actions with the native wrapper. Three platform rendering cases and
a real provider/controller test cover fresh recording, sending, navigation and
return to Browse. Controlled tests do not prove native tab or keyboard clearance.

## M114 — Browse tag filters leave empty results unexplained

P2, source-confirmed at7017709b. BrowseFiltersScreen rendered an empty section
both when the inventory had no tag options and when the search matched nothing.
Two controlled tests reproduced the missing explanation. The candidate uses the
existing section footer to distinguish `No tags available` and `No matching tags`;
Back and Show results remain available. Clearing native search restores choices
with the existing selected draft intact. No tag-creation task is added to filters.

Eight focused Browse/filter-footer tests, TypeScript and structural checks pass
remotely. Expiration already renders a no-match message and is unchanged. Native
search focus, message visibility and assistive-technology announcement remain
unverified; source text assertions do not establish those properties. R016/S072
recovery and search; implementation ready for native acceptance in the larger batch.

## M115 — Browse filter verification can outlive its focused visit

P2, source-confirmed at00706b0f. useBrowseFilterNavigation aborted on unmount,
explicit cancellation and scope changes, but not on blur. An outstanding scope
read could therefore navigate or show an error after leaving the sheet while it
remained mounted. Two deferred success/failure cases reproduced the missing abort.

The candidate binds presentation to the shared focused-visit owner and aborts on
blur. Unfocused calls cannot start a read; returning permits a fresh request while
the abandoned read settles. Its late success/failure cannot navigate, annotate or
clear the new request's busy state. Existing scope validation is unchanged.
Fifteen focused filter/navigation tests, TypeScript and structural checks pass
remotely; code tests cover controlled focus events, not native sheet transitions.
Native interruption/return remains pending. R016 navigation/lifecycle.

M111 integration follow-up: source review at e303e4f7 found that the production
ready Browse route still wrapped the filter screen in a View, unlike its direct
native fixture. This violated the specified direct-scroll structure; no claim is
made that it caused the user screenshot. The candidate now returns the screen
directly and places verification error text inside its scroll content. A new
controlled test reproduces missing in-content recovery before the change and
verifies the alert, draft and Apply remain available afterward. Native route
geometry still requires verification.

## M116 — Loading filter sheets lack an explicit Cancel action

P2, source-confirmed at c27b1e7c. Both Browse and Expiration filter routes rendered
only a progress indicator until choices arrived. Their ready/error states had
Cancel, but pending users had to rely on native dismissal gestures/platform Back.
The candidate shares a direct scroll loading body with visible loading text and
a native Cancel command wired to each route's existing dismissal handler. It does
not wait for choices or apply any filter. A controlled regression verifies the
command is usable while the loading body remains mounted. Native sheet dismissal
and query transport cancellation are not established by that component test.

## M117 — Expiration filter search retains the persistent header field

P2, source-confirmed at b15a38d5. ExpirationFiltersScreen requests stacked search
with hideWhenScrolling false on types, tags and locations. Browse uses the shared
NativeNavigationSearch integrated button. This preserves the vertical-space
problem the user asked to remove and duplicates native search lifecycle wiring.
Reuse the shared adapter while preserving local filtering, selected IDs and
page-return clearing. The candidate now reuses that adapter with a separate lifetime for each page;
13 focused tests, TypeScript and structural checks pass remotely after a failing
regression. Code review found no confirmed blockers. Native acceptance remains outstanding;
verify compact initial presentation, open/search/clear/close, page changes, and
retained selections on iPhone and iPad. This is a project consistency requirement,
not a claim that Apple forbids stacked search in every context.

M117 native follow-up updates the existing Expiration keyboard journey to open
the Search button, capture the collapsed header, verify the full query, and
observe a previously present nonmatching choice disappear while the matching
choice remains. Footer reachability assertions remain. Two fixture-preparation
checks pass remotely; this does not compile Swift or establish native acceptance.

## M118 — Expiration results retain permanently expanded search

P2, source-confirmed at4548bac1. ExpirationWorkspaceScreen still configured stacked
search after Browse and filter selections adopted integrated-button search. The
candidate changes native placement and disables toolbar integration, preserving
the existing debounced route-query hook and its flush-before-navigation behavior.
Unlike the immediate-query selection adapter, this screen needs pending text
available to its filter action without waiting for debounce.

A failing regression reproduced the stacked placement. Three focused tests,
TypeScript and structural checks now pass remotely, including pending text passed
to Filters and immediate clear with no later debounce. Native compact header,
search focus, keyboard, filter return and scrolling acceptance remain pending.
R018/S074 search and keyboard; this is not a full results-surface audit.

## M119 — Expiration Retry starts the pull indicator and cannot recover mismatched scope

P2, source-confirmed at50d48f3c. The error button receives the same onRefresh as
the RefreshControl, calling usePullRefresh.refresh despite no pull gesture. A
successfully loaded but mismatched inventory produces an error with that same
Retry; its callback performs no read when matches is false. Separate command
retry from gesture presentation and give inventory mismatch an actionable return
path instead of an ineffective retry. Preserve loaded pages during recoverable
errors and test command, pull, mismatch and navigation-return paths independently.
The candidate separates query retry from gesture presentation and supplies a native
Return to Home command for mismatch. Retry is disabled and labeled while reads
are in flight. Eight focused component/refresh tests, TypeScript and structural
checks pass remotely after a failing recovery regression. These checks prove the
component action separation, not full route navigation or native spinner geometry.
Native acceptance remains outstanding. R018/S074 loading/recovery.

## M120 — Pending Expiration search can update a departed screen

P2, source-confirmed at4eb45d62. Search debounce was cancelled on unmount but not
blur, and hidden callbacks could still apply route queries. A controlled regression
reproduced the hidden update. The candidate cancels pending timers on blur, ignores
hidden input, and restores retained text with a fresh debounce on return. External
query changes while blurred supersede retained input. Explicit focused flush is
unchanged. Eight focused checks, TypeScript and structural checks passed remotely;
the strengthened focus test also covers the timer alone and external replacement.
Code review found no confirmed blockers. Native focus/text restoration and actual
route transitions remain pending. R018/S074 search/lifecycle.

## M121 — Sharing iOS email typing candidate after native truncation

P2, native-observed in run349789 atb375d4: final iPad email remained
`a@example.invalid` after typing `audit@example.invalid`. The field was visible
and focused. The candidate removes per-keystroke controlled value feedback only
for iOS Sharing email, using a mount-stable native seed. Scope changes and
successful creation replace its lifetime; failed creation/metadata reads retain
it. Android stays controlled and does not remount after success.

A source regression first failed on the native-owned contract. Twenty-two Sharing
tests, TypeScript and structural checks pass remotely, covering identical retry
submission, native field retention, successful clear and scope replacement on
both platforms. This is a candidate, not proof that controlled feedback caused
the native truncation or that it is fixed. The existing full-speed native email
and creation-recovery journey is unchanged and must pass on iPhone/iPad.

M121 review caught a candidate mismatch after a403 temporarily hid the form: an
empty native field could accompany retained submission state. The final candidate
seeds remounted fields from the same-scope draft and gates first-render seeds for
scope replacement. Both platform cases now cover denial/hide/recovery before retry
and successful clear. This correction is not shipped native evidence.

## M122 — Shared native search accepts callbacks while hidden

P2, source-confirmed at01df88ac. NativeNavigationSearch only deactivated when
disabled or unmounted, so a retained hidden route could receive input/submit/close
and change its caller's query. A mounted regression reproduced all three events.
The candidate binds active state to route focus and enabled state, ignoring hidden
callbacks and re-enabling interaction on return. Reviewed consumers: Browse list
(SearchScreen), InventoryMapScreen, Browse tags, Expiration selections, asset
contents, and TimeZonePicker. Caller-owned timers/requests are not cancelled by
this adapter fix and need separate ownership review.

All1,619 tests/265files, TypeScript and structural checks pass remotely in
/tmp/mobile-shared-search-full.log. Native search text restoration, focus delivery,
keyboard transitions and each consumer's return behavior remain unverified.

M122 critic found no confirmed source blocker. Native return must specifically
check that the field text still agrees with retained results if UIKit clears its
field during dismissal while the callback is ignored; callback availability alone
does not establish that consistency.

## M123 — Browse debounce submits after leaving its route

P2, source and mounted-test confirmed atcd0ed36b. A typed query's300ms timer
survived blur and started its search while hidden. The candidate cancels the
timer on blur, retains text, and resumes using current submission callbacks on
return. Hidden scheduling/submission is rejected. Existing List/Map handoff and
filter-route settling semantics remain covered. Twenty-three focused tests,
TypeScript and structural checks pass remotely after the reproduced failure.
This does not cancel previously started query reads or certify native navigation;
actual focus delivery, returned field text and external route transitions remain
native acceptance work.

M123 follow-up covers external route replacement while blurred: the repository
receives the replacement query, never the abandoned draft, and no deferred
setParams overwrites the replacement on return. All10 mounted Browse cases,
TypeScript and structural checks pass remotely; critic found no confirmed issue.
This is controlled route-prop evidence, not native navigation acceptance.

## M124 — Map search updates its path after navigation away

P2, source and mounted-test confirmed. The map search debounce survived blur and
selected a path while the route was hidden. Focus cleanup now cancels the timer;
unfinished searches resume on return, while deliberate branch-navigation
cancellation stays cancelled. Hidden manual submission is ignored. The regression
failed before correction; all13 focused Map/Browse cases, TypeScript and mobile
structural checks pass on paul. Code critic found no confirmed blocker. Native
focus delivery, returned search text and map scroll/highlight behavior remain
pending acceptance.

## M125 — Map search has no no-match feedback

P2, S069, source-confirmed at608b1e54. submitSearch clears highlightedAssetId when
findInventoryMapSearchMatch returns undefined, then returns without any status.
The previous branch remains visible, so a completed unsuccessful search is
indistinguishable from an unchanged map. Provide scoped no-match feedback while
preserving navigation context; clear that feedback on query clear or deliberate
navigation. No correction or runtime acceptance is claimed yet. See map-axis.md.

## M126 — Map recovery and empty-column commands bypass native actions

P2, S068, source-confirmed at608b1e54. Retry map invokes refreshMap, which owns the
explicit pull indicator, and its text-only Pressable has no minimum hit area.
Empty-column Add is a custom Pressable with minHeight40. Reuse native command
adapters and separate Retry from the pull gesture. Verify loading/duplicate retry,
empty-column navigation and native target geometry. Not yet corrected.

M126 candidate now uses NativeCommandButton for Retry and empty-column Add,
retaining the existing view-model permission gate. Retry owns separate pending
state, stays visible/disabled during retry, rejects duplicate and hidden callbacks,
and never starts the pull indicator. The regression failed before correction;
21 Map behavior/presentation tests, TypeScript and structural checks pass on paul.
Critic caught the initially omitted focused-start guard; it is now restored and
tested. No remaining confirmed source blocker. Native recovery/command geometry
remains pending.

M125 candidate displays inline no-match guidance or matched title/placement,
scoped to the searched query and map snapshot. Changed/cleared query, unavailable
or refreshed data and deliberate navigation suppress stale outcomes. The mounted
case failed before correction;9 focused Map tests, TypeScript and structural
checks pass remotely. Critic also caught a retained pan cancellation callback
recording an older query. Stable cancellation now reads the current query; a
regression reproduced the old blur/refocus rerun and passes after correction.
Critic reports no remaining confirmed blocker. Native outcome layout, announcements
and gesture/return behavior remain pending.

## M127 — Appearance save failures outlive their selection

P2, source/mounted confirmed atbb3fc720. AppearancePicker showed a save error after
leaving its screen or choosing a newer preference. Both cases failed before the
correction. Focus and selection ownership now suppress obsolete feedback and
hidden starts, preserving global provider rollback and queued persistence.
Both consumers were inspected: inline Settings and the Appearance detail route.
Seven focused picker/controller tests plus TypeScript/structural checks passed;
an additional blur/refocus-before-rejection case also passes. Critic found no
confirmed blocker. Native menu, navigation, error placement and theme-transition
acceptance remain pending.

## M128 — Home expiration entry lacks an inventory-scope gate

P2, S063, source-confirmed at8b0dfc7b. ExpirationHomeEntry defaults unresolved
tenant/inventory IDs to empty strings but still supplies active onOpen navigation.
ExpirationHomeSection renders See all while data is loading or errored. A tap can
therefore push /expiration with empty scope. Separately, the entry hides resource
access failures but not inventory-scope errors; cached rows can remain visible
when inventory scope has failed. Gate navigation on usable scope and hide data
while scope is unavailable/errored, preserving explicit retry. Verify unresolved
scope, cached-data scope failure, retry and recovered navigation at the mounted
entry boundary. No server authorization bypass is claimed; this is client
presentation/navigation correctness. Not yet corrected.

M128 candidate extracts injected query-driven content from the bootstrap wrapper.
Usable inventory scope now gates queries, cached presentation and See all. Scope
failure and its pending retry keep counts hidden; successful recovery restores
scoped navigation. Mounted regression reproduced empty IDs before correction.
Critic caught retained ready-row callbacks bypassing the initial render guard;
committed scope ownership now rejects them, including after unmount, with a
reproduced/passing failure regression. Three focused tests, TypeScript and mobile
structural checks pass on paul. Critic found no remaining confirmed blocker.
Native loading/error/recovery layout and actual inventory switching remain pending.

## M129 — Asset load recovery still uses a custom command

P2 platform consistency, source-confirmed at801eb1e6. AssetDetailRouteErrorState
used a hand-styled Pressable. It now uses NativeCommandButton labeled Retry asset,
preserving canRetry and the flexible scrollable explanation. Both asset route
wrappers share this screen. A mounted test replaces legacy mocked tree invocation
and covers retry execution, non-retryable absence and scroll content. It failed
before correction; the focused test, TypeScript and structural checks pass on
paul. Critic found no confirmed blocker. Actual native spacing/wrapping remains
pending and is not implied by the mounted test.

## M130 — Home and Add recovery controls still use custom buttons

P2 platform consistency, source-confirmed at8fb0a54d. Home dashboard load error,
Home expiration refresh error and Add context load error now reuse the shared
native command adapter. Existing accessible labels, callbacks and readiness/scope
gates are retained; unused Home button styling is removed. Existing Home and
expiration28 tests plus Add/Home styles15 tests pass remotely, with TypeScript
and structural checks. Critic found no confirmed blocker. Native spacing and
reachability remain pending; this bounded adapter migration introduces no new
recovery semantics or claim of full screen acceptance.

## M131 — Browse inline and photo recovery use custom commands

P2 platform consistency at09c3506b. BrowseHeader's inline retry and AssetDetailView's
failed-photo retry now use NativeCommandButton. Eligibility, callbacks and Retry
labels remain unchanged. Browse explanation and command stack vertically so the
full-width native host does not compress its error text. Unused custom styles are
removed. Sixty-two existing Browse/asset behavior and view tests, TypeScript and
structural checks pass remotely. Critic found no confirmed blocker. Native spacing,
contrast and reachability remain pending. No new photo persistence semantics are
introduced by this adapter change.

## M132 — Browse empty and pagination commands bypass the native adapter

P2 platform consistency, source-confirmed atf8677508. BrowseResultStates retained
custom commands for empty inventory, empty search/refinements, initial load failure
and pagination failure. All now use NativeCommandButton, retaining specific labels,
callback targets, primary/standard prominence and viewer Add absence. Explanations
stack above commands with stretch alignment for the native host. Mounted tests
replace mocked tree traversal and verify recovery callbacks and viewer eligibility;
their initial missing-label failure reflects adapter wiring, not proof of a prior
runtime accessibility defect. All13 focused Browse tests, TypeScript and structural
checks passed on paul. Critic found no confirmed regression. Native full-width
geometry and actual retry/clear navigation remain acceptance work.

## M133 — Invitation acceptance commands remain custom

P2 platform consistency, source-confirmed at6f45c031, R019. InventoryInvitationScreen
uses hand-styled Pressables for Join/Open, start-over, Not now, retry, account switch
and Done. The existing native command adapter should supply these controls, while
preserving pending labels, disabled states, explicit acceptance and recovery. The
review route itself fits the task; no extra selection menu is recommended. See
invitation-acceptance-axis.md for all24 axes and evidence limits. Implementation
and normal-text native acceptance are pending. This is not a claim that source
inspection established a particular rendering defect.

M133 implementation now uses NativeCommandButton for all invitation commands.
Join/Open retain stable names and disabled state, with adjacent named busy progress;
start-over also has pending feedback. Opening failure preserves accepted access
and re-enables Open. The legacy mocked-hook suite is replaced by mounted tests
covering the same9 behavioral/adaptation cases. Retry/Done recovery is newly covered;
its initial missing-label RED is wiring evidence, not a runtime accessibility
finding. All17 focused invitation/route tests, TypeScript and structural checks
pass on paul; the3 progress cases also verify start-over status. Critic found no
confirmed blockers. Native spacing, announcements and reachability remain pending.

## M134 — Disabled expiration year still accepts a text callback

P2, source/mounted-confirmed atcd6f6556, S093. A year event delivered to the disabled
render changes its local year and publishes expiration, unlike the already guarded
date callback. The year handler now rejects disabled events. The regression failed
before the correction and verifies unchanged publication/local draft plus resumed
editing. Shared Add/Edit consumers were inspected. All10 focused expiration tests,
TypeScript and structural checks pass on paul. Critic found no confirmed issue.
This covers callbacks delivered to the disabled render, not arbitrary retained
closures or proof of native event timing. Native keyboard and pending-save behavior
remain unverified. See expiration-entry-axis.md for the full field review.

## M135 — Voice photo-source callbacks outlive the review

P2, source/mounted-confirmed ate295fc6f. VoicePlanPhotoDrafts reported selection
errors without checking the active visit/plan, and a retained source choice could
start camera/library selection after leaving. VoiceSessionSheetScreen now captures
the shared visit owner keyed by photo service and plan ID/status; only proposed
plans can open the chooser. The chooser rejects obsolete starts/errors, and photo
results update the draft only while that owner remains current. Add and asset-detail
chooser consumers retain their existing separate guards. The initial inactive-start
regression failed; tests now cover suppression versus current-error feedback and
mounted blur/refocus or plan replacement. All11 focused photo-draft/presentation
tests, TypeScript and structural checks pass on paul. These checks do not establish
native camera/library permission timing or full VoiceSession screen acceptance.
Critic found no implementation blocker; caller-inventory line references were
refreshed after its documentation correction. Full VoiceSession integration and
native permission timing remain pending.

## M136 — Form-sheet notice overlaps its visible native header

P1, runtime-observed on iPad mini in run34992079258, actual artifact revision
514032e0e290a7970262458ae4970c6d6d1edba4 (headf109567c). Inspected screenshot
`evidence/ipad-notice-header-overlap-349920.png` shows the notice behind the title
and Close control. Hierarchy27B6DF21 places the content at y236.5, header at
246.5–300.5 and notice at246.5–316.5. AppNoticeScreenLayout uses zero header
clearance for a nontransparent header; this formSheet's content extends underneath
that header. The fixture declares presentation=formSheet and headerShown=true.
Correction must account for this presentation while retaining ordinary pushed
screen positioning and header-hidden behavior. Keep the native full-notice bounds
and action/Close reachability checks; do not weaken the assertion. Not yet fixed.

M136 candidate uses the reported header height for iOS formSheet notices with a
visible header. Ordinary pushed screens, Android and hidden-header safe-area
placement retain their branches. The mounted regression failed at10 versus74;
12 focused notice tests, TypeScript and structural checks now pass on paul.
Critic found no production blocker; the hidden-header test now asserts the exact
zero-inset offset and its4-test suite passes again. Native full-bounds and action/
Close hit tests are unchanged and must verify the correction in a newer build.

M51 follow-up at062fd211: run34992079258 iPad actual514032e0 again failed center
activation but passed the trailing-well probe. Inspected captures are retained as
ipad-color-row-inactive-349920.png and ipad-color-well-open-349920.png. The native
row exposed704×36 bounds while the actionable well was at its trailing edge.
The candidate keeps SwiftUI ColorPicker but separates its visible label and hides
the picker's own visual label, constraining its named target to44×44. Selection and
disabled behavior remain unchanged. The original failing center/open/close/clear
journey is unchanged; the diagnostic now requires44-point compact bounds and center
activation instead of a trailing-coordinate workaround. All15 color tests,
TypeScript, structural checks and2 fixture-preparation tests pass remotely. Critic
found no blocker. Actual native target bounds, label layout and VoiceOver activation
remain unverified; M51 is not closed by source geometry.


## M137 — iPad search keyboard dismissal is visible but not hittable

P2, S127 and place-detail search. Run34992079258, actual source514032e0,
normal text on iPad mini (A17 Pro): native search accepts `19`, includes Tool19
and excludes Tool0. The next assertion, `Dismiss keyboard.isHittable`, fails
before any dismissal tap. Inspected final-state image and hierarchy are retained
as evidence/ipad-place-search-dismiss-349920.png and .txt.

The hierarchy exposes the button at x680,y739,width44,height44. An ancestor
occupies x0,y749,width372,height44; the child extends outside that ancestor.
This is a plausible hit-testing cause, not a verified root cause. The visible
chevron and presence in the accessibility tree do not establish operability.
The system Hide keyboard action also exists, but substituting it would bypass
the project-required accessory acceptance check. Keep the failing assertion.

Next acceptance: reproduce on the current candidate; verify the accessory's
actual hit region and dismiss without clearing `19`, then open the matching
result and return. Confirm phone, iPad, ordinary inputs and sheet consumers.
No production change or native closure is claimed. The phone place-search test
also failed in this run, at a different line; it needs separate screenshot triage.


## M138 — Retained item-type confirmation can overwrite newer edits

P2, AssetExpirationEditor type-change confirmation. Source68c4daec guards disabled
state when opening the alert but not its retained Change type callback. That
callback can replace a newer draft and clear its date after disabling, asset or
settings replacement, navigation return or unmount. Seven mounted regression
cases failed before correction (six obsolete-owner cases and one duplicate apply).

The editor now uses the shared visit owner with the serialized asset/draft/type
settings/disabled state and consumes a valid acceptance once. Current confirmation
preserves title, notes and tags while clearing the type-dependent expiration.
Fifteen focused expiration tests, TypeScript and structural checks pass on paul.
Critic found no confirmed blocker. Native dialog timing/return verification remains
pending; this is source and mounted evidence, not native visual acceptance.


## M139 — Stored-photo removal confirmation outlives its viewer

P2, AssetPhotoViewerSheet. Atce586b03 the retained destructive alert directly
calls onRemove after viewer close/reopen, changed selection or collection, remove
access loss, pending removal, focus return or unmount. The detail route has an
operation guard, but that does not establish consent for the current viewer visit.
Eight mounted cases failed before correction, including repeated valid acceptance.

The viewer now captures the shared focused presentation owner keyed by collection
IDs, selected index and removal availability, and consumes acceptance once.
The existing deletion command and native destructive alert remain unchanged.
Thirteen focused viewer tests, TypeScript and structural checks pass on paul.
Native alert timing and gallery navigation acceptance remain pending. Add's draft
photo removal is a separate call site and is not covered by this correction.

Critic found no production blocker and requested independent collection coverage.
Added a ninth passing confirmation case that appends a different photo while
preserving the selected photo and index; collection invalidation is now exercised
without relying on selection change. Native acceptance is still pending.


## M140 — Account confirmation and failure feedback survive departure

P2, AccountSettingsScreen and ConnectionSettingsScreen. At0b8e4102 native sign-out
and change-server confirmation callbacks can start after navigating away and back;
pending failures publish feedback in the new visit. Four mounted regression cases
failed before correction. Current acceptance must remain available even when
inventory or identity reads are unavailable; no new prerequisite is introduced.

The existing focused owner now binds confirmation to settings query and principal
or server identity. Acceptance is single-use; late failures suppress their notice
but release the existing pending lock for retry. Current-visit failure feedback
and session-changing callbacks are preserved. All53 settings behavior tests,
TypeScript and mobile structural checks pass on paul. Existing tests retain
legitimate sign-out while identity is pending/failed and inventory is unavailable.
No authentication, authorization or session persistence implementation changed.
Native alert/focus acceptance remains pending.

Critic found no blocker and identified replacement coverage as a useful addition.
Two added query-replacement cases pass while preserving current acceptance (55
settings tests total). Direct principal/server-value changes remain source-covered;
native acceptance remains pending.

## M141 — Initial Account read failure claims retained details

P2, R024 recovery. At570b804c initial principal failure displays the shared refresh
notice's claim that previously loaded values are shown. No values have loaded.
Account now supplies explicit unavailable-details text only when principal data
is absent. Other consumers and retained principal refresh keep existing copy.
A mounted failing-first retry journey now passes; sign-out stays available.
61 settings tests, TypeScript and structural checks pass on paul. Native pending.

## M142 — Account and Connection command rows bypass native adapter

P2 platform-pattern review, R024/R026. SettingsActionRow implements Sign Out and
Change Server with custom Pressable/Text styling although these are commands and
the project has NativeCommandButton. The repository requires actual native adapters
unless a concrete limitation is documented. No limitation was identified in these
callers. Review shared consumers before migration; preserve accessibility labels,
pending state, confirmation semantics and current-visit retry. This is a source
finding against project policy, not an assertion that Apple prohibits grouped rows.
Implementation and native normal-size verification remain pending.

M141 critic found no confirmed blocker; query/session behavior remains unchanged.

M142 implementation follow-up: Account and Connection now reuse the unchanged
NativeCommandButton adapter. Native visible labels are also accessible names;
subject context remains in the value row and confirmation. Pending labels, disabled
state, one-shot ownership and recovery callbacks are retained. The spec records
this deliberate naming change instead of adding a nested accessibility wrapper.
61 settings behavior/server-state tests, TypeScript and structural checks pass on
paul. Native normal-size geometry and VoiceOver acceptance remain pending.
Critic found no confirmed issue; native/assistive acceptance remains open.


M111 native follow-up: run349983 iPhone actual54714ab4 passes last-tag scroll,
selection, Back and Apply. Inspected screenshot/hierarchy retain final row entirely
above the opaque footer and both contained buttons. See native-evidence.md and
phone-last-tag-clear-footer-349983 evidence. iPad/dark/keyboard acceptance pending.


### M143 — Draft photo removal outlives its selection

P2 source and mounted behavior, S090/R007. Add's native Remove photo alert kept
callbacks after selection/collection changes, closing and reopening the viewer,
navigation return, and unmount. It could remove a draft photo and change the viewer
from an obsolete confirmation; repeated acceptance repeated those effects.
Seven mounted cases reproduced failure before the ownership guard.

DraftPhotoPreviewModal now owns the confirmation by collection, selected index,
busy state and focused visit, and accepts it once. Extraction preserves the existing
viewer and native confirmation. Current removal still selects the next valid index
or closes after the only photo. Eight confirmation cases and20 existing Add tests
pass on paul, with TypeScript and structural checks. Native acceptance remains
pending; this is a behavioral correction, not a claim of visual verification.

M143 critic found no confirmed blocker. Combined remote validation at this change
passes1663 tests across269 files. Source checksum comparison against paul showed
no differences before the run. Log: /tmp/mobile-batch-draft-photo.log. Native
run35003739726 remains in progress and predates this final confirmation change.


### M144 — New-conversation confirmation can reset newer work

P2 source/mounted finding, Voice processing and plan editing. The reset alert
retained onReset directly, allowing acceptance after navigation or newer edits,
photo drafts and plan/status changes. Seven mounted cases failed before correction.
The extracted useNewConversation preserves the native alert and immediate reset
policy, owns acceptance by meaningful conversation state and visit, and accepts
once. Meter-only updates keep the confirmation valid.

Nine focused cases include replacement plan with unchanged status/drafts, as
requested by critic, plus status, photos, edits, visit, unmount, meter, current and
empty paths. Existing presentation/history/navigation tests are retained. Native
confirmation timing remains pending. No controller or API behavior changed.

M144 verification: nine focused cases pass on paul; the17 existing presentation,
history and navigation tests also pass. TypeScript and structural checks pass.
Critic found no confirmed blocker, and its additional plan-identity coverage is
now included.


### M145 — Edit discard acceptance outlives its draft

P2 source/mounted finding, R009/S133. Edit's Discard alert checked the operation
lock but accepted after a changed draft or blur/refocus, and repeated acceptance
navigated twice. Mounted tests reproduced those three failures; existing unmount
protection already passed. Acceptance now belongs to asset, draft, saving state
and focused visit and executes once, preserving the pending-operation guard.
All21 Edit/Move behavior cases, TypeScript and structural checks pass on paul.
Native alert timing remains pending.

M145 critic found no confirmed issue; native acceptance remains open.


### M146 — Failed provider Archive confirmation can be reused

P2 mounted behavior, R052. Existing focus/pending guards reject earlier visits and
simultaneous commands, but a retained Archive acceptance could submit again after
a failed request settled. The new mounted case reproduced that second call.
Acceptance now consumes its confirmation before starting the request. The same
test verifies a freshly confirmed retry remains available. All57 settings behavior
tests, TypeScript and structural checks pass on paul; critic found no confirmed
issue. Native alert interaction remains pending.


### M147 — Mounted asset sheets present completion on a later visit

P2 source/mounted finding, Edit/Move/Move Here/destination creation. The shared
operation helper tracked mounting and pending work but not focused visit. New
Edit tests reproduced late-success navigation and a late-error alert after
blur/refocus. Completion ownership is now captured when the mutation begins.
All four consumers check it before notices, navigation, error alerts, created
destination selection and partial-tag draft replacement. The lock still lasts
until settlement and the mounted current form is unlocked afterward.

All26 Edit/Move behavior tests pass on paul, including returned-visit failure
for Move, Move Here and destination creation, current-visit success/failure and
the new Edit success/failure return cases. TypeScript and structural checks pass.
Native presentation/lifecycle acceptance remains pending.

M147 critic found no confirmed blocker. It noted that dedicated returned-visit
success tests for Move/Move Here/destination creation are not yet present; current
success tests and Edit return-success cover the shared helper, but do not replace
those per-branch scenarios. That coverage gap and native timing remain open.


M147 coverage follow-up: dedicated Move and Move Here tests now compare current
and returned-visit success, asserting both navigation and completion records.
Destination creation returned-success asserts the original query remains and the
created destination is not inserted. All31 Edit/Move cases, TypeScript and structural
checks pass on paul. Critic confirmed the added assertions cover those branches.
This closes the per-branch mounted success coverage gap above; native lifecycle
acceptance remains open.

### M148 — Provider screens blur navigation and command controls

P2 source pattern finding, R052/R054/R055. Provider screens used the same custom
action row for editor navigation and mutations despite existing task-specific
adapters. Add Profile, Replace Credential and Prompt Guidance now disclose their
editor destinations with SettingsNavigationRow. Create recommended draft, Test,
Enable/Disable and Archive reuse unchanged NativeCommandButton. Creation labels
name the action; native command accessible names match visible labels. Subject
context remains in the profile heading and Archive alert, whose destructive
semantic is unchanged. This follows project native-adapter policy; it is not a
claim that Apple prohibits all command rows.

Three pending-operation cases failed against the new accessible names before
migration. All57 settings behavior tests now pass, preserving retry, navigation
locks, confirmation ownership and late-result behavior. TypeScript and structural
checks pass on paul. Native spacing, targets and assistive acceptance remain open.

M148 critic found no confirmed regression; native acceptance remains pending.


### M149 — Failed item-type queries still claim loading

P2 source/mounted finding, Add and Edit. Their recovery controls could coexist
with AssetExpirationEditor's “Loading expiration settings…” when no type data
was available after failure. New Add and strengthened Edit cases reproduced both
contradictions. The consumers now hide only the unavailable editor on failed
initial reads; cached usable arrays remain visible during refresh failure. Add
Retry uses NativeCommandButton and respects the pending draft operation lock.
All42 Add/Edit cases, TypeScript and structural checks passed on paul. Added
background-refresh coverage also passes, preserving the selected type/date field
and dirty name after a later failure. Critic found no confirmed regression. Native
error, retry and keyboard acceptance remains open.

### M150 — Item-type search has no empty-result explanation

P2 source/mounted finding, shared Add/Edit AssetExpirationEditor. A nonmatching
query left a blank choice list. A failing behavior case now requires visible
no-match feedback and verifies that clearing search restores the checked type
without publishing a changed draft. The list now shows a polite no-match message.
It stays in the current form and does not add navigation or clear the selection.
Native empty-state visibility/announcement remains pending.

M150 verification: all10 shared editor cases, TypeScript and structural checks
passed remotely. Critic found no confirmed issue in the implementation or audit
classification. The final Add case also asserts the name after cached refresh
failure; all11 Add cases pass. Native acceptance remains open.


M51 native target follow-up: iPad350037 at818c3f38 exposes the compact well as
36×36 despite the44-point SwiftUI wrapper. Retained ipad-color-target-350037
image/hierarchy confirms the named button at x684,y326.5,width36,height36.
Center activation is exercised separately; the minimum-target assertion remains
failed. The candidate now requests SwiftUI's large control size and a minimum
44-point frame, allowing native sizing rather than scaling the drawing. Apple
documents controlSize as the platform sizing mechanism, but does not establish
that this ColorPicker will honor the requested minimum; runtime evidence is
required. No assertion is removed or relaxed.

Consumers remain Add staged-tag color, Edit staged-tag color and Settings tag
customization through TagColorPicker/FullSpectrumTagColorPicker.15 color behavior
tests, TypeScript and structural checks pass remotely. Native bounds, hit testing
and system-picker behavior remain pending; M51 is not closed.

M51 critic found no confirmed source blocker in the native large-size candidate.
The existing target-size/activation gate remains open pending an actual native
run containing this change.


### M151 — Unsupported photo selection silently disappears

P2 source/test-confirmed at fc8f7cd2. ExpoPhotoSelectionProvider skipped known
unsupported MIME types. A selection containing only those images returned an empty
array (indistinguishable from cancellation); mixed selection silently returned a
subset. All three new public-provider cases failed before correction.

The provider now rejects that new selection with a supported-format explanation.
This is an explicit atomic-selection policy, not a claim that Apple requires
rejecting mixed selections. It avoids silently accepting a different set and adds
no format conversion or mislabeled bytes. User cost: a mixed selection must be
chosen again with supported photos; the error names JPEG, PNG and WebP.

Shared consumers inspected: Add catches selection failure before appending to its
photo draft; existing-asset detail catches before invoking upload; voice plan
source chooser catches the rejection and only appends after a successful return.
Existing photos/drafts are therefore untouched by rejection. Native notice/alert
visibility and actual formats returned by each OS picker still require acceptance.

55 related provider, Add-dismissal and asset-workspace tests pass remotely after
correction. Native picker cancellation and mixed-format device journeys are not
claimed verified. Verify supported selections still attach, unsupported selection
shows the explanation, existing photos remain, and choosing again succeeds.

M151 verification: the4 voice photo behavior cases also pass; TypeScript and
structural validation passed remotely. Critic found no confirmed blocker. The
existing absent-MIME JPEG fallback is unchanged: this finding covers reported
unsupported types, not byte-sniffing or full media-format validation.


### M152 — Old connection expiry callback replaces a new session

P1 source/mounted finding in AppServicesFeedbackGate at bcaeadfa. The auth-required
callback captured its connection profile but had no composition ownership check.
After completing a replacement session, invoking an old callback still called the
real onboarding expiry command. Regression cases failed after sign-out, server
change and expiry transitions.

The callback now checks its composition identity before initiating expiry and
before publishing completion/error. Its dialog dismissal has no session mutation; root
cleanup and successful transitions to onboarding invalidate callbacks. This prevents an obsolete callback from starting a
new credential mutation. It does not cancel credential work already started before
replacement, nor certify all possible overlapping sign-in/sign-out operations.

The37 related gate/onboarding cases passed remotely, including the three previously
failing transitions and existing OIDC boundary cases. Additional controlled cases
exercise resolved/rejected late completion and callbacks after teardown. Native
session-expiry presentation and full account transition acceptance remain open.

M152 critic identified the between-session interval before replacement creation.
Two added assertions failed when old callbacks ran after successful sign-out or
server change; ownership now retires on successful onboarding transitions while
failed push cleanup retains the current composition.

M152 final focused validation:39 cases across5 files passed remotely, followed
by TypeScript and structural checks. Critic confirmed retirement now covers the
previously missed interval and found no remaining confirmed blocker in this fix.

### M153 — Inherited definition shows the inventory as its owner

P2 source/mounted finding at438bd902. Details used the screen scope instead of
loaded ownership, contradicting the inheritance explanation. The label now uses
effective inherited ownership. Four cases cover fields/types and household/local
ownership with misleading route hints; the inherited cases failed before correction.

### M154 — Read-only asset type presents an inactive tracking switch

P2 source/mounted pattern finding at438bd902. Inherited/viewer/archived detail used
a disabled mutation control even though the settings spec requires static values.
Read-only tracking now displays Enabled or Disabled. Editable types retain their
switch and pending-operation behavior. Both inherited values failed before the
correction.49 related tests and static checks pass remotely; native presentation
and assistive reading order remain pending. See inherited-definition-axis.md.

### Native follow-up: M51 and text-entry isolation

See [native-350129-followup.md](native-350129-followup.md) for final phone/iPad
results at22a4a80d, the focused color-picker failures, and the next ordinary-input
comparisons. These findings remain open; no production workaround or native
acceptance is claimed by adding diagnostic fixtures.

### M155 — Stale settings Discard can navigate after return

P2 mounted recovery finding at0f378691. The customization Discard callback used
the current workflow without capturing the focused resource that opened the alert.
Both blur and blur/return regressions dispatched OLD_BACK before correction.
The callback now checks captured focus/resource identity and authorizes only its
original workflow. Current confirmations still dispatch once. A replacement-resource
case verifies the new draft remains. Native alert/focus acceptance is pending;
see [settings-exit-axis.md](settings-exit-axis.md).

### M156 — Settings collections retain custom search/Add chrome

P2 source pattern finding atd458d896, affecting R029/R032/R037/R040/R046.
The custom permanent input and scroll-content Add diverged from the accepted
native-header pattern. The shared collection now uses NativeNavigationSearch and
stable native header actions; lifecycle controls and grouped results remain in
content. Initial loading/error/denial removes header controls. Native phone/iPad
acceptance is pending; see customization-collections-axis.md.

### M157 — Cached collection permissions retain Add after revocation

P2 mounted interaction finding discovered during M156. Rows can remain cached while
the permission query changes; Add used the older context state. The new regression
reproduced the stale action. Mutation affordances now use the current permission
snapshot and removed native handlers do nothing. Read-only rows stay visible when
view permission remains. This is an affordance fix, not a server-authorization change.

### M158 — Retained photo-source choice survives a departed visit

P2 mounted finding at39a846d6. Add and asset detail accepted old camera/library
choices after blur/return. All four iOS regression cases opened selection before
correction. Source choosers now capture their opening visit; the shared adapter
requires and checks current ownership before presentation/acceptance. Fresh
choosers still work after return, with both Android choices also covered. Voice
retains its existing ownership guard. See confirmation-review.md; physical picker
acceptance remains pending, and already-started pickers are outside this correction.

### M159 — Add photo tile has no accessible action name

P2 source/mounted finding at39a846d6. The icon-only tile exposed a hint but no
action label. The semantic-label regression failed before adding Add photos.
Selection/geometry is unchanged. VoiceOver naming/order remains a native check.

### M160 — Home pull failure outlives its focused visit

P2 mounted finding at81e91f74. The pull hook retired its spinner on blur, but Home
passed a default always-true predicate to error feedback. A late failure displayed
Could not refresh Home after leaving or returning. Both navigation cases failed
before correction; the current-visit failure case passed and remains supported.
Home now captures focused visit/resource ownership through useTaskPresentation
when a pull starts. Return-command reconciliation keeps its existing predicate.
All28 Home and2 pull-hook cases pass on paul. Native navigation/error presentation
acceptance remains pending; this does not claim to solve every historical spinner
or inset symptom. See home-dashboard-axis.md for the full route review.

### M161 — Asset lists silently swallow explicit refresh failures

P2 source/mounted atfc3bb7a3. Inventory assets, location contents and the retained
LocationsScreen awaited non-throwing refetch results and rendered errors only
without cached data. A failed pull therefore stopped spinning without explaining
that visible results were unchanged. Three current-visit cases failed before the
fix. The shared usePullRefreshFeedback hook now catches throwing reads, keeps
notices scoped to visit/resource and preserves ordinary cached cards. Initial-load
recovery and access suppression still belong to the query adapter. All18 focused
cases pass remotely; native notice placement/reachability remains pending. The
legacy LocationsScreen has no current route consumer and is not claimed as a
third shipped route. See asset-lists-axis.md.

### M162 — History refresh feedback and Retry have the wrong owner

P2 mounted at153a6be3. A delayed failed pull reported over a departed/returned
visit, while inline cached-error Retry activated the pull indicator. Two visit
tests and one Retry test failed before correction. History now uses shared scoped
pull feedback, including history-view identity, and Retry refetches directly with
pending controls disabled. Existing inline failure context, cached pages and
access suppression remain. All12 focused History cases pass remotely; native
acceptance remains pending. See history-list-axis.md.

### M163 — Item-detail pull failure reports after navigation

P2 mounted ata7180e13. AssetDetailRouteScreen's progressive refresh caught an
error and displayed a global notice without checking its original visit. Both
departed and returned cases failed before adding the existing captureCommandVisit
guard. Current-visit feedback keeps the original useful error detail; cached core
content remains visible. The two production imports are asset detail (R012) and
location-context asset detail (R020); both use this component unchanged. All92
focused detail cases, TypeScript and structural checks pass on paul. This is a
targeted loading/recovery/lifecycle review, not a full24-axis detail pass or native
navigation acceptance. Progressive section-specific errors remain independently
owned by their query state.

### M164 — Scoped settings loading does not identify its task

P2 source/mounted at53975079. Household and inventory settings displayed only an
unlabeled spinner while resolving scope. Both mounted task-label expectations
failed before reuse of SettingsLoadingRow with scope-specific copy. The shared
control supplies progress semantics and visible text, then disappears when rows
load. Native announcements/layout remain unverified. See scoped-settings-axis.md
for the full source review and its remaining acceptance limits.

### M165 — Customization Save bypasses the native command adapter

P2 source/mounted at9e61709c. Tag/type/field editors painted a primary Pressable
despite an existing native primary command adapter. They now use NativeCommandButton
with explicit Save/Saving names, existing validation and pending lock, and a wrapper
preserving content insets. The named-command regression failed before correction;
all54 mounted customization cases and static checks pass remotely. Native sizing,
keyboard and reachability remain pending; this does not close M51 color behavior.

### M166 — Customization editor Back and lifecycle actions remain custom

P2 source-confirmed pattern gap at9e61709c. CustomizationEditorScreen installs a
Pressable/Chevron Back and CustomizationLifecycleSection uses custom action rows.
These are commands, not category navigation. No concrete native limitation is
documented for them. Replace with native adapters while preserving collection
replacement, dirty-exit interception, destructive semantics and operation locks.
Inherited Manage action also needs the command/navigation distinction reviewed.
Implemented candidate after M165: native leading Back uses the existing stable
header adapter; lifecycle and inherited Manage commands use NativeCommandButton.
Destructive role is native SwiftUI on iOS and semantic native-button colors on
Android. Collection replacement, Keep Editing/Discard, confirmation and pending
locks remain intact. New adapter assertions failed before implementation. The
mounted Back test renders the installed navigation header and verifies Keep
Editing and exactly-once Discard. Remote full suite: 1,734 tests/271 files, type
check and mobile structural checks passed. Code critic found no blockers.
Native geometry, appearance and interaction acceptance remain open.

### M167 — Unsubmitted custom-field option can be lost on exit or Save

P1 source/mounted confirmed at0065b1f7. The New enum option input was excluded
from editor snapshots and validation. Typing only an option then leaving skipped
the discard prompt; Save could silently omit it after another edit.

Candidate fix includes pending option text in dirty detection, keeps it through
Keep Editing, and blocks enum Save with inline add-or-clear guidance. Adding the
option clears the pending input and saves the complete options list. The mounted
regression failed before implementation; all76 customization tests, TypeScript
and structural checks pass on paul. Code critic found no blockers. Native input,
announcement and exit acceptance remain pending.

### M168 — Switching away from Enum leaves hidden options in the create payload

P2 source-confirmed in the field editor audit at0065b1f7. Changing field Type from
Enum to Text hides the option editor but retains enumOptions. The create command
passes those to ManageCustomFields, whose validation rejects options on non-enum
fields. The user sees an enabled Save followed by an avoidable validation failure.
Candidate fix preserves dormant option draft for switching back, but submits an
empty options list for non-enum creation. Mounted regression first reproduced the
hidden options in the outgoing payload; it now verifies switching-back retention
and the corrected payload. All77 customization cases, TypeScript and structural
checks pass on paul; critic found no blockers. Native acceptance remains pending.

### M169 — Return-details error heading is obscured after failed Save

P2 screenshot-confirmed on iPhone17 run35029854251 at
e8b3d42dccf3f13428fb26bbb1cfd85ea8b0dd9e. In
[the failed-save capture](phone-return-error-350298.png), the retained note and
buttons are visible, but the error heading sits partly under the navigation blur.
Initial presentation is readable. The XCTest asserts error existence and retry,
not the complete error's position, so its passing outcome does not close this.
Candidate fix reveals each mounted operation error after layout using the native
header inset; repeat layouts do not reset user scrolling. The native-owned note
is retained. The mounted regression failed before implementation and now verifies
one reveal and unchanged retry details. All31 Home/presentation tests, TypeScript
and structural checks pass on paul. XCTest now checks the complete error frame
below navigation, rather than existence alone. Native verification must repeat
the failed-save/retry sequence before this finding closes.
Review identified pinned React Native's scroll-offset clamping: iOS also needs
scrollToOverflowEnabled, as already used by Add. This was added after a failing
regression assertion. The final full remote suite passes1,736 tests/271 files,
TypeScript and structural checks. This does not replace the pending native rerun.

### M170 — Conversation response and decision commands bypass native adapters

P2 source-confirmed at6edc7e85. VoiceConversationExchange paints Previous/Next and
Retry photos as text Pressables; VoiceSessionSheetScreen paints Approve/Cancel
decision buttons. They issue commands rather than select values or navigate to
settings. The composer already uses native controls, so the surrounding command
styling is inconsistent. Candidate now uses NativeCommandButton for rail and photo
retry, and NativeSheetActions for approval/cancellation. Rail bounds/count are in a
small presentation component; position retention and reduced-motion scrolling stay
in the parent. Plan IDs and pending-decision presentation are unchanged. All46
focused rail/presentation/composer/lifecycle tests, TypeScript and structural checks
pass on paul; code critic found no blockers. Native review actions are taller than
the former custom row: short windows, keyboard and reading order need acceptance.
Native layout verification remains open; this is not a full conversation audit.

### M171 — Response navigation can outlive its originating conversation visit

P2 source-confirmed at722ca46c. VoiceSessionSheetScreen's response-link handler
awaits pauseMedia and then unconditionally dismisses/pushes asset details. If the
user leaves or changes scope while pausing, completion can navigate from another
screen. Capture visit and scope ownership before awaiting and check before
navigation. Verify current completion, leave/return, and scope replacement with
a delayed media fake. The candidate now binds handlers to a focused visit and
scope, retires that visit on cleanup, and checks again after media shutdown.
Four mounted scenarios cover current, departed, returned and replaced scope;
retained old callbacks cannot pause a new session, while fresh actions work.
Tests reproduced three late-navigation failures before correction and a retained
callback failure during review. All four now pass on paul with TypeScript and
structural checks. Code critic found no remaining blockers. Native transition
acceptance remains open.

### M172 — Pending native recorder startup can outlive capture cancellation

P2 source-confirmed at3f601d5d. RealtimeVoiceSessionController.pauseMedia only calls
recorder.cancel when recordingStarted is true. ExpoVoiceAudioRecorderCore.start
awaits permission, audio mode and preparation before calling record; it has no
startup cancellation check. Leaving during those waits therefore allows record()
before the controller notices its obsolete generation and cancels. This is a
brief unintended capture, not evidence that audio is sent or capture persists.
The existing test named permission readiness delays the provider readiness port,
before recorder.start, and does not cover native permission/preparation.

Make startup cancellation reach the native adapter before capture begins, retaining
fresh-start behavior and preventing an obsolete start from cancelling a new one.
Acceptance needs delayed permission/mode/preparation fakes, fresh restart and
normal stop/cancel, plus native permission dismissal/return. The candidate passes
an AbortSignal through the recorder port and checks cancellation before capture at
each native startup boundary. Prepared resources are stopped/deleted and audio
mode restored. Controller startup and cancellation cleanup share a queue; old
cleanup cannot disable a fresh recording. Nine startup cases cover the three native
boundaries plus initial/follow-up pause/cancel/disposal. Three additional delayed
active-cleanup cases reproduced a review finding and now pass. All98 focused
recorder/controller tests, TypeScript and structural checks pass on paul. Critic
found no remaining blockers in this change. Native permission/interruption
acceptance remains open; no physical microphone verification is claimed.

### M173 — Approval silently omits a name still being edited

P2 source-confirmed at4b988ab0. Inline name text lives in titleEditor, while Approve
only serialized commandDraftState. The candidate merges the visible normalized
name into approval edits, preserving placement and other commands, and stores it
locally before submission. Blank names block approval with visible guidance;
Cancel remains available. A mounted real-provider/controller test failed before
the correction and now verifies blank rejection, latest name and retention after
transport failure. Native keyboard/decision-area acceptance remains pending.

### M174 — Proposal destination selection uses an incomplete custom panel

P2 source-confirmed at4b988ab0. ParentPicker in VoiceSessionSheetScreen is a flex
sibling below the conversation and approval footer. It uses custom search/close
controls and chevrons on immediate value-selection rows, without indicating the
current destination. The screen passes candidates.data ?? [] and discards lookup
loading/error state, so failure and empty search are indistinguishable.

Use an existing native searchable selection pattern with current-value indication,
bounded loading/error/no-match feedback and retry. Hierarchy, existing/proposed
destinations and descriptions justify a selection view rather than a flat menu.
Preserve the proposal, selected destination and name draft across entry/Back;
scope/plan replacement must retire pending choices. Native layout failure is not
claimed from source alone. The candidate now opens voice-plan-location through
the native stack with native search and Back. It shows current value, checkmarks,
eligible earlier proposed parents, existing candidates, disabled reasons and
loading/error/retry/no-match states. Selection commits once before Back; scope,
plan, pending approval and command validity gate editing. The old inline panel
and its layout/styles are removed. Eight tests cover choice modeling, lookup
states, disabled existing choices, successful Back and obsolete callback rejection
after scope/plan mismatch or leaving. Back preserves the proposal/name draft.
The route adds R142 to the audit inventory; its full source-axis review is in
voice-location-axis.md. Native sheet-to-stack, keyboard and assistive-technology
acceptance remain pending.

### M175 — Inline proposal name commands remain custom narrow icons

P2 source pattern finding at4b988ab0. Save/Cancel in EditablePlanCommandFields use
custom Pressables with36-point widths and44-point heights, with no hitSlop.
Reuse native command controls and preserve explicit save/cancel semantics alongside
M173 final approval. Source geometry is not a measured native hit-region result.
Acceptance includes blank disabling, cancel restoring the committed name, Done,
keyboard visibility and native targets at normal text size. The candidate extracts
VoicePlanNameEditor with a full-width field and native Cancel/Save commands below.
Save and keyboard Done share normalized nonblank validation; Cancel only closes
the editor, preserving the committed draft. Four mounted callback cases and42
related proposal checks pass on paul, with TypeScript and structural checks.
Code critic found no blocker. The initial test failed because the extracted
component did not yet exist; it is not a reproduction of native target geometry.
Native keyboard/targets and parent integration acceptance remain pending.

### M176 — Partial photo failure drops its recovery detail

P2 source-confirmed atcd353ee2. The controller replaces safe upload failure reasons
with a count-only message when at least one attachment succeeds. The shared
progress presentation also appends reasons only for total failure and shows a
success checkmark for either terminal failure. Users can see incomplete counts
without the available explanation, in both the active proposal and history.

The candidate preserves the existing allowlisted reason for partial failure,
renders it alongside counts, and uses a warning symbol while keeping the inventory
change marked saved. Two controller regressions verify safe detail/redaction and
retrying only the failed photo; a mounted shared-progress test verifies visible
and accessible detail. All three failed before implementation. No server boundary
or authorization behavior changes. Native reason visibility, contrast and
announcements remain pending; see voice-progress-axis.md for all24 source axes.

### M177 — Initial conversation failure has no in-place retry

P2 source-confirmed at9e11dfe1. SessionErrorState renders Voice unavailable and
the failure message without a retry command. The underlying scoped query already
supports context and inventory-scope recovery, but the conversation does not
expose it. Users must leave the task or depend on background refetch.

The candidate exposes that existing scoped refetch through the interaction
provider and uses a scrollable failure view with native Retry conversation.
Pending retries reject duplicates; visit ownership rejects retained callbacks
after leaving or returning. Errors stay available if retry fails again. Two real
provider/recovery tests cover scope/context failure and recovery, and a component
test covers lock/visit ownership. Whole-workspace runtime, compact detents and
VoiceOver remain pending. See voice-processing-axis.md for24 source axes.

### M178 — Conversation header is custom and duplicate reset bypasses protection

P2 source-confirmed at828b87ba. Close and New conversation are custom40-point
Pressables despite an available native header adapter. A second Reset session
command calls onReset directly in terminal states, bypassing the confirmation
used by New conversation when photo retries remain. Source dimensions are not
native hit-region measurements; the bypass is visible in callback wiring.

The candidate enables the native sheet header with Close and a system compose
action. Inventory context remains body text. The duplicate inline reset is removed;
the sole reset entry retains the existing plan/photo confirmation policy. A
mounted header case verifies retryable-photo confirmation and independent Close.
An iOS adapter case verifies compose semantics; navigation-feedback coverage
verifies current committed handlers. Composed options are memoized after critic
review; the feedback harness passed before memoization, so no reproduced loop is
claimed. All existing header mappings remain unchanged; shared consumers inspected
include Home, Browse, Add, notifications, settings editors/collections, reminder
timing, checkout history, provider editors and inventory switcher.

All1,774 tests across280 files and static checks pass on paul before final options
memoization (`/tmp/mobile-m178-full.log`). Nine focused cases and static checks pass
after it (`/tmp/voice-header-reviewed.log`). The new native fixture covers header
reachability, declining reset, closing and reopening the proposal; its compilation
and execution are pending. Compact detents, keyboard layout and destination Back
must be rechecked on the changed native header before claiming runtime acceptance.

### M179 — Provider recovery uses an unnecessary custom button

P2 source pattern finding at63d17e79. The conversation's provider recovery command
is a styled Pressable with its own fill/font/shape despite the shared native command
adapter. The candidate uses NativeCommandButton with the same label and navigation
callback and removes unused styles. Existing failure-presentation, native-command
and navigation-contract checks verify retained semantics; native geometry remains
pending. See voice-route-axis.md for the consolidated24-axis route review.

### M180 — Shared provider states lose the current task identity

P3 source-confirmed at684c54d5. ProviderStateView shows an unlabeled full-screen
spinner and labels profile/credential/prompt failures Could not load Voice Setup.
The household-context bridge also omits loading text. This is a task clarity/copy
finding, not a claim that Apple requires a label on every spinner.

The candidate requires a taskLabel at every shared consumer, reuses the labeled
settings progress row, and identifies the task in error headings. Credential
settings text refers to metadata; secrets are not read back. Existing query,
permission and Retry behavior is unchanged. Six RED mounted cases reproduced
missing task text. All110 related checks/static validation pass, plus six final
copy checks; critic found no blocker. Native announcements/layout remain pending.
See voice-readiness-axis.md for all24 source axes and affected consumers.

### M181 — Failed provider tests reported as successful

P1 source-confirmed at d39e4233. The API legitimately fulfills a test request with
status `failed`, but the mobile command previously accepted every fulfilled result.
Both profile details and voice-stage settings then showed Connection tested.
Existing success fakes incorrectly used `success` instead of the API's `succeeded`.

The command now accepts only `succeeded`. Other statuses produce safe configuration
and credential guidance without exposing arbitrary returned text. Two mounted RED
cases reproduce the false success on both screens and verify successful retry;
four command cases reject failed, unknown, empty and invented success statuses.
All82 focused tests, TypeScript and structural checks pass remotely on paul
(`/tmp/provider-test-result-green.log`). Critic found no blockers. Authorization,
transport and cache invalidation are unchanged. Native feedback layout remains open.

### M182 — History detail exposes an opaque actor fallback

P3 source-confirmed at953bb1ff. History list uses Someone with access when email
is unavailable; detail instead shows the internal principal ID in its primary
actor line. The candidate makes those surfaces consistent, retaining trimmed email
when supplied. Three mounted RED cases reproduce absent, empty and whitespace
email; a contrasting real-email case preserves its existing presentation. Audit
identity, permissions and transport are unchanged. See history-detail-axis.md for
the complete source review and outstanding native acceptance.

### M183 — Diagnostics disappears when remote discovery is unavailable

P2 source-confirmed at997bfa50. Diagnostics' shared readiness wrapper hides local
URL/version behind household loading and errors, and can show an empty principal
ID before account discovery finishes. That removes useful troubleshooting context.
The candidate keeps local connection/application sections visible and gives account
and household identity independent named loading/error/retry states. Access failures
suppress cached identity; transport text is not shown. Two mounted RED cases
reproduce the hidden context;68 settings tests and static checks pass remotely.
See about-diagnostics-axis.md; native geometry and announcements remain open.

### M184 — Departed switcher callbacks can still change inventory

P2 source-confirmed atfc93c151. Selection completion was guarded, but a retained
selection callback could begin a new command after blur, including after return.
The candidate binds rendered selection callbacks to the active visit and retires
that visit on blur, session/command replacement and Close. The pending request lock
and existing cancellation behavior remain. A mounted RED case proves the unwanted
post-blur command; coverage also verifies fresh return actions and immediate Close.
See inventory-switcher-axis.md for all24 source axes and remaining native gates.

### M185 — Browse continuation uses a bespoke command control

P3 source-confirmed at2f861b25. Sparse-page continuation is a custom Pressable while
pagination Retry and other result commands already use NativeCommandButton. The
candidate uses that adapter with the same label/callback and footer position.
The existing mounted sparse-page journey still reaches the matching second page;
all13 related tests and static checks pass remotely. Native geometry remains open.
See browse-route-axis.md for the full route/list/search source review.

### M186 — Inbox retains private rows after access is denied

P1 source-confirmed at17f9e2b7. Inbox uses local row state outside the shared query
cache. Its error handler kept loaded titles/paging after authentication or permission
failure. The candidate clears rows, cursor, read markers and loaded state on typed
access loss while retaining safe error/Retry. Ordinary failures retain useful data;
failed retries after denial cannot restore it. Eight mounted HTTP-adapter RED cases
cover401/403 across refresh/open/read-state/mark-all. All31 related checks/static
validation pass remotely; critic found no blocker. Native presentation remains open.

### M187 — Inbox read-state accessory still uses a custom command

P3 source-confirmed at17f9e2b7. Per-row read/unread used a
Pressable plus Lucide envelope while surrounding inbox commands use platform
controls. No concrete native limitation was documented. The candidate adds a
SwiftUI borderless icon button and Compose IconButton, preserving independent
row-open behavior, spoken action, disabled state and read-state reversal. It
reserves 48 points for the accessory. Twenty focused tests plus TypeScript and
structural checks pass remotely. Native hit bounds, row fit and assistive grouping
still require verification; the preview Pressable is only the non-native fallback.
### M188 — Rejected enum option entry disappears

P2 source-confirmed at 9c2c7246. Add option cleared the draft even when normalization
produced no usable value or the option already existed. Two mounted RED cases
reproduced the loss. The candidate preserves rejected input, explains the reason
inline, and clears obsolete feedback on editing or successful addition. Existing
pending-option Save validation and immutable saved values remain intact. All 70
related checks plus TypeScript/structural validation pass remotely; critic found
no blocker. Native keyboard and feedback presentation remain pending. See
[custom field options](custom-field-options-axis.md) for the 24-axis review.
### M189 — Photo-removal failure alert crosses navigation visits

P2 source-confirmed at 53e3915c. A still-mounted Details route checked the asset
lifetime but could show a failed-removal alert after leaving or leaving/returning.
Both cases reproduced before the change. The candidate captures the existing task
presentation predicate and gates the error alert; pending cleanup and fresh retry
remain intact. Current and unmounted cases remain covered. Native alert timing and
navigation acceptance remain pending; see system-dialog-axis.md.
### M190 — Camera and microphone denial omit recovery guidance

P2 source-confirmed at 91c22aa3. Errors said access was required without explaining
how to enable it or use an alternative. Candidate copy describes device Settings
and library/typed alternatives. Explicit retry rechecks permission; 109 related
checks plus static validation pass. Direct Settings shortcuts are recommendations,
not implemented by this copy change. Physical recovery and message fit remain open.

### M191 — Notification Settings opening has no failure recovery

P2 source-confirmed at 91c22aa3. Both device-Settings actions discarded
Linking.openSettings promises. The candidate catches rejection, shows safe manual
guidance and permits retry without changing preferences. Launch identity and
foreground/focus generation reject late errors and keep an old completion from
unlocking a newer attempt. Ten focused checks plus TypeScript/structural validation
pass remotely; critic found no blocker. The RED run reproduced unhandled rejection;
physical Settings launch/return and visual feedback remain pending. See
permission-recovery-axis.md.
### M192 — Proposal location chooser has no explicit Back control

P2 runtime-observed on iPhone 17, run35042066124/source d39e4233. After selecting
and reopening a proposal location, the sheet navigation bar has title/Search but
no Back. Screenshot and hierarchy are retained in native-phone-350420.md. The
candidate installs an explicit native left Back using the existing stable header
adapter; Back preserves proposal/title drafts and direct entry falls back to Voice.
Ten related checks plus static validation pass; critic found no production blocker.
The unavailable-route mounted case also has a body Back button, so its label check
does not uniquely certify the header. The valid-route test does, and native
hittability/return assertions remain required on the next build. No Swift/runtime
pass is inferred from Linux validation.

### M193 — Sharing keyboard obscures cancellation recovery

P2 runtime-observed on phone run35042066124. The keyboard remains over the
invitation cancellation menu after submission recovery. Creation now explicitly
ends keyboard editing while retaining failed email drafts. The regression test
failed with zero dismissals before implementation and passes afterward. This is
a candidate correction; actual keyboard absence and cancellation require retest.

### M194 — Native ellipsis menu has an undersized hit region

P2 runtime-observed in the same Sharing capture: native button bounds are
21.7 by 6.7 points despite a 44-point React Native host. The native image label
now owns a 44-point frame and rectangular content shape. Sharing and Asset
overflow are affected; label and sort-icon variants remain unchanged. Native
bounds and menu interaction assertions must pass before closing this finding.

### M195 — Photo acquisition errors outlive their Details visit

P2 source/mounted confirmed at fa504041. Picker and upload exceptions could
publish a global failure notice while Details remained mounted behind another
route or after returning to a new visit. Both rejection cases failed first.
The candidate gates only the exception notice with the existing focused-visit
predicate; results, failed-photo state, reconciliation and pending cleanup are
preserved. Current failure plus fresh retry remain covered. All112 related
Details checks, TypeScript and structural validation pass remotely. Critic found
no blocker. Native picker navigation focus and notice placement remain pending;
the guard is not an AppState/background policy.

### M196 — Move destination creation uses a custom selection-like row

P2 source-confirmed at bba8c1e3. Creating a destination is a command, but its
custom Pressable looked like the surrounding parent rows and combined action/help
text. It now uses NativeCommandButton with a complete action label and separate
help explaining automatic selection. Existing creation/move locking remains.
Three named-command cases failed before implementation; all82 related checks,
TypeScript and structural validation pass remotely. Critic found no blocker.
Native long-label wrapping, disabled appearance and keyboard reachability remain
pending. This does not migrate parent selection rows or change creation semantics.

### M197 — Add Clear draft remains a custom command

P2 source-confirmed at eac9c3ac. A custom muted text Pressable bypassed the shared
native command adapter for explicit draft removal. The candidate uses the native
destructive command inside More details, retaining the pending guard and scoped
reset. The named-command test failed first; recovery checks now clear a stored
title and unfinished tag/color before creating a new draft. Native appearance,
target bounds and focus remain pending. No confirmation or Close semantics changed.

### M198 — Retained History confirmation can submit an obsolete operation

P2 source/mounted confirmed at f80a4b7e. After opening Revert, a failed refresh or
replacement operation on the same activity did not invalidate the dialog callback.
Both regression cases submitted the old operation before the fix. Confirmation
now captures activity-snapshot ownership through useTaskPresentation; a fresh
record needs a new confirmation. Started reversals retain their existing completion
behavior. Fresh recovery and old-callback retirement pass. Critic identified a
nullable entry type mismatch, corrected with entry ?? undefined. Native dialog
timing, focus and reachability remain pending.

### M199 — Reminder settings retain editable data after access loss

P1 source/mounted confirmed at6729c57f. NotificationSettingsScreen retained its
preferences and editor after authentication/permission denial during refresh or
save, including denial of the supporting type collection. The candidate clears
displayed preferences/type data and locks save entry until an authorized reload.
Ordinary transient failures preserve drafts; failed retries after denial cannot
restore revoked data. Five adversarial recovery cases exercise the real client
and notification adapter against controlled HTTP responses, plus typed collection
denial. All36 related checks, TypeScript and structural validation pass remotely.
Code critic found no blocker; the supporting-query case is explicitly a typed
query failure, not an HTTP-boundary test.
Native content removal/focus and backend authorization enforcement are separate
from this UI recovery evidence.

### M200 — Unresolved Add placement claims top-level placement

P2 source/mounted confirmed at716f23b1. Typing a placement clears the selected ID,
but the disclosure subtitle still said Top level in this inventory. It now says
Not selected yet until selection or clearing. Existing exact-match resolution on
Save is unchanged. The search/retry case failed first on the missing unresolved
status and now verifies both that status and recovery to intentional top-level
placement when cleared. Native wrapping and screen-reader reading remain open.

### M201 — Staged Add tag removal has no explicit action name

P2 source/mounted confirmed at327391bf. A staged new-tag chip's inferred accessible
name was only its tag text, concealing its remove action. It now names Remove new
tag {name} and exposes disabled state. The draft workflow verifies that invoking
the command removes only that staged definition, preserving another staged tag
and unfinished name before saving. This does not delete persisted inventory tags.
Actual screen-reader output and visual target acceptance remain native work.

### M202 — Browser sign-in errors are reported as user cancellation

P2 source/boundary confirmed atc7b64036. All non-success native auth results threw
Sign-in was cancelled, including error, locked and unknown results. Only explicit
cancel/dismiss now get cancellation feedback; other outcomes get safe retry
guidance without provider parameters. Three composed onboarding/OIDC cases failed
first and now verify rejection without session/tenant discovery and a valid fresh
retry. Cancel/dismiss checks remain. Live OS browser-return acceptance is pending.

### M203 — Restore interrupts a routine reversible command with confirmation

P2 design/source confirmed at46f23136. Restore from archived Details asked users
to confirm returning the item to active work. It now dispatches directly from the
native menu, retaining synchronous locking, current-visit feedback and retry.
Archive/Delete confirmations remain. Three RED cases reproduced the required
extra alert; all115 related tests pass with direct Restore and departed outcomes.
Critic found no blocker. Native menu activation/status/focus remains pending.

### M204 — Unsaved field expansion cannot be reversed in place

P2 source confirmed at1695cab3. Expand to all assets removed the targeted choice
from an edited field before Save, forcing users to discard other edits to recover.
The shared native applicability picker now permits a draft roundtrip when the
persisted field targets selected types. Saved all-assets fields remain static;
immutable targets and unsaved additions survive the roundtrip. Two RED cases
preceded implementation;68 related tests and static checks pass remotely. Code
critic found no confirmed blocker.
Native acceptance remains pending. See field-applicability-axis.md.
