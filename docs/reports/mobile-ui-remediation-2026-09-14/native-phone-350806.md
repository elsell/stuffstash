# iPhone native run350806

September16 run35080602419, iPhone17 fixture job104756068757, source da1a3e38
(merge c04b6a5bac882cf50df6a7926e415a421e8c13d7). Terminal result:61/81 tests
pass,20 fail. This predates the subsequent Android cards, fixture seed lifetime,
and asset completion guard changes. Local retained log: `/tmp/native350806-phone.log`.
Artifact10444113937 is retained on paul at `/tmp/native350806-phone.zip`.
Selected screenshots, hierarchies and the color tap event were inspected below.

Normal-text production Place search and settings collection search pass, including
header geometry and search navigation. The preconfigured Place comparison still
fails at Search hittability. Do not close M207 for all shared consumers from this
sample. The pending managed-options comparison remains useful for the discrepancy.

Both Conversation journeys pass. The location journey asserts context below the
native header before lookup failure/retry, searching, selection and Back. This is
new phone evidence for M216; iPad and resizing remain unverified.

The color well single-name, visible-target activation and all nine delivered-touch
probes pass. The separate ordinary opening/clear journey fails waiting for Sliders.
Pre-tap diagnostics were added after this revision and remain pending; do not infer
a cause or replace ordinary tap with a passing coordinate probe.

Account recovery passes; separate notice-placement journeys pass for pushed and
sheet navigation. Sharing
still fails because Cancel invitation is not hittable. Controlled and ordinary
text-entry comparisons lose characters; paced diagnostics do not replace the
normal typing acceptance. Add hidden-header entry and tag disclosure fail before
the requested editing workflow; later warm URL entry diagnostics are not in this run.
Provider credential recovery fails waiting for an interactive keyboard.

Other failures include the normal/full-sheet geometry comparisons and enlarged-text
cases. Retain these results, but prioritize normal-text findings per the user's
sequence. No TestFlight acceptance is claimed. Phone onboarding separately fails
its keyboard-dismissal observation; iPad onboarding passes. The iPad fixture job
was still live when this report was opened.

## Inspected native evidence

- [Production Place](evidence/phone-place-header-search-350806.png) has More and
  Search in the header, without a bottom search bar. The passing test exercises
  search and return. Its [preconfigured comparison](evidence/phone-place-preconfigured-bottom-350806.png)
  instead shows a bottom SearchField at y803–841. This is an observed discrepancy,
  not proof that preconfiguration is the cause.
- [Conversation entry](evidence/phone-voice-context-350806.png) exposes the context,
  proposal location, Approve and Cancel. Context minY245.1 is below header
  maxY243.2. The full location journey passes. This supports the phone entry fix;
  it does not certify iPad, reduced detents or all proposal lengths.
- [Failed color opening](evidence/phone-color-closed-350806.png) remains on the
  settings fixture with no system picker. Its synthesized event records touch-down
  and touch-up at(360,374.667), the center of the observed well frame
  (346,360.7,28,28). A video frame near the tap shows the well pressed. This
  rules out an obvious off-target coordinate in this attempt, but does not identify
  why presentation failed. Keep the pending pre-tap geometry diagnostic and the
  original direct-opening assertion; neither extend the timeout nor substitute
  the passing coordinate probes without stronger evidence.

Selected local artifacts are `/tmp/phone350806-color.{png,txt}`,
`/tmp/phone350806-color-tap.png`, `/tmp/phone350806-selected/` and
`/tmp/phone350806-manifest.json`. Video/event originals remain in the ZIP.
