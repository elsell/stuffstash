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
- Reviewable action-plan proposal, approval, cancellation, and approved execution for supported write requests.

This spec does not include:

- Unreviewed or model-direct write actions.
- Production-polished approval UI beyond the safe review controls described here.
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

Language inference must be phase-aware. Provider adapters must not rely on prompt text alone when a provider offers native constrained output. For Gemini, final-response turns and action-plan planner turns must request `application/json` output with a provider-native response schema. Read turns may use native function calling with a required tool-call configuration when the loop needs inventory context. The planner turn must not expose `propose_action_plan` as a provider-callable function; instead, the provider returns a schema-constrained action-plan object that the API converts into the same internal action-plan proposal path used by other adapters. The API remains responsible for authorization, visible-ID validation, tenant and inventory scoping, action-plan persistence, approval, audit, and execution.

For clear write requests, the loop should follow an explicit state sequence:

1. Gather required inventory context with read-only tools.
2. Ask the language provider for a constrained action-plan object when enough context exists to propose a safe reviewable plan.
3. Validate and persist the action plan through the application service.
4. Pause the session and emit `action.plan.proposed` until the user approves or cancels.

The loop must not continue calling the language model after a valid action plan has been proposed. If a constrained planner output is invalid, the loop may retry with safe structured repair context, but it must not let a later final answer override a previously valid plan.

Approved voice create-asset action plans must execute through the same application create command semantics as REST and mobile Add. If a reviewed create command places the new asset inside an existing item, execution must promote that parent item into a container in the same authoritative unit of work, with audit history for the parent promotion and child creation. Voice adapters and websocket handlers must not perform client-side or transport-layer kind mutation.

The API may complete obvious unsafe or under-specified transcripts locally after transcription and before language inference. This includes provider credential requests, broad destructive database or inventory wipe requests, and vague deictic move destinations such as "over there" or "to the side" when no concrete place is named. Broad destructive inventory requests include clear requests to delete, erase, remove, clear, empty, reset, or purge all assets, items, things, stuff, records, entries, or inventory contents. These local completions must return structured safe final or clarification responses and must not call the language provider.

The internal agent loop may use the project-owned tool catalog specified by `specs/agent-model/mcp-agent-tools.spec.md`, but it must call tools in-process through application services and ports. It must not call the public MCP transport for this first mobile loop.

## User Workflows

The first supported mobile workflow was a read-only question about the selected tenant and inventory, and that remains the direct-completion path.

Examples:

- "Where are my tools?"
- "Where is the fertilizer?"
- "What is in the garage?"
- "Do I have any batteries?"

Read-only workflows must not create, update, move, archive, restore, delete, import, export, share, or configure anything.

Supported state-changing workflows must use action-plan review. If the user asks for a supported create, move, archive, restore, checkout, or return action, the system must gather enough safe inventory context, propose a reviewable action plan, pause for explicit approval or cancellation, and execute only the approved plan through application services. Unsupported, unsafe, or under-specified state-changing requests must produce a safe clarification or refusal rather than executing a change.

## Mobile Interaction

The mobile app must expose realtime voice as a global interaction layer anchored to the native bottom voice accessory. Voice is not a primary navigation destination in the production mobile experience.

The collapsed voice accessory is the persistent voice affordance. Activating it must expand an active voice session surface over the current mobile screen without leaving the user's current tab, asset, location, search, or add context. The expanded surface should use a platform-native detent sheet when the mobile runtime provides one. On platforms where the native bottom accessory or native sheet detents are unavailable or constrained, the app may render the same collapsed voice control and session surface through equivalent custom UI, but the product behavior must remain the same.

When a realtime session is active or has a terminal result, the collapsed voice accessory should reflect the latest safe session state instead of only generic stage labels. During active processing, it should prefer the project-owned `agent.progress` phase vocabulary when available, such as understanding, exploring, planning, answering, and recovering, so the user can see the loop moving without exposing internals. It may show the current safe progress label, review-needed status, safe failure summary, and the final user-facing spoken response. It must not show partial transcript text, raw transcripts, diagnostics, tool labels, tool arguments, provider errors, prompts, credentials, internal IDs, or stack traces.

When a session completes with a clarification response and the same-session follow-up window is still available, mobile must present that state as needing detail rather than as a completed answer. The collapsed accessory, expanded sheet title, bottom hint, and microphone accessibility label should make the next expected action clear: answer the follow-up in the same conversation. Mobile state must model same-session follow-up availability separately from the assistant response kind so a stale or closed WebSocket cannot keep advertising an unavailable follow-up. If the follow-up socket closes before the user sends the next audio turn, mobile must remove the same-session follow-up affordance and treat the microphone as starting a fresh voice interaction.

When the API ends a clarification chain with `session.failed` code `clarification_turn_limit`, mobile must present a safe conversation-specific failure state rather than a generic provider failure. The user-facing copy should explain that Stuff Stash needs a fresh voice request, and it must not direct the user to provider setup or diagnostics unless another provider-specific failure code is present.

The expanded session surface must:

- Visually read as an expansion of the bottom voice accessory.
- Keep the current screen context behind the session surface.
- Use platform-native backdrop, grabber, detent, drag, scroll-expansion, and dismiss behavior when rendered through a native sheet.
- Provide at least compact and expanded detents when the native sheet implementation supports multiple detents.
- Provide a clear close/collapse control that does not reset a completed answer unless the user starts a new session.
- Support tap-to-start and tap-to-stop recording for the first slice.
- Keep the primary mic control reachable with one thumb.
- Keep the primary mic/progress control in a bottom action area so compact detents do not strand the main action in sparse scroll content.
- While recording, present the primary control as a single send/finish action that also displays live audio activity. The activity signal must be driven by native recorder metering when the runtime exposes it; the UI must not use arbitrary or decorative bouncing that is unrelated to microphone input.
- After the user sends audio and before the session reaches review, completion, cancellation, or failure, keep a visible working indicator in both the sheet and collapsed accessory. The primary control must not return to a generic ready/start microphone while Stuff Stash is transcribing, checking inventory, applying an approved plan, preparing a response, or playing speech.
- When a proposed action plan is awaiting review, the bottom action area must replace the mic/new-session action with explicit approve and cancel controls so a compact detent keeps the decision reachable and the user cannot accidentally reset the pending review by starting another voice session.
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

Safe agent progress events should be summarized as lightweight user-facing status rather than exposed as raw event logs. The active voice sheet may render a compact bounded progress trace for multi-step understanding, exploration, planning, answering, and recovery while no action-plan review body is present, so users can see the smart loop moving without enabling developer diagnostics. The active voice sheet must not render progress as an expanding table above the action-plan review because that can push the approval prompt and command list below the visible area. When a proposed action plan exists, the confirmation prompt and command list must remain the primary visible body content, with current progress represented by compact chrome such as the bottom status label, a slim activity bar, or an activity indicator. Tool-call events may be displayed in a developer diagnostics panel only when developer diagnostics are explicitly enabled. Diagnostics must be visually secondary, collapsed by default, placed after the review content, and must not expose hidden resource data, raw query text, raw transcripts, raw prompts, raw model responses, provider credentials, internal IDs, or internal stack details.

