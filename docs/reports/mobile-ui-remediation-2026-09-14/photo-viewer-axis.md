# Photo viewer source review

## Android pan and gesture dismissal follow-up

On September 16, the retained APK SHA-256
`d9b1a7b5196c970f45131b6d960adea3ae02bc46ea5f4d772dff39ec8a00bb72`
ran on paul's Pixel 6/API 36 emulator at normal font scale. In the synthetic asset
photo recovery fixture, double-tap enlarged the image; a diagonal drag changed its
visible crop while the footer remained available. A second double-tap restored the
fitted image. Captures `/tmp/asset-pan-{before,after,fit}.png` on both hosts record
these states.

Two downward gestures and a shorter upward gesture did not dismiss the viewer.
A longer upward gesture from (540,1900) to (540,200), over 250 ms, did dismiss it.
The resulting hierarchy positively showed Back to audit menu, Removal attempts: 0,
and Photos remaining: 1. Evidence: `/tmp/asset-pan-swipe.png` on both hosts and
`/tmp/asset-pan-close.xml` on paul. This establishes one successful upward dismissal,
not bidirectional or short-swipe acceptance. Pinch, assistive operation and iOS
gesture acceptance remain open. No live media or removal service was used.

The original source review below was made at82669f4b, including the installed,
pinned react-native-image-viewing0.2.2 implementation, before the photo recovery
fixture ran. Subsequent sections record fixes and Android runtime evidence;
the original table is not the current verification status.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Dedicated full-screen inspection is appropriate for image detail/zoom; removal requires confirmation. |
| Navigation/modality | Explicit Close plus system request-close/swipe dismissal; persisted photo identity maps to the index. Native close/alert layering pending. |
| Layout/adaptation | Footer takes safe-bottom inset, neutral canvas, four46×54 targets. Dependency restricts modal to portrait and captures screen dimensions at module load. Phone/tablet/large-text/rotation still need verification; app config also says portrait, so no new orientation regression is inferred. |
| Typography/localization | Filename and metadata are limited to one line; numbered position uses English “of”. Long-name, large-text and locale acceptance remains pending. |
| Appearance/imagery | Explicit neutral black canvas, white labels, amber destructive icon; gallery excludes bitmap color inversion. Third-party viewer image inversion and contrast are not verified. |
| Targets/gestures | Footer supplies Close, Previous, Next and Remove as buttons; previous/next disable at boundaries. Pinch/double tap zoom and swipe-close belong to the dependency. Assistive zoom and toolbar visibility while zoomed are not established. |
| Keyboard/accessibility | No text entry in viewer. Buttons have labels/disabled states. Focus placement/return, bitmap description, hardware keyboard and TalkBack still need runtime inspection. |
| Motion | Wrapper uses fade; dependency translates header/footer by300points with200ms timers and no Reduce Motion preference. M84 tracks this source gap. |
| Content/search | Indexed photo sequence with explicit alternatives to swipe. No viewer-specific search; this does not make app search/navigation coverage complete. |
| Loading/recovery | Dependency waits for image dimensions/loading; failed dimensions become0×0, and rendered image lacks onError. No error message or retry replaces loading. M85 tracks this. M83 covers removal failure dialog separately. |
| Editing | Confirmed removal freezes repeat Remove, preserves failed photo and re-enables retry; operation owner suppresses late route completion. Remote tests cover these behaviors, not native alert placement. |
| Privacy/media | Viewer receives authorized URI/header data from its owning screen and forwards headers. No permission request is made merely to view a photo. This is not a security-boundary test or proof of token safety across redirects. |
| Lifecycle/notifications | Route teardown suppresses removal failure, but process restore, app-switch interruption, session changes during image requests and notification entry while viewing remain pending. |

Relevant source: FullScreenPhotoViewer.tsx, AssetPhotoViewerSheet.tsx,
AssetPhotoWorkspacePresentation.ts, AssetDetailPhotoGallery.tsx,
AssetDetailRouteScreen.tsx; dependency dist/ImageViewing.js,
hooks/useAnimatedComponents.js, hooks/useImageDimensions.js and
components/ImageItem/ImageItem.ios.js/ImageItem.android.js.

Repair must address the dependency behavior, not merely change wrapper props that
the dependency ignores. Choose a reviewed pinned patch or replacement through the
spec; preserve image headers, explicit controls, zoom and platform behavior. Add
failed-image and Reduce Motion native scenarios before claiming acceptance.

