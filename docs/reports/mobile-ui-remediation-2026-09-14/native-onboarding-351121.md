# Phone onboarding — run35112198520

Source1c2f8173, merge7b5627db6ebf613330b48bc20430ae3ad5d804c7, phone job104851680347.
Terminal failure: one applicable journey fails, two iPad-only journeys skip. The
other three jobs were still running when reviewed; do not cancel or restart them.

The connection-help test fails at OnboardingAuditTests.swift:94. Server address
entry eventually becomes available; the test verifies the help button hittable,
synthesizes a tap, and waits five seconds for the explanatory text. The final
capture and hierarchy both show the collapsed form, with the help button at
(24,322),152.3×44 points. No keyboard or overlay obscures it. Help did not open;
this is not a text-selector-only discrepancy. See
[final capture](evidence/phone351121-onboarding-help-closed.png).

Source inspection shows a React Native Pressable inside the handled-taps ScrollView,
with a functional local-state toggle and no navigation/mutation side effects.
Nothing in this inspection identifies the missed event's cause. The job records
slow initial readiness and element observations; that is context, not evidence that
the runner caused this failure. The previous run351041 passed phone onboarding.
Do not lengthen timeouts, add a second tap or claim a source fix without evidence.

The test stops before address typing, keyboard dismissal or sign-in recovery; those
scenarios are not established by this run. Phone fixture and iPad results remain
separate. Artifact10455761027 retains the complete result; selected captures and
hierarchies are extracted, preserving the original on GitHub. Downloaded zip is
69,268,489 bytes on paul.

Next acceptance: a single help activation opens in place, a second closes, both
preserve the address and leave sign-in reachable. Reproduce on the queued native
revision and inspect event/layout evidence if it recurs. Track M240 as open.
