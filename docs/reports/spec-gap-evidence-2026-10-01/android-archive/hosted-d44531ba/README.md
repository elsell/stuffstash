# Android archive runtime — October 3, 2026

Scoped fixture acceptance at `d44531ba49ee840c8ad9e1b991952f95a7efd0a1`.
[Hosted build 37080427792](https://github.com/elsell/stuffstash/actions/runs/37080427792)
succeeded and produced the APK identified in [installation-result.json](installation-result.json).
The APK was installed on the isolated API 36 emulator on Paul. No production
credentials or inventory were used.

| Acceptance gap | Observed result |
| --- | --- |
| Native upload integrity | Passed: real native file upload, 1 MiB length/hash and synthetic authentication verified by loopback peer |
| Rejection and recovery | Passed: redirect and oversized response rejected; subsequent upload and cancellation verified |
| Restore review and approval | Passed on corrected selector: preview, close without approval, reopen, approve and destination callback |

The [native transfer screenshot](transfer-result.png) and UI tree
[transfer-result.xml](transfer-result.xml) show the fixture's completed checks.
The [review screenshot](review.png) was visually inspected: inventory name,
counts, omission summary and bottom actions are visible. This is representative
normal-text fixture evidence, not whole-app visual acceptance.

Automation stopped at approval. The selector matched the non-interactive page
title `Restore inventory` before the identically named bottom action, then climbed
to the XML root looking for a clickable parent. The raw error `{'rotation': '0'}`
in [result.json](result.json) is that root's attributes, not a product error.
[approve.xml](approve.xml) preserves both nodes. No approval tap was issued.
[review-retained.xml](review-retained.xml) records returning from the first review.
Do not infer a broken restore button or a successful restore from this run.

One bounded correction selected only matches with enabled clickable ancestors.
A regression against the saved XML chose the bottom action at
`[53,2034][1028,2159]`, skipping the page title. The same APK was reused without
rebuilding. [The corrected result](corrected-selector/result.json) passed both
transfer and restore-review checks. The inspected
[approved screenshot](corrected-selector/approved.png) shows the synthetic
`Restored: Camping equipment` destination callback. This is not a real server
inventory opening. The initial failed result above remains unchanged.
 The runtime peer and owned emulator process groups
were cleaned up by the installer, and its downloaded APK was removed. The other
project's emulator was not managed by this script.

The restore UI uses a synthetic repository and picker. Authenticated server
restore, physical document-provider behavior and TalkBack remain separate,
unverified requirements. The earlier local Gradle failure remains preserved in
[the parent report](../README.md).
