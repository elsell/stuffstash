# Mobile Realtime Voice Query Spec

## Conversation quality revision

`model-led-voice-loop.spec.md` defines the current model/tool contract, workflow migration, response generation and conversation continuity. This spec defines the mobile wire, audio, review, security and evaluation requirements. The retired investigation stages and response fallback are not supported runtime paths.

## Purpose

Stuff Stash needs a first production-shaped mobile voice slice that proves the core conversational loop with real audio input, speech-to-text, language inference, typed inventory reads, structured terminal outcomes, text-to-speech, and streamed audio playback.

The first testable user experience is:

1. The user taps the mobile Voice control.
2. The user asks a read-only inventory question, such as "Where are my tools?" or "What is in the garage?"
3. The mobile app streams audio to the core API.
4. The core API transcribes the audio through a speech-to-text port.
5. The core API runs the agent loop with the project-owned typed read dispatcher.
6. The agent loop may request one or more bounded, authorized inventory reads.
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
- Typed internal read dispatcher.
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

Language inference uses the project-owned conversation port. The provider returns natural text or typed tool calls. The API authorizes reads, validates observed references and domain commands, persists proposals, and controls approval and execution. A clear write request may require several model-selected reads or questions before a complete proposal; there is no mandatory phase sequence.

Once a proposal is persisted, the loop pauses at `action.plan.proposed`. No later model call or response can override that plan while review is pending.

Approved voice create-asset action plans must execute through the same application command semantics as REST and mobile Add. The compiler may target only inventory root or a grounded destination whose kind is location or container; it must not reinterpret or promote an existing item into a container. Voice adapters and WebSocket handlers must not mutate asset kinds or containment semantics.

The model may use the scoped inventory tool catalog through application ports. It cannot call the public MCP transport, raw repositories, provider configuration or direct execution tools. Unsupported requests and ambiguous language are handled in the conversation rather than through local transcript classifiers.

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

The mobile state layer must apply safe realtime session events incrementally as they arrive. It must not wait for the WebSocket session to finish before updating visible progress, final transcript, safe tool progress, final response, speech playback state, or safe failure state. This allows the sheet and collapsed voice accessory to reflect transcription, inventory reads, evidence assessment, response preparation, and speech playback while the server-side loop is still running.

Safe agent progress events should be summarized as lightweight user-facing status rather than exposed as raw event logs. The active voice sheet may render a compact bounded progress trace for multi-step understanding, exploration, planning, answering, and recovery while no action-plan review body is present, so users can see the smart loop moving without enabling developer diagnostics. The active voice sheet must not render progress as an expanding table above the action-plan review because that can push the approval prompt and command list below the visible area. When a proposed action plan exists, the confirmation prompt and command list must remain the primary visible body content, with current progress represented by compact chrome such as the bottom status label, a slim activity bar, or an activity indicator. Tool-call events may be displayed in a developer diagnostics panel only when developer diagnostics are explicitly enabled. Diagnostics must be visually secondary, collapsed by default, placed after the review content, and must not expose hidden resource data, raw query text, raw transcripts, raw prompts, raw model responses, provider credentials, internal IDs, or internal stack details.

When a speech, language, or speech-output provider fails, client copy identifies the failed capability without implying that earlier successful work never happened. Opted-in diagnostics include bounded model/tool-call counts, safe tool names and a safe error category. They exclude raw provider bodies, prompts, transcripts, credentials, endpoints, stack traces and hidden data.

The mobile state layer may maintain a bounded ephemeral progress timeline for the active session for compact summaries and developer inspection. The visible sheet should show the current milestone rather than all milestones by default. The timeline may include safe client and server milestones such as audio upload, API connection, transcription completion, safe `agent.progress` messages, response preparation, speech playback, cancellation, completion, and safe failure. It must not include raw tool arguments, raw provider output, internal IDs, provider errors, prompts, credentials, bearer material, audio bytes, stack traces, or partial transcript history. The mobile application boundary must still redact obvious unsafe terms from progress labels before they reach visible state, even though the API is also responsible for safe progress messages. Duplicate adjacent progress labels should collapse into one visible step so compact sheets stay readable.

