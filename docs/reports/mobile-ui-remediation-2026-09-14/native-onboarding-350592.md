# Onboarding native run 35059267607

Source 79c32758, tested merge b94efa84bc2e2e15914c393c23c965fbef2e085f.
Phone job104676456100 passes its applicable connection-help/address/keyboard
journey; two iPad-only cases are skipped. iPad job104676455996 passes the same
main journey and landscape, but fails the separate form-column comparison before
typing. Logs are retained at `/tmp/native350592-onboarding-phone.log` and
`/tmp/native350592-onboarding-ipad.log`.

The inspected [phone action capture](evidence/phone-onboarding-action-350592.png)
shows the complete `https://example.invalid` address, dismissed keyboard and
unclipped Connect and sign in command at normal text size in light appearance.
This verifies the named journey, not sign-in against a real provider or all
appearance/accessibility combinations.

The iPad comparison fails line71, the initial `Server address` existence wait.
The log shows slow automation snapshots during launch; it never reaches typing
or the alternate dismissal gesture. Its [final capture](evidence/ipad-onboarding-entry-350592.png)
and hierarchy show the empty field present, labeled Server address, at
(96,246),552×34, with the surrounding form visible. This is an observation/readiness
failure, not evidence that the alternate gesture failed. The cause remains
unproven; no timeout relaxation or production change follows from this result.

Artifacts10431384932 (phone) and10432333074 (iPad) retain complete evidence.
The separate full fixture and Home-header runs remain active at this checkpoint.

## Passing follow-up — run35063391045

Source65acfd86, tested merge bd089623f61c94f1e9f9c0e7f15d398717cc1673:
phone job104689596375 passes its main journey (two iPad-only skips), and iPad
job104689596852 passes all three tests. This includes the previously failing
initial-readiness/inside-column keyboard-dismissal comparison without relaxing
its timeout or changing production after the previous observation failure.

Inspected iPad [dismissal](evidence/ipad-onboarding-dismiss-350633.png) and
[landscape](evidence/ipad-onboarding-landscape-350633.png) captures retain complete,
unclipped form controls and primary command. The dismissal capture shows the
exact entered address and no keyboard; landscape shows the empty-address disabled
command with explanatory text. These are normal-size, light-appearance checks,
not real-provider sign-in, VoiceOver or a whole-onboarding certification.

Artifacts10433788455(phone) and10434365601(iPad) retain the evidence. Logs:
`/tmp/native350633-onboarding-phone.log` and
`/tmp/native350633-onboarding-ipad.log`. The iPad success resolves this named
journey's latest failure; the older failure stays recorded above.

## New phone observation failure — run35080602419

Phone onboarding job104756069195 fails the five-second keyboard nonexistence
expectation after the margin drag (line110); the two iPad-only cases skip.
The address-entry assertion passed before the gesture. Its inspected
[final capture](evidence/phone-onboarding-final-350806.png) and accompanying
hierarchy show no keyboard, the entered address retained, and Connect and sign in
at (24,463),354×46, within the viewport. This does not prove dismissal completed
within the required interval: the screenshot was attached about ten seconds after
the failed expectation. Automation snapshots and event delivery were slow in the
log. Delayed dismissal and delayed observation remain unresolved alternatives.

Retain the original assertion and this failure; do not relax the timeout or claim
a production fix from the later screenshot. The iPad onboarding job104756068979
passes. The two fixture jobs remain running at this checkpoint. Artifact10442577020
and `/tmp/native350806-onboarding-phone.log` retain the phone evidence. This run
uses source da1a3e38, tested merge c04b6a5bac882cf50df6a7926e415a421e8c13d7.

Run350950 phone onboarding fails at OnboardingAuditTests.swift90: ordinary address
entry does not reach its complete expected value within the bounded five seconds.
The later keyboard assertions are not reached. One failure, two iPad-only skips;
log `/tmp/native350950-onboarding-phone.log`. This differs from the preceding350919
phone pass. Artifact inspection is still needed before assigning a product cause
or changing the acceptance check.
