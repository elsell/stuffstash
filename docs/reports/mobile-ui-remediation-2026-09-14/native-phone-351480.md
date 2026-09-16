# Full phone results — run35148054814

Source a01fc760, tested merge acd67fc086b563f8e9a6e83fb4acd35c180ec8a9.
Phone job104969140649 completes69/86 cases, with17 failures in3585 seconds.
Artifact10470643862 is retained only on paul at `/tmp/native351480-full-phone.zip`;
the local complete log is `/tmp/native351480-phone-complete.log`.
The full iPad job remains active at this inspection. Both onboarding jobs pass.

Normal-size failures retain ordinary color activation, controlled text/address
loss and preconfigured Place search. The two full-sheet comparison layouts and
eight enlarged-text/accessibility cases also fail. Their scope remains distinct
from normal-size shipped composition; enlarged-text work follows normal-size work.

Two other failures require careful interpretation:

- Add's unfinished tag survives collapse/reopen, stages successfully, and clears
  the field. The subsequent immediate assertion rejects an absent AX value as
  `missing`, even though the field exists and is empty. The inspected
  [screenshot](evidence/phone-add-staged-351480.png) and
  [hierarchy](evidence/phone-add-staged-351480.txt) show Camping staged, New tag name
  present with placeholder New tag and no value, and Save available. Native
  acceptance now waits for all three conditions: staged tag, existing cleared
  field, enabled Save. The remaining Clear draft/Close sequence still needs rerun;
  this capture does not establish completion of the whole case.
- Settings dirty-editor protection stops at keyboard readiness before typing.
  Its final capture shows Name focused and a visible keyboard. This does not prove
  a draft-protection regression. The separate focused351529 readiness failure
  exposes an invalid key frame; the candidate excludes invalid frames before
  asking XCTest for hittability, retaining the bounded wait and typing checks.
  It may not resolve every readiness timeout.

The search ownership and settled-keyboard inset candidates M250/M249 postdate this
build and cannot receive acceptance from it. The focused filters run35154627907
is active on both targets. See [later input traces](native-text-entry-351529.md)
for key delivery and text-change evidence that narrows the input investigation.
