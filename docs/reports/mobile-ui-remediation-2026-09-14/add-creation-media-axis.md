# Add quick creation and photo selection — 24-axis source review

S087 and S090 at73ed44fb plus M208. Sources: AddAssetScreen, PhotoSourceChooser,
DraftPhotoPreviewModal, PhotoSelectionQuery and ExpoPhotoSelectionProvider;
placement details also use add-parent-axis.md. These are source reviews, not
native acceptance. Prior findings and their outstanding runtime checks remain.

| Axes | Evidence and remaining acceptance |
| --- | --- |
| Task, navigation | Quick creation makes a real place and selects it inside the item draft; abandoning the item does not delete that place. Photos use the system camera/library and return to the draft. Neither task needs a separate app navigation stack. |
| Selection, editing | Successful creation stores the returned identity and selection context. Selecting another place or typing clears the created-place marker. Photo cancellation returns an empty selection and leaves existing photos intact; removal and reorder change only the draft. M208's numbered removal commands follow current positions. |
| Modality, lifecycle | Busy operations lock editing, prevent route removal and dismiss the keyboard. Photo-source callbacks are scoped to their opening visit; preview-removal confirmation captures selection and visit ownership. Already-started system pickers, background interruption and process death still need device checks. Draft storage is in-memory, not durable recovery. |
| Layout, adaptation | Creation is inside the bounded nested parent list. Photos use a horizontal rail with108-point previews. M208 removes the28-point overlay and lets each preview's shell grow for a native command below it. Verify narrow phone/iPad command wrapping, rail height, keyboard occlusion and actual hit areas. |
| Typography, appearance | Semantic palette and inherited text scaling; native removal labels may wrap at the fixed preview width. Existing fixed photo ordinals and drag hints need runtime inspection; normal-size checks take priority. Increased contrast is not established by token use alone. |
| Localization | English commands and numbered positions; user-entered place names and image filenames can be long. RTL reorder direction and localized labels remain unverified. |
| Imagery | Local image URIs preview selected originals and ignore color inversion. No media belongs to quick-place creation. The thumbnail itself lacks an explicit load-error recovery path; full-screen viewer recovery is reviewed separately. Treat unreadable local thumbnails as an unverified scenario, not a proven production failure. |
| Targets, accessibility | Quick creation uses the native command adapter. Photo Add has an explicit name; M208 gives removal a separate native target/name. Preview exposes adjustable position and activate/reorder/delete actions. VoiceOver/TalkBack grouping, pronunciation and reachability require native checks. |
| Gestures, motion | Photo drag has accessibility increment/decrement alternatives and bounded movement. Long-press/scroll arbitration and drag feedback under Reduce Motion are unverified. Quick creation is an explicit command; it requires no gesture. |
| Keyboard, search | Parent query is debounced, scoped and retryable; unknown results cannot offer creation. Search focus scrolls the form; operation start dismisses the keyboard. Photos delegate search to the system picker; the app does not implement a second photo search. |
| Content, loading | Place suggestions are capped; quick creation remains explicit. Photo order is draft order. The synchronous operation guard prevents duplicate requests; creation exposes Creating place… while the system picker owns its own loading UI. Large selections and memory pressure are not performance-certified. |
| Recovery, privacy | Creation/picker failures retain the draft and use the shared form error. Camera denial gives Settings/library recovery guidance; library selection does not request broad library permission. Unsupported formats fail explicitly, including mixed selections. Scope and mutations continue through existing application ports; this review changes no authorization boundary. |
| Notifications, media | Neither task owns notification entry. Place creation has no media permission. Photo adapters support JPEG/PNG/WebP and preserve metadata/base64 for later upload; selection itself does not upload. Physical camera, limited library access and interruption remain device acceptance items. |

## M208 — Separate native photo removal from preview

Source-confirmed P2: Add used an absolute28-point X overlay despite the shared
native command adapter already used by voice photo drafts. This is a project
pattern/target correction; it is not a claim that Apple mandates removal below
every photo. Removal now sits below each preview, with position-based labels.

The mounted regression failed on the old implementation, then passed with the
change: both photos can be removed in sequence without losing the item's title
or removing the wrong identity. All41 related tests, TypeScript and mobile
structural checks pass on paul (`/tmp/add-photo-remove-green.log`). Critic found
no blocker; native wrapping at108 points, rail geometry and reordered numbering
remain acceptance checks. No native run has verified this change yet.

Apple's [Buttons guidance](https://developer.apple.com/design/human-interface-guidelines/buttons)
is the applicable reference; this pass's HTML retrieval was JavaScript-only, so
no new detailed Apple requirement is inferred from it. The implementation follows
the already recorded platform interaction standard and existing native adapter.