The unavailable-photo runner scenario now uses a missing bundled-file sibling and
asserts a readable failure, reachable Retry, and reachable Close. It is an
acceptance test awaiting the M85 repair and native execution, not a passing result.
The scenario deliberately keeps failing after Retry; separate controlled tests
must establish fresh requests and successful recovery. Structural validation and
the two fixture-preparation tests passed on the remote Linux validation host.

M85 dimension lifecycle candidate: the pinned adapter now clears old dimensions,
accepts an explicit retry generation, forwards the original headers, and ignores
obsolete or duplicate completions. Its URI-only JavaScript cache was removed.
Two regression tests failed against the previous dependency and pass against the
patch on the remote validation host. This does not yet add visible Retry/error
controls; M85 remains open until image-decode handling and native acceptance pass.

The subsequent M85 candidate adds the visible failure renderer through the pinned
dependency's optional ErrorComponent. Stuff Stash supplies its native command-button
adapter; no extra route or modal is introduced. The shared loading hook makes
errors terminal until Retry and rejects obsolete events. The active failed photo
restores the viewer chrome. Review caught inline projections in both callers that
defeated memoization; those are now stable, with a real AssetPhotoViewerSheet
regression failing before the caller fix and passing afterward. Eleven focused
checks, eighteen Add checks, TypeScript and structural validation passed remotely.
An iOS Metro export validates bundling, not layout or native behavior.

Current follow-up at8a256a2a rechecked the wrapper and gallery consumer across the
24 axes above. M85 now renders Photo unavailable and a native Retry command;
explicit Close remains available, current image headers are preserved, and Remove
stays disabled while pending. Filename truncation, VoiceOver focus, zoom and
background interruptions remain runtime acceptance. This completes the remaining
source cells for S100 without closing native findings. Gallery preview failures
are separately tracked as M211; the full-viewer repair did not cover that surface.

## M233 Android recovery command visibility

On Android16 Pixel6, normal text/light, APKd3fc5d55763fc45cb8f6e5c93cd116440570d1613b0646c884c2fad636e0bd32,
the [unavailable-photo capture](evidence/android-photo-commands-before.png) exposes
dark Retry text on the black canvas and a duplicate top Close overlapping status
icons. Retrying the deliberately missing resource remains unavailable; the footer
Close returns to the fixture with Photos remaining:1 and Removal attempts:0.
The shared candidate uses native filled Retry and omits the library header, keeping
the existing safe-area footer Close and system/swipe dismissal. Both asset and draft
consumers pass31 remote behavior/presentation tests; TypeScript and structural
checks pass. Critic found no source blocker. Rebuilt native verification is pending.

Baseline: `/tmp/android-photo-{unavailable,retry,closed}.xml` and
`/tmp/android-photo-retry.png`. The fixture has no live media or deletion service.

Rebuilt APK426c83b838cb99e80835aa10309c10514c7165d76c09554e814dca7b186c9fa9
shows [filled Retry and no overlapping header Close](evidence/android-photo-commands-after.png).
Retry leaves the deliberate missing-image error recoverable; the remaining Close
positively returns with Photos remaining:1, Removal attempts:0 and Back to audit
menu. Evidence: `/tmp/android-photo-fixed{,-retry,-closed}.xml` and
`/tmp/android-photo-fixed.png`. This confirms the Android unavailable-asset sample;
iOS, draft-photo native acceptance, zoom chrome restoration and TalkBack remain open.

### Asset removal consumer follow-up

On the same426c83b8 Android candidate, Remove opens the native confirmation. Accepting
produces the fixture's delayed rejection and [native acknowledgment above the
viewer](evidence/android-photo-remove-recovery.png). OK returns to the viewer;
footer Close positively returns with Removal attempts:1 and Photos remaining:1.
This confirms reachable error recovery and retained media in the asset consumer,
not real deletion or pending-interaction protection. Evidence:
`/tmp/android-photo-remov{al-entry,e-confirm,e-error,e-retained}.xml` and
`/tmp/android-photo-remove-error.png`.

The existing Audit draft photos fixture renders VoicePlanPhotoDraftStrip and does
not open DraftPhotoPreviewModal. It must not be counted as native acceptance for
that full-screen consumer. At the initial review, AddAssetFixture returned no selected photos. The fixture
extension and actual Add preview acceptance below close that sampling gap on
Android. Its 8 mounted draft-preview cases remain distinct.

