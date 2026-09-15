# System keyboard accessory review

S127, source91a48f09. Reviewed AppKeyboardAccessory iOS/default, provider variants,
AppTextInput and its tests. This is the shared dismissal control, not verification
of every editable field. The project spec explicitly requires the iOS extension
and preserves Android's system dismissal; it is not a claim that Apple mandates
an app-wide accessory. Field inventory remains text-input-sites.csv.

| Axis | Source result and remaining acceptance |
| --- | --- |
| Task | Explicitly hides the keyboard without committing or clearing input. This is the project-specified supplement to drag dismissal. |
| Navigation | No route change. Current field remains owned by its screen. Menu/sheet transitions and return need native checks. |
| Selection | No value selection is owned by the accessory. |
| Modality | KeyboardExtender attaches to the keyboard rather than opening another sheet. Native first-responder behavior across presentations remains a runtime gate. |
| Layout | Absolute root host consumes no normal-flow height. Bar is44 high; right inset is max20/safe-right. Keyboard and sheet boundary composition requires native evidence. |
| Adaptation | iOS extension only; default renderer returns null. Landscape right inset is considered; floating/split iPad keyboard and hardware keyboard remain unverified. |
| Typography | Icon-only action, no visible label to truncate. Spoken label is explicit. Native magnification/large content viewer is unverified. |
| Appearance | Link PlatformColor follows system. Transparent bar relies on keyboard extension rendering; theme/contrast/material composition needs native review. |
| Localization | Spoken label/hint remain English; trailing placement uses physical right. RTL placement remains open. |
| Imagery | ChevronDown communicates dismissal; no media. Native keyboard-down symbol versus generic chevron remains a platform-fit review item, not a verified rendering defect. |
| Targets |44×44 Pressable plus hitSlop4. Mounted test verifies declaration; actual runtime target/keyboard clipping is separate. |
| Gestures | Tap is explicit alternative to drag dismissal. No custom gesture is required for the accessory. |
| Keyboard | Uses KeyboardController.dismiss for actual native first responder. Provider disables preload. Tests verify no submit/change callback; cross-field native behavior still requires verification. |
| Accessibility | Button name and hint explain dismissal without submit. Host is accessibility-hidden while keyboard is closed. VoiceOver focus after hide and external keyboard activation remain pending. |
| Motion | Accessory adds no custom animation. Keyboard-controlled transitions and Reduce Motion remain runtime checks. |
| Content | One command, no list or pagination. |
| Search | No search query or result state is owned here; dismissal must not submit a focused search. |
| Loading | Keyboard visibility follows willShow/didHide, initialized from isVisible. No network loading state. Native interrupted transitions remain unverified. |
| Recovery | Dismiss promise has no explicit error UI. Native failures should be diagnosed from runtime evidence before adding an alert to this low-risk action. |
| Editing | No save, clear or validation is invoked. Mounted test preserves supplied field value and receives no edit/submit callback. Real field draft retention remains a native check. |
| Privacy | No text is read, retained or transmitted by this control. It only dispatches dismissal. |
| Notifications | Owns no notification feature. Push-driven presentation during keyboard visibility remains an interruption test. |
| Media | Owns no camera/audio/file operation. Voice takeover while keyboard is open requires native testing. |
| Lifecycle | Visibility listeners are removed on unmount. Closed host rejects pointer interaction and is hidden from accessibility. Rapid keyboard replacement/hide/show and background return remain native checks. |

Run349853 phone's Expiration keyboard journey passed, but this is one consumer and
not blanket accessory acceptance. The ordinary multiline comparison failed before
entry due to fixture placement (now corrected). No whole-input or iPad keyboard
pass is inferred from either result.


Run34992079258 iPad place-search evidence confirms M137: the dismissal control
is visible and exposed but fails native hit testing after a successful search.
See findings.md and retained ipad-place-search-dismiss-349920 image/hierarchy.
The source target-size declaration does not resolve the native ancestor geometry.
