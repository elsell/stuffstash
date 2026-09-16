# Phone native run 35050407693 — terminal log evidence

Source802e4955, iPhone17 job104649681763:50/72 fixture cases passed,22 failed.
The iPad fixture job104649681823 remains active at this observation. Both
onboarding jobs passed their applicable cases. Local retained terminal log:
`/tmp/native350504-phone.log`. New screenshots have not yet been inspected.
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
