# Onboarding evidence — run35140471580

Source a1b827e0, tested merge a14a86aaa33e5a8169feab902988bea15a9bac15.
iPad mini job104943442121 passes3/3 tests in136.934 seconds: connection/help and
keyboard reachability, keyboard dismissal from within the form column, and
landscape adaptation. The previous launch failures did not recur in this job.

Inspected `B2181DBE-0E8D-40C5-9024-D023F55790D4.png`: exact
https://example.invalid is retained, and enabled Connect and sign in remains above
the software keyboard without overlap. Inspected
`4A36B425-88DB-4819-8DDA-740F96A80BDA.png`: the landscape form, empty-value guidance
and disabled action fit within the viewport. This is scoped normal-size/light
onboarding evidence, not real authentication or general text-input clearance.

Log `/tmp/native351404-onboarding-ipad.log`; artifact10465890785 (4,936,579 bytes)
at `/tmp/native351404-onboarding-ipad.zip`; selected captures under
`/tmp/native351404-onboarding-ipad-selected`. Both fixture jobs remain active at this checkpoint.

Phone job104943442251 passes its applicable connection/help/keyboard journey in
51.902 seconds, with two iPad-only skips and zero failures. This is log evidence;
phone captures from this run have not been inspected. Log is retained at
`/tmp/native351404-onboarding-phone.log`; artifact10466210779 is3,239,777 bytes
and remains available in GitHub Actions. Both onboarding jobs are now successful,
independently of the unresolved fixture interactions.
