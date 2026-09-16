# Onboarding native verification — run35131892834

Source dfad17ab; tested merge21939a989d89e3f0e3a4bf10133f3dab63c8b073.
iPhone17 job104914725508 passes its applicable connection/help/keyboard journey;
two iPad-only tests skip, with zero failures. Logs are retained at
`/tmp/native351318-onboarding-phone.log`. Artifact10462681611 is3,214,831 bytes,
retained at `/tmp/native351318-onboarding-phone.zip`.

Inspected phone captures:

- `55896409-C09D-40E1-A273-CFBB6750F865.png`: expanded help fits and the empty-address
  command remains disabled with corrective guidance.
- `AE55E85E-2FC6-47DE-803D-E79660235FC5.png`: the complete synthetic address
  `https://example.invalid` remains visible and Connect and sign in is above the
  software keyboard without overlap.
- `7BE8EB0D-E5C8-4FDA-9F32-9B69D54AF7E9.png`: after keyboard dismissal, the complete
  address and enabled command remain visible.

Selected files are under `/tmp/native351318-onboarding-phone-selected`.
This verifies the named normal-size, light-mode phone scenario. It does not verify
real OIDC sign-in, every orientation, accessibility text sizes or assistive output.
The phone/iPad fixture jobs remain active at this checkpoint.
The ExpoUI lock path correction has passed deployment installation on all four jobs;
that integration success is distinct from the remaining UI acceptance gates.

## iPad result

iPad mini job104914725527 completed with one pass and two setup failures.
The connection/help/keyboard test timed out launching the app through Xcode.
The inside-form keyboard-dismissal test failed to acquire the app's background
assertion after a launch-service process error. Neither reached its interaction
assertions. These are missing acceptance evidence; the logs do not establish a
product UI defect or its underlying cause. No assertions have been relaxed.

The subsequent landscape test passed. Inspected capture
`DEFB2E3B-FFAA-45B4-BF42-0DCC34330B76.png` shows the centered form, address field,
help link, empty-address guidance and disabled connection action within the
landscape viewport. The preceding failed test's final capture
`E3B621ED-E803-44E6-9871-7D47331EF411.png` shows the empty portrait form, not the
keyboard-dismissal journey; it cannot clear that scenario.

Log: `/tmp/native351318-onboarding-ipad.log`. Artifact10462697292 is104,314,911
bytes and is retained on paul at `/tmp/native351318-onboarding-ipad.zip`.
Only selected captures were copied locally to conserve disk. The already queued
run35132227322 provides another opportunity to verify the two missing scenarios;
the active fixture jobs remain undisturbed.
