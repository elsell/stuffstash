# Phone native audit — run35104157358

Source14d7e06f, merge767d677340b3144968c155bb0f1a4b84d0548e1f;
iPhone17 job104821415261. Complete XCTest log:63/84 pass,21 failures,
3390.513 seconds; job failure. Onboarding passes separately (one applicable test,
two device-inapplicable skips). Artifact10453232053 is retained on paul as
`/tmp/native351041-phone.zip`; selected captures are `/tmp/phone351041-selected`
on both hosts. Complete log: `/tmp/native351041-phone-complete.log` locally.

## Normal-size product journeys

- Add form-sheet draft: immediate name equality reads `Nat`. Screenshot0.33 seconds
  later and hierarchy0.82 seconds later both show exact `Native draft name`.
  [Capture](evidence/phone351041-add-name-after-assertion.png). Candidate acceptance
  now waits at most five seconds for exact equality before saving; save/recovery
  assertions remain. This is evidence of observation timing for this native input,
  not permission to accept corrupted controlled fields.
- Add-photo paging/removal: all photo assertions and return from Add pass; the
  final root-menu entry is outside the retained scroll viewport. Capture shows
  Native UI audit at its previous scroll position. Candidate now requires the root
  navigation bar, bounded reveal and entry hittability. Full rerun remains required.
- Move Here: fails old StaticText candidate lookup. Newer88fe7499 uses the observed
  Button; this run cannot validate that correction.
- Preconfigured Place search: expected top search is unavailable. Production Place
  and managed/static comparisons pass in this run; do not generalize that success
  to the failed variant or other runs.
- Settings collection: expected top search fails; retained hierarchy exposes the
  Search tags field at the bottom. Layout expectation remains unresolved.
- Sharing: exact email entry and creation/link-unavailable recovery pass. Keyboard
  absence passes before the invitation menu tap. Cancel invitation then becomes
  unhittable and keyboard is visible in final capture. M193 remains unresolved;
  this is not simply missing keyboard dismissal at creation. Menu trigger bounds
  are44×44, narrowing M194's size concern without proving menu usability.
  [Failure capture](evidence/phone351041-sharing-cancel-failure.png).

## Diagnostic and deferred failures

Controlled address, controlled name without accessory, ordinary controlled name,
ordinary multiline keyboard readiness and seeded uncontrolled address fail.
The seeded address remains wrong in the final hierarchy (`hps://example.inva…`),
unlike the delayed Add name. Keep these distinctions.

Footer-full-sheet and nested-full-sheet are retained nonproduction layout
comparisons. Seven enlarged-text tests fail: asset-region recovery, command height,
Details commands, Edit metadata, Edit tags, footer appearance and Move Here.
Expiration overview's accessibility audit reports Text clipped; retain its
separate report and identify the affected size/element before labeling a
normal-size production issue. Normal-size findings remain the priority.

The native run is failed and release remains gated. Test-observation corrections
pass six remote fixture-preparation tests and structural checks; critic finds no
blocker. They must pass native execution and do not retroactively change63/84.
