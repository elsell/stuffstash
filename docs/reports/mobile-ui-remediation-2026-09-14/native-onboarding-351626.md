# Onboarding regression — run35162604601

Source c40e7965, tested merge cdaafb1a6598dbe0e068fad34bd1a64ba6f3a76c.
Phone job105023220045 passes the standard connection/help/keyboard journey; two
iPad-only checks skip. iPad job105023219935 passes all three cases: standard entry,
keyboard dismissal from inside the form column, and landscape adaptation.

Reviewed iPad [form-column result](evidence/ipad-onboarding-form-action-351626.png)
retains the complete https://example.invalid address with keyboard dismissed and
Connect and sign in fully visible. The [landscape result](evidence/ipad-onboarding-landscape-action-351626.png)
shows the whole centered form, empty-address explanation and disabled action within
the landscape viewport. This is not an authenticated OIDC session or a production
server connection; it verifies entry and presentation before that boundary.

The intermittent M240 help-opening issue does not reproduce in this run. No
production fix for that issue is established. This source predates the Sharing
email correction, which does not change onboarding.

Logs: `/tmp/native351626-onboarding-phone.log` and
`/tmp/native351626-onboarding-ipad.log`. iPad artifact10474788496 remains on paul
at `/tmp/native351626-onboarding-ipad.zip`; selected captures are retained above.
