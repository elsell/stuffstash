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

## Inspected captures

The purported ordinary Add failure shows **Loading inventory / Audit query state**,
an unrelated fixture, after the menu-label tap. It did not reach Add and cannot
establish an Add product regression. Also, audit-add/audit-add-push intentionally
start with hidden headers; production Add preconfigures its header as does the
passing audit-add-header comparison. Isolated Add tests now open their fixture URL
with Apple's XCUIApplication.open API; their draft assertions remain unchanged.
Native execution of this setup correction is pending.
[Capture](evidence/ipad-add-wrong-fixture-350702.png).

Color's final capture remains on the correct settings-control fixture with its
picker closed; its well frame is x684/y326.5/36×36. All nine delivered-touch probes
pass in the same run, so ordinary activation remains unresolved. A pre-tap capture
and target geometry attachment now supplement that scenario without replacing
its ordinary activation with a coordinate tap or relaxing its open assertion.
[Capture](evidence/ipad-color-unopened-350702.png).

Apple documents [XCUIApplication.open](https://developer.apple.com/documentation/xcuiautomation/xcuiapplication/open(_:))
as URL-based application entry. This is test setup, not proof of production deep-link
or navigation correctness.

Critic found no setup blocker. The color pre-tap capture adds time before activation;
a changed outcome must not be attributed to a product fix, since no picker code
changed. Structural checks passed on paul; Swift compilation and native execution
are pending. The diagnostic change does not resolve the ordinary-tap finding.
