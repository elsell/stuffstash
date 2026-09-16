# Home native evidence — September 16

Completed run [35140471580](https://github.com/elsell/stuffstash/actions/runs/35140471580),
source a1b827e0, tested merge a14a86aaa33e5a8169feab902988bea15a9bac15.
Phone job104943442022 and iPad job104943442379 each pass all seven Home scenarios:
header position/action order while scrolling; Add, notification and Profile touch
probes; Return cancellation; failed details save followed by retry; and tab return
with the voice accessory retained. These are fixture-scoped results, not overall
suite success (phone69/86, iPad79/86).

The final header and tab-return captures from both targets were inspected. All
three actions remain visible in Add/Notifications/Profile order. The inventory
selector has inset text and disclosure spacing, with its long name truncated.
The tab composition shows content passing under the native soft header edge, with
controls remaining legible. The phone keeps bottom tabs; iPad presents its native
tab strip at the top. The voice accessory remains visible after tab return.
No refresh spinner is visible in these final states; this is not a new test of
slow network refresh or every navigation interruption.

| Target | Header after scrolling | Tab return composition |
| --- | --- | --- |
| iPhone17 | [Capture](evidence/phone-home-header-351404.png) | [Capture](evidence/phone-home-tabs-351404.png) |
| iPad mini | [Capture](evidence/ipad-home-header-351404.png) | [Capture](evidence/ipad-home-tabs-351404.png) |

Each independent header action probe tests the center and eight edge/corner points
21pt from it. This verifies delivered taps at those points, not a continuous-area
measurement. Add/Profile destinations are fixture markers reached by production
header handlers; their full destination workflows are separate. The tab scenario
uses production tab layouts with synthetic Home data and a Browse placeholder.

The Return recovery test verifies full native note text, retained error state,
retry, dismissal and return to Home. Its synthetic repository does not assert the
submitted note payload; transport payload coverage is not inferred from this test.
See [Return review](home-return-axis.md) for ownership and payload-evidence limits.

Android Home runtime evidence also exists: [native icons](android-native-icons.md)
verifies visible actions and Add/Profile marker navigation; [dark appearance](android-compose-appearance.md)
verifies the header and counted notification activation. Earlier statements that
Android runtime was unavailable describe the source-review checkpoint, not the
current audit. These Android samples do not establish complete tab-shell behavior.

Remaining coverage includes physical-device behavior, assistive-technology output,
RTL, reduced transparency/motion, adverse loading/access transitions and broader
deep-link/background recovery. Enlarged-text work remains behind normal-size
remediation. These results do not close unrelated input/color/search failures or
clear the TestFlight gate.

Complete logs are `/tmp/native351404-phone-complete.log` and
`/tmp/native351404-ipad-complete.log`. GitHub artifacts10466544534 and10468360267
retain result bundles. Small selected raw Home attachments remain under
`/tmp/home351404-phone` and `/tmp/home351404-ipad` on both hosts; the four captures
above are retained in the repository without copying the large result archives.
