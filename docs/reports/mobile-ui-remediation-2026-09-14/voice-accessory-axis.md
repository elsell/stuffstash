# Voice accessory review

S117 reviewed at c02ac21c, September15. Sources: VoiceAccessoryContent,
VoiceBottomAccessory, VoiceTabContent, VoiceSessionPresentation and the conditional
VoiceConversationReturn. This is the entry/status control review, not an audit of
the full conversation, provider configuration, microphone adapter or model output.

| Axis | Source conclusion and required acceptance |
| --- | --- |
| Task | One primary action starts, sends or opens the current conversation according to state. M113 restores first entry on platforms without native accessory support. |
| Navigation | Status opens /voice; a ref prevents repeat navigation until return. Controlled start/send/return behavior passes. Failed navigation, rapid taps across native accessory copies and tab changes remain unverified. |
| Selection | This is a command, not a setting picker. Listening sets the primary control's accessibility selected state; inspect whether assistive technology conveys recording clearly. |
| Modality | Entry opens the existing root conversation sheet. The accessory does not create nested dialogs; sheet coverage and return focus remain native acceptance. |
| Layout | iOS26 uses native placement. Other platforms reserve a sibling area inside tab content, not an absolute overlay. Actual tab/safe-area/keyboard clearance remains pending. |
| Adaptation | Regular placement has status plus primary control; inline placement retains only the primary control. M113 adds unsupported-platform entry. Test phone/tablet and both native placements. |
| Typography | Status uses system text, one line each,15/12pt; truncation is bounded. Normal-size long response readability remains pending; enlarged text is a later remediation phase. |
| Appearance | Status and buttons use palette tones. The primary glyph uses onAction with action/success/amber/danger backgrounds; contrast in every tone needs runtime inspection rather than assuming one text color works for all. |
| Localization | English labels and route-based context strings remain. No date/number formatting in this control. RTL order and translated length remain unverified. |
| Imagery | Microphone/send glyphs represent commands; listening includes a level meter. Color dots supplement text in regular placement; inline uses the primary accessibility label. |
| Targets | Regular primary is54pt; inline is44pt; status has52pt minimum height. The inline size differs from the shared48pt target preference and needs native placement review before changing system-constrained geometry. |
| Gestures | Actions have explicit press controls; no gesture-only entry. Native tab minimization/placement transitions remain unverified. |
| Keyboard | No text input is owned here. Search/composer keyboard overlap, dismissal and focus return remain native acceptance, especially fallback layout. |
| Accessibility | Both press regions have labels/roles; listening selected state is explicit. Meter is non-accessible. Actual traversal of status-dot children and simultaneous native placement copies requires VoiceOver/TalkBack inspection. |
| Motion | Press feedback scales to0.98; meter height follows level and processing uses a spinner. Reduced-motion treatment of these effects remains an unverified risk, separate from the earlier result-rail motion fix. |
| Content | The compact status is not a transcript. Responses are redacted/normalized for the subtitle; the full sheet is the reading destination. Verify long status clipping does not hide the action. |
| Search | No search control belongs in the accessory. Context differs between Home and Browse; actual query scope is owned by the voice provider, not this presentation. |
| Loading | Loading opens voice status; processing/speaking open the session rather than start another recording. Presentation cases are tested; native busy-state semantics remain pending. |
| Recovery | Failed/cancelled/completed states open their session; error text is safely summarized. No destructive retry occurs directly from the status area. Native failure entry remains pending. |
| Editing | No editable draft or plan approval belongs in this control. Starting/sending delegates to the shared provider; cancellation and draft preservation require conversation-screen audit. |
| Privacy | Subtitle helpers redact response/failure text, and diagnostics affect error detail. Real account/inventory changes, lock-screen exposure and stale session display cannot be certified by this component review. |
| Notifications | This control owns no OS notification permission or push route. Interrupting recording with notification navigation remains a lifecycle scenario. |
| Media | M113 entry uses the same startRealtime/stopRealtime context methods as iOS26. Controlled recorder/transport evidence is not proof of physical microphone permission, audio routing or interruption behavior. |
| Lifecycle | The provider owns session lifetime; component navigation refs are per instance. Controlled same-instance return works. Native regular/inline copies, fallback tab copies, background and covering sheets remain open. |

Evidence: the combined suite at c02ac21c passes1,606 tests/264files remotely,
including three fallback platform cases and a real provider/controller with fake
recording/transport. No native rendering, assistive-technology or physical audio
pass is claimed from these tests. M113 remains a finding with an implemented
candidate pending native acceptance; other risks above are not invented defects.
