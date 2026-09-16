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
The iPad onboarding and phone/iPad fixture jobs remain active at this checkpoint.
The ExpoUI lock path correction has passed deployment installation on all four jobs;
that integration success is distinct from the remaining UI acceptance gates.
