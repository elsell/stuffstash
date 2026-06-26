# Conversational Provider Profiles Spec

## Purpose

Stuff Stash needs tenant-scoped provider profiles so tenant administrators can configure speech-to-text, language model, and text-to-speech providers through the product without changing deployment configuration.

Provider profiles make conversational inventory usable for self-hosted Docker deployments, local model deployments, and remote model provider deployments while keeping provider details behind ports and adapters.

## Scope

This spec covers conversational provider profile management, credential sealing, authorization, audit behavior, and testing requirements.

This spec does not define the final UI layout, provider SDK, realtime message schema, or provider-specific request format.

## Architecture

Provider profiles belong to the agent/model bounded context.

Provider profiles are tenant-scoped application configuration, not inventory domain entities. Asset, location, inventory, audit, identity, search, media, and other product domains must not depend on provider profile types.

Provider profile application services must expose project-owned typed operations for create, update, disable, enable, delete or archive, list, detail, and connection test behavior once the management API is specified.

Provider profile persistence must live behind an explicit repository port. Realtime adapters and provider adapters must resolve profiles through application services or narrow query ports, not by reading persistence directly.

The first implementation slice must introduce the tenant-scoped provider profile model, repository port, GORM persistence adapter, and application service before adding REST or realtime provider-profile resolution. The application service must be the only write boundary for profile lifecycle changes and must require tenant configuration permission.

## Profile Model

Each provider profile must include:

- Provider profile ID.
- Tenant ID.
- Capability: speech-to-text, language inference, or text-to-speech.
- Provider kind, such as OpenAI-compatible HTTP, Gemini, local HTTP, or another specified adapter kind.
- Display name.
- Endpoint URL when required by the provider kind.
- Model name or deployment name when required by the provider kind.
- Non-secret runtime options as typed project-owned values.
- Project-owned capability metadata.
- Credential status metadata.
- Lifecycle state.
- Creation, update, and last-tested timestamps.

Provider profile IDs must be stable. Provider kinds, capabilities, lifecycle states, and credential status values must be enumerations, not magic strings.

Provider profiles must not store raw provider credentials, raw prompts, raw transcripts, raw provider responses, raw audio, generated speech bytes, or provider-specific realtime session identifiers.

The first repository record must store non-secret runtime options and capability metadata as project-owned JSON blobs. The application service must validate that these blobs are syntactically valid JSON objects when supplied. Raw credentials must not be accepted in provider profile create or update inputs; credential replacement is a separate operation through the credential-sealing boundary.

## Lifecycle

The first lifecycle states are:

- `enabled`: available for realtime session provider resolution.
- `disabled`: retained for configuration and audit history, but not eligible for new realtime sessions.
- `archived`: hidden from normal selection and not eligible for new realtime sessions.

Deleting a provider profile may be implemented as archival unless a later spec requires hard delete. Hard delete must not remove audit records.

The first lifecycle commands are create, enable, disable, and archive. Create starts profiles in `disabled` unless the caller explicitly requests `enabled`. Archived profiles cannot be re-enabled in the first slice. Realtime provider resolution must ignore disabled and archived profiles.

Disabling, archiving, deleting, or replacing credentials for a profile must not affect already completed action plans or audit history. New realtime sessions must fail safely when their required provider profile is disabled, archived, missing, unsupported, or has unusable credentials.

## First Management API

The first REST management API must live under the tenant scope:

- `POST /tenants/{tenantId}/provider-profiles`
- `GET /tenants/{tenantId}/provider-profiles`
- `GET /tenants/{tenantId}/provider-profiles/{providerProfileId}`
- `POST /tenants/{tenantId}/provider-profiles/{providerProfileId}/enable`
- `POST /tenants/{tenantId}/provider-profiles/{providerProfileId}/disable`
- `POST /tenants/{tenantId}/provider-profiles/{providerProfileId}/archive`
- `PUT /tenants/{tenantId}/provider-profiles/{providerProfileId}/credential`

All provider-profile management endpoints must require the same bearer-token authentication boundary as the rest of the API and tenant configuration permission for the requested tenant. Viewers, editors, unrelated users, unauthenticated users, wrong-tenant users, expired-token users, and malformed-token users must be rejected.

Provider profile responses must include safe profile metadata only: profile ID, tenant ID, capability, provider kind, display name, endpoint URL, model name, non-secret runtime options, capability metadata, credential status, lifecycle state, and timestamps. Responses must never include raw credentials, sealed credential ciphertext, nonce material, encryption key IDs, provider account details, provider session tokens, provider-specific realtime URLs, raw prompts, raw transcripts, raw model responses, raw audio, or generated speech.

Credential replacement requests must accept raw credential material only in the request body for the duration of that request. The application service must seal the credential through the configured credential-sealing port before persistence, store it through the credential repository boundary, supersede prior active credentials for the same tenant/profile/capability/provider-kind/purpose, update the provider profile credential status to `configured`, and return only safe profile metadata.

## Authorization

Provider profile management requires tenant configuration permission.

Users without tenant configuration permission must not create, update, disable, enable, archive, delete, test, list sensitive metadata for, or replace credentials on provider profiles.

Realtime sessions use provider profiles only after the authenticated principal is authorized for the requested tenant and inventory scope. Provider profiles must not grant additional inventory, tenant, or application permissions to the principal, model provider, agent loop, or realtime adapter.

## Credential Sealing

Provider credentials entered through the UI or API must be sealed before persistence through a credential-sealing port owned by the application layer.

The default infrastructure adapter may store encrypted credential material in the database. This supports Docker-based self-hosted deployments without requiring Kubernetes secrets or an external secret manager.

Credential sealing must follow these rules:

