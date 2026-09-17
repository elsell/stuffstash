# Native color first-tap comparison —35210761064

Source176861cb6119ac4931a2fda9a6cf5a640aeadda5. Focused color selection only:
phone job105167451481 passes5/8; iPad105167451641 passes7/8. Results are in
native-color-first-tap-352107-results.csv. These are assertion/log observations;
this run's screenshots have not been visually reviewed.

Both devices pass the pre-tap-capture comparison and fail the otherwise identical
no-pre-tap-capture comparison. Both pass the original capture-before-opening
workflow this time. Therefore pre-tap capture is not a necessary trigger for the
failure; removing capture does not fix it. The fresh-launch comparison strengthens
timing/readiness as an investigation direction but cannot prove which native
presentation or mount condition is responsible. No arbitrary delay or retry tap
has been introduced into production or acceptance tests.

Phone also fails the standalone visible-well target and delivered-touch probe6
(top-right21,-21), after probes0–5 succeeded. iPad passes both. This means the issue
is not confined to one test's assertion sequence or necessarily to first-ever
presentation: presentation/dismissal transitions and hit testing still need
separate investigation. Do not erase the formerly successful nine-probe coverage,
but do not use it to claim universally reliable activation either.

Native RGB-retention and locked-well/unlock workflows pass on both; the115.1
fixes remain supported. M51 stays open. The next diagnostic must distinguish
native receipt of touch from successful presentation readiness, with a concrete
lifecycle signal rather than arbitrary sleeping or selection mutation.

Ten fixture installer checks passed on paul before dispatch, and code critic
found no blocker. The observer used sleep120 and collected terminal logs without
restarting either job. No product code changed in this diagnostic batch.
