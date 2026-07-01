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
- Optional tenant-managed prompt template for language inference profiles.
- Credential status metadata.
- Lifecycle state.
- Creation, update, and last-tested timestamps.

Provider profile IDs must be stable. Provider kinds, capabilities, lifecycle states, and credential status values must be enumerations, not magic strings.

Provider profiles must not store raw provider credentials, raw final prompts, raw transcripts, raw provider responses, raw audio, generated speech bytes, or provider-specific realtime session identifiers.

Language inference provider profiles may store an optional tenant-managed prompt template. A prompt template is non-secret configuration used to tune provider-specific phrasing, examples, or response style. It must not include credentials, transcripts, provider responses, audio, generated speech, or tenant inventory data. The API-owned agent loop must still append mandatory safety, tool-use, authorization, structured-output, and no-invention instructions after the tenant template so tenant configuration cannot remove the server-owned agent contract.

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
- `PATCH /tenants/{tenantId}/provider-profiles/{providerProfileId}`
- `POST /tenants/{tenantId}/provider-profiles/{providerProfileId}/enable`
- `POST /tenants/{tenantId}/provider-profiles/{providerProfileId}/disable`
- `POST /tenants/{tenantId}/provider-profiles/{providerProfileId}/archive`
- `PUT /tenants/{tenantId}/provider-profiles/{providerProfileId}/credential`
- `POST /tenants/{tenantId}/provider-profiles/{providerProfileId}/test`

All provider-profile management endpoints must require the same bearer-token authentication boundary as the rest of the API and tenant configuration permission for the requested tenant. Viewers, editors, unrelated users, unauthenticated users, wrong-tenant users, expired-token users, and malformed-token users must be rejected.

Provider profile responses must include safe profile metadata only: profile ID, tenant ID, capability, provider kind, display name, endpoint URL, model name, non-secret runtime options, capability metadata, prompt template when configured for language inference profiles, credential status, lifecycle state, and timestamps. Voice configuration summaries may include the safe credential purpose (`api_key`, `server_adc`, or `oauth_bearer`) so clients can present the correct setup action without inferring from capability alone. Responses must never include raw credentials, sealed credential ciphertext, nonce material, encryption key IDs, provider account details, provider session tokens, provider-specific realtime URLs, raw final prompts, raw transcripts, raw model responses, raw audio, or generated speech.

Provider profile update requests may change only non-secret configuration: display name, endpoint URL, model name, non-secret runtime options, capability metadata, and the language-inference prompt template. They must not change provider profile ID, tenant ID, capability, provider kind, credential status, lifecycle state, creation timestamp, or raw credentials, and callers must not provide an arbitrary last-tested timestamp. Updating configuration must clear `lastTestedAt` because the previous diagnostic result no longer proves the changed configuration. Archived profiles cannot be updated in the first slice.

Credential replacement requests must accept raw credential material only in the request body for the duration of that request, except for explicit server-managed credential purposes that do not require user-supplied secret material. The requested credential purpose must be supported by the selected provider profile's provider kind and capability before the profile can be marked configured. The application service must seal the credential or non-secret credential marker through the configured credential-sealing port before persistence, store it through the credential repository boundary, supersede prior active credentials for the same tenant/profile/capability/provider-kind/purpose, update the provider profile credential status to `configured`, and return only safe profile metadata.

Provider test requests must require tenant configuration permission and return safe metadata only: provider profile ID, capability, provider kind, status, safe message, and test timestamp. Provider test requests may be run for `enabled` or `disabled` configured profiles so tenant administrators can verify a profile before enabling it for realtime sessions. Provider test requests must reject archived, missing, unsupported, malformed, credential-missing, or credential-unusable profiles safely. Successful tests must update the provider profile `lastTestedAt` timestamp and write audit history. Failed tests must return a safe failure and must not expose provider credentials, raw provider responses, stack traces, endpoint internals, prompts, transcripts, audio, generated speech, hidden inventory data, or provider account details.

Provider test credential selection must be based on the provider profile capability and provider kind, not incidental repository order. Gemini provider profiles that use Vertex AI runtime options must support `server_adc` and `oauth_bearer` credentials. `server_adc` uses the API process's Application Default Credentials or workload identity and stores only a non-secret marker so tenant administrators can choose server-managed Google authentication through the UI without pasting a short-lived access token. Tenant-managed `server_adc` profiles must not allow a tenant to spend or authenticate the server identity against arbitrary Google projects: the adapter must use an operator-configured default or allowlist for project, location, and quota project, and must reject profile runtime options outside those bounds. Gemini provider profiles for speech-to-text and language inference must also support Google AI Gemini API `api_key` credentials by sending `x-goog-api-key` to the Gemini API `generateContent` endpoint, and should prefer active `api_key` credentials before `server_adc` before `oauth_bearer` credentials when more than one exists. Google Cloud Text-to-Speech provider profiles support `server_adc` and `oauth_bearer` until a separate provider kind or adapter is specified for API-key-backed speech synthesis.

