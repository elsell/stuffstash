# Realtime Interaction Spec

## Purpose

Stuff Stash needs a realtime interaction channel for conversational inventory so users can speak, receive feedback, approve plans, and hear responses without waiting on a page-style request cycle.

## Scope

This spec covers realtime conversational transport expectations for web and mobile clients.

This spec does not define the final wire protocol, message schema, streaming codec, provider SDK, or deployment topology.

The first mobile realtime voice query slice is specified in `specs/agent-model/mobile-realtime-voice-query.spec.md`.

## Architecture Decision

Realtime conversational interaction must run through the Stuff Stash core API.

The initial and intended architecture is API-mediated streaming:

- Web and mobile clients open an authenticated realtime connection to Stuff Stash.
- Clients stream audio or text input to Stuff Stash.
- Stuff Stash resolves the tenant's configured speech-to-text, language model, and text-to-speech provider profiles.
- Stuff Stash streams provider requests and responses through provider ports and adapters.
- Stuff Stash emits transcript, model, speech, clarification, confirmation, execution, cancellation, and failure events back to the client.

Clients must not receive provider credentials, provider session tokens, or provider-specific realtime URLs from Stuff Stash. Direct client-to-provider realtime streaming is out of scope and must not be introduced without a future spec change.

This keeps realtime behavior inside the same authentication, authorization, audit, and observability boundary as the rest of the core API.

## Requirements

- Realtime voice and text interaction must be supported behind ports and adapters.
- WebSockets are the preferred initial transport for bidirectional client interaction.
- The realtime transport must support streaming speech-to-text input and output events.
- The realtime transport must support streaming model responses when the selected provider supports them.
- The realtime transport must support text-to-speech output events when the client experience uses spoken responses.
- The realtime transport must support action plan events, clarification events, approval events, cancellation events, execution progress events, and final result events.
- Realtime interaction must use the authenticated user's session or token.
- Realtime interaction must preserve tenant and inventory authorization boundaries.
- Realtime interaction must not grant extra permissions to the model provider, agent loop, or transport adapter.
- Realtime sessions must be observable with safe metadata.
- Realtime sessions must have explicit timeout, cancellation, retry, and failure behavior before implementation.

## Realtime Session Ownership

The core API owns realtime session lifecycle.

Each realtime session must include:

- Session ID.
- Tenant ID.
- Inventory scope.
- Authenticated principal ID.
- Client source, such as mobile voice, mobile text, web voice, or web text.
- Selected provider profile IDs for speech-to-text, language inference, and text-to-speech when those capabilities are used.
- Connection start time, last activity time, timeout deadline, cancellation state, and final outcome.
- Safe correlation metadata for observability and audit.

Realtime sessions may persist durable metadata needed for audit, debugging, timeout recovery, and action plan linkage. Raw audio, raw provider prompts, raw provider responses, and generated speech must not be durably persisted unless a future retention spec defines a secure policy for doing so.

Action plans created during a realtime session must be stored through the action plan persistence boundary and linked to the session by ID. Approval of a plan must still authorize and execute commands at execution time.

The first durable realtime session metadata slice persists only safe operational metadata:

- Session ID.
- Tenant ID.
- Inventory ID.
- Principal ID.
- Source.
- Lifecycle state: `started`, `completed`, `failed`, or `cancelled`.
- Selected provider profile IDs for speech-to-text, language inference, and text-to-speech.
- Start, last-activity, and end timestamps.
- Safe failure code when a session fails.

The first session metadata repository must live behind a project-owned port. Application services may save the initial record after authorization and provider resolution succeed, and may update the final outcome when the session completes, fails, or is cancelled. The repository must not store raw audio, raw transcripts, raw provider prompts, raw provider responses, generated speech bytes, provider credentials, bearer tokens, provider-specific session identifiers, or hidden inventory data.

