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

The mobile app must expose realtime voice as a global interaction layer anchored to the native bottom voice accessory. Voice is not a primary navigation destination in the production mobile experience.

The collapsed voice accessory is the persistent voice affordance. Activating it must expand an active voice session surface over the current mobile screen without leaving the user's current tab, asset, location, search, or add context. The expanded surface should use a platform-native detent sheet when the mobile runtime provides one. On platforms where the native bottom accessory or native sheet detents are unavailable or constrained, the app may render the same collapsed voice control and session surface through equivalent custom UI, but the product behavior must remain the same.

The expanded session surface must:

- Visually read as an expansion of the bottom voice accessory.
- Keep the current screen context behind the session surface.
- Use platform-native backdrop, grabber, detent, drag, scroll-expansion, and dismiss behavior when rendered through a native sheet.
- Provide at least compact and expanded detents when the native sheet implementation supports multiple detents.
- Provide a clear close/collapse control that does not reset a completed answer unless the user starts a new session.
- Support tap-to-start and tap-to-stop recording for the first slice.
- Keep the primary mic control reachable with one thumb.
- Keep the primary mic/progress control in a bottom action area so compact detents do not strand the main action in sparse scroll content.
- Reserve the scroll body for transcript, response, progress details, diagnostics, and future approval artifacts.
- Avoid using a standalone full-screen Voice route for normal user interaction. An internal modal route may be used as the native sheet implementation detail when the router requires a route to present a platform-native sheet.

Any internal Voice route used for native sheet presentation must not appear as a primary navigation destination. Direct entry to the internal route must fall back to a valid product screen when closed.

The mobile app must show:

- Current tenant and inventory context.
- Recording or listening state.
- Transcription progress when available.
- The full final transcript in the active voice session view when available.
- Safe agent progress events.
- Safe tool-call debug events.
- Final text response.
- Audio playback state.
- Cancellation and failure states.

The mobile state layer must apply safe realtime session events incrementally as they arrive. It must not wait for the WebSocket session to finish before updating visible progress, final transcript, safe tool progress, final response, speech playback state, or safe failure state. This allows the sheet and collapsed voice accessory to reflect transcription, tool execution, model response preparation, and speech playback while the server-side loop is still running.

Safe agent progress events should be summarized as user-facing progress steps rather than exposed as raw event logs. Tool-call events may be displayed in a developer diagnostics panel only when developer diagnostics are explicitly enabled. Diagnostics must be visually secondary, collapsed by default, and must not expose hidden resource data, raw query text, raw transcripts, raw prompts, raw model responses, provider credentials, internal IDs, or internal stack details.

The active voice session view may display the final transcript to the user as ephemeral UI state. This transcript display is not debug history and must not be written to local storage, logs, crash reports, analytics, audit records, or observability metadata before a transcript retention and redaction policy is specified.

Developer diagnostics in the mobile voice surface must be disabled by default and enabled only through explicit runtime configuration. Enabling diagnostics may allow the mobile UI to render sanitized tool-call progress labels and statuses that have already passed through the mobile application redaction boundary, but it must not alter provider prompts, tool availability, authorization, model inputs, session persistence, or raw event retention.

Mobile realtime voice is local-development testable before production mobile authentication exists. A production mobile rollout requires a specified mobile authentication flow and must not rely on `EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN`.

## Realtime Session

The mobile app starts a realtime voice session by opening an authenticated WebSocket to the core API.

Before starting local audio capture, mobile must run a safe provider-profile readiness check through the mobile application layer when tenant-managed provider profiles are available in the composition. The check must use only safe provider profile metadata and must require enabled, credential-configured, successfully tested profiles for speech-to-text, language inference, and text-to-speech. If readiness fails, mobile must not start the recorder or open the realtime WebSocket. It must surface a safe, actionable error that names the missing capabilities without exposing provider credentials, endpoint URLs, raw prompts, raw provider responses, raw audio, or internal IDs.

When mobile can navigate to tenant provider-profile management, a provider-readiness failure in the voice sheet should include a direct safe action to open the Voice providers screen. That action must not include provider IDs, endpoint URLs, credentials, prompt text, raw provider responses, or internal error details in the voice sheet.

Mobile WebSocket handling must tolerate the server closing normally immediately after sending a terminal `session.completed` or `session.failed` event. The transport must drain already queued server messages before treating an `onclose` notification as premature, and premature close errors may include the safe numeric close code but must not surface raw close reason text.

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
- The inventory does not belong to the requested tenant, even if the principal can view both resources independently.
- Required provider profiles are missing, disabled, archived, unsupported, malformed, or have unusable credentials.
- The requested audio format is unsupported.
- The server is unable to enforce timeout, cancellation, or safe observability behavior.

