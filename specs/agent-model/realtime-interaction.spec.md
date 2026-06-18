# Realtime Interaction Spec

## Purpose

Stuff Stash needs a realtime interaction channel for conversational inventory so users can speak, receive feedback, approve plans, and hear responses without waiting on a page-style request cycle.

## Scope

This spec covers realtime conversational transport expectations for web and mobile clients.

This spec does not define the final wire protocol, message schema, streaming codec, provider SDK, or deployment topology.

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

## Provider Interfaces

- Speech-to-text, language inference, and text-to-speech must be separate ports.
- Each provider category must expose a common project-owned interface so providers can be swapped without changing application or domain behavior.
- Providers may be remote or local.
- Providers may support streaming or non-streaming behavior.
- Provider capability differences must be represented explicitly instead of leaking provider-specific APIs into application code.
- The system must be able to compose streaming speech-to-text, streaming inference, and streaming text-to-speech when available.
- The system must also support providers that only return complete responses.

## Security And Privacy

- Realtime messages must not expose hidden tenant or inventory data.
- Audio, transcripts, prompts, model responses, and generated speech must not be logged by default.
- Realtime provider failures must fail safely.
- Realtime connections must reject unauthenticated, unauthorized, expired-token, wrong-tenant, wrong-inventory, and privilege-escalation attempts.

## Testing

- Tests must use fake realtime transports and fake providers.
- Tests must cover connection authentication, authorization, streaming events, cancellation, provider failure, malformed messages, approval, and command execution.
- Security-sensitive realtime behavior must have adversarial end-to-end tests.

## Open Questions

- What exact WebSocket message schema should be used?
- Should realtime interaction support resumable sessions?
- What audio format should mobile clients stream first?
- Which realtime provider capabilities are required for the first release?
