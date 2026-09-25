# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery

Latest verified TestFlight is **0.24.41 (136.1)**. PR196 mergedc2248215;
release36098436371, iOS upload107957316184 and Apple changelog readback passed.
Native commands now contain multiline labels while retaining ordinary density.
[Release evidence](evidence/release-136.txt).
Prior132 corrected inventory collection footer clearance above tabs/voice.
[Prior release](evidence/release-132.txt).
Prior130 delivered History summaries, Before/After values, mode selection and
footer clearance. [Prior release](evidence/release-130.txt).
Prior129 delivered Settings labels/grouping/insets and Home inventory name spacing.
[Prior release](evidence/release-129.txt).
Prior128 delivered Home collection shortcuts through Browse and scoped Android
photo dialog contrast. [Prior release](evidence/release-128.txt).
Prior127 delivered Browse retention/alignment and clearer container detail grouping
with tab/voice clearance. [Prior release](evidence/release-127.txt).

## Shipped batch acceptance

PR182 included PR181: Browse search retention/appearance, neighboring card
alignment, detail header insets and grouped container actions. Standard CI36000041858,
focused source tests, TypeScript, structural checks and critic review pass.

Native acceptance: Browse/filter workflows35979200078 pass6/6 both. Frozen
35987800972 passes phone15/16 and iPad14/16: its32pt Move-items target is corrected
by PR182 and passes in hierarchy35994194799 (six existing cases pass both).
The initial footer test had missing fixture metadata, not demonstrated clipping.
Corrected populated Detail/Sharing35999231112 at28bc9b3c passes2/2 both; iPhone
needed one retry after Xcode failed before app launch. Reviewed screenshots show
final content above delivered tabs/voice chrome. Product code is unchanged since
that native revision. Sharing needs no speculative inset change on this evidence.
[Evidence](evidence/container-organization.txt).

Android appearance and connected Edit/Move checks pass on the named audit APK.
[Appearance](evidence/android-live-appearance-header.txt),
[alignment](evidence/checkout-row-alignment.txt),
[detail entry](evidence/detail-entry-insets.txt).

Normal-text selection refresh36100666455 at4f130a93 completes all four iPad
workflows, including Add destination creation/retry and draft retention. Phone
passes all three tag workflows; destination entry misses its five-second UI
observation deadline, then the final full-screen capture shows the correct Put in
screen. That phone replay remains failed; it is not a demonstrated product failure
or a reason for another unchanged run. Reviewed creation/error and tag-selection
captures retain clear native headers, grouped choices and unobscured footer actions.
[Scoped decision](evidence/selection-current-361006.txt).
Inventory-list inset correction is shipped in132.1. Invitation acceptance is a
root-stack route outside tabs; the same tab-overlap diagnosis does not apply.
Reviewed existing phone/iPad invitation entry/recovery captures show clear
hierarchy and safe-area clearance. These predate the shared button replacement;
they establish baseline composition, not current rendering acceptance.
[Scoped baseline](evidence/invitation-visual-baseline.txt).
Broad refresh35980094051 completed121 cases per device (phone14 failures, iPad6),
including old diagnostics and enlarged-text cases; it is not a release gate for
this verified batch. [Terminal decisions](evidence/native-full-359800.txt).

## Home collection batch acceptance

Home Recently changed → See all now targets existing Browse List with explicit
all-active/recent ordering and clears prior query/tags/kind/availability criteria.
This removes the everyday entry to the older reduced-function asset grid; its
legacy route remains available. Home navigation test failed before implementation;
Home28, route6 and mounted Browse13 tests pass. TypeScript/structural checks pass;
critic found no source blocker. Native connected acceptance remains required:
start with refined Browse, return Home, use See all, open an asset and return;
verify selected tab, criteria, scroll clearance and ordinary tab context retention.
Checked out now uses the same explicit reset contract, differing only in availability.
Home28 and shared fixture3 tests pass; fixture preparation11 tests pass. Focused
`home-collections` acceptance uses production tab paths and shared real screens,
with no competing root Search route or placeholder. Critic's root-header and
stale-state observation findings were corrected. Native36013639520 stopped in fixture setup; corrected36017420318 passes on both
devices at3f7485e3. Reviewed captures show the correct List/tab,25 recent items and6
checked-out items with only the intended filter. CI36017423895 passes the same
revision. PR184 merged8bdcc289; release36020469905 delivered128.1 with verified notes. [Evidence](evidence/home-collections.txt).