When a speech, language, or speech-output provider fails after earlier session work has succeeded, the client-facing failure copy must describe the failed stage without implying that no model or tool calls happened. A language-inference failure can occur on a continuation turn after successful tool calls. In developer-diagnostic sessions, the API must emit a sanitized diagnostic before returning the safe failure so the mobile sheet can show the failed turn number, whether the call was final-only, prior safe tool names, and a safe error category. This diagnostic must not include raw provider response bodies, prompts, transcripts, credentials, endpoint URLs, stack traces, hidden inventory data, or bearer material.

The mobile state layer may maintain a bounded ephemeral progress timeline for the active session for compact summaries and developer inspection. The visible sheet should show the current milestone rather than all milestones by default. The timeline may include safe client and server milestones such as audio upload, API connection, transcription completion, safe `agent.progress` messages, response preparation, speech playback, cancellation, completion, and safe failure. It must not include raw tool arguments, raw provider output, internal IDs, provider errors, prompts, credentials, bearer material, audio bytes, stack traces, or partial transcript history. The mobile application boundary must still redact obvious unsafe terms from progress labels before they reach visible state, even though the API is also responsible for safe progress messages. Duplicate adjacent progress labels should collapse into one visible step so compact sheets stay readable.

The active voice session view may display the final transcript to the user as ephemeral UI state. This transcript display is not debug history and must not be written to local storage, logs, crash reports, analytics, audit records, or observability metadata before a transcript retention and redaction policy is specified.

When the realtime voice sheet displays a proposed action plan, each visible plan row that represents an item, container, or location may offer an inline `Add photos` action only when the row is a newly created asset or a move row that resolves to a concrete reviewed asset after execution. Archive, restore, checkout, and return rows must not offer photo staging. This action must use the same native photo-selection capability as the Add flow, including camera and photo-library choices, native permissions, supported image MIME types, and in-session image previews. Selected photos are draft UI state scoped to the active proposed plan and command row; they must not be uploaded, persisted, logged, added to diagnostics, or sent to the language provider before the user approves the plan.

Mobile must treat action-plan review text as user-visible structured output. Before confirmation summaries, command summaries, command titles, parent titles, and risk text enter visible mobile state, the mobile application boundary must redact obvious unsafe structured values such as credentials, bearer material, provider session IDs, URLs, raw prompt assignments, raw transcript assignments, raw provider-response assignments, and internal asset or tenant ID assignments. This redaction must preserve ordinary user inventory titles that merely contain words such as "stack trace" or "raw query" without credential-like or assignment-like structure, because those may be legitimate asset names.

When the user approves a plan with staged photos, mobile must send only safe photo metadata with the approval message: reviewed command ID, file name, content type, and byte size. It must not send image bytes, base64 content, local file URIs, provider data, raw prompts, transcripts, credentials, internal stack details, or unrelated inventory data in the approval message.

After the user approves a plan, the API execution event must include a safe command-result mapping for each executed asset command, including create, move, archive, restore, checkout, and return commands. Each command result must include only the reviewed command ID, operation, concrete asset ID, and asset kind. Command results are safe execution feedback for the realtime client and are also the only source from which photo upload intents may be derived. When approved photo metadata was provided for a command result that can receive photos, the API must also return an attachment upload intent for that exact command. Each upload intent must include only the reviewed command ID, concrete asset ID, safe file metadata, upload ID, reserved attachment ID, upload method, opaque upload URL, upload headers/form fields, and expiration. The event must not include credentials, provider data, hidden assets, raw prompts, raw transcripts, asset titles, local device URIs, image bytes, base64 content, or unrelated inventory data.

Mobile must upload staged photos only after receiving upload intents for the same active plan and command row. Uploads must use the existing authorized asset attachment API and direct-upload completion flow, so server-side authorization, tenant scoping, inventory scoping, validation, audit, and media constraints still apply. When the API returns an HTTPS direct-upload target, mobile must upload the original selected file to that target and then complete the direct upload through the API instead of sending base64 content through the JSON attachment route. Private local-development HTTP direct-upload targets for Garage testing and the explicit local-development fallback sentinel `stuffstash-local://direct-uploads/{uploadId}` are allowed only when mobile runtime configuration explicitly enables local direct-upload targets. Mobile may fall back to the JSON attachment route only for that local sentinel while the local target setting is enabled, or when direct upload is unavailable. The mobile realtime voice transport must reject attachment upload intents that use any other URL scheme, public cleartext HTTP target, or local-development target while the local setting is disabled before they enter voice session state. Because iOS and Android picker URIs may be permission-scoped or temporary, mobile may keep bounded base64 fallback content captured by the native picker at selection time for this local fallback. It must not compress selected images, send image bytes to the language provider, put image bytes in diagnostics, or persist image bytes outside the authorized attachment upload request. If one or more photo uploads fail after the plan itself succeeds, mobile must keep the inventory change as applied, surface a partial-success photo status, preserve a safe retry path when practical, and show a safe stage-specific photo failure message instead of a generic successful update label. Photo upload failures must not mark the approved inventory action plan as failed.

When `transcript.delta` events are available, mobile may display the latest partial transcript as ephemeral UI state while processing continues. The final `transcript.final` event must replace the partial transcript in visible state and must not append partial transcript history to diagnostics or durable storage.

Developer diagnostics in the mobile voice surface must be disabled by default and enabled only through explicit runtime configuration. Enabling diagnostics may allow the mobile UI to render sanitized tool-call progress labels and statuses that have already passed through the mobile application redaction boundary, but it must not alter provider prompts, tool availability, authorization, model inputs, session persistence, or raw event retention.

Mobile realtime voice must use the same mobile OIDC token provider as REST API calls. It must not rely on `EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN`.

## Realtime Session

The mobile app starts a realtime voice session by opening an authenticated WebSocket to the core API.

Before starting local audio capture, mobile must run a safe provider-profile readiness check through the mobile application layer when tenant-managed provider profiles are available in the composition. The check must use only safe provider profile metadata and must require enabled, credential-configured, successfully tested profiles for speech-to-text, language inference, and text-to-speech. If readiness fails, mobile must not start the recorder or open the realtime WebSocket. It must surface a safe, actionable error that names the missing capabilities without exposing provider credentials, endpoint URLs, raw prompts, raw provider responses, raw audio, or internal IDs. The failed voice session state must preserve the current safe tenant and inventory names when they are already known, so recovery UI still tells the user which inventory context the blocked voice action would have used.

When mobile can navigate to tenant provider-profile management, a provider-readiness failure in the voice sheet should include a direct safe action to open the Voice providers screen. That action must not include provider IDs, endpoint URLs, credentials, prompt text, raw provider responses, or internal error details in the voice sheet.

Mobile provider management must use the voice pipeline as the primary setup model:

- Speech input.
- Agent brain.
- Spoken output.

Each slot must show its selected profile, readiness state, and the next best action. The setup view must distinguish selected profiles from unselected duplicates, missing credentials, disabled profiles, archived profiles, and profiles that need a fresh test. A flat list of provider profile cards may exist only as a secondary profile inventory or advanced view.

The setup screen must help users recover from the real failure classes that block voice:

- No selected profile for a required slot.
- Selected profile is disabled or archived.
- Selected profile has missing credentials.
- Selected profile has not been tested or was changed since the last successful test.
- Multiple eligible profiles exist for the same capability and the user has not made an explicit selection.
- Provider-stage failures returned by the API, such as `speech_to_text_failed`, `language_inference_failed`, or `text_to_speech_failed`.

When the API exposes tenant voice provider configuration, mobile must use that configuration and diagnostics rather than inferring readiness solely from the profile list. Mobile may keep a compatibility inference path only for older API builds, and that path must visually indicate that selection is implicit.

When the API returns a safe provider-stage failure code such as `speech_to_text_failed`, `language_inference_failed`, or `text_to_speech_failed`, mobile should preserve the code in voice session state, show a user-actionable stage-specific message, and offer the same Voice providers recovery action. The sheet must not render raw provider errors, prompts, transcripts, audio, generated speech, credentials, endpoint URLs, provider IDs, or stack traces.

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
- Follow-up audio after a same-session clarification uses the same `audio.chunk` and `audio.end` messages with the existing `sessionId`; it must not start a new realtime session unless the prior session reached a terminal non-clarification outcome. Each completed clarification turn must settle back to the mobile UI even when the server keeps the socket open for another bounded clarification follow-up. If that open socket closes before the user sends follow-up audio, the mobile transport must stop advertising same-session follow-up availability so the next microphone action cannot write to a stale session.
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
- Application Default Credentials are the preferred runtime credential mode for local development and production-shaped deployments. Local development should use `gcloud auth application-default login` plus a configured quota project, while hosted deployments should resolve ADC from the attached workload identity or service account.
- For local smoke testing, the API may accept a short-lived Google OAuth bearer token through runtime configuration only when an explicit access-token credential mode is selected. This token path is process-local, must not be persisted, must not be the default, and is not a tenant-managed provider-profile mechanism.
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
- `action.plan.approve`
- `action.plan.cancel`
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
- `action.plan.approved`
- `action.plan.cancelled`
- `action.plan.executed`
- `action.plan.failed`
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

`action.plan.proposed` contains the safe persisted action-plan review payload for mobile. It must include plan ID, confirmation summary, command summaries, risk summaries, and no raw transcript, raw prompt, raw model response, credentials, provider session IDs, hidden resource data, or approval claims. For existing-asset commands such as move, archive, restore, checkout, and return, the API should enrich the mobile review command with the authorized visible asset title and kind instead of relying only on provider-written command summary text.

When the API emits `action.plan.proposed`, the mobile app must enter the `review` stage and show the proposal in the voice sheet. The first mobile slice may show disabled or not-yet-wired approval actions, but it must not silently execute the plan. The final spoken response for a proposed write may explain that the user should review the suggested change.

The next mobile review slice must let the user approve or cancel the proposed action plan from the voice sheet using the existing realtime WebSocket session. The mobile client must send `action.plan.approve` or `action.plan.cancel` with the server-created session ID, the proposed plan ID, and monotonic client sequence metadata. The client must not send command arguments, approval claims, tenant IDs, inventory IDs, credentials, prompt text, transcript text, or model output in the review decision message.

When the API receives `action.plan.approve`, it must call the action-plan application approval boundary scoped to the authenticated principal, session tenant, and session inventory. Approval transitions the persisted plan from `proposed` to `approved`. The first execution slice may then execute an approved single create command through the action-plan execution service. The API must emit `action.plan.approved` with the safe plan ID and status after approval succeeds, then emit `action.plan.executed` or `action.plan.failed` after execution finishes.

When the API receives `action.plan.cancel`, it must call the action-plan application cancellation boundary scoped to the authenticated principal, session tenant, and session inventory. Cancellation transitions the persisted plan from `proposed` to `cancelled`. The API must emit `action.plan.cancelled` with the safe plan ID and status after the transition succeeds. Cancellation messages must not include photo attachment metadata or other approval-only payloads.

The server must reject review decisions with stale sequence numbers, forged session IDs, missing plan IDs, plan IDs that are not scoped to the session tenant and inventory, unauthorized principals, terminal plans, or malformed message types. Safe rejection must not expose hidden plan existence, raw model output, command arguments, credentials, prompt text, transcript text, provider response details, stack traces, or authorization internals.

After receiving `action.plan.approved`, mobile must show that the approved change is being applied and keep duplicate decisions disabled. After receiving `action.plan.cancelled`, `action.plan.executed`, or `action.plan.failed`, mobile must leave the pending review state and show a safe terminal state. Terminal review states must not keep stale pending-decision flags that can suppress normal terminal controls or misrepresent the session as still waiting on the user's earlier tap. Failure states must not expose command arguments, stack traces, hidden resources, provider output, prompts, transcripts, or authorization details. Cancellation messaging must make clear that no change was made.

## Transcript Events

Speech-to-text may emit partial and final transcript events.

`transcript.delta` may contain a partial transcript. Partial transcripts are for immediate mobile feedback only and must not be treated as final user intent.

`transcript.final` contains the transcript text the agent loop may use for intent interpretation. If the speech-to-text provider cannot stream partials, the server may emit only `transcript.final`.

Raw user transcripts must not be durably persisted before a transcript retention and redaction policy is specified.

User transcripts are ephemeral UI and in-memory agent-loop state only in the first slice. Raw user transcript text must not be written to mobile local storage, debug event history, crash reports, analytics, audit records, observability metadata, API session metadata, logs, or provider profile test records before a retention and redaction policy is specified.

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

The Go application service owns the realtime voice agent loop. The loop should use a small graph-like state machine inspired by durable agent runtimes: explicit state, named steps, bounded turns, safe progress events, model-visible tool results, and terminal outcomes. The implementation must remain project-owned Go code behind Stuff Stash ports and must not depend on Python, JavaScript, provider-hosted agent runtimes, LangGraph, LangChain, or provider-specific agent SDKs for core control flow.

The default realtime voice path must be this graph-like smart loop. The mobile voice experience must not maintain a separate one-shot query path that bypasses clarification handling, server-selected exploration, action-plan review, loop repair, or tool-result grounding for normal user requests.

The first graph-like loop nodes are:

- `transcribe`: convert audio into an ephemeral final transcript.
- `understand`: classify the safe request shape and reject obvious unsafe or under-specified local requests before provider calls.
- `explore`: run bounded read-only inventory lookup through project-owned tools before final answers or planner turns when inventory context is needed.
- `agent`: request the next model turn from the configured language provider.
- `tools`: validate and execute one or more requested project-owned tools.
- `plan`: request a constrained action-plan object for supported write requests once enough context exists.
- `finalize`: validate a structured final response, synthesize speech, and complete the session.
- `recover`: produce a bounded safe final response when the loop cannot make more progress but can still explain the outcome without leaking internals.