The active voice session view may display the final transcript to the user as ephemeral UI state. This transcript display is not debug history and must not be written to local storage, logs, crash reports, analytics, audit records, or observability metadata before a transcript retention and redaction policy is specified.

When the realtime voice sheet displays a proposed action plan, each visible plan row that represents an item, container, or location may offer an inline `Add photos` action only when the row is a newly created asset or a move row that resolves to a concrete reviewed asset after execution. Archive, restore, checkout, and return rows must not offer photo staging. This action must use the same native photo-selection capability as the Add flow, including camera and photo-library choices, native permissions, supported image MIME types, and in-session image previews. Selected photos are draft UI state scoped to the active proposed plan and command row; they must not be uploaded, persisted, logged, added to diagnostics, or sent to the language provider before the user approves the plan.

Every proposed row that creates an item, container, or location must expose direct review editing. Tapping the proposed name once must replace it inline with a focused text field, with accessible save and cancel controls and no separate edit screen. Tapping the containing-parent control once must open a searchable parent selector. The selector must include inventory root, authorized visible parent candidates from the active inventory, and eligible earlier create rows in the same plan. Selecting a parent must return directly to the review with the revised placement visible. Draft title and parent edits must be scoped to the active plan and command row, reset when the plan changes or leaves `proposed`, and remain local until approval.

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

Each slot must show its selected profile, readiness state, and the next best action. The setup view must distinguish selected profiles from unselected duplicates, missing credentials, disabled profiles, archived profiles, and profiles that need a fresh test. Setup labels for readiness, capability, selection source, credential status, lifecycle state, and testing status must use bounded product-owned labels and must not render raw backend strings, provider session identifiers, prompts, endpoints, credentials, stack traces, or provider response material. Unknown setup values must degrade to a safe neutral label such as `Needs attention` or `Unknown`. A flat list of provider profile cards may exist only as a secondary profile inventory or advanced view.

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
- `client.ack`: `seq`, `sessionId`, `ackSeq`. Until explicit server-side flow control is implemented, the API may accept valid acknowledgement messages as no-op protocol metadata after applying the same sequence and session-binding checks as other client messages.

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
- `assistant.response.delta`: reserved and must not be emitted by the first implementation. Mobile clients must reject it as a safe protocol error if it appears and must not render partial model response text.
- `assistant.response.completed`: `seq`, `sessionId`, structured final response with bounded authorized asset-reference artifacts.
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

Once mobile enters a terminal `completed`, `failed`, or `cancelled` state for the current transport turn, later queued server events from that same turn must not move the visible session back into processing, speaking, review, or another terminal state. This preserves the safe terminal outcome when WebSocket delivery races leave already-buffered messages behind a terminal event. Same-session clarification follow-up turns are separate user-initiated transport turns and must start from a fresh active state before their events are reduced.

Every client message after `session.start` must be bound to the authenticated WebSocket connection and server-created session. The server must reject forged session IDs, stale client sequence numbers, replayed audio chunks, messages from cancelled sessions, messages from completed sessions, and any attempt to change tenant or inventory scope after session authorization.

Client messages must include monotonic per-session sequence metadata once a session is established. The server must use that metadata only for ordering, replay rejection, flow control, and safe diagnostics; it must not treat client sequence metadata as authorization.

## Action Plan Events

`action.plan.proposed` contains the safe persisted action-plan review payload for mobile. It must include plan ID, confirmation summary, command summaries, risk summaries, and no raw transcript, raw prompt, raw model response, credentials, provider session IDs, hidden resource data, or approval claims. For existing-asset commands such as move, archive, restore, checkout, and return, the application-compiled review command must include the authorized visible asset title and kind.

Mobile must not enter review for a malformed `action.plan.proposed` event whose plan ID is missing or empty after normalization. It must fail the visible session safely instead of showing an approval surface that cannot send a valid review decision.

The mobile review sheet must treat application-compiled command titles and kinds as verified display context for existing-asset changes. If an existing-asset review command is missing a verified title, mobile may show its application-generated command summary as supporting text, but it must not present that summary as the resolved asset title. The primary review row must use a neutral fallback such as `Selected item`, `Selected container`, `Selected location`, or `Selected asset` until the API supplies verified context.

