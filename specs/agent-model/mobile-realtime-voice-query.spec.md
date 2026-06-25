# Mobile Realtime Voice Query Spec

## Purpose

Stuff Stash needs a first production-shaped mobile voice slice that proves the core conversational loop with real audio input, speech-to-text, language inference, tool calls, structured final output, text-to-speech, and streamed audio playback.

The first testable user experience is:

1. The user taps the mobile Voice control.
2. The user asks a read-only inventory question, such as "Where are my tools?" or "What is in the garage?"
3. The mobile app streams audio to the core API.
4. The core API transcribes the audio through a speech-to-text port.
5. The core API runs the agent loop with the project-owned tool catalog.
6. The agent loop may call read-only inventory tools one or more times.
7. The agent loop produces a structured final response.
8. The core API streams the final spoken response through a text-to-speech port.
9. The mobile app plays speech audio as soon as useful audio chunks are available.

## Scope

This spec covers the first mobile realtime voice query slice.

This spec includes:

- Mobile voice session states.
- API-mediated WebSocket interaction.
- Client audio upload.
- Speech-to-text events.
- Agent progress and tool-call events.
- Read-only internal tool loop.
- Structured final response.
- Text-to-speech streaming back to mobile.
- First mobile display behavior for debug/progress events.

This spec does not include:

- Write actions.
- Approval UI.
- External MCP write tools.
- Long-term transcript retention.
- Raw audio retention.
- Offline voice behavior.
- Direct client-to-provider streaming.
- A final production voice visual design.

## Architecture

The mobile voice query slice uses the API-mediated realtime architecture in `specs/agent-model/realtime-interaction.spec.md`.

The mobile app connects only to the Stuff Stash core API. It must not receive provider credentials, provider session tokens, provider-specific realtime URLs, or direct provider bootstrap payloads.

The core API owns:

- Session authentication.
- Tenant and inventory authorization.
- Provider profile resolution.
- Speech-to-text streaming.
- Language inference.
- Tool-call orchestration.
- Text-to-speech streaming.
- Safe progress events.
- Safe final response events.
- Safe observability.

The internal agent loop may use the project-owned tool catalog specified by `specs/agent-model/mcp-agent-tools.spec.md`, but it must call tools in-process through application services and ports. It must not call the public MCP transport for this first mobile loop.

## First User Workflow

The first supported mobile workflow is a read-only question about the selected tenant and inventory.

Examples:

- "Where are my tools?"
- "Where is the fertilizer?"
- "What is in the garage?"
- "Do I have any batteries?"

The workflow must not create, update, move, archive, restore, delete, import, export, share, or configure anything.

If the user asks for a state-changing action during this first slice, the system must answer safely that changes are not available through voice yet. The response may explain the closest available non-voice action if that is safe and useful.

## Mobile Interaction

The mobile app must expose the first realtime voice query from the existing Voice bottom accessory or full Voice route.

The first interaction may use tap-to-start and tap-to-stop recording. Push-to-talk may be added later if the UI spec chooses it.

The mobile app must show:

- Current tenant and inventory context.
- Recording or listening state.
- Transcription progress when available.
- Safe agent progress events.
- Safe tool-call debug events.
- Final text response.
- Audio playback state.
- Cancellation and failure states.

Tool-call events may be displayed in a simple developer/debug panel for the first slice only when developer diagnostics are explicitly enabled. They must not expose hidden resource data, raw query text, raw transcripts, raw prompts, raw model responses, provider credentials, internal IDs, or internal stack details.

Mobile realtime voice is local-development testable before production mobile authentication exists. A production mobile rollout requires a specified mobile authentication flow and must not rely on `EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN`.

## Realtime Session

The mobile app starts a realtime voice session by opening an authenticated WebSocket to the core API.

The session start message must include:

- Tenant ID.
- Inventory ID.
- Source set to mobile voice.
- Requested capability set: speech-to-text, language inference, and text-to-speech.
- Client-supported audio input format.
- Client-supported audio output formats.
- Optional client correlation ID.

The server must respond with a session-started event or a safe failure event.