Safe `agent.progress` statuses must use a bounded product-owned vocabulary so mobile can render phase-aware UI without provider-specific details. The first statuses are `understanding`, `exploring`, `planning`, `reviewing`, `answering`, and `recovering`. The server may add safe message text, but it must not expose raw prompts, raw transcripts, raw tool arguments, raw model output, provider errors, credentials, internal IDs, or hidden inventory data in progress events.

The `understand` node must emit the `understanding` progress status before any terminal local completion, including locally detected unsafe requests, unsupported provider-credential requests, destructive inventory/database requests, and under-specified deictic move destinations. These local completions must still avoid calling the language provider, but mobile should see the same graph phase vocabulary as provider-backed sessions.

Recoverable tool-call failures must not fail the whole voice session by default. If a model asks for an unknown tool, malformed tool arguments, unsafe proposal arguments, a duplicate exact tool call, or another expected validation failure, the loop must emit a safe `tool.call.failed` event and append a structured error tool result back to the model. The model then gets another turn to repair the tool call, ask for clarification, or produce a safe unsupported-action response. If the model produces a final clarification asking whether to create a clear missing destination for a write request, such as "I can't find the second shelf in the big cabinet in the kitchen. Do you want me to create it?", after a prior authorized read result has already shown the requested source item, the loop must treat that as a recoverable agent-contract failure rather than a terminal final response. It should append safe repair feedback and give the model another bounded tool-capable turn so it can call `propose_action_plan`. If the source item itself is not visible in prior read results, the loop must not force an action plan. Fatal failures such as provider transport errors, authentication or authorization failures, context cancellation, text-to-speech failures, and persistence failures may still terminate the session with a safe failure code.

For move-style utterances, if a read tool search for the requested source item returns no visible matches, the loop must fall forward with a structured clarification before planner mode. It must not ask the model to create the source as a side effect of a move request, and it must not spend repeated repair turns on action plans that invent or create the missing source. Missing destinations are different: once the source item is visible, clear missing destination names should become reviewable create-and-move action plans.

The structured error tool result must be provider-independent JSON and must include only safe fields such as the tool name, a stable error code, a short safe message, and whether the model may retry. It must not include raw provider output, stack traces, raw prompts, raw transcript text, credentials, bearer tokens, hidden resource data, exact unauthorized IDs, or unredacted tool arguments.

The loop must maintain a bounded remaining-step budget. If the model continues asking for tools when no more tool execution budget remains, the loop should request one final-only model turn using the accumulated tool results and no tool catalog. If the model still does not produce a valid structured final response, the `recover` node must complete with a safe final response instead of returning a generic transport failure whenever text-to-speech is available.

If a model turn returns neither a valid structured final response nor any tool calls, the loop must treat that as a recoverable agent-contract failure. It should enter the `recover` node, emit safe recovering/answering progress, synthesize the bounded safe-failure response when text-to-speech is available, and complete the session rather than surfacing a generic provider input failure to mobile.

The loop should optimize for low-friction autonomy. For write-like requests, it should gather enough visible context through read tools to propose a complete approval plan in one session when practical, including multiple dependent commands. It should ask a clarification only when the requested object, destination, or command cannot be resolved safely or unambiguously from visible authorized data.

Clarification is a continuation state, not a forced session reset. When a final structured response has kind `clarification`, the WebSocket handler may emit the response and speech, then keep the same session open for a bounded follow-up audio turn. The follow-up transcript must be evaluated by the same application loop with the same tenant, inventory, principal, provider profiles, and safe session context. The session must close after a non-clarification terminal response, action-plan review, cancellation, failure, timeout, or a bounded clarification turn limit. The first clarification continuation limit is three user audio turns per realtime session, after which the API must fail safely or produce a safe final response rather than continuing indefinitely.

Same-session clarification follow-ups must carry bounded safe conversation context through the project-owned language inference port. The context may include prior final user transcripts and assistant response kind/display text from the same realtime session. The first context window is the last six raw same-session conversation turns before role filtering, then filtered to user and assistant roles only. Each retained turn must bound text to 500 characters and kind to 80 characters after redacting obvious credential, bearer, and provider-session material. It must not include raw audio, partial transcripts, raw prompts, raw model responses, provider diagnostics, hidden inventory data, tool arguments, tool results, credentials, bearer material, endpoint URLs, stack traces, internal IDs, malformed roles, or older out-of-window turns. Provider adapters may render this context into provider-native prompt/content shapes, but provider-specific conversation message types must not cross the language inference port.

For same-session clarification follow-ups, the application loop may derive a bounded effective transcript from the latest safe prior user intent and the current final transcript. The prior intent may be a read question or a write request when it was followed by an assistant clarification. The effective transcript is for loop classification, tool selection, and language inference input only; the client-facing `transcript.final` event must still contain the exact current final transcript. The derived text must use only bounded safe same-session context and must not include raw audio, partial transcripts, raw prompts, raw model responses, provider diagnostics, hidden inventory data, tool arguments, tool results, credentials, bearer material, endpoint URLs, stack traces, or internal IDs. If there is no prior actionable user intent followed by an assistant clarification, the loop must use the current final transcript alone.

When the effective transcript combines a generic read question with a concrete follow-up answer, such as "Where is it? Follow-up answer: Water bottle", "What's in it? Follow-up answer: Toolbox", or "When did I move it? Follow-up answer: Water bottle", read-tool query selection should use the concrete follow-up answer as the lookup object, contents target, or history target. It must not search for generic pronouns or scaffolding words such as "it", "follow-up", or "answer" when the follow-up answer contains the requested item, location, or container name.

Server-selected exploration is part of the smart loop and must remain bounded. When the model's first read misses a specific singular object that appears to have been distorted by speech-to-text or provider query phrasing, the application may run one narrow read-only retry derived from meaningful transcript object words. It must not convert a plural category question or broad inventory question into repeated broad list/search calls, and it must not execute write proposals from this repair path.

The first model turn in a realtime voice session should be a forced context-gathering turn when provider-native function-calling controls support it. On that first turn the application should expose only read tools and request that the provider choose one of them, so inventory requests do not bypass authorized lookup because a structured final-response schema was also present. Later turns may expose the approval-plan proposal tool and let the model either call another tool or produce a structured final response. Provider-independent ports should express this as a tool-choice requirement rather than leaking provider names such as Gemini `ANY` mode into the application layer.

The first query loop exposed only read-only tools to the model. The first approval-backed mobile slice may add exactly one non-mutating write-intent tool: `propose_action_plan`.

The first read-only tools are:

- Search authorized assets.
- Get asset detail.
- List assets in a location.
- List root-level assets in an inventory.
- List authorized assets by safe filters such as kind, lifecycle state, and optional parent or location title.
- List safe audit history for an already-visible asset.

