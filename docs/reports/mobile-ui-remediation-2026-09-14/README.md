# Comprehensive mobile UI audit and remediation

Active goal, started 2026-09-14 at source revision 12b5cb5a.

Earlier full source checkpoint:3d20d0ff passes1,869 tests across288 files, TypeScript
and mobile structural checks on paul. Tracked mobile source/config/native fixture
content matches the remote tree by checksum. React act warnings remain. Log:
`/tmp/mobile-audit-3d20-full.log`. This includes the native Add/tag fields and M216
conversation layout candidate. This does not establish native acceptance.

Latest Android filter candidate:1,876 tests across291 files passed on paul before
the final keyboard wrapper; focused tests, TypeScript and structural checks passed
after it. See `android-filter-sheet.md` for native crash rejection and replacement
evidence; acceptance remains incomplete.

Inventory: **142 route/layout and nested-task surfaces × 24 axes = 3408 review cells**
(retained customization completion added September 15; initial inventory: 132 surfaces; four asset editing/moving and three Home return subtasks added during source inspection).
This is a review worklist, not a count of completed checks. Overlapping shared tasks
are intentional: route coverage and interaction coverage are independent.

- `invitation-acceptance-axis.md`: invitation review across24 axes, command migration and native gaps.
- `add-creation-media-axis.md`: quick place creation and photo selection across24 axes; native removal correction and remaining device checks.
- `asset-edit-route-axis.md`: Edit route review, loading exits and outstanding action-eligibility finding.
- `native-phone-350504.md`: terminal phone fixture results, notification hit probes and unresolved failures.
- `native-onboarding-350592.md`: passing main phone/iPad journey and a distinct iPad entry-observation failure.
- `native-fixtures-350592.md`: phone56/75 and iPad64/75, confirmed phone bottom search, and paired production Place diagnostic.
- `native-fixtures-350633.md`: phone57/79 and iPad67/79; passing action probes/invitation acceptance on both, phone Home tab return, and unresolved iPad tab-container selection. Inspected Account notice contradicts its phone geometry timeout.
- `native-fixtures-350695.md`: phone59/81 and iPad71/81; Add tag journey, Home header and color touch probes pass on both. Phone captures retain bottom Place search and keyboard-obscured Sharing cancellation, while those journeys pass on iPad. Conversation overlap remains on both.
- `expiration-entry-axis.md`: exact-date and month/year entry, draft validity and native gaps.
- `confirmation-scope.md` and `confirmation-call-sites.csv`: native-dialog caller inventory and review boundaries.
- `surfaces.json`: route and nested task enumeration.
- `axes.json`: named review dimensions.
- `matrix.csv`: per-cell state/evidence/finding tracking; initially pending.
- `tag-color-axis.md`: all24 color-selection axes, shared consumers and phone/iPad activation evidence.
- `field-applicability-axis.md`: all24 applicability axes and reversible unsaved expansion (M204).
- `push-permission-axis.md`: all24 device-setup axes, command naming and physical verification limits.
- `onboarding-creation-axis.md`: all24 household/inventory/recovery axes and native command migration.
- `native-search-placement-350465.md`: captured phone bottom-search mismatch and pending static placement comparison.
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
An isolated Android16 emulator now runs the synthetic app on `paul`.
[Initial Home icon findings and verification](android-native-icons.md) record the
first native Android checks; broader Android acceptance remains pending. See also
[runtime preparation](android-runtime-preparation.md). No local builds/tests,
per session constraint.

The [350420 iPad onboarding inspection](native-onboarding-350420.md) records a
pre-typing readiness timeout despite a visible keyboard in the final capture.
The follow-up [350465 onboarding run](native-onboarding-350465.md) passes its
phone case (two device-inapplicable skips) and all three iPad cases. The 350420
fixture jobs are complete; see `native-phone-350420.md` for failures and limits.
The [350465 iPad fixture job](native-ipad-350465.md) is now terminal:58/71 cases
pass, with13 failures. The [phone fixture result](native-phone-350465.md) is49/71
with22 failures. Both targets pass paced typing diagnostics while original
controlled-input comparisons still fail. Paced typing is diagnostic evidence only
and does not replace normal acceptance. Newer run35050407693 is active.

A source-reviewed cell never implies a runtime pass. Add discovered internal
surfaces during inspection. Record justified N/A per cell; do not default missing
coverage to pass. This effort includes fixing findings and TestFlight release.

