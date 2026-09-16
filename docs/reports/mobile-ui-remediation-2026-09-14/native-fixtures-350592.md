# Native fixtures — run35059267607

Source79c32758, tested merge b94efa84bc2e2e15914c393c23c965fbef2e085f.
iPhone17 finishes56/75; iPad mini(A17 Pro) finishes64/75. These are test counts,
not product compliance. Both jobs failed. Onboarding is recorded separately.

## Normal-size evidence

Both devices pass Add in a navigation stack and Add with its header configured
before presentation. The old hidden-header Add diagnostic fails initial field
readiness on both. This run predates the native inline-tag candidate f7b09835.

Phone Place and settings collection fail before search activation. Inspected
[Place](evidence/place-bottom-search-350592.png) and
[Tags](evidence/settings-bottom-search-350592.png) captures show persistent bottom
search fields while More/Add remain in the header. The static search comparison
passes. Both production search journeys pass on iPad. M215's stable option identity
therefore does not establish corrected phone placement; M207 remains open.

The pinned native implementation assigns the search controller before its placement
and toolbar-integration properties. This source ordering is an investigation lead,
not proof of a UIKit defect. A new runner-only comparison uses the same production
Place fixture with search configured at route registration. Both variants share
the complete search/result/clear/cancel/Back journey and assert header placement.
No production workaround or dependency patch is introduced.

Color selection has a single accessible name on both devices. Direct opening,
clear and parent-draft retention pass on phone; iPad still fails the direct-opening
assertion. Separate target tests stop at28-point(phone)/36-point(iPad) accessibility
frames before activation. These failures do not measure the delivered touch region.

Sharing fails differently: phone reaches a Cancel invitation menu item that is not
hittable; iPad loses email characters before submission (`aample.invalid` instead
of `audit@example.invalid`). The iPad voice-location journey also fails and needs
its own capture classification. Do not infer its cause from unrelated search tests.

Both Home layout tests stop at36-point accessibility frames. The later independent
touch evidence and corrected layout assertions are documented in
`native-home-350598.md`; this run predates that correction and the real tab-shell
fixture. It cannot verify either change.

## Remaining diagnostic and enlarged-text failures

Phone text comparisons fail controlled address, controlled/no-accessory,
ordinary controlled and ordinary/no-accessory exact entry. iPad fails controlled
address and controlled/no-accessory. Preserve exact assertions; paced typing or
disabled text assistance is not acceptance for ordinary typing.

Phone also fails enlarged asset-region, command-height, detail-command, edit
metadata, edit tags, expiration overview and Move Here checks, plus full/nested
sheet layout checks. iPad fails enlarged edit metadata, edit tags and Move Here.
Normal-size work remains first; none of these failures is silently waived.

## Evidence and next verification

Phone job104676455804/artifact10433550778; iPad job104676455958. Full logs were
retrieved, and the phone search PNGs and accessibility hierarchies inspected.
Other failure descriptions above are log-level evidence unless explicitly stated.
The new paired Place fixture passes2 installer tests, TypeScript and mobile
structural checks on paul after an observed missing-route RED test. Critic found
no blocker. Swift compilation and native execution of the new journey are pending.

## Inspected iPad follow-up

Artifact10432808528's [color capture](evidence/ipad-color-unopened-350592.png)
shows the parent controls and no system picker after activation. The hierarchy
contains the single-name color well but no Sliders/popover. This is an unopened
picker, not merely a changed Sliders selector. M51 remains unresolved; do not
infer a touch-region defect solely from the36-point accessibility frame.

The [voice capture](evidence/ipad-voice-before-location-350592.png) shows the
Conversation sheet before destination navigation. Its review viewport is
`(98,642.5),548×75.5`, while the location command is `(168.5,557.5),451×44`, above
that viewport. The first app ScrollView belongs to the dimmed background.
The test used that first scroll view and always swiped up. It never reached
location search, so this failure does not demonstrate broken native search.

The corrected procedure scopes scrolling to the location command's containing
scroll view and chooses direction from current bounds. It requires the complete
control frame inside the sheet viewport below the Conversation header on entry,
after selection and after Back. All lookup/retry/search/proposal checks remain.
Mobile structural validation passes on paul. Native execution remains pending;
this is an acceptance-procedure correction, not a product-layout fix or a pass.

Critic identified that gesture coordinates must also exclude the header; the
helper now clips both visibility and gesture bounds consistently. In this capture,
the header ends at677.5, leaving only40.5 points of unobscured scroll viewport for
the44-point control. Native scrolling/detent behavior must resolve that constraint
or the full-frame requirement will still fail. The context line also visually
overlaps the header. This is a remaining sheet-layout investigation, not evidence
that correcting the test alone fixes the conversation experience.
