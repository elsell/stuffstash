# Onboarding evidence — run35156794515

Source2043abb0, tested merge2209ff1c347187e92eef215f965b5fe68e040df2.
iPad job105001197814 finishes2/3 cases. Standard help/keyboard/action journey and
landscape adaptation pass. The inside-form-column comparison fails at help opening,
before address entry or its alternate keyboard-dismissal gesture. This is not
observed failure of that gesture. Phone results remain pending at this inspection.

The [final screenshot](evidence/ipad-onboarding-help-351567.png) and
[hierarchy](evidence/ipad-onboarding-help-351567.txt) show collapsed help at
96,292,152.5,44, without a keyboard or overlay. The address remains empty and
Connect is correctly disabled. This matches the class of intermittent M240
activation observed earlier on phone; the specific cause is unresolved.

## Frozen-batch decision

Retain M240 as noncritical audit follow-up rather than adding it to the frozen
release gate. The baseline-to-cutoff OnboardingScreen diff leaves the help handler,
control and help content unchanged; this batch changes the footer commands. The
standard help/open/close/complete-address/dismiss/action journey passes on this same
build. No evidence currently attributes the isolated comparison failure to the
changed footer. This is a scoped release judgment, not a claim that the help bug is
fixed or that every onboarding repetition succeeds. New evidence of input loss,
unreachable sign-in or a batch-caused navigation failure would still block release.

Artifact10472545793 (74.9MB) remains on paul at
`/tmp/native351567-onboarding-ipad.zip`. Local log:
`/tmp/native351567-onboarding-ipad.log`. Only compact selected evidence is copied
into the repository; the main full-run jobs continue unchanged.
