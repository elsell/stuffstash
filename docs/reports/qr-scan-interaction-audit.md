# QR scanning interaction audit

Reviewed 2026-10-13 against main `a561ab8be`. Scope: mobile Browse scan entry,
camera/manual entry, externally opened labels, sign-in return, resolving and
recovery, normal asset navigation, and browser landing. This is an interaction
review plus source evidence, not a claim that physical scanning was tested.

## Reference patterns

| Reference | Documented behavior | Application to Stuff Stash |
| --- | --- | --- |
| [Apple Camera and Code Scanner](https://support.apple.com/en-us/102680) | Aim the camera and follow the detected link. Code Scanner also provides lighting. | Capture should dominate; entering a URL is a fallback rather than a competing form. An external link tap already expresses navigation intent. |
| [Pixel Camera](https://support.google.com/pixelphone/answer/16561572) | Detect a QR code through Camera and follow its link. | Preserve the same short capture-to-destination path on Android. |
| [1Password](https://support.1password.com/one-time-passwords/) | Scan in context, with manual entry when scanning is unavailable; iOS Camera can hand verification codes into the app. | Separate capture and manual entry. Continue the intended task after external handoff rather than asking the user to scan again. Its credential-save confirmation is appropriate for mutation, not a precedent for confirming every asset read. |
| [Apple HIG: Modality](https://developer.apple.com/design/human-interface-guidelines/modality) | Keep modal tasks focused and short; provide an obvious exit. Camera views may justify an immersive presentation. | Retain one bounded scanner task and native dismissal. Asset details belong in normal navigation, not inside another scanner screen. |

These are documented interaction examples, not screenshots inspected on current
versions of those apps. The decision to auto-open only on known safe root routes
is a project safeguard for unsaved drafts, not an Apple requirement. Apple's
universal-link developer page required JavaScript and its Markdown endpoint could
not be retrieved; no additional HIG claims rely on that inaccessible page.

## Findings and corrections

| ID / priority | Before (source-confirmed) | User impact | Correction / acceptance |
| --- | --- | --- | --- |
| QR01 / high | `LabelScanScreen` combined a fixed 280-point preview, Use camera, permanent URL form, and two possible Open label buttons. | The task resembled configuration; controls competed with aiming. | Flexible preview with aiming guidance; compact native toolbar for flashlight and Enter label link. Manual mode has one labeled field and one submit. Denial exposes that fallback. |
| QR02 / high | `PendingLabelNavigation` offered Open label, then `LabelScanScreen` required Open label again. | Redundant confirmation after an explicit external tap. | Safe Home/Browse entry opens the resolving task directly; pending identity resolves once. Other routes retain a single offer so editing is not replaced. |
| QR03 / medium | All errors exposed Change server and Switch account. | Invalid input and camera denial suggested unrelated, disruptive recovery. | Syntax: retry camera/manual entry. Camera: Settings in the toolbar/manual entry. Instance mismatch: Change server. Lookup failure: one primary Retry plus a contextual account toolbar action. |
| QR04 / medium | Ready-to-open notice also appeared for invalid links and repeated on pathname changes. | Misleading status and repeated interruptions. | Honest invalid copy; one handling decision per incoming revision. |
| QR05 / high | Cold native intent opened `/scan-label` before asynchronous initial-link lookup settled. | Camera permission could precede resolving an already-scanned link. | Mark external entry and wait for initial-link readiness before mounting scanner. Never request camera merely to open an external label. |

Source locations: `apps/mobile/src/ui/labels/LabelScanScreen.tsx`,
`PendingLabelNavigation.tsx`, `useLabelLinkSnapshot.ts`,
`apps/mobile/src/app/scan-label.tsx`, `+native-intent.tsx`, and
`apps/mobile/src/adapters/labels/ExpoLabelCamera.tsx`.

## Coverage and retained behavior

- **Entry and hierarchy:** Browse List and Map share the scanner route. No new
  tab or general barcode-product lookup is added. Native Cancel remains available.
- **Capture:** focused foreground capture only; one permission request per camera
  attempt; repeated frames resolve once. Torch is off on a new attempt. Native toolbar placement keeps commands out of the camera/form content; every state has at most one content action. A simulator
  cannot establish physical decoding, focus quality, lighting or torch behavior.
- **Manual input:** explicit mode except camera failure; URL keyboard, visible
  label, native paste, submit key and submit button. No clipboard surveillance.
- **Loading and recovery:** one progress state, actionable errors, retained link
  for retry, cancellation and rejection of stale results. Backgrounding stops
  capture and aborts lookup; no late asset navigation.
- **External entry:** cold/warm link, sign-in readiness, invalid or superseded
  link, safe root vs potentially editing route. Server/account transitions retain
  pending identity before lookup, including automatic session expiry. Automatic opening must not run twice for an in-task capture.
- **Security:** existing parser rejects unsupported input; `OpenLabel` verifies
  connected instance before resolving and selecting an authorized scope. Printed
  origins never become credential-bearing network destinations. Existing boundary
  tests remain required; this redesign does not add a new endpoint or authority.
- **Asset destination:** scanner replacement uses existing detail navigation.
  Archived targets remain governed by existing detail behavior; no unsupported
  claim is made that archived navigation was broken.
- **Browser handoff:** `LabelLanding.svelte` already resolves authenticated links,
  preserves sign-in entry, and offers validated Open in app. Retained. Actual OS
  association depends on deployment and needs a real-device check. Browser-local
  `LabelScanner.svelte` still exposes manual input beside camera; this is a tracked
  web consistency follow-up, outside this mobile implementation batch.
- **Appearance/adaptation:** shared colors, spacing and native command controls;
  flexible preview, wrapping actions, scroll/keyboard safe area handling. Runtime
  phone/tablet appearance, VoiceOver/TalkBack and enlarged-text checks remain
  distinct from source assertions. Normal-size task structure is the priority.

The retained October 4 phone denial screenshot was also inspected: it shows
Use camera, Open label, Change server and Switch account stacked below an error.
This historical image corroborates the old recovery hierarchy; it is not evidence
of the new build.

## Verification

The focused local run passes 45 tests across 11 suites. Mobile TypeScript and
structural checks pass. Code-critic review has no remaining source findings.
Critical automated checks cover direct opening without camera permission,
manual fallback, failure-specific actions, repeated-frame suppression, stale
lookup cancellation, initial-link races, draft-safe external entry and existing
parser/authorization boundaries. The existing `label-scanner` native audit is
updated to expect automatic sign-in return. Retain one named phone/iPad run and
record its terminal result; do not repeat unchanged runner failures.

Physical printed-label scans, real Camera-to-app association and visual acceptance
are consolidated in `docs/reports/user-testing-checklist.md`. Missing user checks
do not block delivery, but they are never reported as passed.

### Named native evidence

[Run 38339579818](https://github.com/elsell/stuffstash/actions/runs/38339579818),
revision `ff806568`, passed four scanner journeys each on iPhone 17 and iPad mini.
The screenshots were inspected: denial/manual entry, lookup retry and foreign
instance recovery show contextual native toolbar actions and a single content
action without overlap. The iPad manual form stretched across the window; a
follow-up caps non-camera content at 640 points. This width refinement requires
its own screenshot confirmation and does not invalidate the passed behaviors.
The fixture cannot exercise a physical camera preview or torch.
