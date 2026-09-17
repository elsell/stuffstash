# iPad fixture run350465 — terminal result

Job104638923397 finished with58/71 cases passing and13 failing. Source
`b6321dcb4e5022aa13209894788c1c4c3e8f1229`, tested merge
`5a1e23d69395ff52f003e75d2fb61be3b7e6c472` into main438bd902.
The phone fixture job is still running at this inspection; do not restart it.
[Run](https://github.com/elsell/stuffstash/actions/runs/35046586497).

Authoritative job log: `/tmp/native350465-ipad.log`. Artifact10427689939 holds the
native screenshots, hierarchy exports and XCTest results. Later PR153 fixes are
not tested by this older source. A passing case is scoped to its assertions, not
the whole related screen or all24 axes.

## Normal-size failures, first priority

- Add draft in navigation stack.
- Add retained text and recovery after rejected Save.
- Add unfinished tag across details disclosure: expected Camping, observed Camp.
- Color-well target: accessibility frame36, required44. This is geometry evidence;
  actual expanded native hit-region behavior remains separately unproven.
- Controlled text without accessory: expected Native draft name, observed
  Nativdraft namee followed by a space.
- Ordinary controlled text: expected Native draft name, observed Nate draft nameiv.
- Home header target geometry: accessibility frame36, required44. The later
  separate hit-region diagnostic is not included in this71-case build.
- Place contents native search/navigation.
- Settings collection native search/Add.
- Voice proposal location search/return: fails at its Back assertion. The later
  explicit native Back fix still needs its own runtime result.

## Deferred enlarged-text failures

Edit metadata recovery, Edit tag disclosure, and Move-here recovery. These remain
failures, sequenced after normal-size remediation per the user's direction.

## Typing diagnostic changes the next investigation

Both new paced cases pass, asserting the complete string in the native field and
observed app value. Original whole-string assertions remain unchanged. In this
run ordinary uncontrolled, uncontrolled without accessory, multiline, and both
no-assistance cases also pass. Controlled default-assisted and controlled without
accessory still fail, and Add tag typing is truncated.

This establishes sensitivity to entry cadence in this run. It does not establish
whether XCTest, prediction/correction, React Native reconciliation or another
native interaction causes the failure. Earlier uncontrolled failures remain
counterevidence against treating controlled state as the only cause. Do not replace
the failing acceptance method with paced typing or disable user typing assistance
globally on the strength of these two diagnostic passes.

Home return details, direct footer layout, color open/clear, Browse's last-tag
footer, reminder-mode menu and settings Save/return pass their cases in this run.
Those results do not certify the current branch, phone equivalents, visual quality
outside the assertions, or physical-device behavior.

## Inspected search evidence and procedure correction

Retained `ipad-integrated-search-350465.png/.txt` show Tools filtered correctly,
Add visible, a toolbar SearchField and its Clear text accessory. There is no
Cancel/Close search control. `ipad-place-search-350465.png/.txt` show the same
pattern with query19 and only Tool19 matching. Both tests fail specifically when
demanding the phone's Cancel affordance, after their filtering assertions pass.

The candidate uses Clear text on iPad, dismisses any remaining keyboard through
the explicit accessory, and still verifies restored results and reachable actions.
Phone Cancel/collapse assertions remain. This changes the audit procedure, not
production search behavior; the13 failures above remain the actual run result.
Remote structural validation passes (`/tmp/ipad-search-audit-structural.log`) and
critic found no unjustified weakening. Swift compilation and the revised native
journeys remain pending behind the existing run.

Controlled default and paced screenshots were also visually inspected: both the
field and observer show the reordered string in the former and the full expected
string in the latter. Selected captures are in `/tmp/native350465-selected/`;
the complete1.27GB artifact is retained on paul at `/tmp/native350465-ipad.zip`.
