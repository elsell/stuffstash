# Existing native run35170728754 — post-release collection

Source8bd07cc2 contains released production cutoff223d6d0a and corrected calendar
navigation targeting, exact native-query observation and per-case timeouts. It
does not contain the subsequent M251 status-bar candidate. A sleeping collector
waited for this already-queued run; no new native build was dispatched.

Phone job105056396340 passes70/87; iPad105056396424 passes80/87. The frozen matrix
records51/55 required phone and54/55 required iPad passes. Every outcome remains
in release-batch-checks.csv; totals are assertions, not visual certification.

Calendar dismissal/Apply/Back, complete Sharing email recovery, production-title
Tags Add/Search and ordinary color activation pass on both devices. The corrected
calendar test now completes its intended navigation instead of rejecting targets
behind the modal. Exact Tools query observation passes before both phone filter
journeys fail the full-action keyboard-clearance assertion. M249 remains open;
351689 provides its coordinate evidence.

Other required failures: phone Place integrated search and Expiration accessibility
audit (Text clipped; issue details need review before classifying this occurrence),
and iPad Account command cancellation/recovery at Swift line2445. These remain
tracked rather than being erased by earlier passes. Diagnostic input comparisons
and enlarged-text failures stay separate. The Account failure needs retained-state
inspection before attributing a new product regression.

Both standalone onboarding jobs pass applicable cases: phone105056396435 passes
one with two iPad-only skips; iPad105056396423 passes3/3. This does not establish
live OIDC sign-in. No new capture review was performed during log collection.

Logs: /tmp/native351707-JOBID.log. Job index: /tmp/native351707-jobs.json.
TestFlight0.24.24(113.1) remains the explicitly authorized interim release;
this evidence does not claim comprehensive audit completion.


## Account failure boundary

iPad job105056396424 fails at FixtureAuditTests.swift:2445 waiting for the
Account navigation bar after tapping the audit launcher. The log records event
synthesis, followed by the ten-second header wait and failure. Sign Out, Cancel
and recovery assertions are never reached. This run therefore provides no
acceptance evidence for those Account commands, and does not demonstrate a
Sign Out implementation failure. Inspect the retained final-state image and
hierarchy before changing product code or the test target. The launcher failed
to establish the expected destination; its cause remains unclassified.