Separate follow-up: Home's inventory switcher now uses available header space up
to320pt instead of always capping at180pt, avoiding unnecessary truncation when
only Profile is visible. Viewer regression failed before the change; six width
checks, structural checks and critic review pass. Native36018428143 at452265ab passes five cases on each device; reviewed captures
confirm one-action/three-action spacing and stable actions on scroll. This is
delivered by PR186 in129.1. [Evidence](evidence/home-header-space.txt).

## Next connected review

Settings overview review confirmed iPhone root labels crowded out by long values,
Appearance flush with the group edge, and repetitive scope subtitles. Candidate
uses descriptive subtitles, shared inset rows and scope context once.83 focused
tests, TypeScript,12 preparation tests, structural checks and critic review pass.
Native36019282915 reached Diagnostics but stopped on duplicate selectable-text
AX nodes; the locator is corrected without relaxing geometry checks. Corrected36023217961 passes all three cases on each device. Reviewed final
captures confirm hierarchy, footer clearance and tab return; Account/Connection
recovery checks also pass. PR186 delivered these changes in129.1.
[Diagnosis and evidence](evidence/settings-overview.txt).

## Current workflow investigation

M281 is confirmed: enlarged iPad sheet labels grow while their backgrounds stay
54pt high. The candidate reuses the measured UIKit command adapter for full-width
footer actions.15 focused tests, TypeScript, structural checks,17 fixture checks
and critic review pass. Native acceptance across footer, Conversation and long
Tags consumers is required before release. The earlier fixture-only run is
superseded; it is not a reason for another unchanged investigation.
[One current diagnosis](evidence/sheet-footer-current.txt).

The UIKit command batch shipped in136.1 using combined same-product native
evidence:36093575172 passes phone4/5 and iPad5/5;36095743496 passes phone5/5 and
iPad4/5. Every scoped workflow has a full pass on each device with identical product
code. Full-screen review confirms ordinary density and bounded multiline labels.
The replay iPad Sharing assertion read a partial email immediately after typing;
its final capture shows the full address. Its earlier complete pass supports the
batch; the failed replay remains a fixture timing gap, not a retroactive pass.
Critic review found no concrete release blocker. CI36095747484 passes88b9bbfd.
The rejected SwiftUI candidate never shipped. [Diagnosis](evidence/edit-large-text-entry.txt).

Customization collection baseline36085164693 atbe336c60 passes phone/iPad.
Reviewed full-screen entry/search/footer captures confirm header clearance and
final long-name row reachability above persistent controls. Existing viewport
ownership is retained; no product correction or TestFlight build is needed.
[Decision and checks](evidence/customization-collection-clearance.txt).

Notification tab journey36057807116 at4b64bbd7 passes on phone/iPad. Reviewed
full-screen detail, pagination and final-row captures confirm tab/Back continuity,
read-state reconciliation and final content/action clearance. No product change
is justified. Physical push and route guard behavior are outside this fixture.
[Scoped evidence](evidence/notification-tab-journey.txt).

History baseline36027545902 confirmed iPhone final metadata hidden under persistent
chrome and a selected-mode trigger collapsed into an ellipsis. iPad failed Xcode
launch before UI interaction. Android confirmed dense change paragraphs; candidate
uses concise list summaries and separated Before/After values. Candidatef8cea8e0
also adopts automatic scroll insets and shared native choice picker.30 focused
tests, TypeScript,13 fixture tests, structural checks and critic review pass.
Corrected native36031360544 passes on both devices after one iPad Xcode-launch
retry. Reviewed list/detail/final-metadata/pagination captures confirm hierarchy
and clearance. Exact-head CI36034561755 passes. PR188 mergedb45c5e96; release
36037058030 delivered130.1 with verified Apple changelog readback.
Accessory follow-up is resolved as an evidence limitation below; no renderer fix is justified. [Current diagnosis](evidence/history-structure.txt).

Accessory review correction: full-screen baseline36035027845 shows the microphone
on list, settled detail and tab return. The element-only crops omit it and were
incorrectly treated as proof of a product defect. Native-symbol candidate36041316716
shows the same screenshot discrepancy. Withdraw the renderer change; retain
full-screen observation and document the evidence limitation. PR190 now contains
verification/process changes only. [Diagnosis](evidence/voice-symbol-navigation.txt).

