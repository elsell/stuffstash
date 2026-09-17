# Filter search timing diagnostic —35240908816

Source0e00208db8da672846c215eb67cde66f10568b8d. Focused ordering-candidate
selection: phone105268802473 passes1/2 (Expiration), iPad105268802170 passes0/2.
No production change or timeout increase. Exact outcomes and raw monotonic timing
attachments are retained alongside this report.

Every failed keyboard-readiness wait performs a single false evaluation:
- iPad Browse:6.5335 seconds evaluating;7.6207 seconds total wait.
- iPad Expiration:4.6531 seconds evaluating;5.7098 seconds total wait.
- Phone Browse:5.1423 seconds evaluating;6.2092 seconds total wait.
- Phone Expiration passes readiness:1.3496 seconds evaluating;2.3857 seconds total.

The first evaluation starts about one second into each five-second wait. Thus the
failed samples consume the observation window in a single expensive evaluation;
they are not evidence of five seconds of repeated negative observations. This
proves observation cost, not app responsiveness or keyboard usability at tap time.
Phone Browse then fails while resolving all keyboard keys for failure diagnostics
(line580), before the original assertion can report. Timing evidence was already
attached and retained. Do not expand that exhaustive diagnostic further.

Next reduce the number/cost of accessibility queries without removing real input,
exact-query, filtering or action-clearance checks; distinguish framework query
failure from app behavior. Do not change production keyboard lifecycle from this
evidence. The runner-only search ordering candidate remains unpromoted and M207
open. Ten remote fixture checks and critic review passed before dispatch. The
observer slept120 seconds and did not restart jobs. Artifact range reads totaled
under2MB; whole archives were not downloaded.
