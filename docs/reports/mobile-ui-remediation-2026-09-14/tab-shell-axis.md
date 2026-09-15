# Tab and nested stack shell review

Source review at e3367204, September 15: R001 Home stack, R003 tab shell,
R004 Browse stack, NativeTabHeader, VoiceBottomAccessory and the installed
Expo native-tab adapter. This reviews shell responsibilities, not the Home/Browse
content or voice service. No new runtime acceptance is claimed.

| Axis | Source result and remaining acceptance |
| --- | --- |
| Task | Home and Browse are persistent peer destinations; Add and Settings remain stack tasks. Appropriate source structure. |
| Navigation | Each tab owns a native stack. Selected-tab reselect, retained position and Back need native acceptance. |
| Selection | Native tab triggers own selected state; no custom tab selection overlay. Verify restored selection. |
| Modality | Shell does not present its own modal; voice routes to /voice. Root sheet dismissal belongs to separate route review. |
| Layout | NativeTabHeader enables transparent iOS headers and soft iOS26 scroll edges, material on older iOS; Android uses opaque background. M103 notice placement remains separately tracked. |
| Adaptation | Native tab placement delegates to Expo. BottomAccessory is documented in the installed adapter as iOS26-only; other-platform voice entry needs review. |
| Typography | Native tab labels and titles are explicit. Long/localized titles and normal-size toolbar crowding need runtime inspection. |
| Appearance | Header colors use palette; native tab appearance is system-owned. Verify live theme transitions and scroll-edge blending. |
| Localization | Home/Browse labels are English literals. Existing localization review applies; no localized runtime result inferred. |
| Imagery | R003 supplies only SF symbols. Installed Android conversion accepts drawable/md/src, not sf: M112. Home/Browse stacks own no icon assets. |
| Targets | Native tab items own touch geometry. Accessory primary button has44/54pt dimensions; actual platform hit regions remain unverified. |
| Gestures | Tab taps and explicit voice control are available in the iOS26 shell. Tab reselection/pop/scroll behavior remains pending. |
| Keyboard | Shell has no input. Keyboard dismissal while changing tabs and native search interaction belong to runtime acceptance. |
| Accessibility | Native labels identify tabs; accessory labels include state. Selected announcements, traversal and duplicate native accessory copies need inspection. |
| Motion | Native stack/tab transitions are delegated. Accessory has a pressed scale; shared motion review remains applicable. |
| Content | Shell contains two destinations, not a data list. Child pagination cannot be certified here. |
| Search | Browse owns search; tab shell deliberately has no search-field state. Search restoration belongs to R005. |
| Loading | Stacks do not bind a refresh spinner. Voice accessory derives progress from voice stage; child loading remains separately reviewed. |
| Recovery | Shell contains no data retry; accessory can open voice while not ready. Verify failure-state entry on supported platforms. |
| Editing | No shell draft. Switching tabs during child mutation requires child visit ownership; no blanket acceptance from native stacks. |
| Privacy | Shell relies on outer authenticated composition; it does not implement authorization. Account change and retained tab contents need integration verification. |
| Notifications | Shell does not own notification routing. Notification return/tab state belongs to root notification acceptance. |
| Media | Accessory starts/stops voice through interaction context. Permission, audio and non-iOS26 entry are not verified by this source review. |
| Lifecycle | Accessory guards repeated navigation until pathname leaves voice. Background, restoration, rapid tab changes and failed navigation require runtime checks. |

M112 is a source-confirmed configuration omission, not an Android screenshot claim.
Use platform-provided Android tab icons alongside the existing iOS symbols. Do not
copy SF symbols into Android assets or change the two-destination information
architecture to address it. The voice accessory compatibility gap remains an
unverified risk until all alternate entrypoints and supported OS scope are checked.

Apple's tab-bar guidance URL was checked during this review, but its HTML required
JavaScript; no new normative claim is attributed to inaccessible text. Platform
capability evidence comes from the installed Expo types and implementation.
