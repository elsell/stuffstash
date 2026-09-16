# Photo viewer source review

Reviewed at82669f4b, including the installed, pinned react-native-image-viewing0.2.2
implementation. This is source evidence; the new photo recovery fixture has not run.

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
that full-screen consumer. AddAssetFixture currently returns no selected photos;
controlled photo selection needs extending before native preview/paging/Close
acceptance can be claimed. Its8 mounted draft-preview cases remain distinct.

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
