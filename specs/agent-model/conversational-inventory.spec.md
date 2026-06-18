# Conversational Inventory Spec

## Purpose

Stuff Stash must make inventory updates fast enough that people will actually use it while they are doing real work.

Home inventory is hard because no one is dedicated to maintaining it. People move, use, consume, and discard items while they are focused on another task. The mobile app must make those updates possible with a quick voice interaction instead of forcing a full manual workflow.

## Scope

This spec covers the first product direction for language-model-assisted inventory interactions:

- Voice-first inventory commands.
- Text command parity for the web app and mobile app.
- Speech-to-text integration.
- Text-to-speech integration.
- Language model integration.
- Agentic planning and confirmation.
- Pluggable model providers.
- Safety boundaries for automatic actions.

This spec does not define a specific model provider, prompt format, streaming protocol, mobile screen layout, or persistence schema.

## User Experience Goals

- The mobile app must put the conversational inventory action front and center.
- A user must be able to press a button, speak a command, review any needed confirmation, and continue with their real-world task.
- The web app must offer feature parity for text-based conversational inventory commands.
- The web app should support voice commands when browser capability and permissions allow it.
- The system must reduce friction without hiding important consequences from the user.
- The clients must preserve authenticated sessions so routine inventory updates do not require a sign-in detour.
- Common commands should feel like natural language, not a command-line syntax.

Example command:

> Move my fertilizer off of the shelf in the garage to the wire rack in the garage.

## Required Capabilities

- The system must accept spoken commands from the mobile app.
- The system must transcribe spoken commands into text through a speech-to-text port.
- The system must accept typed commands from web and mobile clients.
- The system must interpret inventory intent through a language model port.
- The system must support an agentic loop that can inspect current inventory state, plan actions, ask clarifying questions, request confirmation, and apply approved changes.
- The system must support text-to-speech responses through a text-to-speech port where the client experience calls for it.
- The system must allow users to complete simple, low-risk updates with minimal taps after the original voice command.
- The system must maintain feature parity between mobile and web for the underlying actions, even when the interaction mode differs.

## Model Provider Pluggability

- Speech-to-text, language model, and text-to-speech capabilities must be represented as ports.
- Provider implementations must be adapters.
- The domain and application layers must not depend on provider SDKs, HTTP APIs, model-specific types, or prompt framework types.
- Users or operators should be able to choose model providers through environment-backed configuration.
- The design must support remote providers such as Gemini API.
- The design must support local models where practical.
- Provider selection must not require changing domain logic.
- Provider credentials, endpoints, model names, and runtime options must come from configuration, not hard-coded values.
- Provider adapters must expose enough metadata for observability, debugging, and cost/performance analysis without leaking secrets or sensitive user content.

## Agent Boundary

- The agent may propose actions, ask questions, and call application ports.
- The agent must not bypass domain services, repositories, authorization, tenancy checks, validation, or audit behavior.
- The agent must not write directly to persistence.
- The agent must not trust model output as authoritative domain state.
- All state-changing operations must be performed by application services that enforce domain rules.
- All state-changing operations must be authorized for the current principal and tenant.
- The agent must preserve a structured action plan before applying changes.
- The action plan must be observable and testable without relying on provider-specific model behavior.

## Inventory Actions

The conversational flow must eventually support at least these actions:

- Create an asset when the user refers to an item that does not exist.
- Create a location when the user refers to a location that does not exist.
- Move an asset from one location to another.
- Record that an asset was consumed, discarded, sold, donated, or otherwise removed.
- Update asset details.
- Query where an asset is located.
- Query what is in a location.
- Ask clarifying questions when the command is ambiguous.

Domain specs must define the exact commands, entities, lifecycle states, and authorization rules before implementation.

## Confirmation Rules

- The system must ask for confirmation before creating an asset from an uncertain reference.
- The system must ask for confirmation before creating a location from an uncertain reference.
- The system must ask for confirmation before creating a new asset.
- The system must ask for confirmation before destructive or hard-to-reverse actions.
- The system may create a missing location without confirmation when the location name, tenant, parent location, and user intent are unambiguous.
- The system may move an existing asset to an existing or newly created location without extra confirmation when the identified asset, source location, destination location, tenant, and principal are unambiguous.
- The confirmation must describe the planned action in human language.
- The confirmation must be backed by a structured action plan.
- The system must support cancellation at confirmation points.

## Ambiguity Handling

- The system must ask a clarifying question when multiple assets match the user's words.
- The system must ask a clarifying question when multiple locations match the user's words.
- The system must ask a clarifying question when the requested action is unclear.
- The system must prefer a short clarification over making a risky guess.
- Clarifying questions must preserve the user's original intent so they do not need to start over.

## Security And Privacy

- Conversational inventory is a security-sensitive interaction surface.
- Every conversational action that touches authenticated or authorized behavior must have adversarial end-to-end tests.
- The agent must operate only within the authenticated user's tenant and permissions.
- Cross-tenant lookup, inference, or action execution must be rejected.
- Model providers must not receive more tenant, user, asset, or location data than needed for the task.
- Sensitive data sent to model providers must be minimized and observable through safe metadata.
- Provider request and response bodies must not be logged by default.
- Error messages must not expose provider secrets, prompts, tokens, tenant data, or hidden inventory records.
- If a provider fails, the system must fail safely and explain the next useful user action.

## Observability

- Conversational flows must use domain-oriented observability through ports.
- Observability must record safe events for transcription, intent interpretation, clarification, confirmation, action execution, provider latency, provider failure, and user cancellation.
- Observability must support fan-out to console logging, OpenTelemetry, audit events, or future sinks.
- Observability must not record raw audio, raw transcripts, prompts, model responses, credentials, or sensitive inventory details unless a future spec explicitly defines a secure redaction and retention policy.

## Testing

- Tests must use fakes for speech-to-text, language model, text-to-speech, inventory, location, authorization, and observability ports.
- Tests must verify real behavior through application boundaries.
- Tests must not mock provider SDKs.
- Tests must cover successful commands, ambiguous commands, missing assets, missing locations, confirmation, cancellation, provider failure, malformed model output, unauthorized access, unauthenticated access, cross-tenant attempts, and privilege escalation attempts.
- Model-dependent flows must be testable with deterministic fake model responses.
- Security-sensitive flows must include adversarial end-to-end tests for mobile, web, REST, and future MCP entry points where applicable.

## Open Questions

- Which model provider should be the first remote adapter?
- Which local speech-to-text, language model, and text-to-speech runtimes should be supported first?
- Which actions are safe enough for one-tap confirmation?
- Should voice audio ever be sent through the backend, or should clients call speech-to-text providers directly through short-lived credentials?
- What retention policy should apply to transcripts, action plans, and audit events?
- How should the UI show confidence without making the user think too hard?
