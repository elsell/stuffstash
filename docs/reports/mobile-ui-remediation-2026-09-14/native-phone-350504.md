# Phone native run 35050407693 — terminal log evidence

Source802e4955, iPhone17 job104649681763:50/72 fixture cases passed,22 failed.
The iPad fixture job104649681823 remains active at this observation. Both
onboarding jobs passed their applicable cases. Local retained terminal log:
`/tmp/native350504-phone.log`. Sharing screenshots/recording were inspected in the
follow-up below; other new captures remain unreviewed.
This source excludes the later static search comparison, onboarding command
migration, photo-removal migration and M209/M210 corrections.

Important changes in the evidence:

- Voice destination search/retry/return passes with the explicit Back candidate.
  This is a named journey pass, not whole-voice acceptance.
- Place contents search passes this time; settings collection search still fails.
  The earlier bottom-field captures remain valid for their run. Do not assume a
  deterministic universal search failure or claim the placement defect fixed.
- The Home notification edge probe passes all nine center/edge/corner taps around
  a centered44-point square, checking one activation per tap. The old header test
  still fails on an AX height of36. The AX rectangle therefore does not establish
  the notification's actual hit region. Add/profile and the full scrolling-header
  journey need their own evidence before changing those acceptance assertions.
- Both Home Return cancellation and failed-save recovery pass again.
- Sharing still fails while trying to activate Cancel invitation: the logged menu
  button is42 points high and not hittable. Screenshot inspection is required to
  distinguish overlap, keyboard state and menu presentation; no root cause is
  asserted from the log alone.
- Ordinary controlled typing and Add draft cases still lose characters; paced
  controlled/uncontrolled diagnostics pass. Do not relax complete-text assertions.
- Phone color-picker activation still fails, and its AX-height assertion reports28.
  No full hit-region conclusion follows from that geometry alone.

Other failures include the existing enlarged-text journeys, Expiration text
clipping and retained Footer/NestedFullSheet comparison cases. Normal-size
production findings remain the priority. This run is not release acceptance.

## Sharing recording and hierarchy inspected

Artifact10429678833 is retained on paul at `/tmp/native350504-phone.zip`.
The Sharing recording and selected attachments were extracted to
`/tmp/sharing350504` there; only selected images/text were copied locally.
The [pre-menu capture](sharing-before-menu-350504.png) shows the complete failure
message, native Create command and44-point invitation-menu trigger with keyboard
absent. The [terminal capture](sharing-keyboard-menu-350504.png) and
[hierarchy](sharing-keyboard-menu-350504.txt) show the keyboard after opening that
menu. The menu action occupies y614.7–656.7 while the keyboard starts at y583;
the action is obscured. Recording frames27–33 seconds confirm that transition.
The existing pre-menu keyboard-absence assertion passed, so this is not simply
the earlier missing dismissal on submission (M193).

Upstream leads, not established local root cause:

- [Expo47343](https://github.com/expo/expo/issues/47343) reports Menu/Picker reopening
  the keyboard after input focus. Its [closing comment](https://github.com/expo/expo/issues/47343#issuecomment-4840735652)
  says the reporter could reproduce only on Simulator, not physical hardware.
- [Expo43198](https://github.com/expo/expo/issues/43198) also describes menu-triggered
  keyboard reopening. Its proposed search-controller explanation is insufficient
  for this input-only Sharing screen.
- [Screens maintainer reproduction](https://github.com/software-mansion/react-native-screens/issues/3677#issuecomment-4067063758)
  reproduced similar behavior without react-native-screens and suspected React
  Native focus handling. This does not prove a fix for our pinned stack.

Retain the original recovery assertion. Do not add a test-only keyboard dismissal
after opening the menu, switch to custom controls, or apply the unrelated search
patch based on this evidence. Next verification needs to isolate native menu/input
focus on the pinned stack and compare physical-device behavior. M194/menu recovery
remains open; the44-point trigger correction itself is visible in the capture.
