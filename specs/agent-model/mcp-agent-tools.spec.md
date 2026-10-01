# MCP Agent Tools Spec

## Purpose

Stuff Stash needs a native Model Context Protocol server so external agents can inspect and manage inventory through the same secure application boundary used by REST, mobile, web, and conversational flows.

The MCP server also gives the internal conversational agent loop a useful tool-catalog shape without making MCP transport the core application boundary.

## Scope

This spec covers the first architecture for Stuff Stash MCP tools, authentication, authorization, tool catalog ownership, action-plan interaction, audit behavior, and tests.

This spec does not define the final MCP SDK, exact tool schemas, every tool, generated client behavior, external agent UX, or deployment topology.

## Architecture Decision

The Stuff Stash MCP server is an adapter over application services.

MCP must not become the domain boundary, application-service boundary, authorization boundary, persistence boundary, or audit boundary. MCP tools must call project-owned application services and ports in the same way REST, realtime, mobile, web, CLI, and future adapters do.

The internal conversational agent loop may reuse the same project-owned tool catalog and JSON-schema-like tool descriptors, but it does not need to call the public MCP server over HTTP. The public MCP transport is for external MCP clients. Internal voice and text orchestration should call the tool catalog or application services directly through in-process ports to avoid self-calls, duplicate authentication, avoidable latency, and transport-specific coupling.

## Transport

The first remote MCP transport should use Streamable HTTP unless a future spec chooses another MCP transport before implementation.

Stdio MCP is not the first Stuff Stash server target because Stuff Stash is a multi-tenant service with authenticated users, shared inventories, audit history, and central authorization. Stdio may be considered later for local developer or single-user self-hosted workflows only after its security and operational model is specified.

The MCP server may be hosted by the same API process as the REST and realtime adapters or by a separate process using the same application composition. Either deployment must preserve the same authentication, authorization, observability, configuration, and application-service boundaries.

## Authentication

The remote MCP server must use the same authentication boundary as the main Stuff Stash application.

MCP requests must authenticate with bearer tokens issued by the configured Stuff Stash authentication flow. In production-shaped deployments, this means OIDC ID token verification through the same authentication adapter family used by the API. In local development and tests, MCP may use the explicit `local-dev` authentication mode when the API is also configured for local development.

MCP authentication must follow these rules:

- MCP must not introduce a separate user database, API-key identity model, static shared token, or MCP-only principal.
- MCP must not accept unauthenticated tool calls.
- MCP must reject local development bearer tokens unless the MCP adapter's own environment-backed authentication mode explicitly enables local development authentication.
- Production-shaped MCP deployments must reject local development bearer tokens.
- MCP must normalize provider-specific OIDC claims into the same project-owned principal shape used by REST and realtime adapters.
- MCP must reject missing, malformed, expired, wrong-issuer, wrong-audience, unsigned, or otherwise invalid tokens.
- MCP authentication failures must be bland and must not reveal whether a tenant, inventory, user, provider account, or credential exists.
- MCP server metadata required by the MCP authorization model must point clients at the same configured authorization issuer or protected-resource metadata used for Stuff Stash authentication once the exact MCP authorization metadata shape is selected.
- MCP clients are responsible for obtaining user authorization through the configured OIDC flow. The MCP server validates bearer tokens; it must not exchange provider credentials with model providers or external agents.

MCP authentication middleware may share implementation with HTTP authentication middleware where practical, but authentication must remain behind the project-owned authentication port.

## Authorization

Every MCP tool call must authorize at execution time using the authenticated principal, tenant, inventory, and target resource.

MCP tools must not receive independent authorization grants. A language model, MCP client, MCP server, external agent, generated SDK, or internal agent loop must never gain permissions beyond the authenticated principal who initiated the call.

Authorization must follow these rules:

- Tenant-scoped tools must require tenant authorization.
- Inventory-scoped tools must require tenant and inventory authorization.
- Resource-scoped tools must require tenant, inventory where applicable, and target-resource authorization.
- Viewer principals may use read tools only where their inventory permissions allow it.
- Editor principals may use edit tools only where their inventory permissions allow it.
- Tenant and inventory configuration tools require the same configuration or sharing permissions as REST and web workflows.
- Cross-tenant and cross-inventory access attempts must fail safely without leaking hidden resource existence.
- Authorization checks must use the existing authorization port and SpiceDB-backed relationship model when configured.

