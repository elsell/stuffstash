# Onboarding follow-up: run35046586497

Source b6321dcb. iPad job104638923255 passes all3 cases: connection help/keyboard,
keyboard dismissal inside the centered form and landscape adaptation. The
keyboard-readiness correction after run350420 is included here. This establishes
those XCTest journeys on this simulator build, not all onboarding visual axes.

Phone job104638923469 passes the connection-help/keyboard case. Two cases are
explicitly skipped: iPad centered-column comparison and landscape (the shipped
iPhone configuration is portrait-only). Do not report this as three phone passes.

Completed logs are retained at `/tmp/native350465-onboarding-ipad.log` and
`/tmp/native350465-onboarding-phone.log`. Fixture jobs subsequently completed;
see native-phone-350465.md and native-ipad-350465.md for their failures.
Later branch changes and TestFlight readiness are not validated by these results.

## Next run35050407693

At source802e4955, phone job104649681618 passes its connection-help/keyboard case
with two platform-inapplicable skips. iPad job104649681924 passes all three cases.
Logs `/tmp/native350504-onboarding-phone.log` and
`/tmp/native350504-onboarding-ipad.log` were inspected. Fixture jobs remain active
at this checkpoint. Neither onboarding job contains M206's later native command
migration at2edf4dbd; repeat those native scenarios on the migrated controls.