### Actual Add preview consumer

The Add fixture now returns two distinct IDs/filenames backed by the bundled glyph
from its library port once per fixture lifetime; subsequent selection and camera
return empty. It exercises the real Add chooser/thumbnail/preview composition.
TypeScript and structural checks pass on paul; critic found no blocker. No real
library, provider or upload is used.

Android16 Pixel6, normal text/light, APK
`5660fb54a10d43f04f9be57288832bf8532f558ca4a07186429ff6623528348c`:
Add photos → Choose from library creates two thumbnails. Opening the first shows
1 of2 and its filename. Next shows the second filename and2 of2. Remove's native
confirmation identifies the new-item draft; Cancel preserves the second photo.
Accepting removal then [shows the first photo as1 of1](evidence/android-add-preview-remaining.png)
without paging buttons. Close positively returns to Add item with Remove photo1
and without Remove photo2. The duplicate library header stays absent. Distinct
filenames/counts prove paging because the synthetic images share pixels.

Evidence: `/tmp/android-add-photo-{entry,source,selected}.xml` and
`/tmp/android-add-preview-{first,second,remove-confirm,cancel,removed,closed}.xml`.
Last-photo removal, zoom chrome restoration, iOS, TalkBack and physical media
selection/upload remain outside this sample. No asset was saved.

Last-photo follow-up on the same5660fb54 Android build: reopening the remaining
thumbnail and accepting draft-photo removal closes the viewer automatically.
The positive destination is Add item, with Add photos and Asset name present and
no Remove photo1 thumbnail command. Close Add then returns to the audit root. No
asset was saved. Evidence: `/tmp/android-add-last-{entry,confirm,removed,close}.xml`.
This closes the last-photo runtime gap for this Android sample, not zoom or iOS.

The matching iOS acceptance scenario now enters through Add's library chooser,
checks both filenames during paging, cancels then accepts removal, closes/reopens
the surviving thumbnail, and removes the last photo. Final dismissal requires
Add's field to disappear, the root menu to be hittable, and the app to remain
foreground; a background menu's mere existence cannot pass. The scenario is
included in the focused Add-draft suite as well as the complete suite. Remote
structural validation and critic review passed; Swift compilation and native
execution remain pending, so this adds coverage, not iOS acceptance evidence.

### Android zoom and system escape

Same5660fb54 build, Android16 Pixel6, normal text/light: Add's selected synthetic
photo opens with its filename, position and footer commands. Double tap visibly
[enlarges the image and hides the footer](evidence/android-photo-zoomed.png). A
single tap while enlarged does not reveal controls. A second double tap restores
the fit image and [the full footer](evidence/android-photo-zoom-restored.png),
including Close, paging and Remove. Zooming again and using Android system Back
returns to Add item with both draft thumbnails and both removal commands intact
(`/tmp/android-zoom-back.xml`). No asset was saved or photo removed.

This confirms double-tap zoom/reset and system escape in the Add consumer. It
does not establish pinch/pan behavior, access to commands while still zoomed,
TalkBack, iOS gestures, or the unavailable-image transition during zoom. The
single-tap observation is retained as an interaction limitation, not dismissed
because zoom reset provides another path.

### M234 candidate: commands independent of zoom

The pinned viewer patch now leaves zoom responsible only for image scale/scroll
behavior. A distinct single-tap callback toggles the footer without React state
remounting the image; opening the viewer, selecting another index and image-load
failure restore commands. The iOS recognizer owns its delayed tap per image scope
and retires stale callbacks. Android filters double taps, drags, long presses,
multitouch and canceled gestures before dispatching a single tap; its responder
cleanup cancels pending work. Review caught effect-replay activity and same-index
reopening gaps; both were corrected.

Single-tap regressions failed against the prior installed hooks on paul. Five
focused checks now pass, including iOS scope/unmount ownership and independent
viewers, and Android StrictMode, drag, double-tap and termination behavior. The
full mobile suite passes1,893 tests across295 files; TypeScript and mobile
structural checks pass. These are source checks, not native acceptance.

The patch applies cleanly to pristine0.2.2 whose tarball matches the committed
SHA-512 integrity. The lock's three patch identities were updated to the patch's
SHA-256. Existing paul pnpm installs left the old package symlink in place, so
validation explicitly applied the patch to pristine source and copied that result
into the validation dependency. A fresh CI install remains necessary to verify
package-manager materialization; native Android rebuild is in progress.