## Tool Catalog

Stuff Stash must define a project-owned tool catalog in the agent/model application layer.

The tool catalog owns:

- Stable tool names.
- Human-oriented tool descriptions suitable for model use.
- Project-owned input and output schemas.
- Required tenant, inventory, and permission metadata.
- Read/write classification.
- Confirmation and action-plan requirements.
- Safe error categories.
- Observability metadata keys.

Tool catalog permission metadata is descriptive planning and routing metadata. It is not an authorization policy source. Tool execution must still call the owning application service and authorization port at execution time.

Tool descriptors must not contain raw prompts, provider credentials, hidden tenant data, provider-specific model response shapes, provider-specific tool-call formats, framework types, GORM models, Huma types, or MCP SDK-specific structs.

The MCP adapter maps project-owned tool descriptors to MCP tool definitions. The realtime conversational adapter maps the same catalog to the selected language-provider tool format when a provider supports tools or structured output. Providers that do not support native tools may still use deterministic structured-output prompting behind the language inference port.

## First Tool Slice

The first MCP and internal-agent tool slice should be read-only:

- Search authorized assets.
- Get asset detail.
- List assets in a location.
- List root-level assets in an inventory.
- List currently checked-out assets.
- List checkout history for an authorized asset.
- List accessible tenants and inventories for the authenticated principal.

The first write-capable slice should come only after action-plan confirmation is implemented:

- Create asset.
- Create location-like asset.
- Move asset.
- Update asset details.
- Archive and restore asset.
- Check out and return asset.

Write tools must map to application commands, not persistence operations.

## Action Plans And Confirmation

External MCP tool calls are direct adapter calls from an authenticated client. They do not automatically prove user approval for risky state changes.

Write tools used by the internal conversational voice/text loop must produce or execute through structured action plans according to `specs/agent-model/conversational-action-plan.spec.md`.

For external MCP clients:

- Read tools may execute immediately after authentication and authorization.
- Write tools must require an explicit approved action plan, confirmation token, or equivalent user-mediated approval flow unless a future spec explicitly whitelists a narrow low-risk write contract for direct MCP execution.
- Destructive or hard-to-reverse tools must not be exposed for direct execution until their confirmation and undo behavior are specified.
- Approval of one action plan must not authorize unrelated MCP tool calls.

Model output, MCP tool input, and external-agent reasoning must not be treated as authorization or approval.

## Audit And Observability

MCP tool calls must produce domain-oriented observability through ports.

State-changing MCP tool calls must produce audit history with the same action names and metadata standards used by equivalent REST, web, mobile, and realtime workflows.

Safe read audit history must be produced where the relevant domain spec requires read audit behavior.

Audit and observability metadata may include:

- MCP session ID or request ID.
- Tool name.
- Tenant ID.
- Inventory ID when present.
- Authenticated principal ID.
- Read/write classification.
- Safe outcome category.
- Latency and provider-independent failure category.

Audit and observability metadata must not include raw bearer tokens, OIDC claims, provider credentials, raw prompts, raw model responses, raw transcripts, raw audio, generated speech, hidden inventory data, or MCP client secrets.

## Security And Privacy

MCP is a security-sensitive adapter.

The MCP server must:

- Reject unauthenticated calls.
- Reject unauthorized calls.
- Preserve tenant and inventory isolation.
- Return bland authentication, authorization, and hidden-resource errors.
- Avoid exposing internal package names, stack traces, SQL errors, SpiceDB internals, OIDC provider internals, secrets, tokens, filesystem paths, or infrastructure details.
- Rate-limit remote MCP traffic through the same rate-limiter boundary once rate limiting exists.
- Avoid tool descriptions that reveal hidden tenant data, credentials, internal prompts, or implementation details.
- Treat tool inputs as untrusted external input and validate them before application-service execution.
- Defend against tool-input attempts to smuggle unapproved commands, hidden resource IDs, cross-tenant references, or prompt-injection instructions.

## Testing

Focused application and adapter tests may use fakes for MCP transport, authentication, authorization, application services, repositories, audit history, and observability.