## Authorization

Provider profile management requires tenant configuration permission.

Users without tenant configuration permission must not create, update, disable, enable, archive, delete, test, list sensitive metadata for, or replace credentials on provider profiles.

Realtime sessions use provider profiles only after the authenticated principal is authorized for the requested tenant and inventory scope. Provider profiles must not grant additional inventory, tenant, or application permissions to the principal, model provider, agent loop, or realtime adapter.

## Credential Sealing

Provider credentials entered through the UI or API must be sealed before persistence through a credential-sealing port owned by the application layer.

The default infrastructure adapter may store encrypted credential material in the database. This supports Docker-based self-hosted deployments without requiring Kubernetes secrets or an external secret manager.

Application services and provider-resolution services must depend on a project-owned provider credential vault port for preparing new credentials and reading active raw provider material. The vault port hides the concrete sealing adapter and credential repository from callers. The first vault adapter composes the AES-256-GCM sealer with the database credential repository so raw credentials are sealed before persistence and unsealed only when constructing provider adapters or running provider diagnostics.

Credential replacement must still remain atomic with provider profile status updates and audit history. To preserve that invariant, the vault may prepare a sealed `ProviderCredentialRecord`, but the provider profile unit of work remains responsible for persisting the profile update, superseding prior active credentials, inserting the new encrypted credential row, and writing the audit record in one transaction.

Credential sealing must follow these rules:

- Raw provider credentials must never be persisted.
- `server_adc` credential records must persist only a non-secret marker. The marker exists so normal profile readiness, audit, testing, and credential replacement semantics remain tenant-scoped; the provider adapter must resolve fresh server ADC tokens at provider call time.
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

The provider credential vault port must support:

- Preparing a sealed credential record from caller-supplied raw credential material, a generated credential ID, authenticated scope, and timestamps.
- Reading active raw credential material by tenant ID, provider profile ID, capability, provider kind, and credential purpose.

The first database-encryption vault adapter must:

- Seal newly prepared credentials through the configured AES-256-GCM sealer.
- Read only active credentials through the credential repository port.
- Unseal active credentials with authenticated associated scope before returning raw material.
- Treat missing credentials as a safe `found=false` result and malformed, mismatched, empty, or undecryptable credentials as invalid provider input.
- Never expose raw credentials through profile responses, realtime events, audit records, observability metadata, logs, or OpenAPI examples.

Startup validation may still query the credential repository port directly to determine whether active encrypted credentials exist before constructing the vault. This allows the API to fail closed when a database contains encrypted credentials but no usable decryption key is configured.

## Provider Resolution

Realtime session startup must resolve provider profiles through a tenant-scoped provider resolution service.

The realtime application service must depend on a project-owned provider resolver port, not direct concrete speech-to-text, language inference, or text-to-speech adapters. The resolver returns the selected speech-to-text, language inference, and text-to-speech provider ports plus the safe selected provider profile IDs for the session. The session must carry that resolved provider set so a profile change after session start does not silently change an in-flight session.

When the selected language inference provider profile has a prompt template, the resolver must return that safe template metadata with the provider set. The realtime application service must pass the resolved template into language inference calls for that session, including final-only recovery turns. It must not persist the rendered final prompt or send the prompt template to the client.

The first compatibility implementation may wrap a process-configured provider set behind the same resolver port for development fakes and the transitional Google smoke-test bridge. That compatibility resolver must not expose provider credentials to clients and must be replaceable by the tenant-profile resolver without changing realtime transport code or the agent loop.

The tenant-profile resolver must:

- List tenant provider profiles through the provider profile repository port.
- Select enabled, configured profiles for speech-to-text, language inference, and text-to-speech.
- Read active credential material through the provider credential vault port using the tenant/profile/capability/provider-kind/purpose scope.
- Build provider adapters through a provider factory interface owned by the provider adapter boundary.
- Return safe profile IDs and provider ports only; raw credentials may live only in resolver/provider-adapter memory for the duration needed to construct or call the provider.

Resolution must verify:

- The profile belongs to the requested tenant.
- The profile has the required capability.
- The profile lifecycle state is `enabled`.
- The selected adapter supports the provider kind.
- Required endpoint, model, runtime options, and credentials are present and valid enough to attempt provider use.
- The authenticated principal is authorized for the tenant and inventory scope requested by the realtime session.

For Gemini speech-to-text and language inference, runtime profile construction must support three credential modes:

- `server_adc`: uses Vertex AI Gemini with project, location, optional quota project, model name, and a fresh token from server Application Default Credentials. Project, location, and quota project must be resolved through operator-controlled defaults or allowlists before tenant runtime options are applied.
- `oauth_bearer`: uses Vertex AI Gemini with `projectId`, `location`, optional `quotaProject`, model name, and bearer authorization.
- `api_key`: uses the Google AI Gemini API with model name, optional endpoint URL, and `x-goog-api-key` authorization. It must not require `projectId`, `location`, or `quotaProject`.

