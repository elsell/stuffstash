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
