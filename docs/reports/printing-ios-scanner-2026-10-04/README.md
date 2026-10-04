# iOS scanner acceptance

[Native run 37179475027](https://github.com/elsell/stuffstash/actions/runs/37179475027)
passed all four scanner cases on iPhone 17 and iPad mini (A17 Pro) simulators.
Both downloaded artifacts identify source
`5492bdec734d72fb32c5f1f97ca821f5e45c2dd7`. XCTest reports zero failures:
331.840 seconds on iPhone and 232.197 seconds on iPad.

## Observed behavior

| Case | Result on both devices |
| --- | --- |
| Denied camera and pasted links | Reset camera authorization, observe the actual OS prompt, choose Don't Allow, require prompt disappearance, then see Settings/paste guidance. The real native Paste menu retains the exact complete input. Invalid and foreign-instance links receive the appropriate explanation. |
| Retry and resolution | A recoverable repository failure leaves a retry action; retry resolves an old-host link through the configured instance and navigates to its asset. |
| Cancel while lookup is pending | Native Cancel returns to the prior screen. Completing the delayed reply leaves no pending label and does not select an inventory or navigate to the asset. |
| Pending label and sign-in readiness | A controlled signed-out/ready transition retains the pending link; the real pending-label navigation component then opens the intended asset. This is simulated readiness, not OIDC. |

Selected screenshots and cancellation hierarchies:

| Device | OS prompt | Denied fallback | Late reply after Cancel |
| --- | --- | --- | --- |
| iPhone 17 | [Prompt](phone-scanner-os-camera-prompt.png) | [Fallback](phone-scanner-camera-denied-paste.png) | [Screen](phone-scanner-cancelled-late-reply.png), [hierarchy](phone-scanner-cancelled-late-reply-hierarchy.txt) |
| iPad mini | [Prompt](ipad-scanner-os-camera-prompt.png) | [Fallback](ipad-scanner-camera-denied-paste.png) | [Screen](ipad-scanner-cancelled-late-reply.png), [hierarchy](ipad-scanner-cancelled-late-reply-hierarchy.txt) |

## Scope and earlier failures

The runner uses the production scanner route, native sheet/header, camera adapter,
parser and resolution use case. Its repository is a stateful fake that can complete
a request after cancellation; its destination screen exposes navigation results.
This establishes those interactions, not connected iOS authorization or the full
asset screen. [Connected Android evidence](../printing-android-2026-10-04/README.md)
separately covers real Dex owner/unrelated-account journeys on the normal APK.

Earlier run `37176663968` stopped some cases before lookup because synthesized
bulk typing lost a prefix or failed to focus the keyboard. The acceptance driver
now uses actual Paste with exact text assertions; it does not establish ordinary
typing reliability. Run `37178128691` passed all iPad cases, but its iPhone denial
tap encountered an XCTest interruption-targeting error and left the OS prompt
visible. The corrected driver touches the observed denial-button center and
requires the prompt to disappear; it never retries denial or pregrants access.

No result claims physical camera decoding, printed-label readability, verified
HTTPS app association, real iOS OIDC return, VoiceOver, or enlarged-text coverage.
Later release-metadata synchronization and this report do not change the tested
scanner or fixture source.
