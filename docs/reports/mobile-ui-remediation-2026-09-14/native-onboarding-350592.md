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