When the API emits `action.plan.proposed`, the mobile app must enter the `review` stage and show the proposal in the voice sheet. The loop must stop there without `assistant.response.completed`, text-to-speech events, or `session.completed` until the user explicitly approves or cancels. It must never silently execute the plan.

The next mobile review slice must let the user approve or cancel the proposed action plan from the voice sheet using the existing realtime WebSocket session. The mobile client must send `action.plan.approve` or `action.plan.cancel` with the server-created session ID, the proposed plan ID, and monotonic client sequence metadata. Approval may additionally send only bounded reviewed create-command edits keyed by command ID and safe photo metadata. Cancellation must not send edits. The client must not send replacement command kinds, arbitrary command arguments, summaries, risks, approval claims, tenant IDs, inventory IDs, credentials, prompt text, transcript text, or model output in the review decision message.

Before sending an approval or cancellation decision, the mobile application boundary must validate that the proposed plan ID remains present after normalization. If the review payload is missing a usable plan ID, mobile must fail the decision locally with safe user-facing failure state rather than sending an empty or malformed review decision over the realtime transport.

When the API receives `action.plan.approve`, it must call the action-plan application approval boundary scoped to the authenticated principal, session tenant, and session inventory. The application must validate and atomically persist any reviewed create-command edits while transitioning the persisted plan from `proposed` to `approved`. The first execution slice may then execute an approved single create command through the action-plan execution service. The API must emit `action.plan.approved` with the safe plan ID and status after approval succeeds, then emit `action.plan.executed` or `action.plan.failed` after execution finishes.

When the API receives `action.plan.cancel`, it must call the action-plan application cancellation boundary scoped to the authenticated principal, session tenant, and session inventory. Cancellation transitions the persisted plan from `proposed` to `cancelled`. The API must emit `action.plan.cancelled` with the safe plan ID and status after the transition succeeds. Cancellation messages must not include photo attachment metadata or other approval-only payloads.

The server must reject review decisions with stale sequence numbers, forged session IDs, missing plan IDs, plan IDs that are not scoped to the session tenant and inventory, unauthorized principals, terminal plans, or malformed message types. Safe rejection must not expose hidden plan existence, raw model output, command arguments, credentials, prompt text, transcript text, provider response details, stack traces, or authorization internals.

After receiving `action.plan.approved`, mobile must show that the approved change is being applied and keep duplicate decisions disabled. After receiving `action.plan.cancelled`, `action.plan.executed`, or `action.plan.failed`, mobile must leave the pending review state and show a safe terminal state. Terminal review states must not keep stale pending-decision flags that can suppress normal terminal controls or misrepresent the session as still waiting on the user's earlier tap. Failure states must not expose command arguments, stack traces, hidden resources, provider output, prompts, transcripts, or authorization details. Cancellation messaging must make clear that no change was made.

Mobile must apply action-plan approval, cancellation, execution, and failure events only when the event plan ID matches the currently reviewed plan. A mismatched plan event is stale or malformed for the active mobile session view and must not complete, fail, clear pending review state, upload photos, or change the visible active review.

## Transcript Events

Speech-to-text may emit partial and final transcript events.

`transcript.delta` may contain a partial transcript. Partial transcripts are for immediate mobile feedback only and must not be treated as final user intent.

`transcript.final` contains the transcript text the agent loop may use for intent interpretation. If the speech-to-text provider cannot stream partials, the server may emit only `transcript.final`.

Raw user transcripts must not be durably persisted before a transcript retention and redaction policy is specified.

User transcripts are ephemeral UI and in-memory agent-loop state only in the first slice. Raw user transcript text must not be written to mobile local storage, debug event history, crash reports, analytics, audit records, observability metadata, API session metadata, logs, or provider profile test records before a retention and redaction policy is specified.

## Agent Loop

The project-owned conversation loop starts from the final transcript and uses the configured language-model port. The model chooses authorized reads, follow-up questions, natural answers and reviewable proposals. The server does not classify intent, resolve semantic references, repair a canonical interpretation, enforce prose anchors or generate a fallback factual answer. The complete contract is `model-led-voice-loop.spec.md`.

The application supplies the authenticated principal and fixed tenant/inventory scope, validates tool shapes and observed IDs, enforces per-turn model/tool/time budgets and session context limits, and authorizes every read through the owning application service. Inventory and vocabulary results are untrusted data, not instructions. Household vocabulary is available through the bounded `get_inventory_vocabulary` tool; there is no mandatory preload or interpretation stage.

