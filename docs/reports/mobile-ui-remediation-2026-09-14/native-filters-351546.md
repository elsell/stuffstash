# Focused filter verification — run 35154627907

Source bef2150ac65e087a27f51fe80176b1caf37a51f0. This run includes the
settled-keyboard measurement candidate and Browse search ownership fix.

| Journey | iPhone 17 | iPad mini (A17 Pro) |
| --- | --- | --- |
| Last tag clears footer and applies | Pass | Geometry passes; selected-tag result fails |
| Browse search, keyboard clearance and selection | Fails clearance | Pass |
| In-place availability and apply | Pass | Pass |
| Expiration calendar and bottom actions | Fails popover dismissal | Pass |
| Expiration search and keyboard actions | Invalid activation-frame error | Pass |

Phone job 104993585435 passes 2/5; iPad job 104993585635 passes 4/5.
An assertion pass is scoped to its journey, not a whole-screen visual acceptance.
The [iPad last-row capture](evidence/ipad-filter-last-tag-351546.png) and
[hierarchy](evidence/ipad-filter-last-tag-351546.txt) show the last row at
Y674.5–726.5, above the footer at Y746.5–886.5. Both action frames fit, and all
geometry assertions pass. The failure is the final selected-tag result after
row tap, Back and Show results: the result is `Browse availability: any`, not
`Browse selected tags: audit-last`. This corrects the earlier attribution made
from a line number in the current file instead of the run's source revision.
Inspect row tap delivery and selected-state transition before changing layout.

## Confirmed phone overlap

The [screenshot](evidence/phone-filter-keyboard-351546.png) and
[hierarchy](evidence/phone-filter-keyboard-351546.txt) show Back at Y491–545
and Dismiss keyboard at Y485–529. The accessory visibly obscures Back.
These coordinates match the earlier failure: settled notifications alone do not
resolve M249. The native test now rejects the overlap instead of accepting a
center-point tap. The phone journey stops before its post-selection search checks.
The iPad journey completes those checks, giving scoped native evidence for M250.

Expiration search has the same visible overlapping Back button. Its test aborts
while XCTest calculates an activation point, so it cannot independently certify
the clearance assertion. The screenshot supports the overlap finding regardless
of that automation error.

The calendar failure occurs after tapping the full-screen PopoverDismissRegion;
the captured calendar remains visible and the underlying sheet is displaced.
Inspect the actual synthesized tap location before attributing this to product
dismissal behavior. This is a separate investigation from keyboard clearance.

## Retention

Phone artifact 10472116092 is retained at `/tmp/native351546-filters-phone.zip`
on paul. Only the compact reviewed evidence above is copied into the repository.
The iPad artifact is 10471498466. Full local logs are
`/tmp/native351546-filters-phone.log` and `/tmp/native351546-filters-ipad.log`.
No production fix or new native acceptance is inferred from the failed candidate.

## Next diagnostic

The hierarchy puts the sheet origin at Y62, the footer bottom at Y557 and the
keyboard container top at Y495. The 62-point difference suggests a sheet/window
coordinate mismatch; the extender additionally draws its control 10 points above
its container. Neither number should become a hard-coded inset.

The focused filters fixture now exposes a small, noninteractive geometry sample:
an unmoved sibling bottom boundary measured with measureInWindow, the settled
keyboard event frame and their shared boundary calculation. This does not record
the production hook's applied inset and is not synchronized to XCTest; compare
the sample with final hierarchy frames before drawing a causal conclusion. Full
audit routes do not mount this diagnostic. The fixture lifecycle test and eight
route-isolation checks pass on paul, together with TypeScript and structural checks.
Critic review found no blocker; pending-callback-after-unmount is guarded in source
but is not independently demonstrated by the mounted test.

The retained [synthesized touch summary](evidence/ipad-last-tag-touch-351546.json)
places the last-row tap at372,700.5, inside its exposed98,674.5,548,52 frame.
Back follows at372,847.5. This rules out an obviously outside synthesized target,
not a dropped native press or lost state. The next acceptance observes the checked
AX value before Back and retains a screenshot/value attachment, while keeping the
final applied-tag check. The English pinned RN implementation emits
`checkbox, checked`; native execution is pending. Fixture preparation and structural
checks pass on paul; critic found no blocker. This is not a production selection fix.
