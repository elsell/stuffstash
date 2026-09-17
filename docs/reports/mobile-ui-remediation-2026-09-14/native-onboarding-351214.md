# Native onboarding — run35121454700

Source1a15ca11, tested mergec5b6f1b110135c3d80f1433429c84dcd248d8fdb.
iPad mini job104881256074 completed successfully:3 tests,0 failures,78.687 seconds.
Artifact10457668983 (4,616,186 bytes) is retained locally as
`/tmp/native351214-onboarding-ipad.zip`; selected PNGs are in
`/tmp/onboarding351214-ipad-selected`. Complete log:
`/tmp/native351214-onboarding-ipad.log`.

The native suite passes connection-help/keyboard reachability, keyboard dismissal
inside the iPad form column, and landscape adaptation. Inspected
[keyboard capture](evidence/ipad351214-onboarding-keyboard.png) retains the complete
`https://example.invalid` address with the connect button above the software
keyboard. The [landscape capture](evidence/ipad351214-onboarding-landscape.png)
shows the centered form fitting onscreen and the empty-address hint beside the
unavailable action. No clipping or overlap is apparent in these sampled states.

The actual iOS address adapter uses a seeded SwiftUI TextField; the failing
controlled React Native diagnostic is a separate fixture. This result supports the
production address journey on this iPad configuration, not every text field or
input method. It does not verify real sign-in against a server, dark appearance,
VoiceOver, hardware keyboards, or enlarged text. Both fixture jobs remain live at this checkpoint. Leave them undisturbed.

## Phone follow-up

Phone job104881256446 also completed successfully: the connection help/keyboard
journey passed; two iPad-only tests skipped. There were0 failures across206.227
seconds. Artifact10459165589 (3,237,317 bytes) is retained locally at
`/tmp/native351214-onboarding-phone.zip`; captures are in
`/tmp/onboarding351214-phone-selected`, and the log is
`/tmp/native351214-onboarding-phone.log`.

The inspected [help capture](evidence/phone351214-onboarding-help.png) shows the
expanded guidance fitting without clipping. The
[keyboard capture](evidence/phone351214-onboarding-keyboard.png) retains the complete
address and keeps Connect and sign in visible above the keyboard. This run does
not reproduce M240. No targeted M240 production correction was introduced between
the failed and passing runs, so retain the earlier intermittent observation rather
than attribute the pass to an invented fix. Broader input methods and authentication
remain outside this configuration's scope.