Tool descriptors must use project-owned names, descriptions, read-only markers, and parameter metadata. Provider adapters may translate that metadata into provider-native schemas, but provider-specific tool declaration types must not cross the language inference port.

The asset-detail read tool must be scoped to an asset ID returned by an earlier authorized read tool in the same agent session. It must reject IDs that have not been made visible to the session and must not leak whether a rejected ID exists elsewhere. It must use the authorized asset-detail application query so normal inventory view authorization, safe read audit, and domain validation are preserved. It may return safe bounded asset fields such as title, kind, description, lifecycle state, inventory name, parent title, containing location title, containment path, tag names where available, and current checkout state. It must not return raw custom-field payloads, hidden resources, provider details, credentials, bearer material, raw transcripts, raw prompts, raw audit metadata, internal authorization data, operation IDs, or stack traces. The language model should use this tool when an earlier search or list result identifies the likely asset but the next answer or action needs a more precise current detail snapshot.

The audit-history read tool must be scoped to an asset ID returned by an earlier authorized read tool in the same agent session. It must reject IDs that have not been made visible to the session and must not leak whether a rejected ID exists elsewhere. It must use an asset-scoped audit-history application query backed by an audit repository port filter for tenant, inventory, target type, target ID, newest-first ordering, and limit. It must not scan generic tenant or inventory audit pages in memory to approximate asset history. The application query must authorize inventory view access and emit one intentional safe audit-read record for the history lookup, not one audit-read record per internal page. It may return safe audit fields such as action, source, occurred-at time, target type, target title, asset kind, previous parent title, new parent title, lifecycle state changes, and a concise summary. Historical parent titles must be resolved through an authorization-aware application boundary or omitted as unavailable. The tool must not return raw audit metadata wholesale, provider details, credentials, bearer material, raw transcripts, raw prompts, hidden resource data, internal authorization data, operation IDs, or stack traces. The tool is intended for questions such as "when did I move this item?", "who changed this?", "when was this created?", and "where was this before?" The language model must use this tool, together with normal asset lookup tools, for history and movement questions instead of guessing from current containment alone.

The checkout-history read tool must likewise be scoped to an asset ID returned by an earlier authorized read tool in the same agent session. For specific checkout or return history questions such as "who checked out the drill?", "who had the drill?", "who borrowed it?", "when was the drill checked out?", or "was the drill returned?", the loop must not answer only from generic search metadata after resolving the visible asset. It should call `list_asset_checkout_history` for that visible asset before finalizing. Same-session clarification follow-ups may combine a generic checkout-history question with a concrete item answer, such as "Who had it? Follow-up answer: Drill", and must resolve the visible item before calling checkout history. The checked-out-assets list tool returns authorized visible assets and may establish asset visibility for a later asset-scoped checkout-history read in the same session. A failed or rejected checkout-history tool call must not satisfy the requirement to read checkout history before finalizing. Broad checkout-state questions such as "what is checked out?" should be able to use the checked-out-assets list tool on the initial context-gathering turn.

`propose_action_plan` is not an inventory mutation tool. It may persist a proposed action plan through the action-plan application boundary and return a safe plan summary for mobile review. It must not execute asset, location, tenant, sharing, provider-profile, audit mutation, import/export, or raw repository operations. Its arguments must be bounded, typed, validated by the application boundary, and free of raw prompts, raw transcripts, raw provider responses, credentials, bearer tokens, provider session IDs, hidden resource data, and approval claims.

After `propose_action_plan` succeeds, the realtime loop must pause the agent run and emit the proposed plan for mobile review. It must not ask the language model for another final response, synthesize a spoken response, emit `session.completed`, or let a later model turn override the proposed plan before the user approves or cancels the plan. The WebSocket session should remain open for the explicit approval decision. After approval or cancellation, the server may emit the corresponding action-plan terminal event and close or complete the realtime session.

If one or more `propose_action_plan` attempts failed validation for a write request, the loop must not accept a later final `answer` as terminal. Such a final response must be treated as a recoverable agent-contract failure and converted into safe repair feedback. The model may still produce a final `clarification`, `unsupported_action`, or `safe_failure` response when no valid proposal can be prepared. The spoken response must not say an item was added, moved, created, archived, restored, or otherwise changed unless the approved action-plan execution path actually completed that change.

The same rejected-plan rule applies when the realtime loop reaches its final-only fallback after exhausting tool turns. A final-only answer must not override earlier rejected action-plan feedback or imply that a write occurred; the API must recover safely instead of speaking the unsupported claim.

Write proposals must include executable command arguments as structured JSON. For create or move requests that reference an existing location or container, the loop must first use read tools to resolve the visible resource and then place the returned `assetId` in `parentAssetId`. For requests that require missing parent containers or locations to be created, the loop may propose an ordered multi-command plan and use `parentCommandId` to place later creates or moves inside earlier creates. The agent should assume that a clear named missing destination, such as `Kitchen`, `Living room`, `Garage`, `Box under the TV`, `Big cabinet`, or `Second shelf`, is intended to be created unless the phrase is ambiguous, conflicts with visible inventory, or looks likely to be a speech-to-text mistranscription. A request such as "move my water bottle to the kitchen" when Kitchen does not exist should propose creating the Kitchen location first, then moving the existing water bottle into that newly-created location. A nested request such as "move my water bottle to the second shelf in the big cabinet in the kitchen" when the destination path is missing should propose the full dependency chain, for example create `Kitchen`, create `Big cabinet` inside `Kitchen`, create `Second shelf` inside `Big cabinet`, then move the existing water bottle into `Second shelf`. For a clear write request with a missing destination, the agent must use the reviewable action-plan proposal path instead of asking a final yes/no clarification such as "would you like me to create it"; the mobile approval surface is the confirmation step. Parent or location titles may be used only as read filters; they must not be persisted as executable action-plan arguments.

For nested create or move requests, the agent should resolve named outer locations and containers as separate search terms rather than only searching the whole phrase. If a combined phrase search returns no matches, that does not prove each named path segment is missing. For example, if a user asks to add something to a box under the TV in the living room and `Living room` is visible, the proposal should create only the missing box under the TV inside the existing Living room and then create or move the requested item into that box.

The action-plan tool contract must distinguish create, move, checkout, and return command shapes. New items and new containers must use `create_asset`; `create_location` is reserved for true locations and must not be combined with `kind: container`. `create_asset` arguments must contain a title or name and an asset kind, and must never contain `assetId`; `assetId` is only valid for operations against existing assets returned by authorized read tools. `checkout_asset` and `return_asset` arguments must contain an existing visible `assetId` and may contain safe user-provided `details`. When the user asks to add a new item into a missing container under an existing visible parent, the proposed plan should first create the container with `parentAssetId` set to the visible parent asset ID, then create the item with `parentCommandId` set to the container command ID.