### M234 rebuilt Android and footer contrast

APK22ac03fb9bb1f81fc65b2475ab1986852643cf2b87bf51424906cea360176536
contains the regenerated bundle (the first Gradle assembly reused the old APK;
forcing the bundle task produced this distinct build). Android16 Pixel6 normal
text/light: double tap zooms; separate single taps hide/reveal commands without
resetting the image in the captured sequence. Closing while commands are hidden
via system Back and reopening the same thumbnail restores the footer. Swiping
to photo2 while commands are hidden also restores it (`/tmp/photo-tap-reopened.xml`,
`/tmp/photo-tap-swiped.xml`).

This exposed [white position text over bright image pixels](evidence/android-photo-zoom-label-contrast-before.png).
The complete footer now has the viewer's opaque neutral background. Ten focused
consumer tests, TypeScript and structural checks pass on paul; critic found no
blocker. Rebuilt APKa960ab9282b8188ce996d49f5fcec3d10fad745c85be707dba58abd29a35573f
shows [readable labels while zoomed](evidence/android-photo-zoom-label-contrast-after.png).
The matching [hidden state](evidence/android-photo-zoom-controls-hidden.png) and
revealed state preserve the same enlarged image. Footer Close positively returns
to Add with both thumbnails/removal commands retained (`/tmp/photo-footer-closed.xml`).

An earlier interrupted capture sequence on a960ab92 returned to fit scale between
the zoom capture and subsequent hide/reveal captures; the cause was not observed.
The immediate repeat above preserves scale, but zoom persistence across unrelated
rerenders/elapsed time is still unverified and needs a targeted check. iOS, asset
viewer regression, pinch/pan and assistive operation remain open. This is partial
Android acceptance, not closure of M234.

### M234 zoom persistence follow-up

The installed Android responder reproduced scale2→1 on an unrelated mounted
rerender. Its local mutable gesture state and Animated values were recreated on
every render. The candidate now memoizes a gesture controller for image scope
and initial geometry while committed callback refs remain current. New scopes
retire timers/listeners/animations and stale handlers. Review found that a new
fit-scale controller also needed to publish its zoom state so the parent could
reenable horizontal paging; the new replacement assertions failed before that
correction and pass after it. Activation publishes actual retained state during
effect replay rather than always declaring unzoomed.

All1,896 mobile tests across295 files, TypeScript and structural checks pass on
paul. The final patch applies to pristine pinned0.2.2. Native rebuild is in
progress; elapsed-time zoom retention, image replacement/paging and iOS remain
open. This source reproduction does not prove which event caused the earlier
interrupted native capture to reset.

### M234 retained zoom on the rebuilt Android candidate

APK `d9b1a7b5196c970f45131b6d960adea3ae02bc46ea5f4d772dff39ec8a00bb72`
includes the retained gesture controller from `460007a9` and the opaque footer.
On the same Android16 Pixel6 fixture at normal text size, double-tapped photo1
remains enlarged after45 seconds and after Home followed by warm app return.
The [resumed capture](evidence/android-photo-zoom-retained-after-resume.png) retains
the same image crop and readable commands. Subsequent single taps hide and reveal
commands while preserving the enlarged crop. Next selects filename2 at fit scale;
a horizontal swipe returns to filename1, confirming paging is reenabled. Footer
Close returns to Add with both thumbnails and both removal commands retained.

Inspected captures: `/tmp/photo-retain-{start,delayed,resumed,hidden,revealed,next}.png`.
Native hierarchy evidence on paul: `/tmp/photo-retain-{swipe,close}.xml`.
The fixture uses two synthetic draft photos; no asset is saved. This strengthens
Android Add-preview acceptance but does not verify iOS, asset-viewer consumers,
pinch/pan, assistive operation or a fresh package-manager install.

The same APK also exercises `AssetPhotoViewerSheet` through the synthetic persisted-
photo removal fixture. While double-tap zoomed, Remove opens the native confirmation;
accepting produces the expected fake removal error. After OK, the image remains
zoomed and both Close and Remove are visible. Close returns to the fixture with
`Removal attempts: 1` and `Photos remaining: 1`. Inspected screenshot:
`/tmp/asset-zoom-recovered.png`; hierarchy captures on paul:
`/tmp/asset-zoom-{confirm,error,closed}.xml`. This verifies the shared viewer's
failure recovery on Android, not a real service deletion or complete asset flow.
