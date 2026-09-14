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