When the transcript names a destination, a proposed root-level item create is invalid unless the transcript explicitly asks for a root or top-level item. For example, "add a phone charger to the office" must either create the phone charger under the visible Office asset or ask a clarification; it must not create the phone charger at the inventory root.

The voice tool adapter may canonicalize a dependent command reference when a provider places an earlier command ID in `parentAssetId` instead of `parentCommandId`. This compatibility normalization is allowed only when the value exactly matches an earlier command ID in the same proposed plan. Stored action-plan commands must still use canonical executable arguments.

The voice tool adapter may also canonicalize a provider-produced `create_asset` command whose arguments specify `kind: location` into the domain command kind `create_location`. This compatibility normalization preserves model tolerance while keeping stored action plans aligned with Stuff Stash domain language.

The voice tool adapter may canonicalize provider-produced `create_item` and `create_container` command kinds into the domain command kind `create_asset` with `kind: item` or `kind: container` in the structured command arguments. This compatibility normalization is allowed because item and container are asset kinds in the Stuff Stash domain, not separate action-plan command kinds. It must not create new domain command kinds or persist provider-only command names.

The voice tool adapter may reorder commands inside a proposed action plan when a command references another command in the same plan through `parentCommandId`. The stored plan must use an executable order where the referenced create command appears before the dependent create or move command. This compatibility normalization is allowed only within the proposed plan and must not invent commands, bypass validation, or reorder unrelated commands in a way that changes user intent.

The exact utterance `Move my water bottle to the kitchen.` is a required Gemini live-regression scenario. Given a visible active item named `Water bottle`, an existing containing location named `Office`, and no visible location or container named `Kitchen`, the Gemini language adapter must drive the agent toward a reviewable action plan that creates `Kitchen` and moves the existing water bottle into it. It must not end the session with a final answer such as `I can't find a water bottle` after the tool results have shown the water bottle. This scenario should be covered by deterministic application tests and by an opt-in live Gemini adapter test that can run with Application Default Credentials or an explicit test access token.

The Gemini live-regression suite must include a small realistic voice corpus that exercises the agent loop with phrases a home user is likely to say or receive from speech-to-text. The corpus is an evaluation artifact, not an exhaustive prompt example list. It should include normal spoken requests, casual phrasing, nested containment, ambiguous or likely mistranscribed phrases, and adversarial or unsupported requests. Each scenario must assert one of these outcomes:

- A factual answer completes with speech and uses visible tool results.
- A reviewable action plan is proposed and the loop pauses before final speech.
- A clarification, unsupported-action response, or safe failure completes with an actionable next step rather than a dead-end provider error.

The corpus must cover at least these behaviors:

- Asking where a known item is.
- Asking where a category-like household phrase is, such as tools, when the answer must be grounded in visible matching assets and containment or must fall forward with a useful no-match response. The loop must not turn a narrow no-match category phrase into a broad unfiltered item list unless a future category/tag/search spec defines that behavior.
- Asking for a broad list of visible items.
- Asking what is inside a known location or container.
- Asking what is inside an existing visible container.
- Creating an item in an existing visible location.
- Creating an item in an existing visible container using casual spoken wording such as "I got..." or "put it in...".
- Creating an item inside a missing nested container path under an existing visible location.
- Moving an existing item into an existing visible location.
- Moving an existing item with casual location language, such as "out to the garage", while still using only visible destination IDs.
- Moving an existing item into a missing single destination that should be created.
- Moving an existing item into a missing nested destination path that should be created.
- Archiving an existing visible item.
- Restoring an existing visible archived item.
- Checking out an existing visible item.
- Returning an existing visible checked-out item.
- Asking for a change where the source item is not visible.
- Asking for a change where speech-to-text produced an unclear or unlikely destination.
- Asking for a dangerous, destructive, provider-configuration, credential, or unrelated system action.

Live corpus tuning must optimize the general contract, tool metadata, loop repair behavior, and concise prompt rules. It must not add a one-off instruction for every corpus utterance. Adding a representative example is acceptable only when it teaches a general class of behavior, such as clear missing destinations being represented as dependent create commands before a move. A live corpus pass is successful only when every expected-success scenario succeeds and every expected-non-success scenario falls forward with a user-actionable response instead of producing a generic voice failure, transport/provider error, hidden diagnostic, or contradictory final response.

## Voice Evaluation Skill And Harness

Stuff Stash must maintain a repo-local Codex skill for evaluating conversational inventory quality. The skill must guide an agent through running the live Gemini voice corpus, preserving full traces, reviewing the actual model and tool behavior, and deciding whether each scenario is product-good rather than merely test-green.

The evaluation workflow may produce durable artifacts for synthetic opt-in corpus fixtures only. These traces may include the fixture transcript, model/tool diagnostics, and spoken response needed to evaluate agent quality. They must not include arbitrary user transcripts, provider credentials, bearer tokens, raw audio, generated speech bytes, hidden resources, or production tenant data. The evaluation workflow must produce durable artifacts for each run, including raw `go test -json` output, extracted scenario traces, and a summary that distinguishes:

- Runs where no corpus scenarios were extracted, including locally skipped live-provider runs, as non-green execution evidence.
- Hard execution failures, such as provider errors, invalid action plans, or unexpected session completion.
- Assertion failures from the Go regression suite.
- Human/product quality concerns found by an agent reading the trace, such as awkward fall-forward wording, wrong mental model, unnecessary tool turns, brittle planning, or missing next steps.
- Cases that passed deterministic checks but still need product follow-up.

Live corpus tests must log the same full event trace for failed scenarios that they log for successful scenarios. A provider failure, timeout, invalid model response, assertion failure, or unexpected session completion must still leave enough trace evidence for the evaluation harness to extract the transcript, provider stage, safe failure code, tool calls, tool results, and last spoken text when present.

Provider adapters may perform bounded retries for transient language-inference failures, including rate-limited HTTP responses and malformed structured planner/final responses from a provider that otherwise returned successfully. Retries must be internal to the adapter, must not re-execute Stuff Stash tools or inventory writes, must preserve the same transcript, prompt, tool results, and tenant/inventory context, and must remain bounded so voice sessions fail safely instead of hanging indefinitely.

The skill may use the Codex CLI as an optional judge for trace review, but the primary agent remains responsible for evaluating the judge reasoning. A green Codex-judge verdict must not be accepted blindly. The agent must read the judge explanation, compare it to the trace and rubric, and identify any changes needed in prompts, schemas, loop policy, tool metadata, fixtures, or product behavior.

The evaluator must prefer realistic home-inventory utterances and traces. Fast deterministic unit tests remain necessary for safety invariants, but they are not a substitute for live trace evaluation.

The realtime loop must track the set of opaque `assetId` values returned by successful authorized read tools during the current session. `propose_action_plan` must reject any `assetId` or `parentAssetId` that is not in that session-visible set before persisting the proposed plan. This provenance check is in addition to action-plan command validation and execution-time authorization.

