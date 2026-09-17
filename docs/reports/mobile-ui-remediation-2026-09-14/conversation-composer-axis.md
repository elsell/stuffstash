# Typed conversation composer — 24 source axes

S118 at6edc7e85. Reviewed VoiceConversationComposer, its production sheet placement,
VoiceInteractionStateContext submission/pause handling, conversation stage policy,
native conversation adapters, composer and conversation lifecycle tests. This
review does not certify recording hardware, the full response UI, or plan approval.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Enter an inventory request or start recording. Nonempty text changes the main command to Send; listening changes it to Finish recording and send. |
| Navigation | Composer belongs to retained conversation state above the sheet. It does not introduce a separate editing route. Native collapse/return still needs acceptance. |
| Selection | No option selection; text input and commands. |
| Modality | No composer-owned modal. Plan approval replaces it with a separate decision area; that area's custom controls require M170 review. |
| Layout | Flexible input beside a 48-point native command; composer sits in sheet bottom action area. Native keyboard and safe-area overlap remain open. |
| Adaptation | Multiline input has 44 minimum and120 maximum height. Real long-input scrolling and narrow windows need runtime checks. |
| Typography | 16-point input styling, explicit placeholder and semantic colors. Text scaling/truncation remains unverified. |
| Appearance | Native SwiftUI button or Compose IconButton with semantic tint. Actual appearance and dark/light acceptance remain open. |
| Localization | English command labels and placeholder; request text remains user input. RTL caret, IME and dictation not verified. |
| Imagery | Standard microphone, upward arrow and stop icons; each carries a changing accessible label. |
| Targets | Native hosts declare48×48. Actual UIKit/Compose hit regions need runtime evidence. |
| Gestures | Explicit Send/record/cancel; text editing uses platform gestures. No custom gesture is required to submit. |
| Keyboard | Shared multiline AppTextInput supports native paste and selection. Focusing while listening pauses media. Physical keyboard, IME and native text reliability remain pending. |
| Accessibility | Named Message Stuff Stash input, state-specific named buttons, named activity indicator. VoiceOver focus on changing actions and audio meter semantics remain open. |
| Motion | No composer animation; live recording meter is separate. Reduced-motion meter behavior is not certified here. |
| Content | Placeholder distinguishes ask/add, actions distinguish send/record/cancel. Failure detail belongs to conversation output. |
| Search | Natural-language requests may search, but this is not a list-search field; no local filter semantics are implied. |
| Loading | Processing disables editing and substitutes Cancel only when cancellable. Non-cancellable progress is shown unless plan card owns it. Preview loading/error replaces the body before composer mounts. |
| Recovery | Unaccepted submission failure restores typed text; accepted requests are not restored for accidental duplicate submission. Mounted lifecycle test covers this distinction. |
| Editing | Empty/whitespace text cannot send; request lock is set before awaiting. Text is retained above the route; input has8000-character limit. Native editing remains pending. |
| Privacy | Inventory-scoped provider clears old state and uses generation ownership. This does not replace real API tenant/authorization coverage. |
| Notifications | No composer notification setting or delivery action. Interruptions still require lifecycle testing. |
| Media | Start/finish recording calls parent media flow; focus pauses listening. Microphone permission and hardware cannot be certified by this source review. |
| Lifecycle | Generation prevents late submission results from replacing current session. Sheet focus cleanup pauses media. Collapse/reopen/background and scope-change native journeys remain open. |

Existing mounted composer test verifies a single Cancel command during processing,
disabled typing and absence of Send/Cancel during review. Lifecycle coverage checks
submission-failure restoration and accepted-request protection. These are included
in the1,736-test checkpoint; none proves native typing, audio or layout acceptance.


## Android runtime follow-up, September 16

APK `d2cafbc60c76c60f57526e58075fc5597b27d3e8799cb83ef16b42fe66341a4a`,
Pixel6 Android16 API36, normal font. This is the selectively patched audit candidate
described in M238, not a full-HEAD release build.

After resetting the synthetic conversation, entered an exact two-line request:
“Find the camping tent in the garage” followed by “Include the blue box and all
camping supplies”. The multiline field and native Send control stayed above the
software keyboard. Home/launcher warm return retained the exact text. Tapping
Send while the keyboard was visible displayed that exact request and the fixture's
synthetic proposal; this verifies activation, not the relevance of model output.
Close returned to the audit root. Warm deep-link reentry to Conversation retained
the request and proposed change. System dark appearance retained the review with
legible native header and decision controls. Restored system light after capture.

Evidence: [keyboard-visible composer](evidence/android-voice-composer-keyboard.png)
and [dark review after return](evidence/android-voice-review-dark.png). Hierarchies
on paul: `/tmp/voice-composer{,-return,-sent}.xml`, `/tmp/voice-close-root.xml`,
`/tmp/voice-route-return.xml`. These observations narrow the keyboard, multiline,
appearance, gesture and lifecycle gaps for this Android candidate. They do not
establish iOS/iPad, physical audio, screen-reader, alternate IME, hardware-keyboard,
process-death, very long input or enlarged-text acceptance.
