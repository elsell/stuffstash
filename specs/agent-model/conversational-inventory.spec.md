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

## Realtime Architecture Decision

Conversational inventory realtime interaction must be mediated by the Stuff Stash core API.

Clients must not stream conversational audio, transcripts, prompts, model context, tool calls, or generated speech directly to speech-to-text, language model, or text-to-speech providers as part of the Stuff Stash product flow.

The core API owns:

- Realtime client connection authentication.
- Tenant and inventory authorization checks.
- Tenant provider profile resolution.
- Speech-to-text, language model, and text-to-speech adapter selection.
- Streaming proxy behavior to configured providers.
- Agent loop orchestration.
- Action plan creation, clarification, confirmation, cancellation, and execution.
- Audit history, undo metadata, and safe observability.

This decision keeps provider credentials out of clients, keeps all model-assisted actions inside the same authorization boundary as REST and future MCP interactions, and makes self-hosted Docker deployments viable without provider-specific client credential flows.

Provider adapters may stream to local or remote providers from the API process, but those provider streams are infrastructure details behind ports. Provider streaming protocols, SDKs, request types, response types, and authentication mechanisms must not leak into domain, application command, REST, realtime transport, web, or mobile client code.

## Architectural Boundary

- Conversational inventory is part of the core product experience, not the domain core.
- Language model, speech-to-text, and text-to-speech integrations are interaction mechanisms.
- The inventory domain must remain usable through non-model interfaces such as REST, MCP, CLI, tests, background jobs, imports, and future adapters.
- The domain core must not know whether a command came from a spoken request, typed chat, REST call, MCP tool, mobile screen, or another adapter.
- The domain core must expose clear application operations for inventory behavior.
- Conversational flows must translate user language into those application operations.
- The same domain rules, validation, authorization, tenancy checks, observability, and audit behavior must apply no matter which adapter initiated the operation.
- The system must not create model-specific domain concepts unless those concepts are genuinely part of the business domain and are specified outside the provider integration layer.

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
- The system must support real-time voice and text interaction once the client interaction protocol is specified.
- The initial realtime client transport should use WebSockets unless a future spec chooses a better fit before implementation.
- The system must allow users to complete simple, low-risk updates with minimal taps after the original voice command.
- The system must maintain feature parity between mobile and web for the underlying actions, even when the interaction mode differs.

## Model Provider Pluggability

- Speech-to-text, language model, and text-to-speech capabilities must be represented as ports.
- Provider implementations must be adapters.
- The domain and application layers must not depend on provider SDKs, HTTP APIs, model-specific types, or prompt framework types.
- Provider and prompt details must not leak into asset, location, identity, expiration, or other product domain models.
- Tenant administrators must be able to configure conversational provider profiles through Stuff Stash UI and API surfaces once those management surfaces are specified.
- Provider profiles must be scoped to a tenant.
- Provider profiles must identify the supported capability: speech-to-text, language inference, or text-to-speech.
- Provider profiles must capture endpoint, model name, provider kind, non-secret runtime options, capability metadata, and encrypted credential material where the provider requires credentials.
- Provider profiles must support local HTTP endpoints for self-hosted speech-to-text, language model, and text-to-speech runtimes.
- Provider profiles must support remote provider endpoints.
- Gemini remains a planned remote provider target, but the first common language inference adapter should support OpenAI-compatible endpoints because common local model runtimes expose that shape.
- The design must support OpenAI-compatible local endpoints.
- The design must support local models where practical.
- Provider selection must not require changing domain logic.
- Provider credentials, endpoints, model names, and runtime options must come from tenant provider profiles and runtime configuration, not hard-coded values.
- Provider adapters must expose enough metadata for observability, debugging, and cost/performance analysis without leaking secrets or sensitive user content.
- Provider profile management, lifecycle, authorization, audit, credential sealing, and provider testing are specified in `specs/agent-model/provider-profiles.spec.md`.

## Provider Profile Credential Handling

Provider credentials entered through the UI must be encrypted before persistence.

The first design uses a credential-sealing port owned by the application layer and implemented by infrastructure adapters. The default adapter may store encrypted credential material in the database so Docker-based self-hosted deployments can manage providers through the UI without a separate secret manager.

Credential storage must follow these rules:

- Raw provider credentials must never be returned by API responses after creation or update.
- Raw provider credentials must never be logged, audited, included in observability metadata, or exposed in generated OpenAPI examples.
- The persistence layer may store ciphertext, nonce or initialization vector material, key identifier, encryption algorithm identifier, creation timestamp, update timestamp, and safe credential status metadata.
- The encryption scheme must provide authenticated encryption, not only confidentiality.
- Encrypted credential material must be authenticated to tenant ID, provider profile ID, capability, provider kind, and credential purpose so ciphertext cannot be moved across tenants, profiles, capabilities, provider kinds, or purposes.
- The root encryption key or key-encryption key must come from environment-backed runtime configuration or a mounted runtime secret.
- The service must fail closed at startup when encrypted provider credentials exist but no usable decryption key is configured.
- Profile create, update, credential replacement, and provider test operations that require credentials must fail before accepting raw credentials when the credential-sealing adapter is unavailable or misconfigured.
- Key rotation must be supported by recording a key identifier with each encrypted credential and by providing a re-encryption path before production use.
- UI and API responses may show safe metadata such as whether a credential is configured, when it was last updated, and which provider profile uses it.
- Deleting or replacing a credential must remove or supersede the prior encrypted credential material so inactive credentials are not accidentally used.
- Tests must use deterministic fake credential sealers rather than real provider secrets.

External secret managers may be added later behind the same credential-sealing or credential-store port, but the product architecture must not require Kubernetes secret references or operator-managed environment variables for tenant-level provider configuration.

## Agent Boundary

- The agent may propose actions, ask questions, and call application ports.
- The agent is one adapter-facing orchestration path, not the owner of inventory behavior.
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
- Tests must verify raw transcripts are not durably persisted before a transcript retention and redaction policy is specified.

## Open Questions

- Which local speech-to-text, language model, and text-to-speech runtimes should be supported first?
- Is an external agent-loop framework needed, or are direct application-level orchestration services enough?
- Which actions are safe enough for one-tap confirmation?
- What audio formats and chunk sizes should the API-mediated realtime transport accept first?
- What provider profile management API shape should tenant configuration screens use?
- What retention policy should apply to transcripts, action plans, and audit events?
- How should the UI show confidence without making the user think too hard?