## First Wire Protocol

The first realtime voice wire protocol uses WebSocket over HTTPS or HTTP in local development.

The WebSocket path is:

- `/v1/realtime/voice`

The client authenticates with the same bearer token header used by REST:

- `Authorization: Bearer <token>`

The first implementation uses text WebSocket frames containing JSON messages. Binary frames may be introduced later for lower overhead after the JSON protocol is proven.

Every client and server message contains:

- `type`: stable message type enumeration.
- `sessionId`: omitted only on `session.start` because the server has not created the session yet.
- `seq`: monotonic sender sequence number.

Client message fields:

- `session.start`: `seq`, `tenantId`, `inventoryId`, `source`, `requestedCapabilities`, `inputAudio`, `outputAudio`, optional `clientCorrelationId`.
- `audio.chunk`: `seq`, `sessionId`, `chunkId`, `audioBase64`, `isFinalChunk`.
- `audio.end`: `seq`, `sessionId`.
- `session.cancel`: `seq`, `sessionId`, optional safe `reason`.
- `client.ack`: `seq`, `sessionId`, `ackSeq`.

Server message fields:

- `session.started`: `seq`, `sessionId`, `acceptedInputAudio`, `acceptedOutputAudio`.
- `session.failed`: `seq`, optional `sessionId`, `code`, safe `message`.
- `transcript.delta`: `seq`, `sessionId`, `text`.
- `transcript.final`: `seq`, `sessionId`, `text`.
- `agent.progress`: `seq`, `sessionId`, `status`, safe `message`.
- `tool.call.started`: `seq`, `sessionId`, `toolCallId`, `toolLabel`, `status`.
- `tool.call.completed`: `seq`, `sessionId`, `toolCallId`, `toolLabel`, `status`.
- `tool.call.failed`: `seq`, `sessionId`, `toolCallId`, `toolLabel`, `code`, safe `message`.
- `assistant.response.started`: `seq`, `sessionId`, `responseId`.
- `assistant.response.delta`: reserved and must not be emitted by the first implementation.
- `assistant.response.completed`: `seq`, `sessionId`, structured final response.
- `tts.audio.started`: `seq`, `sessionId`, `format`.
- `tts.audio.chunk`: `seq`, `sessionId`, `chunkId`, `audioBase64`, `isFinalChunk`.
- `tts.audio.completed`: `seq`, `sessionId`.
- `session.completed`: `seq`, `sessionId`.
- `session.cancelled`: `seq`, `sessionId`.
- `session.failed`: `seq`, `sessionId`, `code`, safe `message`.

The first implementation may use a development provider profile set supplied at API composition time. The development provider set must be disabled by default and enabled only through explicit runtime configuration such as `STUFF_STASH_VOICE_DEV_FAKE_ENABLED=true`.

Development fake providers may return deterministic transcript, language, and speech-like byte chunks for local end-to-end testing. They must live behind the same project-owned provider ports as real adapters and must not be enabled implicitly in production-shaped configuration.

The first real Google-hosted provider bridge may be enabled through explicit runtime configuration while tenant-managed provider profiles are still pending. This bridge must:

- Use Google Application Default Credentials or equivalent OAuth credentials resolved only in the API process.
- For local smoke testing, the API may accept a short-lived Google OAuth bearer token through runtime configuration. This token path is process-local, must not be persisted, and is not a tenant-managed provider-profile mechanism.
- Require an explicit Google Cloud project ID.
- Use Vertex AI Gemini through the speech-to-text port for the first mobile native audio path because Expo SDK 55 native recording defaults to MPEG-4 AAC (`.m4a`), while Google Cloud Speech-to-Text does not support M4A/AAC as a direct input encoding.
- Use Vertex AI Gemini through the language inference port for the agent loop.
- Use Google Cloud Text-to-Speech through the text-to-speech port and return MP3 chunks to the mobile app.
- Prefer the cheapest fit-for-purpose Google models for local smoke testing: Gemini Flash-Lite for Gemini calls and Standard Cloud Text-to-Speech voices unless quality requirements justify a more expensive provider profile.
- Keep Google SDK, REST, OAuth, endpoint, and response-shape details inside provider adapters.
- Fail closed at startup or session start when required Google configuration or credentials are unavailable.

Tenant-managed provider-profile persistence and UI management remain separate implementation work, but the realtime application service must still depend on project-owned provider ports and must not depend on concrete provider adapters.

## Client Audio Input