The realtime loop must reject a proposed `move_asset` command that moves an asset to inventory root when the transcript appears to name a destination, unless the transcript itself clearly asks for root, top level, no parent, or removal from a container/location. If the model searches for a destination, finds no visible match, and then proposes root as a substitute, the loop must treat that as a recoverable invalid tool request and give the model a chance to ask a clarification or propose a better action plan. This prevents likely speech-to-text errors such as "move my drill to the side" from becoming destructive or confusing root moves.

The realtime loop must also reject approval plans whose existing asset or existing parent references do not align with the transcript. For existing asset moves, the referenced asset title should have meaningful title words present in the transcript or another explicit user-provided reference from the current session. For existing parent locations or containers, the referenced parent title should likewise be named by the user. If the model proposes moving unrelated visible items or substitutes a visible location the user did not name, the loop must return a recoverable invalid tool result rather than showing that plan for approval.

The action-plan proposal tool must reject creating a root location or container whose title and kind exactly match an already visible authorized asset in the selected inventory. The model should repair by resolving that existing parent with read tools and using its `assetId` as `parentAssetId` for dependent creates or moves. This prevents a combined-path miss, such as no result for "box under the TV in the living room", from causing a duplicate `Living room` create when `Living room` already exists.

When prior read tool results in the same session show that a destination phrase returned no visible match, a later action plan must still account for that phrase when it is a clear part of the user's requested destination. It may account for it by creating a location or container command with the missing segment's meaningful words in the command title. It must not silently drop the missing segment and place the item or moved asset in a broader parent.

When `propose_action_plan` rejects provider output for invalid or unprovenanced IDs, the retryable tool result returned to the model must include safe repair guidance. That guidance should state that `assetId` and `parentAssetId` must be copied from successful authorized read tool results, that guessed IDs or titles are not executable IDs, and that missing destinations should be represented as earlier create commands referenced by `parentCommandId`. The repair guidance must not leak hidden assets, raw prompts, credentials, stack traces, or provider internals.

The mobile approval sheet must present multi-step plans as an explicit review surface, not a generic confirmation sentence. It should separate what Stuff Stash will use from what it will create, show nested placement in readable language, keep approve and cancel controls fixed in the bottom action area, and avoid raw IDs, provider terminology, diagnostics, or hidden model details. For dependent creates, the user should be able to see the hierarchy before approval, such as `Living room` as an existing location, `Box underneath the TV` as a new container inside it, and `Apple TV remote` as a new item inside the new container.

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
- Opaque `assetId` values for visible assets when needed for follow-up tool calls or action-plan arguments.
- Custom fields only after a field sensitivity and provider-disclosure policy exists. The first improved catalog must omit custom field values from cloud-provider tool results.
- Match metadata that helps the model understand why a result was returned.

Tool results must not include raw authorization decisions, hidden resources, bearer tokens, provider credentials, raw prompts, raw model responses, raw audio, generated speech, custom field values before a sensitivity policy exists, internal stack traces, or infrastructure details. Internal resource identifiers may be provided to the in-process agent loop only when needed to chain read-only tool calls or prepare an action plan for a visible resource the tool returned. Final user-facing responses, mobile progress events, and mobile action-plan review text must not speak or display those identifiers.

For specific where-is questions, when a search result clearly returns the requested item and includes `locationTitle` or `containmentPath`, the agent should answer from that result without issuing broad follow-up list calls. `containmentPath` is ordered from outermost visible parent to the asset itself; the last element is the returned asset, not a container that contains itself. User-facing answers must not describe an item as being inside itself.

The first implementation may expose structured tool results as compact JSON strings through the language inference port while provider-native tool schemas are still evolving. The JSON shape must remain project-owned and provider-independent.

Gemini language inference must request structured final responses through provider-native structured output controls instead of relying only on prompt wording. The adapter must set JSON response MIME type and a response schema for the final-response envelope on final-only turns, and must set JSON response MIME type and the provider-supported JSON schema field for the action-plan envelope on planner-only turns. The prompt may briefly tell the model to follow the provided response schema, but it must not duplicate the full JSON schema or maintain a separate hand-written final JSON example that can drift from the actual provider schema.

The Gemini `generateContent` adapter must not depend on combining native function calling with structured-output response schemas in the same turn for Gemini 2.5 models. Tool turns are native function-calling turns and should not request a textual JSON response schema. Final turns and planner turns are schema-constrained structured-output turns and must not expose provider-callable tools. The application loop owns this phase transition: gather authorized context through read tools, then call either the constrained action-plan planner for write intents or the constrained final-response generator for answer intents. This follows the current provider contract instead of relying on prompt text to force JSON while tools are also available.

When the loop requires a first context-gathering tool call, the language adapter should use a compact read-only prompt for that turn. The prompt should direct the model to choose an appropriate authorized read tool and should not include the full write/action-plan contract or instructions for tools that are unavailable on that turn. For nested add/create destinations, the first read strategy should prefer resolving the outermost named place or container separately instead of searching the whole path phrase first. Later turns that include `propose_action_plan` must restore the full action-plan contract and structured final-response controls.

The application loop may override a provider-requested read tool call with a server-selected read when the loop has enough safe context to know the provider's requested read is not the next minimal read needed for the user's intent. This override is allowed only for bounded context-gathering phases:

- For contents questions such as "what's in the toolbox", after a read result returns a visible location or container named by the transcript, the loop may force `list_authorized_assets` scoped to that exact location or parent title. If multiple visible locations or containers overlap the transcript, the loop must prefer exact phrase coverage, stronger word coverage, and direct containers over broad locations.
- For add/create requests where the first read only proves a new item is missing and the transcript names a destination, the loop may force one destination search using the likely outer room, place, or container from the transcript before entering the planner phase.
- For nested create or move requests where a combined destination phrase misses, the loop may force one search for the likely outer parent before entering the planner phase.

The application loop may also perform narrowly bounded proactive read-only calls without a provider-requested tool call when the transcript has an unambiguous grammar and the read cannot mutate inventory. This is allowed only for exact context gathering that a provider prompt would otherwise be required to select: resolving the named target of a contents question, and resolving the named parent of a simple non-nested create/add request. It must not proactively execute write tools, broad exploratory lists, nested path expansion, source-item substitution, destructive actions, provider-profile actions, or reads that depend on guessing hidden state. Proactive reads must have deterministic tests and negative coverage for nested or ambiguous phrasing.

Server-selected reads must preserve tenant, inventory, authorization, and visible-resource scoping. They must emit developer diagnostics when developer diagnostics are enabled. They must be minimization-preserving: the loop must not replace a narrow no-match lookup with a broad list of unrelated visible inventory merely to help the model guess a category or synonym.

The first improved implementation may expose a generic filtered list tool that covers list-by-location, list-root-level, and list-by-kind behavior before separate public tool names are added. The generic list tool must provide an explicit root-level filter, such as `parentScope: root`, because omitting a parent or location filter means "all visible assets" rather than "only assets at inventory root." Root-level filtering must not be combined with parent-title or location-title filters. The first voice tool catalog must support at least these user intents accurately for visible inventory data:

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

