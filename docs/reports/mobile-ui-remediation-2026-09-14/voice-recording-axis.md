# Voice recording — 24 source axes

S119 at3f601d5d. Reviewed VoiceConversationComposer, NativeConversationButton
iOS/Android, VoiceLevelMeter, VoiceInteractionStateContext, controller start/stop/
pause, ExpoVoiceAudioRecorderCore, and their tests. Physical capture, permission
dialogs and audio interruption behavior are not established by this review.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Explicit microphone command begins capture; finish sends. Typing cancels capture without sending. |
| Navigation | Sheet blur pauses media. M171 protects response navigation; M172 covers pending native capture startup. |
| Selection | Recording is a command, not a value picker; no selection screen is appropriate. |
| Modality | Native router sheet contains persistent composer; permission is requested by the OS on start. Actual prompt transitions remain open. |
| Layout | Composer reserves a48-point native command beside flexible text and a meter. Keyboard and narrow-window clearance remain native checks. |
| Adaptation | Shared flex row, platform-native command implementations. Tablet/window and Android edge-to-edge acceptance remain open. |
| Typography | Recording command is an icon with a state-specific accessible label. Error copy wraps in the parent scroll. |
| Appearance | Meter uses semantic action foreground (M25); native command uses action tint. Device contrast remains unverified. |
| Localization | English labels; no fixed-width visible command words. RTL composer ordering remains unverified. |
| Imagery | Native microphone, send arrow and stop-square symbols express distinct commands. Meter is decorative rather than an interactive icon. |
| Targets | Native hosts reserve48 points. Actual native hit regions are not established by host dimensions. |
| Gestures | Tap starts and tap finishes; no hold-only gesture. Input remains an alternate interaction. |
| Keyboard | Focusing typing while listening calls pauseMedia. Native keyboard/prompt focus and interrupted startup need acceptance. |
| Accessibility | Explicit Start recording and Finish recording and send labels. Decorative meter is non-accessible. VoiceOver state-change announcement remains unverified. |
| Motion | Meter reflects sampled audio without explicit animated transitions. Physical level updates and reduced-motion experience remain open. |
| Content | Stage controls determine command meaning; captured audio is not submitted merely by dismissing. |
| Search | No recording-owned search; spoken queries enter the conversation workflow. |
| Loading | Duplicate startup is guarded by requestPending; stage remains unchanged while startup awaits. A visible startup indication is an unverified usability risk. |
| Recovery | Denied permission throws readable error and typing remains available. No dedicated Settings recovery action was found for microphone denial; actual permanent-denial flow needs review. |
| Editing | Composer text is retained separately; no waveform trimming or audio editor is specified. |
| Privacy | Permission precedes capture; cancelled audio is deleted without reading. M172 candidate checks cancellation before capture across permission/mode/preparation. |
| Notifications | No recording-owned notifications; interruptions need physical-device acceptance. |
| Media | MP4 capture, measured levels and temporary-file cleanup use injected native ports. Route/audio interruption and competing audio remain physical checks. |
| Lifecycle | M172 candidate aborts native startup and serializes cancellation cleanup before a fresh recording. Mounted/adapter evidence is distinct from physical permission/interruption acceptance. |

The Apple [privacy guidance](https://developer.apple.com/design/human-interface-guidelines/privacy)
is a relevant review lens, but its current HTML returned only a JavaScript shell.
The attempted recording-audio topic could not be retrieved. No detailed Apple
requirement is inferred from those unavailable pages. M172 follows the project's
explicit cancel-capture behavior and inspected implementation.

The baseline88 recorder/controller/composer/lifecycle tests passed on paul before
M172. Its candidate adds12 native-boundary/controller cases, including initial and
follow-up starts and cleanup overlapping fresh recording. All98 focused recorder/
controller tests and static checks pass; code critic reviewed the correction.
Physical permission/preparation timing and interruptions still require acceptance.
