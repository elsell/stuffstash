# iPad fixture run 35070277508

Job104732071655 completed with71/81 checks passing on iPad mini (A17 Pro).
Source e327f843 includes the iPad tab-strip selector; it predates the Conversation
header-height reservation and color accessibility-frame gate correction. Both
onboarding jobs passed. Phone fixtures subsequently completed60/81; full run failed.

Ten failed scenarios: hidden-header Add draft, ordinary Add draft recovery, color
picker clear-parent-draft, color accessibility target geometry, two controlled
text-entry diagnostics, three enlarged-text Edit/Move scenarios, and Conversation
location context below navigation. Home tab return no longer appears among the
failures. Do not call the full suite accepted, or attribute these failures to
later source changes not present in this run. Artifact inspection remains needed
for the new ordinary Add/color-parent failures before diagnosing their causes.

Evidence: terminal GitHub job log, locally `/tmp/native350702-ipad.log`.

Phone failures include ordinary Add recovery, color picker open/target checks,
controlled text diagnostics, enlarged-text cases, intentionally wrapped layout
comparisons, Place and preconfigured Place search, Settings collection search,
Sharing cancellation, and Conversation context. These are scenario results, not
21 independently confirmed product defects. Native Home and notice checks are
absent from the phone failure list in this run.

The next run35080602419 (da1a3e38) is now live; run35084537103 (da5ded17) is pending.
The newer fixes have not yet received their native acceptance result.