- Raw provider credentials must never be persisted.
- Raw provider credentials must never be returned by API responses after creation or update.
- Raw provider credentials must never be logged, audited, included in observability metadata, or exposed in generated OpenAPI examples.
- The encryption scheme must provide authenticated encryption, not only confidentiality.
- The credential sealer must bind ciphertext to tenant ID, provider profile ID, capability, provider kind, and credential purpose as authenticated associated data or an equivalent authenticated scope.
- The credential sealer must reject unseal attempts when authenticated scope metadata does not match the requested tenant, profile, capability, provider kind, and purpose.
- The persistence layer may store ciphertext, nonce or initialization vector material, key identifier, encryption algorithm identifier, creation timestamp, update timestamp, and safe credential status metadata.
- The root encryption key or key-encryption key must come from environment-backed runtime configuration or a mounted runtime secret.
- The service must fail closed at startup when encrypted provider credentials exist but no usable decryption key is configured.
- Create, update, credential replacement, and provider test operations that require credentials must fail before accepting raw credentials when the credential-sealing adapter is unavailable or misconfigured.
- Key rotation must be supported by recording a key identifier with each encrypted credential and by providing a re-encryption path before production use.
- UI and API responses may show safe metadata such as whether a credential is configured, when it was last updated, and which provider profile uses it.
- Deleting, archiving, or replacing a credential must remove or supersede the prior encrypted credential material so inactive credentials are not accidentally used.

External secret managers may be added later behind the same credential-sealing or credential-store port, but the product architecture must not require Kubernetes secret references or operator-managed environment variables for tenant-level provider configuration.

## First Credential Adapter

The first credential-sealing adapter must use Go standard library AES-256-GCM with random 96-bit nonces. AES-GCM is chosen instead of Fernet so the first slice can use authenticated encryption without adding a new cryptographic dependency.

The first adapter must:

- Accept one active 32-byte root key from environment-backed runtime configuration, encoded as unpadded or padded base64.
- Require a non-empty key identifier and persist that key identifier with sealed credentials.
- Persist `AES-256-GCM` as the algorithm identifier.
- Generate a fresh nonce for every seal operation.
- Bind tenant ID, provider profile ID, capability, provider kind, and credential purpose as authenticated associated data.
- Fail closed when the key is missing, malformed, the key identifier is missing, the algorithm is unsupported, the key identifier does not match, the authenticated scope does not match, or ciphertext authentication fails.
- Expose only sealed credential metadata and ciphertext to persistence; raw credentials may exist only in request memory and provider adapter memory after successful unseal.

Provider credentials persisted in the database must live behind a credential repository port. The first GORM adapter stores encrypted credential material in a tenant-scoped table with provider profile ID, capability, provider kind, credential purpose, key ID, algorithm, nonce, ciphertext, creation timestamp, update timestamp, and superseded timestamp. Repository reads for active credentials must require tenant ID, provider profile ID, capability, provider kind, and credential purpose.

## Provider Resolution

Realtime session startup must resolve provider profiles through a tenant-scoped provider resolution service.

Resolution must verify:

- The profile belongs to the requested tenant.
- The profile has the required capability.
- The profile lifecycle state is `enabled`.
- The selected adapter supports the provider kind.
- Required endpoint, model, runtime options, and credentials are present and valid enough to attempt provider use.
- The authenticated principal is authorized for the tenant and inventory scope requested by the realtime session.

Resolution must fail safely with user-safe errors and safe observability when a profile is missing, disabled, archived, malformed, unsupported, or has unusable credentials.

## Provider Testing

Provider profile test operations must run through provider ports and adapters.

Test operations must:

- Require tenant configuration permission.
- Use sealed credentials only after successful unseal.
- Avoid sending real tenant inventory data unless a future spec explicitly permits a scoped test.
- Return safe success or failure metadata.
- Avoid exposing provider credentials, account details, raw provider responses, endpoint internals, stack traces, prompts, transcripts, or hidden inventory data.
- Emit safe observability and audit history.

## Audit And Observability

Provider profile create, update, enable, disable, archive, credential replacement, and provider test operations must produce audit history.

Audit and observability metadata may include provider profile ID, tenant ID, capability, provider kind, lifecycle state, credential configured status, latency, and safe failure category.

Audit and observability metadata must not include raw credentials, sealed ciphertext, encryption keys, raw provider request bodies, raw provider responses, raw transcripts, raw prompts, audio, generated speech, or hidden inventory details.

## Testing

Tests must use fakes for credential sealing, provider adapters, authorization, repositories, audit history, and observability.

Tests must cover:

- Tenant administrator profile create, update, list, detail, enable, disable, archive, credential replacement, and provider test behavior.
- Viewer, editor, unrelated user, wrong-tenant user, unauthenticated user, expired-token user, and malformed-token rejection.
- Cross-tenant and cross-profile ciphertext swapping attempts.
- Wrong capability, wrong provider kind, wrong credential purpose, and wrong profile unseal attempts.
- Missing, malformed, disabled, archived, unsupported, and unusable provider profiles.
- Missing or misconfigured credential-sealing adapter fail-closed behavior for startup, create, update, credential replacement, and provider test operations.
- Provider test failure safety.
- Credential redaction from REST responses, realtime events, OpenAPI examples, logs, audit records, and observability metadata.
- Rejection of any REST, realtime, or provider-profile API behavior that exposes provider credentials, provider session tokens, provider-specific realtime URLs, or direct provider bootstrap payloads to clients.

## Open Questions

- Which exact REST paths should manage tenant provider profiles?
- What key rotation workflow should be required before production deployments store tenant provider credentials?
- Which provider profile fields should be editable after a profile has been used by completed realtime sessions?
