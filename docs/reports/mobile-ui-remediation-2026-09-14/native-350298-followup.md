# Native run35029854251 — terminal evidence

PR source0c26c3bf; tested merge e8b3d42dccf3f13428fb26bbb1cfd85ea8b0dd9e
(artifact revision). Both onboarding jobs passed. iPhone fixture job104585543543
finished with **42/61 passing**; iPad finished with **49/61 passing**.

iPad mini fixture job104585543340 finished with **49/61 passing**. Its12 failures
include Add typing/readiness, controlled ordinary/address character loss, color
well AX height36 versus44, Home action AX height36 versus44, place-content search,
and three enlarged-text scenarios. Normal-size failures remain the priority.
The AX-height assertions do not by themselves establish UIKit's full touch region.

Color open/clear/parent-draft behavior passed in this sample, while its separate
target test failed. Ordinary uncontrolled entry passed here but failed in the
focused phone run350294; no flaky baseline is closed from one passing sample.
Direct/full/nested footer fixtures and Browse last-tag action-footer test passed
on this iPad run. These do not establish phone behavior or full visual acceptance.

Phone normal-size failures include Add typing/readiness, color picker activation,
color well AX height28 versus44, Home action AX height36 versus44, controlled
address entry, controlled/uncontrolled ordinary text entry, full/nested sheet
row reachability, and a non-hittable invitation menu cancellation command.
Seven other phone failures concern enlarged text or accessibility inspection.
These remain tracked separately while normal-size findings take priority.

Evidence is terminal XCTest logs `/tmp/native350298-ipad.log` and
`/tmp/native350298-phone.log`; screenshots have not
yet been inspected for this run. Fixture artifact10422138544 is available. This
build predates M156 and later fixes, including current native Save migration.
