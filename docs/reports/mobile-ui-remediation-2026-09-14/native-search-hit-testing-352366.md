# Search hit-test diagnostics —35236686772

Source880b5d941c839ba26ae53a8bd7c1d55442c37b76; runner-only attachment-order
candidate. Phone105254293258 passes8/8. iPad105254292975 passes6/8; voice
clear/re-entry now passes on both devices. Browse keyboard readiness and
Expiration exact-query observation time out on iPad. Outcomes are retained in
native-search-hit-testing-352366-results.csv.

The new failure diagnostics materially narrow the next investigation. After the
Browse five-second readiness wait, its diagnostic records a hittable keyboard
and every valid key hittable. Invalid padding keys are recorded safely without
asking XCTest to hit-test them. This is later evidence, not proof that keys were
ready throughout the wait. The final Browse capture shows a focused empty input
and visible keyboard without an editing menu. No typing was attempted after the
readiness assertion failed.

Expiration's final capture and hierarchy show the exact query Tools and the
correct filtered Tools row, despite its prior exact-query timeout. They establish
eventual state, not timely acceptance. The job log also records long delays in
ordinary automation operations: opening the fixture waits about22 seconds between
SpringBoard idle checking and event synthesis; its Search tap waits about7 seconds
before returning from synthesis and another6 seconds before field observation.
This makes automation/runner latency a live hypothesis; it does not establish an
app-side lost-keystroke defect or justify increasing all timeouts.

Both failed final captures and hierarchies were reviewed and retained in evidence/
with352366 suffixes, alongside keyboard hit-test diagnostics. Next isolate slow
native observation from product response before repeating the same candidate run.
Do not alter production keyboard handling based only on these timeouts. M207 is
still open and the dependency ordering change remains diagnostic-only.

Remote fixture installer10 checks and critic review passed before dispatch.
The terminal-state observer slept120 seconds; no jobs were restarted. Ranged
artifact extraction transferred under1MB. TestFlight115.1 remains unchanged.