The server must reject the session before audio streaming begins when:

- The token is missing, malformed, expired, invalid, or unauthorized.
- The tenant is hidden from the principal.
- The inventory is hidden from the principal.
- Required provider profiles are missing, disabled, archived, unsupported, malformed, or have unusable credentials.
- The requested audio format is unsupported.
- The server is unable to enforce timeout, cancellation, or safe observability behavior.

## Client Audio Input

The first mobile implementation must choose one Expo-compatible audio capture format that can be streamed in chunks to the API and consumed by the selected speech-to-text adapter.

The exact format remains an implementation choice for the first slice, but it must be specified before coding begins and must include:

- Container or raw encoding.
- Sample rate.
- Channel count.
- Chunk duration target.
- Maximum chunk byte size.
- End-of-utterance behavior.

The mobile app must not record audio before the user intentionally starts a voice session.

The mobile app must stop audio capture when the user cancels, the session ends, the server rejects the session, or the timeout is reached.

Raw audio must not be durably persisted by the mobile app or API in this first slice.

## Realtime Message Families

The exact serialized schema must be specified before coding begins. The first slice must support these message families.

Client-to-server messages:

- `session.start`
- `audio.chunk`
- `audio.end`
- `session.cancel`
- `client.ack` when acknowledgement is needed for flow control

Server-to-client messages:

- `session.started`
- `session.failed`
- `transcript.delta`
- `transcript.final`
- `agent.progress`
- `tool.call.started`
- `tool.call.completed`
- `tool.call.failed`
- `assistant.response.started`
- `assistant.response.delta`
- `assistant.response.completed`
- `tts.audio.started`
- `tts.audio.chunk`
- `tts.audio.completed`
- `session.completed`
- `session.cancelled`
- `session.failed`

All message names must be stable enumerations in implementation code.

All server messages must include session ID and safe sequence metadata so the mobile app can order events and ignore late events from cancelled sessions.

Every client message after `session.start` must be bound to the authenticated WebSocket connection and server-created session. The server must reject forged session IDs, stale client sequence numbers, replayed audio chunks, messages from cancelled sessions, messages from completed sessions, and any attempt to change tenant or inventory scope after session authorization.

Client messages must include monotonic per-session sequence metadata once a session is established. The server must use that metadata only for ordering, replay rejection, flow control, and safe diagnostics; it must not treat client sequence metadata as authorization.

## Transcript Events

Speech-to-text may emit partial and final transcript events.

`transcript.delta` may contain a partial transcript. Partial transcripts are for immediate mobile feedback only and must not be treated as final user intent.

`transcript.final` contains the transcript text the agent loop may use for intent interpretation. If the speech-to-text provider cannot stream partials, the server may emit only `transcript.final`.

Raw transcripts must not be durably persisted before a transcript retention and redaction policy is specified.

Transcripts are ephemeral UI and in-memory agent-loop state only in the first slice. Raw transcript text must not be written to mobile local storage, debug event history, crash reports, analytics, audit records, observability metadata, API session metadata, logs, or provider profile test records before a retention and redaction policy is specified.

## Agent Loop

The agent loop starts after the server has enough final transcript text to attempt interpretation.

The agent loop must:

- Use the authenticated principal.
- Use the selected tenant and inventory scope.
- Use the project-owned tool catalog.
- Provide the language model with only the tools allowed for the first read-only slice.
- Treat model output as untrusted.
- Validate tool-call requests before execution.
- Authorize every tool execution through the owning application service and authorization port.
- Allow multiple tool-call iterations when needed.
- Stop when the model produces a structured final response, a safe failure occurs, cancellation is requested, or the session times out.

The first loop must expose only read-only tools to the model.

The first read-only tools are:

- Search authorized assets.
- Get asset detail.
- List assets in a location.
- List root-level assets in an inventory.

The loop must not expose write tools, provider profile tools, tenant configuration tools, sharing tools, audit mutation tools, import/export tools, or raw repository access.

## Tool Progress Events

The server should emit safe tool progress events during the loop so the mobile app can show that work is happening.

`tool.call.started` may include:

