# Native audit build failures, run350919

Source3dc2bab6; September16. Neither fixture job reached native tests:
iPhone job104781815983 and iPad job104781815954 fail Swift compilation because
`XCUIApplication.open(URL)` requires iOS16.4 while the test target supports older
versions. The two calls serve the managed-search comparison and Add URL entry.
A shared availability-guarded helper now preserves URL entry on supported OSes
and explicitly fails unsupported runtimes. It changes no product behavior or
acceptance assertions. macOS compilation and runtime verification remain pending.

The iPad onboarding job104781815556 independently fails dependency installation.
React Native's prebuilt probes return no artifacts and fall back to source;
`pod install --deployment` rejects the resulting dependency change. The deployment
lock remains intact. This supplies no onboarding runtime evidence. Phone onboarding
was still running when this report was written.

Logs: `/tmp/native350919-phone.log`, `/tmp/native350919-ipad.log`,
`/tmp/native350919-onboarding-ipad.log`. Preserve previous run failures; do not count
these build failures as passed or as newly observed UI regressions.
