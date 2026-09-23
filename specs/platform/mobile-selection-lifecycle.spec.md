# Asset Selection Visit Lifecycle

## Confirmed defects and decisions

M271: run35880132749 phone capture confirms Add returns from Tags with an empty
asset name and disabled Save after successfully typing Tent. The pinned Expo UI
TextFieldView.swift assigns props.defaultValue in onAppear. DraftTextField.ios
freezes that prop to its first mount value, so native reappearance can restore an
obsolete empty name and emit it into the owning draft. This is a data-loss defect.
Keep the native field instance and native-owned typing, but pass the current
committed draft as its reappearance seed. Updating defaultValue does not command
a live field to replace text; the pinned adapter reads it on appearance. Explicit
scope/reset keys remain responsible for deliberate draft replacement. No timers,
keyboard-provider changes or discarded empty-text events are justified.

Shared consumers are Add name/new tag, Edit new tag, Move creation and Move Here
search, Add destination creation, and customization enum option entry. Preserve
current validation hints, busy guards, normal editing and explicit resets.

M272: the same run's Edit capture shows the Tags card behind the still-present
Edit formSheet, leaving its choices untappable. Root-stack child card presentation
cannot be accepted as a visible selection task for modal owners. Choose a native
presentation that appears above both Add and Edit without losing their drafts.
Use a shared fullScreenModal selection presentation on iOS and the ordinary card
route on Android. The searchable selection task has its own Cancel/Done or
tap-to-choose return, appears above a modal or card owner, and retains the parent
form rather than dismissing/recreating it. No sheet detents or duplicate parent
chrome are needed. This is a targeted adaptation of the current root stack, not
a new general modal rule. Do not merely adjust padding or weaken hittability checks. The Add destination
selection route shares this owner/presentation relationship and needs the same
review. Its Android result does not establish iOS modal presentation.

## Verification

Reproduce M271 with the pinned adapter's reappearance behavior through a controlled
native boundary: type a name, observe the committed draft, disappear/reappear with
the current native seed, verify the draft survives, then perform explicit reset.
Run representative shared-consumer tests and critic review. Native acceptance must
retain exact typed-name assertions through selection Cancel, Done and rejected save
on phone/iPad. Keep existing whole-string entry and clear/retry evidence; do not
restart provider or keyboard investigations. M272 needs visible, tappable selection
and predictable return from both parent forms. These fixes are follow-ups; they do
not silently expand frozen M260–M264 acceptance.

The grouped selection-lifecycle native run covers Add destination, Add/Edit Tags,
and unfinished Add tag disclosure. Assert the selection's Done command is hittable
before trying to scroll its choices; accessibility-tree presence alone did not
prove it was above the editor. Use targeted T/C readiness for these Add inputs and the established lowercase t
for native search; literal space was absent in retained iPad evidence, retaining exact typed-value assertions. Staging an unfinished
tag after disclosure may close creation; require the staged tag and enabled Save,
and require any remaining entry to be empty. Do not require a cleared field to
remain mounted. Native acceptance remains pending for M271/M272.

## Native run 35890124446 decision

Add Tags and Edit Tags passed on phone and iPad, including exact Add name retention
and rejected-save recovery. Add destination passed on iPad; phone stopped when the
harness requested task Cancel while native search was still active. Close the
native search submode before requesting task Cancel, then retain the original
parent destination and exact name assertions. Do not clear or alter selection to
make this pass. This is the same observed search-mode distinction as Move Here.

Unfinished tag disclosure remains unverified: phone failed to obtain a keyboard
on the reopened field; iPad entered Camping and reached Add tag, then exceeded the ten-minute case allowance during repeated scroll/geometry observations. Preserve these gates
and inspect retained evidence before choosing a correction; do not repeat input
provider or key-delivery experiments. These results do not close the whole batch.

The disclosure harness uses `for ... where !visible()`, which evaluates expensive
native geometry for every remaining iteration even when the field is already
visible. Replace it with an explicit early return on visibility and one geometry
snapshot per attempt. Preserve bounded scrolling, header/viewport containment,
hittability, ordinary field tap, and exact Tent/Camping checks. This fixes a known
observation cost, not a proved production focus defect. One corrected lifecycle
run distinguishes exhausted observation time from a persistent focus failure. If
phone focus still fails, retain that failed gate and choose an input-lifecycle
correction from its evidence rather than repeating keyboard experiments.
