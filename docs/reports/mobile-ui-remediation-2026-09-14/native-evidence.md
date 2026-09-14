# Native simulator evidence

Run [34882515267](https://github.com/elsell/stuffstash/actions/runs/34882515267),
source a3fefb9761f3677d88b985186b13b8b378c984a0 (PR merge artifact records
merge revision e45599fc022bf89ce42367c25332f2b563807ba5).
Release simulator build on macOS 26 / Xcode 26.6; iPhone 17 and iPad mini (A17 Pro).
Both applications built and launched. Both tests failed the keyboard-dismissal
assertion. This is not a successful smoke run or an authenticated-app review.

The test opened and closed connection help successfully on both devices. Inspected
iPhone screenshots show the entry, help content, and sign-in action above the
visible keyboard without clipping at the default text size and light appearance.
This does not establish large-text, dark-mode, landscape, VoiceOver, or iPad layout quality.

- [Entry screenshot](evidence/onboarding-entry.png)
- [Help screenshot](evidence/onboarding-help.png)
- [Failed keyboard state](evidence/onboarding-failed-keyboard.png)

The initial dismissal check swiped upward over the whole application. iOS interactive
dismissal follows a downward drag from scroll content toward the keyboard; the
procedure is corrected in 23b17cc4 and awaits a new native result.

The entered address appears as `h.invalid` after `typeText("https://example.invalid")`.
M14 tracks this separately. The revised test asserts the full value; a failure
there leaves dismissal untested because XCTest stops after the first failure.
Do not infer a React input defect until reproduced against simulator typing behavior.

PNG files are unmodified screenshot payloads extracted from the XCTest result bundle.
Future runs export named attachments with xcresulttool directly on the macOS runner.


## Filter rerun 34897215957

Revision `254f5820c2c9c3c23ad4d2d8ca7b84f60a73c617` (PR merge of
`b0d7ce0c`). iPhone17 and iPad mini(A17 Pro), default text, light appearance.

- Browse visible labels, native availability selection, and Apply/Cancel passed
  on both devices. Inspected phone screenshot:
  [visible labels](evidence/browse-visible-labels-34897215957.png).
- Persistent actionable feedback passed on both devices.
- Expiration date-page and search-keyboard scenarios passed on iPad. Inspected
  [keyboard screenshot](evidence/expiration-ipad-keyboard-34897215957.png)
  shows both actions above the keyboard.
- Both expiration scenarios fail on iPhone before reaching their task: the
  body is absent. [Blank sheet](evidence/expiration-phone-blank-34897215957.png)
  and native hierarchy confirm the footer is present but ScrollView is absent.
  Removing KeyboardAvoidingView did not resolve M19; the prior causal hypothesis
  is rejected. A separate medium-detent diagnostic retains the failing full-height
  production fixture to avoid masking the issue.
- Onboarding full-string entry fails on both devices (`hmple.invalid` phone,
  `hs://example.invalid` tablet). Gesture dismissal is not reached in this run.
  Controlled/uncontrolled fixture comparison is pending; no production input
  workaround is justified yet.

These scenarios establish only their named interactions. They do not establish
whole-screen accessibility, larger text, dark mode, Android or physical-device behavior.


### Diagnostic run 34903318947

[Run](https://github.com/elsell/stuffstash/actions/runs/34903318947), source
`eeeefd5b`, completed with failures. All eight iPad fixture scenarios passed.
On iPhone, Browse filter selection/apply, persistent feedback, draft enum removal,
medium-detent expiration body and uncontrolled URL entry passed. The full-height
expiration body remained absent in both scenarios; controlled URL entry lost
characters (`hs://example.invalid`). This isolates a detent-sensitive sheet failure
and an input synchronization suspect; it does not prove a framework root cause.
The same controlled-input fixture passed on iPad, so the input problem is not
universal across devices/runs.

Production onboarding URL entry still failed on both devices; keyboard dismissal
was consequently not reached. The iPad landscape adaptation scenario passed.
No additional whole-surface or accessibility claims follow from these scenarios.

Inspected the [iPhone medium-detent screenshot](evidence/expiration-phone-medium-34903318947.png):
all six filter rows and the Apply/Cancel footer are visible. This is diagnostic
presentation evidence, not evidence of expansion or date/keyboard completion.

Remote source regression verification after native settings/input candidates:
1,353 tests in 242 files passed on paul, followed by TypeScript and the mobile
structural check. This includes mounted header tests replacing legacy module
mocks; code critic found no meaningful behavioral coverage lost. Native runtime
results remain separate from this source verification.

### Resizable-sheet run 34905451368

[Run](https://github.com/elsell/stuffstash/actions/runs/34905451368), source
`8bada3d9`, failed. Initial medium-detent content is visible, but expansion
removes the body on both devices. The iPhone expansion assertion confirms actual
upward movement before content reachability fails. The iPhone search keyboard
also removes the body and footer; iPad search action reachability passes.
See the [expanded sheet](evidence/expiration-expanded-empty-34905451368.png).
This rejects the detent configuration as a sufficient fix for M19.

The [calendar screenshot](evidence/expiration-calendar-open-34905451368.png)
shows the popover still open: the test's navigation-bar tap selected a date
inside it. That action-reachability failure does not independently establish
a product defect. Correct the dismissal target before evaluating it again.

Controlled and uncontrolled phone URL fixtures both lost initial characters;
iPad uncontrolled passed while controlled failed. Production onboarding also
lost characters on both devices. This weakens the earlier synchronization
hypothesis and does not establish a cause. Tests currently begin typing before
explicitly verifying keyboard readiness. Full-string assertions remain required.

Browse menu/apply, feedback retention, and draft-option removal passed on both
devices. These results certify those scenarios only, not whole surfaces.

Hierarchy inspection of run34905451368 distinguishes the expanded failure from
ordinary scrolling: the initial sheet contains its ScrollView and Choose tags
button, while the expanded sheet subtree no longer exposes a ScrollView or its
rows. The footer still has a valid frame. This does not establish why the view
vanished. The pinned screens implementation coerces form-sheet scroll frames by
searching direct children (or its own safe-area wrapper); its footer option is
explicitly Android-only. A nested-body versus direct-scroll diagnostic is needed
before adopting another production layout workaround.

Run34906713382: the iPhone onboarding job104186247263 terminated with an Xcode
application-launch timeout before its keyboard scenario could execute. This is
an infrastructure/launch failure, not confirmation or rejection of M14. The run's
other jobs were still active when this job log was inspected.

### Run 34906713382 completed

[Run](https://github.com/elsell/stuffstash/actions/runs/34906713382), source
73ff5bad, failed. Appearance menu selection, compact expiration date entry,
Browse menu/apply, draft option removal and feedback passed on both devices.
The full expiration expansion failure persists on both, with phone keyboard
footer failure and the already identified calendar dismissal test mistake.

The [color picker screenshot](evidence/color-picker-open-34906713382.png) proves
it opened directly. The test incorrectly requested `Close`; the native hierarchy
labels it `close`. Close/clear checks still need completion after correction.

Uncontrolled input preserved the full URL on both devices this run; controlled
phone input still lost initial characters. The onboarding submission fixture
passed on iPad. On phone it preserved the full displayed URL, but its submitted
value remained `none`. The [screenshot](evidence/onboarding-keyboard-action-34906713382.png)
shows most of Connect behind the keyboard despite XCTest reporting hittability.
The command test will explicitly dismiss the keyboard; keyboard layout remains
unresolved. Production phone onboarding failed at app launch; iPad production
stopped at the help-collapse assertion before typing, so neither verifies M14.

Add failed before Asset name appeared on both devices. Phone log checks for an
unexpected app termination and lacks the usual final screenshot/hierarchy. The
recording ends around Add activation. This needs crash diagnostics; it does not
prove a form layout cause. The queued keyboard-readiness run34907820613 has now
started. Later source fixes and the isolated sheet diagnostic were pushed only
after it started, preserving that run.

### Run 34907820613: partial results and keyboard procedure correction

[Run](https://github.com/elsell/stuffstash/actions/runs/34907820613), source
a01407dc (merge cac89c8f), still had its phone fixture job running when reviewed.
Both onboarding jobs failed before typing at the new keyboard-readiness condition.
The inspected phone hierarchy exposes a zero-size `Padding-Left` Key before real
letter keys. The [screenshot](evidence/onboarding-keyboard-ready-34907820613.png)
shows a visible keyboard and focused field. Therefore the first-key hittability
assumption was invalid. Both test helpers now require any hittable key; exact
full displayed/submitted text assertions remain. Critic found no blocker;
compilation and native execution of this correction remain pending.

The completed iPad fixture job passed appearance menu selection, Browse
availability/apply, draft option removal, compact expiration dates, actionable
feedback, and the corrected date-page calendar dismissal/action scenario.
Expiration expansion still lost its body. Add still failed before Asset name
appeared. Color still hit the previously identified uppercase Close selector
mistake (the correction was not in this run). Keyboard scenarios stopped at the
readiness test mistake and provide no new evidence about typing preservation or
keyboard/footer layout. Production iPad landscape passed.