- Tool call ID.
- Stable public tool label.
- Safe generic status, such as `searching`, `looking_up_item`, or `checking_location`.

`tool.call.completed` may include:

- Tool call ID.
- Stable public tool label.
- Safe generic result status, such as `completed`, `no_visible_match`, or `needs_more_context`.

`tool.call.failed` may include:

- Tool call ID.
- Stable public tool label.
- Safe failure category.
- Safe user-facing message.

Tool progress events must use bland denial and not-found behavior. They must not distinguish hidden resources from missing resources.

Tool progress events must not include raw model reasoning, raw prompts, raw transcript text, raw query text, raw tool inputs, raw tool outputs, resource identifiers, exact resource titles, hidden IDs, result counts that can reveal hidden inventory data, credentials, bearer tokens, provider responses, authorization decisions, or stack traces.

## Structured Final Response

Every successful agent loop must produce a structured final response, even for read-only answers.

The first structured final response shape must include:

- `responseId`: stable ID for the response.
- `sessionId`: realtime session ID.
- `tenantId`: tenant scope.
- `inventoryId`: inventory scope when applicable.
- `source`: mobile voice.
- `kind`: final response kind, initially `answer`, `clarification`, `unsupported_action`, or `safe_failure`.
- `spokenResponse`: concise text intended to be spoken to the user.
- `displayResponse`: text intended to be displayed in the mobile UI.
- `artifacts`: optional safe structured artifacts, initially empty or limited to safe asset/location references.
- `toolCallIds`: tool calls used to produce the response.
- `auditMetadata`: safe metadata for observability and audit.

`spokenResponse` is the only field that may be sent to text-to-speech in the first slice.

The model must be instructed that `spokenResponse` is what the user will hear. It must be concise, natural, and free of JSON, Markdown tables, hidden reasoning, provider details, implementation details, and unsafe secrets.

`displayResponse` may be the same as `spokenResponse` in the first slice.

The final response must not include raw chain-of-thought, raw model reasoning, raw prompts, raw provider responses, raw transcripts, raw audio, credentials, bearer tokens, hidden resource data, or stack traces.

## Final Response Streaming

The system should reduce time-to-first-audio where provider capabilities allow it.

Because the canonical final response is structured, the full `assistant.response.completed` event is the authoritative final response.

`assistant.response.started` indicates the agent loop has begun producing the final user response.

`assistant.response.delta` is reserved for a future streaming-safe response contract. The first implementation must not send unvalidated model deltas to the mobile app and must not send unvalidated deltas to text-to-speech.

The first implementation must wait for the validated structured final response before sending text to text-to-speech.

Pre-validation spoken-response streaming may be added only after a future spec defines a separate streaming-safe response contract with:

- A provider-independent delta schema.
- A validator that can reject raw JSON fragments, tool-call syntax, hidden reasoning, prompts, provider metadata, hidden resources, secrets, and unsafe identifiers before playback.
- Adversarial tests proving unsafe deltas never reach mobile display or text-to-speech.
- A clear policy for provider interruption after partial output.

## Text-To-Speech Streaming

The server must send `spokenResponse` text to the text-to-speech provider through the text-to-speech port.

When the selected text-to-speech provider supports streaming output, the server should emit audio chunks as soon as useful audio is available.

When the selected text-to-speech provider does not support streaming output, the server may emit one or more audio chunks after synthesis completes. The realtime protocol remains chunk-oriented so streaming providers can be adopted without changing the mobile client contract.

`tts.audio.started` must include the output audio format selected by the server.

`tts.audio.chunk` must include ordered audio chunk data or a binary frame associated with the session and sequence number.

`tts.audio.completed` indicates no more speech audio is expected for the current response.

The mobile app must play audio chunks in order and must stop playback on cancellation, session failure, or a newer session replacing the current session.

Generated speech audio must not be durably persisted by the mobile app or API in this first slice.

## Cancellation And Timeouts

The user must be able to cancel the session while recording, while transcription is running, while tool calls are running, while final response generation is running, or while text-to-speech playback is running.

Cancellation must:

