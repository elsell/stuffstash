# Native audit and released color batch —35199347918

Source2876cc19df60937ff4c0bf271f3b3a5348074f09; tested merge
333bbb315f6d806f00bfa641fca602dff54e917d. Phone job105130091256 passes73/89;
iPad105130091147 passes82/89. Exact outcomes are in native-full-351993-results.csv.
Standalone onboarding passes its applicable cases: phone1 passed/2 skipped,
iPad3 passed. Assertions do not establish blanket visual acceptance.

M252 RGB retention and M253 native lock/tap/unlock cases pass on both devices.
Locked screenshots were visually reviewed on phone and iPad: selected Green is
retained, controls are dimmed, and the parent draft remains #2E7D32. Hierarchies
expose the native well as Disabled. Delivered center taps do not present Sliders
within the two-second negative observation; unlocking permits normal presentation.
Selected locked captures and hierarchies are retained under evidence/ with351993
suffixes. This accepts the tested normal-text state transitions, not every
consumer route, appearance, rotation or window configuration.

Ordinary color opening still fails on phone and passes on iPad; M51 remains open.
Notification delivered-touch coverage passes on both, while the prior351910 iPad
pre-probe timeout remains historical evidence. Existing diagnostic compositions,
text-entry comparisons and enlarged-text findings remain tracked separately.
No blanket rerun was launched to replace these failures.

## Next investigation order

Continue normal-text production findings first: intermittent color opening (M51)
and inconsistent native search placement (M207). Compare the failed ordinary
color entry with successful delivered-touch and seeded-color entries before
changing production code. Diagnostic-only comparisons and large-text follow-up
must not silently expand a frozen release gate.

## M51: failed entry compared with successful controls

The reviewed phone late-observation capture shows the original settings controls,
no system picker, unchanged `Color value: none`, and an enabled transparent color
well. Its recorded target before the tap is (346,397.67,28,28), hittable=true;
the final hierarchy retains the same frame. No presentation appears during the
initial five seconds or fifteen further seconds. This is not merely a five-second
assertion timeout with a later visible picker. Evidence is retained under evidence/
with color-after-late-presentation-observation, color-ordinary-tap-target and
color-late-presentation names suffixed351993.

In the same run, `testColorWellTargetOpensSystemPicker` opens the same initially
unset control with `picker.tap()`, and the delivered-touch case opens it at all
nine center/edge/corner probes while retaining the unset value. Thus neither a
blank initial value, an intrinsically untappable 28-point accessibility frame,
nor ordinary element tapping alone explains the failure. The failing case also
captures a screenshot and full accessibility hierarchy before tapping; the
successful target test does not. This is a test-sequence difference to isolate,
not proof that capture causes the failure.

The next controlled comparison should preserve assertions and initial unset state,
record pre/post target frames and native presentation state, and compare capture
before first tap against capture after first tap within independent fresh launches.
Do not invent a default persisted color, add retry taps that conceal missed first
activation, or increase timeouts as a claimed product fix. Native presentation
ownership/mount readiness requires direct evidence before altering the adapter.
