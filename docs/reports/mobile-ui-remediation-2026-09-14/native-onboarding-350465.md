# Onboarding follow-up: run35046586497

Source b6321dcb. iPad job104638923255 passes all3 cases: connection help/keyboard,
keyboard dismissal inside the centered form and landscape adaptation. The
keyboard-readiness correction after run350420 is included here. This establishes
those XCTest journeys on this simulator build, not all onboarding visual axes.

Phone job104638923469 passes the connection-help/keyboard case. Two cases are
explicitly skipped: iPad centered-column comparison and landscape (the shipped
iPhone configuration is portrait-only). Do not report this as three phone passes.

Completed logs are retained at `/tmp/native350465-onboarding-ipad.log` and
`/tmp/native350465-onboarding-phone.log`. Fixture jobs104638923397 and104638923406
remain active. Later branch changes, full fixture acceptance and TestFlight
readiness are not validated by these onboarding results.