- Stop mobile recording.
- Stop mobile playback.
- Tell the server to stop processing when possible.
- Prevent late events from changing the visible state of the cancelled session.
- Emit safe cancellation observability.

The server must enforce configured session, silence, provider, tool-call, and idle timeouts.

Timeouts must fail safely and should produce a spoken response only when there is enough time and provider availability to do so safely.

## Mobile Application Boundaries

Mobile voice behavior must preserve the mobile hexagonal organization:

- UI components own rendering and native interactions.
- Mobile application services own voice session state and view models.
- Mobile ports own microphone capture, realtime transport, audio playback, and runtime configuration.
- Mobile adapters own Expo microphone capture, WebSocket transport, and native audio playback details.
- Bootstrap owns composition.

UI code must not call WebSocket APIs, generated API DTO clients, provider SDKs, or native audio modules directly.

## Security And Privacy

The mobile voice query slice is security-sensitive.

The system must:

- Authenticate the realtime connection.
- Authorize selected tenant and inventory scope before processing audio.
- Reject cross-tenant and cross-inventory attempts.
- Reject hidden-resource access through tool calls.
- Reject state-changing tool calls in the first slice.
- Avoid logging raw audio, raw transcripts, raw prompts, raw provider responses, raw model reasoning, generated speech, credentials, or bearer tokens.
- Avoid returning hidden resource data in transcript, progress, tool, final response, or TTS events.
- Map provider and tool failures to safe user-facing errors.

Read tool executions must follow the safe read audit requirements of the underlying application operation. Voice read audit metadata must not include raw audio, raw transcripts, raw query text, raw prompts, raw tool inputs, raw tool outputs, raw provider responses, generated speech, or hidden resource details.

## Testing

Tests must use fakes for speech-to-text, language inference, text-to-speech, tool catalog, inventory application services, authorization, realtime transport, observability, microphone capture, and audio playback where focused unit or adapter behavior is under test.

Realtime boundary tests must exercise the actual API WebSocket adapter with configured authentication and authorization adapters.

Tests must cover:

- Successful read-only question from audio input through spoken audio output.
- Successful typed-transcript equivalent for deterministic agent-loop tests.
- Partial transcript events.
- Final transcript event.
- Multiple tool-call iterations before final response.
- Safe tool progress events.
- Structured final response validation.
- Rejection of unvalidated `assistant.response.delta` text before mobile display or text-to-speech.
- Text-to-speech chunk streaming.
- Non-streaming text-to-speech fallback using the same chunk-oriented protocol.
- Cancellation during recording, transcription, tool execution, final response generation, and audio playback.
- Forged session IDs, stale client sequence numbers, replayed audio chunks, late audio after cancellation, and event crossover between concurrent sessions.
- Missing, disabled, archived, malformed, unsupported, or unusable provider profiles.
- Speech-to-text failure.
- Language inference failure.
- Malformed model tool call.
- Malformed model final response.
- Text-to-speech failure.
- Unauthorized, unauthenticated, wrong-tenant, wrong-inventory, viewer-hidden-resource, expired-token, malformed-token, and privilege-escalation attempts.
- Model attempts to call write tools or unlisted tools.
- Model attempts to smuggle hidden IDs, authorization claims, or approval claims through tool inputs or final output.
- Hidden ID probing, wrong-inventory asset detail attempts, and count leakage through progress events.
- Voice read audit emission for underlying read operations without leaking transcript, provider, or raw tool content.
- Redaction of raw audio, raw transcripts, raw query text, raw prompts, raw tool inputs, raw tool outputs, raw provider responses, raw model reasoning, generated speech, credentials, bearer tokens, hidden resources, and stack traces from mobile state persistence, debug history, crash reports, analytics, API session metadata, audit, observability, logs, progress events, final responses, and TTS.

## Open Questions

- Which Expo-compatible audio input format should be used first?
- Which mobile audio playback adapter should own streamed chunk buffering?
- Should the first implementation use tap-to-start/tap-to-stop or push-to-talk?
- What future streaming-safe response contract would allow spoken-response deltas before full structured final validation?
- What exact artifact shape should safe asset and location references use in final responses?
