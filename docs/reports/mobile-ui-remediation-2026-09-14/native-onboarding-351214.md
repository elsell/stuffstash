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
VoiceOver, hardware keyboards, or enlarged text. Phone onboarding and both fixture
jobs remain live at this checkpoint; M240 phone help activation is not cleared by
iPad success. Leave those jobs undisturbed.