The first mobile implementation records audio through Expo's standard audio recording API and sends the completed recording to the API as one or more `audio.chunk` messages. This preserves the chunked server protocol while acknowledging that the standard Expo audio API records to a local cache file rather than exposing low-latency PCM callbacks to JavaScript.

Expo Audio 55 exposes recorder construction primarily through React hooks. Because the realtime voice recorder is a mobile adapter composed outside React UI components, the first slice may isolate any required non-hook Expo recorder construction inside the Expo voice adapter only. UI code, application services, and other adapters must not import Expo private modules. This compatibility exception must be removed when Expo provides a stable public non-hook recorder factory or the voice recorder is redesigned around a hook-owned adapter boundary.

The first accepted input format is:

- Container or raw encoding: platform-recorded MPEG-4 AAC (`audio/mp4`) for native mobile; deterministic text fixtures may be used only in tests through fake microphone and speech-to-text adapters.
- Sample rate: requested 44.1 kHz when the Expo adapter supports specifying it.
- Channel count: mono.
- Chunk duration target: mobile should flush chunks of approximately 256 KiB after recording completes; future low-latency adapters may target 100-250 ms chunks.
- Maximum chunk byte size: 512 KiB before base64 encoding.
- End-of-utterance behavior: tap-to-stop sends the final audio chunk followed by `audio.end`.

The mobile app must not record audio before the user intentionally starts a voice session.

The mobile app must stop audio capture when the user cancels, the session ends, the server rejects the session, or the timeout is reached.

Raw audio must not be durably persisted by the mobile app or API in this first slice.

When a platform audio API requires file-backed recording or playback, the adapter may use the platform cache directory only as a transient implementation detail. The adapter must delete recorder and playback files after use, must perform best-effort stale cleanup on later voice operations, and must not let a failed player disposal or stale-file delete prevent cleanup of other voice temp files.

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
- `action.plan.proposed`
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

The mobile transport must validate server event sequence monotonicity and session binding before forwarding events to application state. `session.failed` may omit `sessionId` only before `session.started`; after a session is established, server events with missing, stale, or mismatched session metadata must fail the local session safely and must not update UI or play audio.

Every client message after `session.start` must be bound to the authenticated WebSocket connection and server-created session. The server must reject forged session IDs, stale client sequence numbers, replayed audio chunks, messages from cancelled sessions, messages from completed sessions, and any attempt to change tenant or inventory scope after session authorization.

Client messages must include monotonic per-session sequence metadata once a session is established. The server must use that metadata only for ordering, replay rejection, flow control, and safe diagnostics; it must not treat client sequence metadata as authorization.

## Action Plan Events

`action.plan.proposed` contains the safe persisted action-plan review payload for mobile. It must include plan ID, confirmation summary, command summaries, risk summaries, and no raw transcript, raw prompt, raw model response, credentials, provider session IDs, hidden resource data, or approval claims.

When the API emits `action.plan.proposed`, the mobile app must enter the `review` stage and show the proposal in the voice sheet. The first mobile slice may show disabled or not-yet-wired approval actions, but it must not silently execute the plan. The final spoken response for a proposed write may explain that the user should review the suggested change.

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
- Provide the language model with only the tools allowed for the current mobile slice.
- Treat model output as untrusted.
- Validate tool-call requests before execution.
- Authorize every tool execution through the owning application service and authorization port.
- Allow multiple tool-call iterations when needed.
- Stop when the model produces a structured final response, proposes an action plan for user review, a safe failure occurs, cancellation is requested, or the session times out.
- Instruct the model to use tool results as the source of truth and to avoid inventing locations, quantities, or inventory contents that are not present in tool results.

The first query loop exposed only read-only tools to the model. The first approval-backed mobile slice may add exactly one non-mutating write-intent tool: `propose_action_plan`.

The first read-only tools are:

- Search authorized assets.
- Get asset detail.
- List assets in a location.
- List root-level assets in an inventory.
- List authorized assets by safe filters such as kind, lifecycle state, and optional parent or location title.

Tool descriptors must use project-owned names, descriptions, read-only markers, and parameter metadata. Provider adapters may translate that metadata into provider-native schemas, but provider-specific tool declaration types must not cross the language inference port.

`propose_action_plan` is not an inventory mutation tool. It may persist a proposed action plan through the action-plan application boundary and return a safe plan summary for mobile review. It must not execute asset, location, tenant, sharing, provider-profile, audit mutation, import/export, or raw repository operations. Its arguments must be bounded, typed, validated by the application boundary, and free of raw prompts, raw transcripts, raw provider responses, credentials, bearer tokens, provider session IDs, hidden resource data, and approval claims.