Retained inventory collection: baseline36045543667 confirmed the iPhone final tag
behind persistent controls. Automatic insets fix it: candidatebe7d0f9a passes
unchanged native36048916536 on phone/iPad. Reviewed full-screen entry/final captures
confirm heading and final-tag clearance; source checks and critic review pass.
PR191 merged2a447c2b after final CI36051578009 passedfb3a4901; release36052329679
delivered132.1 with verified TestFlight notes. Root invitation acceptance is outside tabs
and is not implicated by this finding.
[Bounded diagnosis](evidence/inventory-collection-clearance.txt).

PR190 verification/policy merged296f7d6a after CI36045074264 passed1398cdad.
It contains no production renderer change and does not justify a TestFlight build.

## Separate unresolved decisions

Physical first-tap observation received2026-09-25: in response to the requested
TestFlight0.24.40 (132.1) tag-editor check, the user reports “opens right away,
all good.” Ordinary custom-color opening is accepted for that tested iPhone/build;
no color-control rewrite is justified by the earlier inconsistent simulator result.
This does not validate the unreleased PR161 candidate, coordinate probes, iPad,
color editing/save, or assistive activation. iPad Add creation/retry now passes the scoped refresh above.
[Device evidence](evidence/color-first-tap-device-132.txt).

- **M20 dismissal timing:** current iPhone onboarding refresh timed out awaiting
  keyboard disappearance; final capture/tree shows it dismissed and address intact.
  iPad passed. Timing remains unverified; no input rewrite or unchanged rerun.
  [Evidence](evidence/onboarding-current-dismissal.txt).

- **M251 Android photo status:** dialog-owned native adapter shipped in128.1 and shows
  readable light glyphs on the dark canvas in reviewed API36 captures. Close and
  Android Back restore the underlying screen in light/dark appearance. Final-photo
  disposal/restoration also passes both. Reviewed captures are retained with the
  evidence. Swipe replay failed both directions and remains unverified after the
  bounded investigation; older Android is not certified. No Activity-level style workaround was restored. [Evidence](evidence/android-photo-status.txt).

- **M51 ordinary color opening:** user verified immediate first-tap opening on
  iPhone/TestFlight132.1 on2026-09-25. The shipped SwiftUI picker matches release
  commit2a447c2b. Alternative UIKit replacement [PR161](https://github.com/elsell/stuffstash/pull/161)
  is closed without merging; its branch and failed candidate gates remain evidence.
  Do not revive it from the old simulator discrepancy alone. This closes the
  reported ordinary-opening concern, not iPad/assistive or arbitrary color-edit
  coverage. [Device evidence](evidence/color-first-tap-device-132.txt).
- **Text-input comparisons:** provider removal and paced entry did not establish
  a general correction. Fix reproduced consumers using the established native
  draft field, preserving reset and ownership semantics. Do not repeat the same
  provider/key experiments. [Consolidated evidence](native-text-entry-352471.md).
- Settings findings reviewed in this reconciliation have source corrections;
  their remaining native/assistive claims stay scoped to the named checks.
  Physical integrations, assistive behavior and wider device adaptations remain
  verification gaps. Historical “candidate” wording is not evidence of a current
  unimplemented defect; passed fixture workflows do not certify those gaps.

## Coverage and evidence limits

[Surfaces](surfaces.json), [axes](axes.json) and [matrix](matrix.csv) enumerate144
surfaces ×24 axes. Matrix classifications describe evidence, not3,456 separate test
requirements. The two recent selection routes now have source reviews for every axis;
[source follow-through](selection-surface-followthrough.md) preserves runtime gaps.
Source review does not establish native acceptance. [Findings](findings.md) retain stable IDs and historical evidence.
The older [whole-workflow review](everyday-workflow-review.md) records design
rationale; this file supplies current acceptance status.

The last full native sweep, [35980094051](evidence/native-full-359800.txt), ran121
cases per device: phone107 passed and iPad115 passed. It predates subsequent
fixes and is neither a current failure count nor whole-app certification. Verify shared controls once, representative
compositions and critical connected workflows; add coverage when ownership differs.


This is the sole current status summary. Update it in place. Keep exact diagnoses,
decisions and durable evidence; historical pending-run statements elsewhere are not
current state. Sleeping scripts collect terminal job results without unchanged
status narration. Do not weaken acceptance to make a batch pass.