## Current checkpoint — September 15

The requested PR150 checkpoint is released as **TestFlight0.24.23 (112.1)**.
[Run35028077706](https://github.com/elsell/stuffstash/actions/runs/35028077706)
uploaded main438bd902 successfully; Apple processing and exact-build changelog
verification completed September15 at22:20:03 UTC. The changelog verifier waits
for Apple's VALID processing state and reads back the published notes. This does
not establish physical-device UI acceptance or external beta review approval.

The release's source `a2ed8342` passed all **1,694
mobile tests across270 files**, TypeScript and structural checks remotely on paul
(`/tmp/mobile-batch-a2ed8342.log`), with a clean source checksum comparison. This
includes item-type failure/search recovery, failed-photo retry, unsupported-format
feedback, retired session callbacks and the earlier interaction fixes.
Native acceptance remains incomplete: full run35012949816 passed43/58 phone
fixtures and47/58 iPad fixtures; both onboarding jobs passed. See
[latest native follow-up](native-350129-followup.md) for actual revisions and limits.
The resumed audit continues in draft PR153. Checkpoint81e91f74 passes all1,710 mobile
tests across270 files, TypeScript and structural checks remotely on paul
(`/tmp/mobile-batch-81e91f74.log`). Its new native collection journey is pending.
The subsequent [Home dashboard review](home-dashboard-axis.md) adds M160,
retiring delayed pull-refresh notices after navigation; focused validation is
recorded separately and does not change that full-suite checkpoint.

[Inventory/location asset lists](asset-lists-axis.md) now have all24 source axes
reviewed. M161 provides feedback for explicit refresh failures while retaining
ordinary cached cards; obsolete-visit failures stay silent. All18 focused cases,
TypeScript and structural checks pass remotely. Code critic found no blockers.
Native acceptance is still pending; the retained LocationsScreen is not counted
as a current shipped route.

Combined post-M161 validation passes **1,722 tests across271 files**, TypeScript
and mobile structural checks on paul (captured `/tmp/mobile-list-batch.log`).
This includes M160 and M161; no native result is promoted by the source suite.

[Asset History](history-list-axis.md) now has all24 source axes reviewed. M162
retires delayed pull notices and separates inline Retry from the native pull
indicator. Its12 focused tests, TypeScript and structural checks pass remotely;
code critic found no blockers. Native acceptance remains pending. This follows
the1,722-test checkpoint above.

The audit ledger retains142 surface IDs ×24 axes, including two absent Add
controls documented as inventory corrections. Its3,408 cells comprise zero pending
source-review cells,2,630 source-reviewed,555 finding,25 runtime-partial and198
not-applicable. This completes source inventory coverage, not native acceptance or
finding remediation. The remaining route/server-entry/notice reviews are linked
in their surface reports; M212 adds direct-entry checkout-history exit recovery.

At ebdd3090, a checksum comparison confirmed the tracked mobile source, fixture
and test configuration matched paul's validation tree. All1,854 tests across286
files, TypeScript and mobile structural checks pass (`/tmp/audit-batch-ebdd-full.log`).
This includes gallery preview and direct-entry checkout-history fixes; it does
not establish native rendering. Text-entry diagnostic35056372549 runs separately
atfdbf30bf while the full350549 fixtures continue. Release remains gated by the
recorded normal-size runtime failures, not by incomplete source enumeration.

Current expiration-entry, gallery and full-viewer follow-ups account for38 more
source cells. M211 adds gallery preview failure recovery while retaining original
photo opening;8 remote tests and static checks pass. Native visual acceptance
remains pending. Run350504's iPad job finished58/72 with14 failures; see
[iPad results](native-ipad-350504.md). Passing source checks do not close these
native failures.

Home Return's optional details and pending/recovery now have complete source
follow-ups. Run350465's phone/iPad captures verify the full failed-save error is
below the native header with the note and commands retained; named overlap M169
is corrected for these normal-size light states. Wider native acceptance remains
open. See [Home Return evidence](home-return-axis.md).

[Retained customization completion](retained-completion-axis.md) and Home's
checked-out Return entry now have source follow-ups across24 axes. Their combined
89-case validation passes remotely; native interaction acceptance remains open.

Appearance selection, History reversal and asset overflow now have complete
source follow-ups in their surface reports. M198 rejects retained confirmations
after an activity changes or refresh fails;34 combined History/appearance checks
and static validation pass. Native confirmation timing remains pending.

[Add and draft recovery](add-draft-axis.md) now cover three surfaces across24
axes. M197 migrates Clear draft to a native destructive command;23 Add checks
and static validation pass. Native text entry and recovery remain open.

The [Move route review](move-route-axis.md) covers both routes, selection and
creation across24 axes. M196 makes creation a native command;82 related checks
and static validation pass. The preceding bba8c1e3 full suite passed1,815 tests
across284 files remotely; native acceptance remains incomplete.

Both [Details routes](asset-detail-route-axis.md) now have a shared24-axis source
review and112 passing related checks. M195 fixes late picker/upload exception
notices crossing navigation visits; native focus and placement remain pending.

[Invitation link intake](invitation-link-axis.md) now covers S131 across all24
axes, with64 remote checks. Physical cold/warm handoff remains pending. Sharing's
native keyboard/menu findings and Home's separate tap-delivery diagnostic are
recorded in their surface reports; source fixes are not runtime acceptance.

[Permission recovery](permission-recovery-axis.md) now has all 24 source axes
reviewed. M190 adds camera/microphone recovery guidance; 109 related checks/static
validation pass. M191 adds notification Settings launch-failure recovery with
10 related checks and static validation passing remotely.
Physical permission, Settings return and native feedback verification remain open.

[Push entry](push-entry-axis.md) now has all 24 source axes reviewed. Eighteen
adapter/application checks pass; mounted root navigation and physical cold/warm
notification journeys remain explicit acceptance gaps. This adds source coverage,
not new runtime verification or a production change.

[System dialogs](system-dialog-axis.md) now has all 24 source axes reviewed, with
caller-specific ownership evidence retained. M189 prevents late photo-removal
alerts crossing navigation visits; 101 related checks/static validation pass.
Native alert activation, interruption and focus return remain open.

[Phone run350420 follow-up](native-phone-350420.md): the completed phone job
passed46/69 cases. Text/target/footer failures remain; the voice location journey
now reaches its final Back command. Artifact inspection confirms the sheet has
no Back; M192 adds an explicit native header action, with retest pending.
The iPad fixture job remains active at this checkpoint. This older
source revision does not verify current-head candidates.

Combined checkpoint at a2ad3163: all 1,808 mobile tests across 283 files,
TypeScript and mobile structural checks pass on paul
(`/tmp/mobile-audit-a2ad3163-full.log`). A checksum dry-run confirms tracked mobile
source, native fixtures and listed package/test configuration match the validation
tree; timestamp/permission differences were excluded. The suite emits React act
warnings. This is source/mounted validation, not current-build native acceptance.

[Custom field type/options](custom-field-options-axis.md) now has all 24 source
axes reviewed for both nested controls. M188 preserves rejected option drafts;
70 related checks and static validation pass remotely. Its new editing finding
supersedes a runtime-partial classification while preserving the earlier evidence.
Normal-size native typing and feedback acceptance remain open.

[Proposal editing](voice-plan-edit-axis.md) now covers all24 source axes. M173
includes the visible pending name on approval and blocks blank names;52 focused
tests and static checks pass. Critic found no implementation blocker, while noting
that the new provider test does not render the actual proposal/failed-review UI.
M174 tracks the custom destination panel and missing lookup recovery; M175 tracks
the narrow custom inline name commands. M175 now uses native commands below a
full-width field;46 focused tests and static checks pass, with critic review.
M174 now replaces the panel with a native stack selection route, search, checked
choices and lookup recovery. The new R142 route now has a full
[24-axis source review](voice-location-axis.md). Its runner-only native journey
uses the real proposal and destination screens to exercise retry, search, selection
and Back. Fixture isolation,20 focused cases and static checks pass remotely;
Swift compilation and native presentation acceptance remain open.

Combined M174 checkpoint: all1,766 tests across278 files, TypeScript and mobile
structural checks pass on paul. Code critic reviewed the route and added selection/
Back/disabled-candidate cases. This remains source/mounted evidence.

[Progress and photo recovery](voice-progress-axis.md) now has all24 source axes
reviewed. M176 preserves safe partial-upload failure reasons in both active/history
progress and replaces the terminal failure checkmark with a warning. Three new
regressions reproduced the lost reason and verify safe text plus retry ownership.
Combined validation passes all1,769 tests across278 files, TypeScript and structural
checks on paul (`/tmp/mobile-m176-full.log`). Critic found no blockers; native
visibility, contrast and announcements remain pending.

[Conversation loading/processing](voice-processing-axis.md) now has all24 source
axes reviewed. M177 adds native in-place Retry for initial context failure, with
duplicate suppression and focused-visit ownership. All45 focused cases, TypeScript
and structural checks pass remotely (`/tmp/voice-preview-green.log`). Critic found
no blocker. Query/provider/component evidence does not establish whole-workspace
native recovery; compact-sheet geometry and announcements remain pending.

M178 replaces custom conversation header buttons with native Close/New conversation
and removes the duplicate Reset path that bypassed retryable-photo confirmation.
Shared header consumers retain their mappings. Combined validation passes1,774
tests across280 files and static checks; nine focused cases/static checks pass
after reviewer-requested options stabilization. The native proposal fixture adds
Close/reopen and declined-reset journeys; execution remains pending.

The [conversation route review](voice-route-axis.md) consolidates all24 source
axes across its nested tasks without promoting native gaps to passes. M179 replaces
the remaining custom provider-recovery button with the shared native command.
All49 relevant presentation/adapter/navigation tests, TypeScript and structural
checks pass remotely (`/tmp/voice-recovery-command.log`). Native acceptance remains
open, including header/keyboard geometry and physical media interruptions.

[Voice setup/readiness](voice-readiness-axis.md) now covers all24 source axes for
the overview, capability route and nested readiness surface. M180 identifies each
provider task in shared loading/error views rather than labeling unrelated editor
errors Voice Setup. All110 relevant tests and static checks pass remotely, plus
six final copy checks. This is a project task-clarity choice, not a blanket Apple
requirement to label spinners. Critic found no blockers; native acceptance remains
open.

M181 fixes false success feedback when a completed provider connection test returns
`failed`. Profile and stage settings now accept only `succeeded`; 82 focused checks
and static validation pass remotely. Native feedback presentation remains pending.

[Native run35038625270](native-350386-followup.md) finished with43/67 phone and
51/67 iPad fixture cases passing; both onboarding jobs passed. The Home Return
visibility check hit duplicate nested text matches; its selector is corrected
without relaxing geometry acceptance. Current-head native verification remains open.

[History detail](history-detail-axis.md) now has a complete24-axis source review.
M182 aligns its missing-actor label with History list. All14 focused cases and
static checks pass remotely; code review found no blockers. Native acceptance
remains open.

[About and Diagnostics](about-diagnostics-axis.md) now have complete source-axis
reviews. M183 keeps local diagnostics available during remote identity failures,
with independent named loading and native Retry controls. Native layout remains
unverified.

[Native run35034075257](native-350340-followup.md) is terminal: onboarding passed
on both devices; fixture suites passed45/64 on phone and52/64 on iPad. Normal-size
text-entry, search and target findings remain open. This predates current fixes.

[Recording](voice-recording-axis.md) now has all24 source axes reviewed. M172
identifies capture starting after cancellation during native permission/preparation;
existing readiness coverage did not test that boundary. The candidate adds native
startup cancellation and serialized cancellation cleanup, with12 regressions and
98 focused recorder/controller cases passing remotely. Review is complete; physical
permission/interruption acceptance remains pending.

Combined M172 checkpoint: all1,753 mobile tests across273 files, TypeScript and
mobile structural checks pass on paul after the final review correction. This
does not establish native microphone, permission or interruption behavior.
These are evidence states, not a compliance score. Finding cells can include
implemented corrections whose native acceptance remains open. Absent controls
are not native passes or claims of product feature parity.

M163 also retires item-detail pull-error notices after navigation, using the
existing visit/resource guard in the shared detail screen. Both asset and
location-context detail routes consume it. All92 focused detail cases, TypeScript
and structural checks pass on paul; full source-axis and native acceptance work
for those routes remains open.

[Household/inventory settings](scoped-settings-axis.md) now have all24 source
axes reviewed. M164 reuses labeled settings progress for both initial loads.
All59 mounted settings cases, TypeScript and structural checks pass on paul;
native announcement/layout acceptance remains pending.

[Type/tag editors](type-tag-editors-axis.md) now have all24 source axes reviewed
across six create/edit routes. M165 moves shared Save to the native primary
command; all54 customization cases, TypeScript and structural checks pass on paul.
M166 migrates Back/lifecycle/inherited Manage to native adapters. The full remote
suite passes (1,734 tests/271 files), with TypeScript and structural checks.
Native acceptance and M51 color behavior remain open. Field editors receive the
shared controls but are not certified by this six-route review.

[Run350298 follow-up](native-350298-followup.md) records both onboarding passes
and iPad49/61, phone42/61 fixtures. Current editor changes are
newer than this build. No screenshot acceptance is claimed from terminal logs.

Production customization editor journeys now cover native Back (Keep Editing and
Discard), Save, and Archive cancellation/completion in the isolated runner fixture.
Fixture installation tests, TypeScript and structural checks pass on paul. Native
execution is pending; these additions do not close editor acceptance findings.

[Custom-field editors](field-editors-axis.md) now have all24 source axes reviewed
across four routes. M167 protects unsubmitted option text; M168 prevents hidden
enum options from breaking non-enum creation after a type change. All77 focused
tests and static checks pass remotely. Field-specific native acceptance is open.

[Home return details](home-return-axis.md) has all24 source axes reviewed. Its
cancel/undo and save-error/retry journeys passed on both devices in run35029854251;
this is partial native interaction evidence, not full visual or persistence acceptance.
Screenshot review found a partly obscured failed-save heading (M169), still open.
Its reveal-on-error candidate passes31 focused tests and static checks; a stronger
native geometry assertion is added, with native rerun pending.
The combined M169 checkpoint passes1,736 tests across271 files, TypeScript and
mobile structural checks on paul, including the review correction for iOS offset
clamping. This remains source validation; native acceptance is not claimed.

The [typed composer](conversation-composer-axis.md) has all24 source axes reviewed;
its native controls and pending request behavior are distinct from recording and
plan approval acceptance. M170 tracks remaining custom response/decision commands.
M170's native command migration passes46 focused tests and static checks. Review
keeps the taller native decision area flagged for keyboard/short-window acceptance.

The [response surface](conversation-response-axis.md) has all24 source axes
reviewed;29 focused tests pass. Historical photo retry is plan-owned. M171 records
an unguarded navigation completion after pausing media. Its candidate now guards
pending completion and retained handlers across visit/scope replacement, with four
mounted scenarios passing and code critic review complete; native acceptance remains.
The combined checkpoint passed1,741 tests across273 files and static checks on paul;
the subsequent retained-handler correction passed its four focused cases and static
checks separately. Neither result is native transition verification.
The [sheet comparison](phone-sheet-comparison-350298.md) distinguishes a blank
nested diagnostic layout from the direct-scroll structure used by production filters.

Post-PR150 review adds [settings exit ownership](settings-exit-axis.md) and M155.
The [latest completed native follow-up](native-350129-followup.md) records full
run35012949816 failures and focused color results. Text-entry isolation is running
at0f378691 in run35029455242; iPad completed4/5, with reproduced character loss,
and phone completed3/5: uncontrolled ordinary text lost characters and controlled
entry failed keyboard readiness. See [text evidence](native-text-350294.md).
Expanded manual diagnostic35031744887 at81e91f74 is now running. No native fix
is claimed from its dispatch; newer PR pushes may supersede pending full runs.

Combined post-PR150 checkpoint0c26c3bf: all1,703 mobile tests across270 files,
TypeScript and the mobile structural check pass on paul
(`/tmp/mobile-batch-0c26c3bf.log`). This includes M153–M155; native acceptance
remains incomplete. Tests/builds were not run on the local host.

The five definition/tag collection routes now have a full source-axis review in
[customization-collections-axis.md](customization-collections-axis.md). M156 moves
search/Add to existing native header adapters; M157 updates Add when edit access
is revoked. Native acceptance remains pending. The code critic's asynchronous
filter assertion finding was corrected; no production blocker was identified.

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

[Keyboard accessory](keyboard-accessory-axis.md) reviews the shared dismissal control across all24 axes without claiming coverage of every input consumer.

[Home expiration and recent summaries](home-summary-axis.md) have source review across all24 axes, with destination and native verification limits retained.

Combined checkpoint75923e21: all1,627 tests across267 files, TypeScript and structural checks passed on paul (`/tmp/mobile-batch-75923e21.log`). A checksum comparison of mobile source was clean before this run. The count decreased because three grouped mounted Browse scenarios replaced six tree tests, while one invitation recovery test was added. This checkpoint includes M129–M133; native acceptance remains pending.


Combined checkpoint `dcbf1fe0`: **1,647 tests across 268 files**, TypeScript and
mobile structural checks passed on paul. Mobile source checksum comparison was
clean before execution; log `/tmp/mobile-batch-dcbf1fe0.log`. This includes the
M134–M139 follow-ups and preserves all native acceptance gaps. M137 keyboard
hit-testing remains unresolved. PR150's review body reflects the current batch.
The active native run34998354801 has passed iPhone onboarding; other jobs were
still running at this checkpoint and its source predates the newest corrections.


Matrix reconciliation at e1b3b7d6 maps M137–M140 onto22 affected review cells,
including Edit and Add consumers of the expiration editor, the stored-photo viewer,
and Account/Connection settings. These are findings with recorded source fixes or
unresolved native behavior, not pass promotions. Existing evidence is retained.
Coverage is still141 surfaces ×24 axes:984 source-reviewed,2086 pending,268 finding,
18 runtime-partial and28 not-applicable. Unique surface/axis pairs match the
inventory exactly. The many pending cells remain work, not implied compliance.

`account-connection-axis.md` reviews R024/R026 across all24 axes. M141 corrects initial identity recovery copy; M142 records custom command rows awaiting native-adapter correction. Source review does not close runtime cells.


Combined checkpoint51fedfc8 passes1,677 tests across270 files, TypeScript and
structural checks on paul (`/tmp/mobile-batch-51fedfc8.log`). Mobile source checksum
comparison was clean before testing. This includes M143–M146 and the non-crashing
accessory-comparison candidate; native acceptance of those latest changes remains
pending. [Confirmation review](confirmation-review.md) distinguishes reviewed
callers from still-pending dialogs. Run350037 onboarding passes on both devices;
phone/iPad keyboard and iPad landscape screenshots have been inspected and retained
in [native evidence](native-evidence.md). Fixture jobs remain running.

[Provider management](provider-management-axis.md) covers list, creation and detail
(R052/R054/R055) across all24 axes at5501152d. M146/M148 retain their finding
status; native command geometry and interaction acceptance remain open.

Combined checkpoint `5501152d`: **1,687 tests across 270 files**, TypeScript and
mobile structural checks passed on paul. Source checksum comparison was clean
before execution; log `/tmp/mobile-batch-5501152d.log`. This includes M143–M148.
No native acceptance status is changed by this validation.

Provider review reconciliation preserves all3,384 unique inventory/axis pairs:
1,022 source-reviewed,2,011 pending,289 finding,22 runtime-partial and40
not-applicable. Critic reviewed the new report and the72 affected matrix rows;
its stale creation-label correction is included.

[Home action header](home-header-axis.md) reviews S066 across all24 axes and
identifies why the existing Home Return fixture cannot verify three-action
visibility or production transparent scroll edges. Those native checks remain open.

[Add item type](add-type-axis.md) covers S088 across24 axes, records M149/M150
recovery fixes and corrects S085/S091 inventory entries for controls absent from
Add. Stable surface IDs remain; absent controls are not native passes.

Upload recovery now has a [24-axis source review](upload-retry-axis.md). A mounted
partial-failure journey uses the real command and confirms failed-only retry
without reopening the picker or duplicating the successful upload. All40 related
cases, TypeScript and structural validation pass remotely; native acceptance
remains open. This is additional focused evidence, not a new full-suite result.

The [root presentation review](root-presentation-axis.md) covers all24 R006 axes
and identifies M152, obsolete connection expiry callbacks acting on a new session.
The ownership correction has mounted evidence; native account and root layering
acceptance remain open.

Contained-item rows now have a current24-axis source follow-up in
[contained-items-axis.md](contained-items-axis.md);15 remaining source cells
were reviewed. Phone/iPad search disagreement remains an open native finding.

Android native controls now share the app appearance override. The normal-size
Cancel contrast regression and rebuilt dark/light evidence are recorded in
[Android Compose appearance](android-compose-appearance.md); other adapter families
retain individual runtime verification gaps.

[Android Add header](android-add-header.md) records M226's missing native commands,
shared presentation fix, and native Save/failure/Close acceptance. The separate
M227 no-history Close fallback and remaining header-dependent Android sheets remain
open; this sample does not certify all Add workflows.