Model-proposed changes use typed domain commands validated by application services. Existing asset references must have been observed in scoped reads and be reauthorized before proposal preparation. Dependent commands may reference an earlier create command. A proposal immediately pauses for explicit approval; the model cannot approve or execute changes. Domain lifecycle, containment, authorization and audit rules remain authoritative.

Safe progress statuses and tool events describe work without exposing raw prompts, provider errors, identifiers or inventory content. Tool argument errors can return safe structured feedback for the model to correct within the remaining budget. Authorization failures and cancellation terminate processing. A provider failure remains a failure; the old investigation executor and deterministic answer fallback must not run.

## Authorized Read Tools

The asset-detail read tool must be scoped to an asset ID returned by an earlier authorized read tool in the same agent session. It must reject IDs that have not been made visible to the session and must not leak whether a rejected ID exists elsewhere. It must use the authorized asset-detail application query so normal inventory view authorization, safe read audit, and domain validation are preserved. It may return safe bounded asset fields such as title, kind, description, lifecycle state, inventory name, parent title, containing location title, containment path, tag names where available, and current checkout state. It must not return raw custom-field payloads, hidden resources, provider details, credentials, bearer material, raw transcripts, raw prompts, raw audit metadata, internal authorization data, operation IDs, or stack traces. The language model should use this tool when an earlier search or list result identifies the likely asset but the next answer or action needs a more precise current detail snapshot.

The audit-history read tool must be scoped to an asset ID returned by an earlier authorized read tool in the same agent session. It must reject IDs that have not been made visible to the session and must not leak whether a rejected ID exists elsewhere. It must use an asset-scoped audit-history application query backed by an audit repository port filter for tenant, inventory, target type, target ID, newest-first ordering, and limit. It must not scan generic tenant or inventory audit pages in memory to approximate asset history. The application query must authorize inventory view access and emit one intentional safe audit-read record for the history lookup, not one audit-read record per internal page. It may return safe audit fields such as action, source, occurred-at time, target type, target title, asset kind, previous parent title, new parent title, lifecycle state changes, and a concise summary. Historical parent titles must be resolved through an authorization-aware application boundary or omitted as unavailable. The tool must not return raw audit metadata wholesale, provider details, credentials, bearer material, raw transcripts, raw prompts, hidden resource data, internal authorization data, operation IDs, or stack traces. The tool is intended for questions such as "when did I move this item?", "who changed this?", "when was this created?", and "where was this before?" The language model must use this tool, together with normal asset lookup tools, for history and movement questions instead of guessing from current containment alone.

The checkout-history read must likewise be scoped to an asset ID already visible in the authorized session. History answers need actual history evidence rather than guesses from current state; failed reads do not establish facts.

The model proposes ordered typed commands through the proposal tool. Application validation and domain services own executable command semantics. Existing resources use observed authorized IDs; dependent creates use `parentCommandId`. Review must preserve the requested placement and distinguish existing resources from proposed new ones.

New items and containers use `create_asset`; locations use `create_location`. Existing-asset commands require authorized existing IDs. The model resolves ambiguity through useful questions, considers existing records before suggesting creation, and distinguishes an explicitly additional physical item from an existing one. These are model responsibilities verified by outcome tests, not intent classifications or exact-match completion rules.

After compilation and persistence, the loop emits `action.plan.proposed` and pauses without final speech or `session.completed`. The WebSocket remains available for explicit approval or cancellation. User-facing text must not claim that a change occurred until approved action-plan execution actually succeeds.

The live-regression suite must cover the semantic family where a user requests moving a visible existing item into a clear missing destination. The expected outcome is a reviewable plan that creates the destination and moves the existing item; a no-match answer or a question asking whether to create the clear destination is a failure. Fixtures should be generated from declarative intent, inventory topology, and expected semantic predicates so production behavior is not tuned to one household title or frozen utterance. Deterministic application tests must cover the invariant, and opt-in live provider tests must use several independently realized phrasings, nouns, names, and destination types.

