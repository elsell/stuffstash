# Conversation loading and processing — 24 source axes

S120, baseline9e11dfe1 plus M177. Reviewed the workspace's initial load/error,
processing feedback, composer cancellation, session presentation, scoped provider
and new-conversation ownership. Recording and applied-plan progress have separate
reviews. Native geometry, speech and interruption behavior remain unverified.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Loading establishes inventory context; processing waits for a response. Text identifies work without inventing measured progress. |
| Navigation | Conversation remains in its sheet. Closing pauses media through the existing lifecycle; content stays in the provider for return. |
| Selection | Not applicable: processing and initial recovery have no value-selection control. |
| Modality | No nested wait dialog. Existing M144 protects new-conversation confirmations; native dismissal remains pending. |
| Layout | Conversation scroll owns body content; composer stays in the sheet action area. M177 makes initial failure recovery scrollable rather than a fixed minimum-height panel. Native compact detents need inspection. |
| Adaptation | Flexible text and scroll recovery avoid fixed viewport assumptions. Phone/tablet widths and split windows are not yet measured. |
| Typography | Status/error text has no single-line restriction. Long safe messages and ordinary font-size wrapping need native acceptance. |
| Appearance | Palette text/status with native spinner, cancel and new Retry. Contrast and platform rendering remain unmeasured. |
| Localization | Status vocabulary and retry/error labels are English; no locale/RTL certification. |
| Imagery | Activity indicator is accompanied by a textual phase; no dependency on photos for processing feedback. |
| Targets | Native Cancel request and Retry conversation. M178 moves Close/New conversation into the native header; actual hit regions and compact detents remain unverified. |
| Gestures | Explicit cancellation during cancellable processing; approved writes cannot be cancelled. No gesture is needed to recover initial context failure after M177. |
| Keyboard | Composer becomes noneditable during processing, while cancellation remains a separate command. Native keyboard/action overlap remains open. |
| Accessibility | Progress text uses polite updates; Retry has a label/disabled state and error content a heading. Native reading order and announcements remain pending. |
| Motion | Native spinner plus textual phase; no artificial time estimate. Reduce Motion runtime behavior requires acceptance. |
| Content | Current transcript/history remain visible while safe progress updates. Action-plan progress is rendered separately to avoid competing save indicators. |
| Search | Not applicable: this is conversational execution, not a search/filter control. Result discovery is audited under response. |
| Loading | Query loading is independent of recording/processing stages. M177 retries scoped context rather than restarting a conversation or recording. |
| Recovery | M177 fixes initial Voice unavailable with no Retry. Failed retry retains error; success restores ready provider state. Subsequent request errors already render safe recovery content. |
| Editing | Processing disables new input submission. Existing M144 confirmation ownership protects pending plan/photo drafts; M178 removes the duplicate unprotected terminal Reset path. Native draft dismissal remains open. |
| Privacy | Progress text passes safe presentation filtering; inventory context stays query-scoped. Retry does not bypass existing access checks or introduce an endpoint. |
| Notifications | Not applicable: this surface does not own notification entry or permission registration. |
| Media | Busy microphone cannot start another request; close pauses media. Physical speech playback/interruptions require separate runtime evidence. |
| Lifecycle | Existing generation/visit guards protect completion and reset. M177 owns Retry per focused visit; callbacks retained after return cannot start work. Process restart has no durable conversation restoration by product design. |

M177 tests mount the real query/provider and recovery component for both failed
inventory scope and failed context, repeated failure, then successful ready state.
They do not mount the entire workspace: its imported Expo native services are not
available in the unit harness. Workspace wiring is source evidence. A separate
component case covers pending lock and retained callbacks after leave/return.
The first test failed because the new recovery component was absent, not because
native failure geometry was reproduced. Code critic found no implementation
blocker; native Retry reachability and announcements remain pending.

Validation:45 focused cases across five files, TypeScript and mobile structural
checks pass remotely on paul (`/tmp/voice-preview-green.log`).
