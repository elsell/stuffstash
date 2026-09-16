# Focused filter verification — run 35154627907

Source bef2150ac65e087a27f51fe80176b1caf37a51f0. This run includes the
settled-keyboard measurement candidate and Browse search ownership fix.

| Journey | iPhone 17 | iPad mini (A17 Pro) |
| --- | --- | --- |
| Last tag clears footer and applies | Pass | Fails footer containment assertion |
| Browse search, keyboard clearance and selection | Fails clearance | Pass |
| In-place availability and apply | Pass | Pass |
| Expiration calendar and bottom actions | Fails popover dismissal | Pass |
| Expiration search and keyboard actions | Invalid activation-frame error | Pass |

Phone job 104993585435 passes 2/5; iPad job 104993585635 passes 4/5.
An assertion pass is scoped to its journey, not a whole-screen visual acceptance.
The iPad footer failure still needs attachment review.

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