The Gemini live-regression suite must include a realistic, declaratively generated voice corpus that exercises the agent loop with phrases a home user is likely to say or receive from speech-to-text. The corpus is an evaluation artifact, not an exhaustive prompt example list or a source of production branching logic. It should include normal spoken requests, casual phrasing, nested containment, ambiguous or likely mistranscribed phrases, and adversarial or unsupported requests. A held-out realization set must vary household nouns, names, topology depth, wording, morphology, and transcription-like substitutions without exposing exact expected titles in approximate-match inputs. Each scenario must assert one of these outcomes:

- A factual answer completes with speech and is grounded in authorized observations.
- A reviewable action plan is proposed and the loop pauses before final speech.
- A clarification, unsupported-action response, or safe failure completes with an actionable next step rather than a dead-end provider error.

The corpus must cover at least these behaviors:

- Asking where a known item is.
- Asking where a category-like household phrase is, such as tools, when the answer must be grounded in visible matching assets and containment or must fall forward with a useful no-match response. A successful answer must say where the likely item or collection is; merely naming a search match is a failure. The loop must not turn a narrow no-match category phrase into a broad unfiltered item list unless a future category/tag/search spec defines that behavior.
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

Live corpus tuning must improve general tool contracts, configured model guidance and conversation handling. It must not add production branching or one-off prompts for individual utterances. A successful corpus requires every expected-success scenario to produce a useful answer or reviewable proposal; provider failures and timeouts remain failures and are reported separately from semantic mistakes.

Evaluation must report results by semantic scenario family and realization source, including deterministic grammar realizations and provider-generated holdouts. Release evidence must distinguish provider or API availability failures from semantic and contract failures, report instrumentation completeness, and retain full safe traces for failed trials. Promotion requires zero invented or cross-reference IDs, zero writes before recorded approval, zero dropped named destination segments, exact semantic command graphs for supported write fixtures, and no dead-end creation-confirmation question for clear missing destinations. Repeated live trials should include pass^k-style stability evidence for representative families rather than relying on a single green run.

## Voice Evaluation Skill And Harness

Stuff Stash must maintain a repo-local Codex skill for evaluating conversational inventory quality. The skill must guide an agent through running the live Gemini voice corpus, preserving full traces, reviewing the actual model and tool behavior, and deciding whether each scenario is product-good rather than merely test-green.

The evaluation workflow may produce durable artifacts for synthetic opt-in corpus fixtures only. These traces may include the fixture transcript, model/tool diagnostics, and spoken response needed to evaluate agent quality. They must not include arbitrary user transcripts, provider credentials, bearer tokens, raw audio, generated speech bytes, hidden resources, or production tenant data. The evaluation workflow must produce durable artifacts for each run, including raw `go test -json` output, extracted scenario traces, and a summary that distinguishes:

- Runs where no corpus scenarios were extracted, including locally skipped live-provider runs, as non-green execution evidence.
- Hard execution failures, such as provider errors, invalid action plans, or unexpected session completion.
- Assertion failures from the Go regression suite.
- Human/product quality concerns found by an agent reading the trace, such as awkward fall-forward wording, wrong mental model, unnecessary tool turns, brittle planning, or missing next steps.
- Cases that passed deterministic checks but still need product follow-up.

Live corpus tests must log the same full event trace for failed scenarios that they log for successful scenarios. A provider failure, timeout, invalid model response, assertion failure, or unexpected session completion must still leave enough trace evidence for the evaluation harness to extract the transcript, provider stage, safe failure code, tool calls, tool results, and last spoken text when present.

Each conversation inference call makes one provider request. Provider adapters must not hide additional semantic or transport attempts inside that call. Tool correction stays in the visible, bounded model/tool conversation; it cannot repeat inventory writes.

The skill may use the Codex CLI as an optional judge for trace review, but the primary agent remains responsible for evaluating the judge reasoning. A green Codex-judge verdict must not be accepted blindly. The agent must read the judge explanation, compare it to the trace and rubric, and identify any changes needed in prompts, schemas, loop policy, typed-read metadata, fixtures, or product behavior.

The evaluator must prefer realistic home-inventory utterances and traces. Fast deterministic unit tests remain necessary for safety invariants, but they are not a substitute for live trace evaluation.

