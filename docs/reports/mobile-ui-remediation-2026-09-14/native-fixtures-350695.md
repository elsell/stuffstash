# Native fixture run35069581807

Source211bdc2bcd7b4e3fa70333791417ae70ef936c16; tested merge
df3486fed51bc7c0044eadca6ac06a87649d8c81. Phone job104708278597
finished with59/81 cases passing and22 failures. iPad job104708278267 subsequently
finished with71/81 passing and10 failures. Both onboarding jobs passed.
The following selected screenshots and hierarchies were subsequently inspected.

## Evidence that changes the next action

- Add's complete unfinished-tag disclosure journey passes on phone with the
  exact-value observation wait. The deliberately hidden-header Add comparison
  still fails; normal navigation-stack and preconfigured-header cases pass.
- Home header scrolling and actual tab-shell return pass, as do all three
  nine-point action probes. The later iPad tab-strip selector is not in this run.
- The color control's nine delivered-touch probes pass, as do opening/clearing
  and its single accessible name. Its separate accessibility-frame check still
  fails at28pt. Probe success does not establish VoiceOver or every touch point.
- Both production Place search and its preconfigured-header comparison fail
  waiting for the Search button. Preconfiguration alone has not resolved M207.
  Settings collection search passes in this run.
- Sharing retains the complete email and gets past keyboard-absence and
 44-point trigger assertions. Cancellation then fails because the menu's
  `Cancel invitation` command is not hittable. Earlier typing failures remain
  unresolved evidence; this one pass does not establish a universal input fix.
- Conversation fails the new context-below-header check: context minY187.46,
  header maxY243.15. M216's larger initial detent and safe-area candidate is
  not accepted. The final capture also shows this overlap, as detailed below.
- Account fails waiting for its navigation header, earlier than the previous
  notice-geometry failure. Sheet notice geometry also fails in this run.

## Remaining failure classes

The22 failed cases comprise Account entry; hidden-header Add; eight enlarged-text
cases (asset region, command height, Details commands, Edit metadata, Edit tags,
Expiration overview, footer, Move); color accessibility bounds; four text-entry
comparisons (controlled address, controlled without accessory, ordinary controlled,
ordinary single-line); two intentionally wrapped sheet comparisons; sheet notice
geometry; both Place search variants; Sharing cancellation; and Conversation
context geometry. Wrapped comparison failures are not automatically production
filter defects; see the classification in `native-fixtures-350633.md`.

Normal-size failures remain first priority. This failed suite is neither native
acceptance nor a release approval. Phone artifact10438472417 retains screenshots,
hierarchies and the result bundle. Full log: `/tmp/native350695-phone.log`.

## Inspected phone captures

- [Conversation](evidence/phone-voice-header-overlap-350695.png): inventory context
  is above/behind the title bar, and the first message extends beneath it. Final
  hierarchy retains context y187.5 and header y191.3–243.2. This persists in the
  failure capture; it is not explained away by the initial observation alone.
- [Sharing](evidence/phone-sharing-keyboard-return-350695.png): the keyboard is
  visible again after the earlier absence assertion. Its hierarchy starts at
  y583; the cancellation command is y614.7–656.7. The full email remains intact.
  M193 is still open: a one-time dismissal before submission has not ensured
  cancellation remains usable. The capture alone does not identify who refocused
  editing or why.
- [Preconfigured Place](evidence/phone-preconfigured-bottom-search-350695.png):
  search is a bottom field, not the requested header button. The hierarchy places
  it at(33,803),336×38. This reproduces the placement problem despite configuring
  the header before presentation; retain that comparison as negative evidence.

## Conversation follow-up candidate

The iOS body now reserves `useHeaderHeight()` rather than relying on top safe-area
padding, retaining side safe areas and the existing footer ownership. Android
retains top safe-area handling. The existing failing native context assertion is
unchanged and must rerun; this is not accepted from source alone. Nineteen focused
tests, TypeScript and mobile structural checks pass on paul after correcting an
initial hook-scope error caught by TypeScript and review. Validation log:
`/tmp/voice-header-reservation-validation.log`.

## Terminal iPad results

The iPad log (`/tmp/native350695-ipad.log`) confirms the Add unfinished-tag journey,
nine color touch probes, Home header scrolling, Sharing cancellation/recovery and
both Place search variants pass. Phone failures for those latter two surfaces must
not be generalized to iPad. Both targets still fail Conversation context clearance:
iPad context minY326.5 versus header maxY378.5. This source predates the measured
header-reservation follow-up2978fe7b.

The ten iPad failures are hidden-header Add, color accessibility bounds36pt,
controlled address, controlled text without accessory, ordinary controlled text,
three enlarged-text journeys (Edit metadata, Edit tags, Move), Home tab-shell
selection, and Conversation context clearance. The Home tab-shell failure uses
the old TabBar selector; its platform-aware correction43e585ef is not included.

Phone Sharing's trace checks keyboard absence at33.98 seconds, taps the menu
trigger at34.14 seconds and then finds its command unhittable. The final screenshot
shows the returning keyboard. Inspection of the app menu adapter finds no explicit
focus call; the native cause is unresolved. Do not replace the menu or suppress
typing assistance solely to evade this failure.

## Notice placement observation gap

The phone sheet notice's five-second geometry assertion failed. Its inspected
final screenshot and hierarchy nevertheless show header maxY132, content
rectangle(0,62,402,812), notice(16,142,370,93), dismiss control(33,153,147.3,71)
and action(190.3,164.7,178.7,48), all within the asserted final bounds.
Capture: [notice sheet](evidence/phone-notice-sheet-350695.png).
This does not establish what the predicate saw earlier or retroactively pass the
journey. Add last-sample geometry diagnostics to the unchanged containment gate;
sample shared bounds once per predicate evaluation and record all three results
without short-circuiting. Native rerun remains required; no product correction is
claimed from the later capture.