For Gemini-backed Google Cloud Text-to-Speech, runtime profile construction must support `server_adc` and `oauth_bearer`. `server_adc` must use the configured `languageCode`, `voiceName`, operator-bounded quota project, and fresh server ADC tokens. `oauth_bearer` remains supported for compatibility but must not be used for long-running local or production deployments unless the operator accepts token-expiration risk.

Resolution must fail safely with user-safe errors and safe observability when a profile is missing, disabled, archived, malformed, unsupported, or has unusable credentials.

## Voice Provider Configuration

Tenant administrators need an explicit voice pipeline configuration instead of relying on incidental provider profile order. The voice pipeline has three required slots:

- Speech input: `speech_to_text`.
- Agent brain: `language_inference`.
- Spoken output: `text_to_speech`.

Each tenant may store one selected provider profile ID per slot. A selected profile must belong to the tenant and must match the slot capability. Disabled, archived, credential-missing, or untested profiles may be selected only as draft configuration; realtime session startup must still reject them until they are ready. This lets the UI explain "selected but not ready" instead of hiding the profile behind implicit selection.

When no explicit selection exists for a slot, the application may choose a safe fallback for compatibility by selecting the oldest enabled, credential-configured profile for that capability. The response must mark that selection source as `implicit` so clients can invite the administrator to save an explicit choice. Missing or duplicate eligible profiles must not be silently hidden from setup diagnostics.

The voice configuration API must live under the tenant scope:

- `GET /tenants/{tenantId}/voice-provider-configuration`
- `PUT /tenants/{tenantId}/voice-provider-configuration`

Both endpoints require tenant configuration permission. The response must include safe metadata only:

- Selected profile IDs for each required capability, when present.
- Selection source per slot: `explicit`, `implicit`, or `missing`.
- Slot readiness: `ready`, `missing`, `disabled`, `archived`, `credential_missing`, `untested`, `duplicate_candidates`, or `invalid_selection`.
- Safe selected profile summaries using the same redacted metadata as provider profile responses.
- Safe issue messages and recommended action keys for UI routing.
- Duplicate eligible profile summaries by capability.

The response must never include raw credentials, sealed credential material, raw provider errors, raw prompts rendered with server instructions, raw transcripts, raw audio, generated speech bytes, provider account details, provider session tokens, or internal stack details.

Updating the voice configuration must validate all provided selected profile IDs before saving any slot. A profile selected for the wrong tenant, wrong capability, missing profile, or archived profile must fail safely. Successful updates must write audit history and safe observability.

Realtime provider resolution must prefer explicit selected profiles when present. If the selected profile is not enabled, configured, tested, or otherwise usable, resolution must fail safely for that capability instead of falling back to another profile. This prevents hidden provider changes after an administrator has made an explicit selection.

Provider setup UI must present the voice pipeline as the primary mental model. The normal view must show the three slots in order, their selected provider, readiness, direct fix actions, and duplicate warnings. A separate profile inventory or advanced tab may list all profiles, including inactive or archived profiles, but it must not be the first or only way to understand voice readiness.

## Provider Testing

Provider profile test operations must run through provider ports and adapters.

Test operations must:

- Require tenant configuration permission.
- Use sealed credentials only after successful unseal.
- Use a project-owned provider test port so the application layer never imports concrete provider adapters.
- Prefer capability-specific safe provider probes when the selected adapter implements them.
- For language inference, perform a minimal final-response probe that sends no tenant inventory data and verifies the provider can return a project-owned structured response.
- For text-to-speech, synthesize a short safe diagnostic phrase and verify non-empty speech bytes are returned without persisting those bytes.
- For speech-to-text, perform the safest adapter-supported diagnostic probe. If the adapter cannot safely send synthetic non-tenant audio, the probe may verify the same provider endpoint, credential, model, and request path without sending tenant audio, but the response message must not imply that arbitrary microphone transcription has been proven.
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
- Tenant-configured `server_adc` provider profiles resolving through the provider factory without persisted bearer-token material.
- Missing, malformed, disabled, archived, unsupported, and unusable provider profiles.
- Missing or misconfigured credential-sealing adapter fail-closed behavior for startup, create, update, credential replacement, and provider test operations.
- Provider test failure safety.
- Credential redaction from REST responses, realtime events, OpenAPI examples, logs, audit records, and observability metadata.
- Rejection of any REST, realtime, or provider-profile API behavior that exposes provider credentials, provider session tokens, provider-specific realtime URLs, or direct provider bootstrap payloads to clients.

## Open Questions

- Which exact REST paths should manage tenant provider profiles?
- What key rotation workflow should be required before production deployments store tenant provider credentials?
- Which provider profile fields should be editable after a profile has been used by completed realtime sessions?