The realtime loop retains scoped observed asset IDs for follow-up reads and proposals. Invented, hidden and wrong-inventory IDs must fail before plan persistence, with fresh authorization at the proposal and execution boundaries.

Dependent proposals must retain the requested containment hierarchy and use valid earlier create-command references. The application validates dependency order and domain containment rules; the model decides what the user meant and which authorized resources fit.

The mobile approval sheet must present multi-step plans as an explicit review surface, not a generic confirmation sentence. It should separate what Stuff Stash will use from what it will create, show nested placement in readable language, keep approve and cancel controls fixed in the bottom action area, and avoid raw IDs, provider terminology, diagnostics, or hidden model details. For dependent creates, the user should be able to see the hierarchy before approval, such as `Living room` as an existing location, `Box underneath the TV` as a new container inside it, and `Apple TV remote` as a new item inside the new container.

The loop must not expose direct write tools, provider profile tools, tenant configuration tools, sharing tools, audit mutation tools, import/export tools, or raw repository access. Any future direct execution must go through an approved action-plan execution service.

Observations provided to the language model must be structured, safe, and useful enough for accurate resolution. For visible assets, they may include:

- Title.
- Kind.
- Description when present.
- Inventory name.
- Lifecycle state.
- Parent title and parent kind when present.
- Nearest containing location title when present.
- Human-readable containment path from outermost visible container or location to the asset.
- Opaque candidate IDs needed for scoped follow-up reads or proposals.
- Custom fields only after a field sensitivity and provider-disclosure policy exists. Until then, cloud-provider observations must omit custom field values.
- Match metadata that helps the model understand why a result was returned.

Observations must not include raw authorization decisions, hidden resources, bearer tokens, provider credentials, raw prompts, raw model responses, raw audio, generated speech, custom field values before a sensitivity policy exists, internal stack traces, or infrastructure details. Final user-facing responses, mobile progress events, and mobile action-plan review text must not speak or display internal identifiers.

For specific location questions, a grounded candidate with a containment path is sufficient; the loop must not broaden the read merely to seek more context. `containmentPath` is ordered from outermost visible parent to the asset itself, and user-facing answers must not describe an asset as being inside itself.

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

Mobile developer diagnostics may expose bounded conversation-loop metadata only when explicitly enabled. The allowlist contains capability stage, model/tool-call counts, tool names and safe failure categories. It excludes transcripts, prompts, vocabulary contents, asset observations, IDs, model responses, reasoning and speech.

Developer diagnostics must not include raw audio bytes, provider credentials, bearer tokens, API keys, encrypted credential ciphertext, provider session identifiers, raw HTTP headers, stack traces, hidden inventory resources, raw tool arguments or results, prompts, transcripts, or model turns. The realtime `session.start` contract must carry an explicit developer-diagnostics opt-in, and the API must stream `agent.diagnostic` events only for sessions that opted in at start. Normal tool progress events must stay safe and bland. The API must construct diagnostics from an allowlisted metadata shape rather than attempting to redact arbitrary provider or inventory payloads after the fact. The mobile app may render diagnostic metadata as selectable text to support debugging poor model behavior, but it must clearly keep it inside the diagnostics section and must not speak diagnostic text.

Developer diagnostics expose bounded model-call and tool-call counts, allowlisted tool names, current capability stage and safe failure attribution. They must not expose provider reasoning or raw model/tool payloads.

Language-inference failures use the same bounded conversation metadata as successful work. Retired intent, evidence-round, resolution and response-generation fields must not reappear in the diagnostic contract.

Tool progress events must not include raw model reasoning, raw prompts, raw transcript text, raw query text, raw tool inputs, raw tool outputs, resource identifiers, exact resource titles, hidden IDs, result counts that can reveal hidden inventory data, credentials, bearer tokens, provider responses, authorization decisions, or stack traces.

## Structured Final Response

Every terminal answer, clarification, unsupported-action response, safe failure, or successful application-owned no-op must produce a structured final response. A proposed action plan is instead terminal at `action.plan.proposed` until approval or cancellation and must not also produce a final response.

The first structured final response shape must include:

- `responseId`: stable ID for the response.
- `sessionId`: realtime session ID.
- `tenantId`: tenant scope.
- `inventoryId`: inventory scope when applicable.
- `source`: mobile voice.
- `kind`: final response kind, initially `answer`, `clarification`, `unsupported_action`, or `safe_failure`.
- `spokenResponse`: concise text intended to be spoken to the user.
- `displayResponse`: text intended to be displayed in the mobile UI.
- `artifacts`: a bounded list of safe structured artifacts. The first non-empty artifact type is `asset_reference` with only `assetId`, exact `title`, `assetKind` (`item`, `container`, or `location`), and an optional application-authored `context` containing a bounded authorized parent or containment label for disambiguation.
- `toolCallIds`: tool calls used to produce the response.
- `auditMetadata`: safe metadata for observability and audit.

`spokenResponse` is the only field that may be sent to text-to-speech in the first slice.

The model authors `spokenResponse` and `displayResponse`. The application validates their encoding and size and presents them without semantic rewriting or deterministic factual fallback.

`displayResponse` may be the same as `spokenResponse`. Neither channel must repeat every search result. The model selects relevant facts and may group them naturally.

Asset-reference artifacts are navigation metadata constructed by the application from observed authorized asset IDs selected by the model. Validate reference membership and the 16-reference bound, then resolve titles, kinds and containment context from trusted metadata. Model-authored titles alone cannot create navigation targets. Plain-text answers remain valid without cards.

Mobile must validate the artifact collection at the WebSocket boundary, preserve it only in ephemeral active-session state, and render case-sensitive exact non-overlapping title occurrences in the visible response as inline accessible links. Tapping a linked item, container, or location must navigate through the existing asset-detail route for that asset ID. Longer exact titles take precedence over overlapping shorter titles. Duplicate-title references must not be assigned to an arbitrary inline occurrence; each must remain available as an explicitly labeled `Open` control in the response card. Duplicate controls must use their authorized context when available and a stable visible ordinal when title and context are still identical, so their labels and accessibility names remain distinct without exposing asset IDs. Any reference that cannot be placed safely inline must remain available through that bounded response-card control. Malformed, duplicate-ID, unsupported, or unbounded artifacts must fail the realtime protocol safely rather than navigating. Artifacts must not be written to diagnostics, analytics, logs, crash reports, or durable local storage.

The provider boundary excludes thought content from user responses. The application rejects malformed or oversized response shapes and unobserved artifact IDs before display or TTS. Credentials and hidden resources must never enter model context. Ordinary inventory prose must not be rejected by keyword lists or required factual-anchor rules.

## Prompt Templates

Configured provider and workflow instructions tune household vocabulary, tone and model behavior through the provider/application ports. Project-owned tool metadata explains the available operations and response protocol. The model handles ambiguous language and selects its own steps; prompts must not prescribe a mandatory investigation/assessment/response sequence.

Prompt customization cannot change fixed tenancy, authorization, observed-reference validation, domain command semantics, approval, budgets, audit or retention boundaries. Provider schemas constrain tool/output shapes without encoding semantic request classifications.

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

The HTTP realtime voice adapter must enforce a per-message client idle timeout while waiting for audio chunks, `audio.end`, or `session.cancel`, separate from the whole-session deadline. The first runtime knob is `STUFF_STASH_REALTIME_VOICE_IDLE_TIMEOUT`, defaulting to 15 seconds. Invalid, empty, zero, or negative values must fall back to the safe default rather than disabling the idle timeout.

The realtime voice application loop must enforce a per-tool-call timeout around every project-owned inventory read, separate from provider HTTP timeouts and the WebSocket idle timeout. The first runtime knob is `STUFF_STASH_REALTIME_VOICE_TOOL_CALL_TIMEOUT`, defaulting to 10 seconds. Invalid, empty, zero, or negative values must fall back to the safe default rather than disabling read deadlines. A read timeout must cancel the read context and emit a safe `tool.call.failed` event. The failed read must not establish evidence or be sent back to the provider as a repair conversation; the application must terminate that stage with a bounded clarification or safe failure instead of hanging the session.

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
- Reject direct or model-authored state-changing calls; state changes may occur only through an approved application-compiled action plan.
- Avoid logging raw audio, raw transcripts, raw prompts, raw provider responses, raw model reasoning, generated speech, credentials, or bearer tokens.
- Avoid returning hidden resource data in transcript, progress, tool, final response, or TTS events.
- Map provider and tool failures to safe user-facing errors.
- Provider-stage failures must use stable safe failure codes that identify only the failed capability stage, such as speech-to-text, language inference, or text-to-speech. They must not expose provider response bodies, prompts, transcripts, generated speech, credentials, endpoint URLs, stack traces, or account details.