## Developer Diagnostics

Mobile developer diagnostics may expose verbose realtime agent-loop details when diagnostics are explicitly enabled for a development build. Developer diagnostics are not normal user-facing product copy and must stay behind the diagnostics flag. They may include the final transcript, the server-owned prompt text sent through the language-inference port, the model turn shape, tool names, tool-call arguments, safe tool result JSON, action-plan proposal payload summaries, and final structured response content.

Developer diagnostics must not include raw audio bytes, provider credentials, bearer tokens, API keys, encrypted credential ciphertext, provider session identifiers, raw HTTP headers, stack traces, or hidden inventory resources. The realtime `session.start` contract must carry an explicit developer-diagnostics opt-in, and the API must stream `agent.diagnostic` events or diagnostic detail payloads only for sessions that opted in at start. Normal tool progress events must stay safe and bland; verbose tool arguments, tool results, prompt text, and model turn content belong only in separate `agent.diagnostic` events. The API must redact unsafe key names, bearer-token phrases, and bounded diagnostic text before streaming diagnostics. The mobile app may render diagnostic details as selectable text to support debugging poor model behavior, but it must clearly keep them inside the diagnostics section and must not speak diagnostic text.

Developer diagnostics should make the loop understandable without drowning out state changes. The first language-model call may include the full server-owned prompt. Later calls should identify the model turn and may elide repeated prompt scaffolding with a short marker while still showing the observable model turn, tool-call request, tool result, action-plan proposal, or final structured response.

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

The final response must not include raw chain-of-thought, raw model reasoning, raw prompts, raw provider responses, raw transcripts, raw audio, credentials, bearer tokens, hidden resource data, stack traces, tool-call syntax, raw JSON envelopes, or internal resource-key names such as `assetId`. The application final-response validator must reject unsafe `spokenResponse` and `displayResponse` content before any text is sent to text-to-speech or mobile response-completed events.

## Prompt Templates

The first real provider adapters may use a fixed project-owned prompt template for the voice loop.

Provider adapters that support native tool or function calling must use the provider-native tool calling mechanism for tool selection instead of asking the model to emit project-defined tool-call JSON in text. For Gemini on Vertex AI, the adapter must send the project-owned read-only tool catalog as `functionDeclarations` with OpenAPI-compatible parameter schemas, parse returned `functionCall` parts into project-owned `AgentToolCall` values, and send tool outputs back as `functionResponse` parts on later turns.

Native provider tool calling is an adapter concern. Application services, domain services, REST adapters, mobile clients, and tool execution code must continue to use project-owned tool descriptors, tool calls, tool results, and structured final responses. Provider-native tool declaration, function-call, and function-response shapes must not leak across the language inference port. The Gemini adapter must treat native function declarations as a read-tool surface only: non-read-only descriptors such as the internal `propose_action_plan` tool must never be exposed as provider-callable functions, even if their project-owned descriptor is marked provider-callable for compatibility with other adapters or internal loop phases.

The agent loop must allow multiple distinct read-only tool calls across turns when needed to answer the user accurately. Loop control must be owned by the application agent loop rather than provider-specific adapter shortcuts. If the model asks for an identical tool name and argument set that has already been executed in the same session, the loop must not execute the duplicate call again, but it must continue executing other unseen tool calls in the same model turn. It may request one explicit finalization-only model turn using the existing tool results and no tool catalog; if the model still does not produce a valid final response, the session must fail safely.

Provider adapters may continue to use structured JSON output for final responses when the provider supports it. Native tool calling must not loosen the final response validator, read-only/write boundaries, tenant and inventory scoping, or redaction rules.

Action-plan proposal is a structured-output phase, not a provider tool-calling phase. For Gemini on Vertex AI, the adapter must force planner turns with `responseMimeType: application/json` and a concrete JSON schema for the Stuff Stash action-plan envelope. Gemini planner turns must use the provider field that supports the required schema shape, such as `responseJsonSchema` on Vertex AI when command-specific branches are needed. The schema must require `actionPlan.intentSummary`, `actionPlan.modelInterpretationSummary`, `actionPlan.confirmationSummary`, and `actionPlan.commands`, and must define `commands[]` with command-specific schema branches for create, move, archive, restore, checkout, and return commands. The planner schema must not be derived from an unconstrained provider tool argument object, because that permits valid JSON that is still not a valid executable plan.

The action-plan schema should use provider-supported structured-output features such as field descriptions, enum values, required fields, and nullable root-parent references where available. Prompt text must not duplicate the full schema. Where a provider accepts but does not reliably follow branch constraints in nested objects, the adapter may prefer simpler required fields with explicit empty-string sentinels for unused references, such as requiring both `parentAssetId` and `parentCommandId` while allowing one to be empty. Provider schema constraints are a quality and safety aid, but the application loop must still validate command semantics through project-owned action-plan parsing, authorization, tenancy, and approval rules before presenting or applying any change.

For nested create or move requests, the application agent loop must prefer gathering enough read context to distinguish existing parents from missing path segments before entering the structured planner phase. If a create/add request names a missing inner container or surface under a named outer room, place, or container, and the first read only proves the inner segment is missing, the loop must request at least one additional read for the likely outer parent before asking the model for an action plan.

Action-plan validation must reject plans that create a requested move source when the transcript phrased the source as an existing item to move. A missing source should fall forward with a clarification or safe response, not silently become a new inventory item.

Action-plan validation must also reject root creates for containers or surfaces when a visible parent location or container named in the transcript was returned by read tools. In that case the plan must either use the visible parent's `parentAssetId` or, if the visible parent is not sufficient, fail safely and ask the planner to repair the parent reference.

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

The same cancellation boundary applies to same-session clarification follow-up turns. If the user cancels while recording a follow-up or while the follow-up audio is being sent over the existing conversation socket, mobile must not send the cancelled audio turn, must not apply later follow-up events to visible state, and must keep the session in a terminal cancelled state.

When the API receives `session.cancel` before `audio.end`, it must mark the realtime session as cancelled and emit a terminal `session.cancelled` server event rather than reporting a user cancellation as `session.failed`. If the connection disappears after `audio.end` while provider work is already in flight, the server should cancel through the request context where practical and record a safe terminal outcome without relying on the client to keep listening for the acknowledgement.

The server must enforce configured session, silence, provider, tool-call, and idle timeouts.

The first Google provider bridge must use a configured provider HTTP timeout, defaulting to 60 seconds for local smoke testing unless runtime configuration overrides it. Tenant-managed Google provider profiles may also carry a provider-runtime HTTP timeout option so slow model turns can be tuned without changing code. Invalid, empty, zero, or negative timeout values must be rejected or ignored safely rather than disabling timeouts.

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
- Provider-stage failures must use stable safe failure codes that identify only the failed capability stage, such as speech-to-text, language inference, or text-to-speech. They must not expose provider response bodies, prompts, transcripts, generated speech, credentials, endpoint URLs, stack traces, or account details.

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