Adversarial MCP boundary tests must exercise the real remote MCP HTTP adapter with the configured authentication adapter family and authorization port. Boundary tests must verify bearer-token parsing, authentication middleware ordering, authorization execution, safe errors, local-development mode separation, and production-shaped OIDC behavior.

Tests must cover:

- Tool listing with valid authentication.
- Missing, malformed, expired, wrong-issuer, wrong-audience, and unauthenticated token rejection.
- Local development authentication when explicitly configured.
- Local development token rejection when MCP is not explicitly configured for local development authentication.
- Local development token rejection in production-shaped OIDC mode.
- Viewer read success and write denial.
- Editor read and allowed edit success.
- Tenant configure permission checks for configuration tools.
- Wrong-tenant, wrong-inventory, hidden-resource, and privilege-escalation attempts.
- Attempts to smuggle authority through MCP-supplied principal IDs, tenant IDs outside the authenticated context, role hints, permission hints, model assertions, tool metadata, or forged approval artifacts.
- Tool input validation and safe error envelopes or MCP error results.
- Read tool behavior for search, asset detail, location contents, root inventory contents, and accessible tenant/inventory listing.
- Write tool rejection before confirmation behavior is specified.
- Action-plan-backed write execution once write tools are enabled.
- Audit and observability emission for successful, denied, failed, cancelled, and malformed tool calls.
- Redaction of bearer tokens, OIDC claims, provider credentials, raw prompts, raw model responses, raw transcripts, raw audio, generated speech, hidden inventory data, and infrastructure details.
- Verification that MCP tools cannot bypass the same application services used by REST and realtime adapters.

Security-sensitive MCP behavior must have adversarial end-to-end tests before implementation is considered complete.

## Open Questions

- Which Go MCP SDK or minimal protocol implementation should be pinned first?
- Which exact MCP authorization metadata endpoints should Stuff Stash expose for remote clients?
- Should MCP tools live under the same public base URL as REST or under a separate configured origin?
- What is the first external MCP client to verify against: Claude Desktop, Claude Code, ChatGPT/OpenAI Responses, Gemini CLI, or a protocol-level test client?
- What confirmation-token or approved-action-plan contract should external MCP clients use for write tools?

## Read-only delivery contract — October 2026

Close G2 with an externally usable, authenticated read-only endpoint, not an
internal tool registry or a transport placeholder. Writes remain explicitly
unpublished until the approval contract above is delivered.

### Deployment and protocol

- Use the official `github.com/modelcontextprotocol/go-sdk` pinned to `v1.8.0`.
  It supports the project's Go 1.25 toolchain and supplies protocol negotiation,
  JSON-RPC validation and bounded transport decoding. Its transport hardening
  makes it preferable to implementing another JSON-RPC/MCP stack locally.
- Register `/mcp` through a dedicated adapter registration function, alongside
  REST in the API process and inside the existing security-header and rate-limit
  middleware. Keep server construction and bootstrap wiring thin.
- Use stateless Streamable HTTP with JSON responses. Explicitly support protocol
  revisions `2026-07-28` and `2025-11-25`; do not advertise resources, prompts,
  server-to-client requests, subscriptions or write tools. Negotiation and the
  legacy initialization handshake belong to the SDK. No transport session may
  carry a principal across HTTP requests.
- `STUFF_STASH_MCP_ENABLED` defaults to false. When enabled,
  `STUFF_STASH_MCP_PUBLIC_URL` must be a canonical absolute HTTPS URL ending in
  `/mcp`, with no user information, query or fragment. Explicit local-dev mode
  may use HTTP on a loopback host. Reject malformed configuration at startup.
- `STUFF_STASH_MCP_AUTH_MODE` must be explicitly set to the same `oidc` or
  `local-dev` mode as the API. No inherited local-dev default. Authenticate every
  protocol request through the existing application authentication boundary;
  reject local-dev tokens in OIDC mode before any tool execution.
- Validate Host against the configured public endpoint and validate any Origin
  against its origin or the explicitly configured CORS origins. Reject malformed,
  multiple and untrusted origins. Do not trust forwarded host headers to choose
  an authentication issuer or public endpoint. Extend CORS only for required MCP
  protocol headers, keeping the existing origin allowlist.