Read tool executions must follow the safe read audit requirements of the underlying application operation. Voice read audit metadata must not include raw audio, raw transcripts, raw query text, raw prompts, raw tool inputs, raw tool outputs, raw provider responses, generated speech, or hidden resource details.

## Testing

Tests must use fakes for speech-to-text, language inference, text-to-speech, the typed read dispatcher, inventory application services, authorization, realtime transport, observability, microphone capture, and audio playback where focused unit or adapter behavior is under test.

Realtime boundary tests must exercise the actual API WebSocket adapter with configured authentication and authorization adapters.

Tests must cover:

- Successful read-only question from audio input through spoken audio output.
- Successful typed-transcript equivalent for deterministic agent-loop tests.
- Partial transcript events.
- Final transcript event.
- Multiple authorized tool reads within per-turn budgets before an answer or proposal.
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
- Malformed model tool output or invalid typed read request.
- Malformed model-authored final response rejected before mobile display or text-to-speech.
- Text-to-speech failure.
- Unauthorized, unauthenticated, wrong-tenant, wrong-inventory, viewer-hidden-resource, expired-token, malformed-token, and privilege-escalation attempts.
- Provider output attempts to introduce executable commands, unobserved opaque IDs, or unsupported read kinds.
- Provider output attempts to smuggle hidden IDs, authorization claims, or approval claims through tool arguments or response artifacts.
- Hidden ID probing, wrong-inventory asset detail attempts, and count leakage through progress events.
- Voice read audit emission for underlying read operations without leaking transcript, provider, or raw tool content.
- Redaction of raw audio, raw transcripts, raw query text, raw prompts, raw tool inputs, raw tool outputs, raw provider responses, raw model reasoning, generated speech, credentials, bearer tokens, hidden resources, and stack traces from mobile state persistence, debug history, crash reports, analytics, API session metadata, audit, observability, logs, progress events, final responses, and TTS.

## Open Questions

- Which Expo-compatible audio input format should be used first?
- Which mobile audio playback adapter should own streamed chunk buffering?
- Should the first implementation use tap-to-start/tap-to-stop or push-to-talk?
- What future streaming-safe response contract would allow spoken-response deltas before full structured final validation?
- What exact artifact shape should safe asset and location references use in final responses?

## Conversation continuity negotiation

Clients may opt into `conversationContinuity: true` on WebSocket `session.start`. For those clients, a completed answer or clarification can retain the session and bounded prior user/assistant context for another audio turn. Approval proposals remain an explicit review boundary and the first slice still terminates after its decision. A `session.completed` event marks a completed audio turn and includes `followUpAvailable` for negotiated sessions; false means the client must start a new session. Legacy clients retain clarification-only continuation and existing terminal-answer behavior.

The pinned workflow permits one initial turn plus its follow-up allowance; the default flow permits three total turns. Recheck access before each utterance and reauthorize references before proposals. Retain bounded model/tool context under `model-led-voice-loop.spec.md`, including paired tool calls and results. Model-call, tool-call and processing-time budgets apply per utterance; pauses consume none of the next turn budget. At the turn limit, the final completion advertises false and closes normally without replacing the answer.

The mobile client opts in, retains the transport only when advertised, and reuses it on the next microphone tap. It never begins recording automatically. Legacy servers without `followUpAvailable` retain clarification-only compatibility. Closed, cancelled, denied or exhausted sessions cannot advertise a follow-up. The UI continues to offer starting a fresh request.

## Additional physical items

The model must distinguish referring to an existing recorded item from adding an explicitly additional physical item. A similar existing title is relevant evidence but does not prohibit an additional item. There is no creation-mode enum or evidence-quote gate. Tests must verify that existing-item moves do not duplicate assets and explicit additional-item requests can propose creation.

The proposal remains reviewable and states which item will be created or moved. It cannot reuse an existing ID as the identity of a new item, approve itself or bypass domain command validation, authorization or audit.