The first implementation may persist only sessions that pass startup authorization and provider resolution. Pre-start failures such as unauthenticated, unauthorized, malformed start messages, or missing provider profiles may remain transient until a future security analytics spec defines safe failure telemetry for rejected sessions.

## Initial Transport Dependency

The first Go realtime WebSocket adapter uses `nhooyr.io/websocket v1.8.17`.

The dependency is allowed because the Go standard library does not include a WebSocket implementation, realtime voice requires actual WebSocket upgrade behavior at the HTTP boundary, and tests must exercise that real boundary. The adapter must keep `nhooyr.io/websocket` types inside the HTTP adapter package. Application services, domain code, provider ports, mobile application services, and tests outside the transport adapter must use project-owned message and session types.

## Provider Interfaces

- Speech-to-text, language inference, and text-to-speech must be separate ports.
- Each provider category must expose a common project-owned interface so providers can be swapped without changing application or domain behavior.
- Providers may be remote or local.
- Providers may support streaming or non-streaming behavior.
- Provider capability differences must be represented explicitly instead of leaking provider-specific APIs into application code.
- The system must be able to compose streaming speech-to-text, streaming inference, and streaming text-to-speech when available.
- The system must also support providers that only return complete responses.
- Provider adapters must authenticate to providers using credentials resolved by the core API from tenant provider profiles.
- Provider credentials must never be sent to clients.
- Provider adapters must expose safe project-owned capability metadata, including whether they support streaming input, streaming output, tool calls or structured output, supported audio formats, and typed limits that affect validation.

## Initial Provider Compatibility

The language inference port must be project-owned and must not be identical to any provider API.

The first language inference adapters should include OpenAI-compatible HTTP shapes because common local runtimes expose OpenAI-compatible Chat Completions and related APIs. Provider compatibility must be represented as adapter behavior, not as a domain or action-plan dependency.

Speech-to-text and text-to-speech ports must also be project-owned. Their adapters may target OpenAI-compatible endpoints, provider-native endpoints, or local HTTP endpoints, but clients and application services must interact only with Stuff Stash realtime messages and project-owned provider ports.

## Security And Privacy

- Realtime messages must not expose hidden tenant or inventory data.
- Audio, transcripts, prompts, model responses, and generated speech must not be logged by default.
- Realtime provider failures must fail safely.
- Realtime connections must reject unauthenticated, unauthorized, expired-token, wrong-tenant, wrong-inventory, and privilege-escalation attempts.
- Tenant provider profile management must require tenant configuration permission.
- Realtime session startup must fail safely when required provider profiles are missing, disabled, unsupported for the requested capability, or have unusable credentials.
- Provider error responses must be mapped to safe client errors without exposing credentials, provider account details, raw prompts, raw transcripts, raw model output, hidden inventory data, endpoint internals, or stack traces.

## Testing

- Tests must use fake realtime transports and fake providers.
- Tests must cover connection authentication, authorization, streaming events, cancellation, provider failure, malformed messages, approval, and command execution.
- Security-sensitive realtime behavior must have adversarial end-to-end tests.
- Tests must cover missing provider profiles, disabled provider profiles, malformed provider configuration, unusable encrypted credentials, provider timeout, provider streaming interruption, and provider output that attempts to bypass confirmation or authorization.
- Tests must verify provider credentials are not exposed in realtime events, REST responses, OpenAPI examples, logs, audit records, or observability metadata.
- Tests must verify clients cannot obtain provider credentials, provider session tokens, provider-specific realtime URLs, or direct provider bootstrap payloads through REST, realtime events, or provider-profile APIs.
- Tests must verify raw audio, raw transcripts, raw provider prompts, raw provider responses, and generated speech are not durably persisted before a retention and redaction policy is specified.

## Open Questions

- What exact WebSocket message schema should be used?
- Should realtime interaction support resumable sessions?
- What audio format should mobile clients stream first?
- Which realtime provider capabilities are required for the first release?
