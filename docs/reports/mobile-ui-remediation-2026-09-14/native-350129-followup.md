# Native follow-up after PR150

Run [35012949816](https://github.com/elsell/stuffstash/actions/runs/35012949816)
checked out `22a4a80d95af147bccc4558144991375add02b2e`, the PR merge revision
for head `96f1ea359287c37c57d60847f8175115f8d0dff2`. Both onboarding jobs passed.
Fixtures executed 58 tests per device: iPhone 17 had 15 failures; iPad mini
(A17 Pro) had 11. This is incomplete native acceptance, not a passed batch.

## Normal-text investigation

The ordinary single-line input lost/reordered characters on both devices. Its
fixture already uses `defaultValue`, not a controlled value. Uncontrolled URL,
system URL and ordinary multiline comparisons passed on both devices; the
controlled URL failed on phone. Different keyboard and assistance settings mean
these results cannot establish controlled state as the cause.

The next diagnostic compares identical ordinary text/default keyboard using the
unchanged uncontrolled baseline, a controlled value, assistance disabled, and the
accessory disabled without unmounting its native host. Exact native and observed
draft assertions remain. No production typing workaround is being introduced.
The fixed `text-entry` workflow selection runs these cases plus multiline on both
devices; it cannot replace full native acceptance.

Other normal-text failures include Home action target height (36 points on both
devices), phone sharing Cancel invitation reachability, phone full/nested sheet
checks, iPad place-contents search, and Add draft entry/readiness. Enlarged-text
failures remain recorded in the job logs and follow the user's later priority.

## M51 color picker remains open

Focused run [35011368960](https://github.com/elsell/stuffstash/actions/runs/35011368960)
recorded selection `color-picker` at `c8b460c2858e467610e5b48b23d053556930910a`.
Both tests failed on both devices: direct opening did not expose the system picker,
and the accessible target measured 28 points on phone / 36 on iPad rather than 44.
The inspected phone failure screenshot shows the settings fixture with the well
still visible and no system picker; its hierarchy exposes a 28×28 button.

The later full run passed open/close/clear on phone but failed it on iPad, and
failed target size on both. Activation is intermittent across runs; neither the
large control-size modifier nor the enclosing minimum frame resolves M51. Do not
infer the underlying presentation failure from geometry alone or weaken assertions.

Apple documents [UIColorWell](https://developer.apple.com/documentation/uikit/uicolorwell)
and [UIColorPickerViewController](https://developer.apple.com/documentation/uikit/uicolorpickerviewcontroller)
as native options. A bridge replacement remains an investigation, not an accepted
implementation or permission to substitute a custom color dialog.