The loop must not expose direct write tools, provider profile tools, tenant configuration tools, sharing tools, audit mutation tools, import/export tools, or raw repository access. Any future direct execution must go through an approved action-plan execution service.

Tool results provided to the language model must be structured, safe, and useful enough for accurate answers. For visible assets, read-only tool output should include:

- Title.
- Kind.
- Description when present.
- Inventory name.
- Lifecycle state.
- Parent title and parent kind when present.
- Nearest containing location title when present.
- Human-readable containment path from outermost visible container or location to the asset.
- Custom fields only after a field sensitivity and provider-disclosure policy exists. The first improved catalog must omit custom field values from cloud-provider tool results.
- Match metadata that helps the model understand why a result was returned.

Tool results must not include raw authorization decisions, hidden resources, bearer tokens, provider credentials, raw prompts, raw model responses, raw audio, generated speech, custom field values before a sensitivity policy exists, internal stack traces, or infrastructure details. Internal resource identifiers may be provided to the in-process agent loop only when needed to chain read-only tool calls, and final user-facing responses must not speak or display those identifiers.

The first implementation may expose structured tool results as compact JSON strings through the language inference port while provider-native tool schemas are still evolving. The JSON shape must remain project-owned and provider-independent.

The first improved implementation may expose a generic filtered list tool that covers list-by-location, list-root-level, and list-by-kind behavior before separate public tool names are added. The first voice tool catalog must support at least these user intents accurately for visible inventory data:

- "Where is my water bottle?" returns the containing location or containment path when available.
- "What items do I have?" returns visible item-kind assets in the selected inventory.
- "What is in the office?" returns visible children of the matching location or container when available.

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

## Prompt Templates

The first real provider adapters may use a fixed project-owned prompt template for the voice loop.

Provider adapters that support native tool or function calling must use the provider-native tool calling mechanism for tool selection instead of asking the model to emit project-defined tool-call JSON in text. For Gemini on Vertex AI, the adapter must send the project-owned read-only tool catalog as `functionDeclarations` with OpenAPI-compatible parameter schemas, parse returned `functionCall` parts into project-owned `AgentToolCall` values, and send tool outputs back as `functionResponse` parts on later turns.

Native provider tool calling is an adapter concern. Application services, domain services, REST adapters, mobile clients, and tool execution code must continue to use project-owned tool descriptors, tool calls, tool results, and structured final responses. Provider-native tool declaration, function-call, and function-response shapes must not leak across the language inference port.

The agent loop must allow multiple distinct read-only tool calls across turns when needed to answer the user accurately. Loop control must be owned by the application agent loop rather than provider-specific adapter shortcuts. If the model asks for an identical tool name and argument set that has already been executed in the same session, the loop must not execute the duplicate call again, but it must continue executing other unseen tool calls in the same model turn. It may request one explicit finalization-only model turn using the existing tool results and no tool catalog; if the model still does not produce a valid final response, the session must fail safely.

Provider adapters may continue to use structured JSON output for final responses when the provider supports it. Native tool calling must not loosen the final response validator, read-only/write boundaries, tenant and inventory scoping, or redaction rules.

Future tenant-managed provider profiles must support model-specific prompt template configuration because smaller or local models may need different instructions, output examples, or schema wording. Prompt templates must be configuration data resolved through the provider-profile/application boundary, not hard-coded provider adapter behavior.

Prompt template customization must preserve required security and product guardrails:

- The structured response contract.
- The allowed tool catalog.
- Tenant and inventory scope.
- Read-only/write confirmation boundaries.
- Safe error behavior.
- Redaction and retention rules.
- Prohibition on exposing hidden identifiers, credentials, raw prompts, raw transcripts, raw audio, generated speech, or hidden resources.

Provider-specific prompt templates may tune wording and examples, but they must not loosen authorization, tenancy, tool validation, action-plan, confirmation, or audit requirements.

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

The first mobile cancellation implementation must expose cancellation through the mobile voice application boundary, not only through sheet dismissal. Cancelling while still recording must stop the recorder and audio player, must not open a realtime WebSocket, and must move the mobile session into a terminal cancelled state. Cancelling after a WebSocket session has been established must send a `session.cancel` client message with the active server session ID when that can be done safely, close the socket, stop local playback, and make later server events unable to update visible state for the cancelled session.

When the API receives `session.cancel` before `audio.end`, it must mark the realtime session as cancelled and emit a terminal `session.cancelled` server event rather than reporting a user cancellation as `session.failed`. If the connection disappears after `audio.end` while provider work is already in flight, the server should cancel through the request context where practical and record a safe terminal outcome without relying on the client to keep listening for the acknowledgement.

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