- Enforce the configured API JSON body limit and request deadlines before SDK
  decoding. Use `STUFF_STASH_HTTP_WRITE_TIMEOUT` as a context deadline, not only
  a socket timeout; default to 30 seconds when constructed without a value. Rate-limit both discovery and tool traffic using the existing port.
  Responses with inventory data are private/no-store. Never reflect raw adapter,
  authorization-provider or repository errors to an external agent.

### Authentication discovery

Publish `/.well-known/oauth-protected-resource/mcp` when MCP is enabled, with its
canonical resource URL and the same configured OIDC issuer in
`authorization_servers`. A 401 response uses a Bearer challenge pointing to that
metadata and a bland response body. Expose `WWW-Authenticate` to allowed browser
origins so they can read the challenge. Metadata must contain no credentials or
principal information. Local-dev metadata does not advertise an OAuth issuer.

This uses the existing Stuff Stash bearer-token contract (verified OIDC ID tokens
with the configured audiences). Do not imply that Google access tokens, arbitrary
MCP client registrations, token exchange or dynamic registration are supported.
Document a configured-token client example and the provider/client registration
needed to obtain a token. Changing the project to a distinct OAuth access-token
resource/audience requires a separate authentication specification; it must not
be smuggled into this adapter.

### Catalog and query behavior

The application-owned catalog lives under the agent/model application package;
SDK types remain in the MCP adapter. The catalog defines these stable tool names,
required scope fields, bounded pagination and read-only annotations:

| Tool | Required scope and behavior |
| --- | --- |
| `list_tenants` | Authenticated principal; only accessible tenants |
| `list_inventories` | `tenantId`; only accessible inventories |
| `search_assets` | `tenantId`, `inventoryId`, `query`; authorized exact/fuzzy search |
| `get_asset` | `tenantId`, `inventoryId`, `assetId`; current detail |
| `list_root_assets` | `tenantId`, `inventoryId`; active assets with no parent |
| `list_location_assets` | `tenantId`, `inventoryId`, `assetId`; active direct children of an authorized container/location |
| `list_checked_out_assets` | `tenantId`, `inventoryId`; current checkouts |
| `list_asset_checkout_history` | `tenantId`, `inventoryId`, `assetId`; checkout history |

- Call the same owning application queries and authorization ports as REST.
  Never authorize through descriptor metadata alone or query a repository from
  the adapter. Parent-filtered listing belongs in the asset application query and
  repository contract, with the parent predicate applied before pagination.
- Every paged tool uses the application's bounded `limit` and opaque `cursor`
  contract. Root/parent filters participate in cursor scope so a cursor cannot
  be reused against another location. Do not load every asset and then filter a
  client-sized page. Check parent access and containment kind before listing.
- Parse IDs and typed query values before execution. Reject unknown input fields,
  invalid limits, malformed cursors and unsupported lifecycle/mutation arguments.
  Inventory search requires an explicit inventory, even if tenant-wide search is
  available through another adapter.
- Project outputs explicitly: user-visible IDs, names/descriptions, containment,
  lifecycle, expiration, tags/custom fields and checkout details as applicable,
  with pagination metadata. Exclude storage keys, signed attachment URLs,
  credentials, invitation tokens, hidden schema and raw transport/domain objects.
  Text from assets is data, never a tool instruction or authorization grant.
- Use `audit.SourceMCP` for domain reads where audit is required. Record tool name,
  safe outcome and elapsed time through injected observability. Never log raw
  arguments, returned inventory content, bearer tokens or model prompts.

### Required completion evidence

Exercise the HTTP endpoint using the official client, not only direct tool
callbacks: discovery, every read tool, pagination, missing/invalid arguments,
unknown write tools, read audit source, revocation and current-principal handling.
Use a real signed-token test issuer for expired, unsigned, wrong-issuer,
wrong-audience and malformed-token rejection. Verify explicit local-dev mode,
production rejection of dev tokens, viewer/editor read success, cross-tenant and
cross-inventory denial, hidden parent/resource denial, limits, Origin/Host checks,
rate limiting and disabled configuration. A real client must retrieve an asset
created through the existing authenticated REST boundary. Preserve bland errors
and prove rejected requests do not execute or publish inventory data.

Sources: [official Go SDK v1.8.0](https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.8.0),
[SDK protocol documentation](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/docs/protocol.md),
and [MCP Streamable HTTP](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports).
