# Roadmap Spec

## Purpose

Stuff Stash needs a durable place to record what work should happen next.

This spec exists so a cold-start agent can recover the project direction from the repository without relying on chat history.

## Scope

This spec captures near-term sequencing, current focus, and known follow-up work.

It is not a full product backlog, release plan, issue tracker, or substitute for domain specs.

## Maintenance Rules

- Keep this spec current whenever project focus, sequencing, completion evidence, or known blockers change.
- Do not update this spec for tiny fixes, formatting-only changes, or routine refactors that do not change project direction.
- Keep entries short, concrete, and ordered.
- Move completed work into the repository history; do not let this spec become a changelog.
- If a next step needs domain detail, create or update the relevant domain spec and link to it from the work item.
- If this spec and a domain spec disagree, update the domain spec first, then update this roadmap.

## Current Focus

The current focus is closing the first conversational inventory gaps through the mobile app first, while preserving the promoted SvelteKit web inventory workspace direction as a later parallel surface.

The conversational goal is to move from the first realtime read-only voice slice toward a phone-testable, production-shaped mobile voice loop backed by tenant-configured provider profiles, encrypted provider credentials, a reusable tool catalog, approval-backed write action plans, and stronger agent-loop/session state.

The web goal remains important but is not the immediate starting point. It still needs a production-shaped path through:

- a web visual system based on SvelteKit and Svelte-compatible shadcn primitives,
- clear separation between generated API DTOs and frontend domain models,
- performance-conscious frontend choices,
- generated OpenAPI/client integration without hand-written API clients,
- tenant-first inventory switching through frontend ports and adapters,
- real tenant and inventory loading from authenticated API discovery,
- session-scoped tenant and inventory selection without cross-principal bleed,
- permission-aware empty states and add/create affordances,
- mobile and desktop access to the same tenant-first context switching model,
- focused web adapter tests for tenant selection, empty tenants, and selected-tenant inventory creation.

## Current Evidence

- `f18a6c6 feat: add postgres repository adapter` added the GORM repository adapter, initial tenant and inventory migrations, repository mode configuration, and Postgres-backed Compose path.
- `2791159 chore: add code critic agent` added the code critic custom agent and made post-implementation critic review part of the process.
- `make test` passed after the Postgres repository adapter was added.
- `make docs-build` passed after the Postgres repository adapter was added.
- Docker Compose verification passed on `paul` with Postgres persistence and SpiceDB authorization enabled.
- Postgres on `paul` contained the tenant and inventory rows created by the verification script.
- HTTP-level adversarial tests now cover protected-route auth rejection, unrelated-user denial, tenant-owner inventory listing, inventory-owner list filtering, and safe missing-tenant errors.
- `make verify-spicedb-adapter` passed on `paul` against the pinned local SpiceDB image.
- The real SpiceDB verifier found and drove fixes for tenant viewer relationships and fully consistent permission checks.
- Tenant and inventory creation now write durable state and authorization grant intent through a transactional outbox before SpiceDB relationship writes are drained.
- Authorization outbox retries now run on startup and on an environment-configured interval, not only after create requests.
- Authorization outbox events now use claim IDs and lease deadlines so multiple API replicas do not update the same event at the same time.
- The pinned migration library is wired into the `stuff-stash` binary, and the same image can run `migrate up` or `migrate status`.
- Authorization outbox events now support a terminal dead-letter state for unrecoverable event data problems while keeping transient SpiceDB failures retryable.
- The first asset REST slice implements asset creation, unified `item`/`container`/`location` kinds, same-inventory containment, cursor-paginated asset listing, and adversarial asset authorization tests.
- Inventory listing now uses cursor pagination after authorization filtering, preserving the API collection contract without exposing hidden inventories.
- Direct inventory sharing now supports owner-created viewer/editor grants, cursor-paginated grant listing, outbox-backed SpiceDB relationship writes, and adversarial API tests proving viewers and editors cannot share.
- Custom field definitions now support tenant and inventory scopes, effective inventory listing, cursor pagination, asset value validation, and adversarial API tests for authorization and scope handling.
- Asset update and same-inventory movement now support title, description, parent, and custom field updates while preserving containment invariants, editor/viewer authorization boundaries, and descendant relationships.
- Durable audit history now records the first state-changing tenant, inventory, sharing, custom asset type, custom field definition, and asset actions behind a repository port, with authenticated and authorized paginated REST reads.
- Custom asset types now exist for tenant and inventory scopes, can be assigned to assets, can be renamed with metadata updates, and custom fields can target all assets or specific custom asset types.
- Custom field definitions can now be renamed and safely evolved by adding enum options, adding active custom asset type targets, or expanding targeted fields to all assets while rejecting incompatible narrowing or removals.
- Asset lifecycle now supports archive and restore operations with audit history, active-only default listing, and authorization checks.
- Asset media attachments now support JSON base64 upload, direct upload initiation/completion behind media ports, cursor-paginated listing, raw content download, thumbnail generation behind an image-processing port, model-image preparation readiness behind media ports, local filesystem blob storage, Garage S3-compatible blob storage, audit history, generated OpenAPI, and adversarial API tests.
- Local Dex OIDC verification now runs the full API user flow with two Dex-issued ID tokens and SpiceDB authorization.
- Authorized asset search now supports exact and fuzzy lookup across asset title, description, custom fields, custom asset type metadata, and attachment metadata, with tenant scoping, inventory authorization filtering, lifecycle filtering, cursor pagination, generated OpenAPI, adapter tests, and adversarial API tests.
- Direct inventory access revocation now removes persisted viewer/editor grants, enqueues SpiceDB revoke events through the authorization outbox, records audit history, exposes a no-content REST endpoint, and has adversarial API tests.
- Inventory invite-link tokens now support pending email-scoped invitations, time-limited one-time acceptance tokens, verified-email acceptance, outbox-backed SpiceDB grant creation, revocation, audit history, and adversarial API tests.
- Custom asset type archive now preserves existing asset and custom field target references while hiding archived types from normal lists, blocking new assignments, blocking new field targets, recording audit history, and exposing adversarial API coverage.
- Full REST lifecycle coverage now exists for tenants, inventories, assets, attachments, custom field definitions, custom asset types, access grant detail, and invitation detail/cancel/delete.
- Lifecycle endpoints emit read/write audit records, preserve tenant and inventory security boundaries, and are covered by OpenAPI generation checks plus adversarial HTTP tests.
- The separate SvelteKit web app exists under `apps/web`, uses Dex OIDC with PKCE, uses runtime configuration, calls the API through the generated OpenAPI client boundary, and proves inventory creation, asset creation, active/archived asset browsing, asset archive, asset restore, and asset hard delete.
- `0c8d7d4 feat(web): add asset lifecycle controls` added the first web lifecycle controls and focused frontend interaction tests.
- The current web screens are disposable tracer-bullet UI. Do not expand them into the real product UI before a dedicated UI spec and design workshop.
- `cac140c feat(web): adopt shadcn component primitives` added the Svelte-compatible shadcn foundation and dependency freshness checks.
- `182b7fb feat(web): add audit history viewing` exposed tenant and inventory audit history through the promoted web settings surface.
- `a265e47 feat(web): add custom schema management` added promoted custom asset type and custom field management, create integration, edit integration, and focused web tests behind frontend ports.
- `7daca8e feat(web): add parent quick-create flow`, `63cd859 feat(web): improve containment browsing`, and `ce98824 feat(web): expose location editing` deepened web location browsing, parent quick-create, duplicate-name-safe active rows, nested location navigation, and permission-gated location edit entry.
- `798927d test(web): add browser smoke coverage` added pinned Playwright browser smoke coverage for the seeded local web workspace, covering desktop shell load, mobile add tray, desktop search, and location/detail/back navigation.
- Sharing and user-management backend hardening now includes an explicit inventory access repository port, paginated invitation listing with status filters, pending-invitation expiration management, generated OpenAPI/client updates, documentation updates, and adversarial API tests for token redaction, tenant/inventory boundaries, and role denial.
- The first audit-backed undo/redo slice now supports asset create, update, move, archive, and restore through operation-scoped compensating commands, dedicated undoable-operation persistence, generated OpenAPI/client updates, and adversarial API coverage.
- Asset state-changing application commands now use a dedicated transactional asset unit-of-work port instead of overloading the read repository port with audit and undoable-operation write concerns.
- Core API hardening now separates read repositories from explicit command/unit-of-work ports across the implemented write surfaces, uses durable blob-deletion intent for attachment hard delete cleanup, routes search visibility through an authorization query port, and applies HTTP security headers, request body limits, and configurable server timeouts.
- The mobile realtime voice slice can record through the mobile app, proxy through the core API, transcribe with Google Gemini, call read-only inventory tools through the application boundary, validate structured final responses, synthesize Google Text-to-Speech audio, and stream the result back to mobile.
- The Gemini language adapter now uses native provider function calling while keeping provider-specific tool declarations and function-call/function-response shapes behind the project-owned language inference port.
- Realtime voice loop control now allows multiple distinct tool calls, including across turns, while preventing exact duplicate tool-call re-execution and requesting an explicit finalization-only turn when needed.
- Provider credential sealing now has a port, AES-256-GCM adapter, encrypted GORM persistence, migrations, startup fail-closed validation, and tests.
- Provider credentials now sit behind a provider credential vault port; the first adapter composes AES-256-GCM sealing with database-backed encrypted credential rows while preserving atomic provider-profile credential replacement through the existing unit-of-work boundary.
- Tenant-scoped conversational provider profiles now have a typed agent/model domain model, application service boundary, memory and GORM persistence adapters, migrations, audit/observability taxonomy, and tests.
- Tenant-scoped provider-profile management now exposes authenticated REST endpoints for create, list, detail, enable, disable, archive, and credential replacement with redacted responses, encrypted credential storage through the sealing port, audit records, generated OpenAPI coverage, and adversarial HTTP tests.
- Realtime voice session startup now resolves session-scoped provider ports through a resolver boundary, carries selected provider profile IDs on the session, supports a tenant-profile resolver backed by provider profile and encrypted credential ports, and keeps the transitional process-configured dev/Google provider set behind the same resolver interface.
- `b60bbe8d feat(api): add provider profile test operation` and `eca22004 feat(api): run safe provider diagnostic probes` added a tenant-scoped provider profile test endpoint, safe success/failure metadata, audit/observability hooks, provider-aware credential selection, and capability-specific diagnostic probes for Google-backed language inference, text-to-speech, and speech-to-text endpoint validation.
- `a69d3f12 feat(api): support api-key Gemini profiles` added Google AI Gemini API-key support for speech-to-text and language-inference provider profiles using `x-goog-api-key`, while keeping Google Cloud Text-to-Speech OAuth-only.
- Tenant-scoped language-inference provider profiles now support bounded prompt templates that round-trip through the management API, persist through GORM migrations, resolve with the selected provider set, and are passed into realtime language model calls while the API appends the mandatory agent contract.
- Tenant-scoped provider-profile management now supports non-secret PATCH updates for display name, endpoint URL, model name, runtime options, capability metadata, and prompt template, with partial-update semantics, audit/observability, generated client coverage, and `lastTestedAt` reset when configuration changes.
- Mobile startup now has a connection/onboarding gate that can save non-secret instance metadata, use the local-development token only from runtime configuration, guide tenant and first-inventory creation, rebuild application services after onboarding, and reset the saved instance from Settings.
- Realtime voice sessions now persist durable safe session metadata through a repository port with memory and GORM adapters, including session scope, selected provider profile IDs, lifecycle state, timestamps, and safe failure codes without storing raw audio, transcripts, prompts, model responses, generated speech, credentials, bearer tokens, or provider session IDs.
- Mobile provider-profile management now exposes safe tenant-scoped provider profile metadata, recommended profile creation, credential replacement, prompt-template replacement, lifecycle actions, safe provider tests, readiness summaries, and a voice-sheet recovery action that opens Voice providers when readiness fails before recording.
- Mobile realtime voice cancellation now has an application boundary, recorder cleanup path, WebSocket abort path that sends `session.cancel` when session-bound, safe terminal cancelled state, API `session.cancelled` response for pre-processing cancellation, and focused mobile/API tests.
- Mobile realtime voice can now expose a bounded `propose_action_plan` native tool, persist a proposed action plan through the application boundary, stream a safe `action.plan.proposed` WebSocket event, and render the proposal in the mobile voice sheet review stage without executing inventory writes.
- Mobile realtime voice can now keep the review WebSocket session open after proposal, accept explicit mobile `action.plan.approve` or `action.plan.cancel` decisions, transition the persisted plan through application services, emit safe review outcome events, and disable duplicate mobile review decisions while awaiting the terminal review outcome.
- Approved mobile voice action plans can now execute the first single create command slice through the existing asset application boundary, atomically persist the asset/audit/undoable operation with the terminal action-plan state, and stream safe `action.plan.executed` or `action.plan.failed` review outcomes back to mobile.
- Approved mobile voice action plans can now execute a single `move_asset` command through the existing asset movement boundary, atomically persist the asset move/audit/undoable operation with the terminal action-plan state in memory and GORM adapters, and stream safe execution outcomes back to mobile.
- Approved mobile voice action plans can now execute a single `archive_asset` command through the existing asset lifecycle boundary, atomically persist the archive/audit/undoable operation with the terminal action-plan state in memory and GORM adapters, and stream safe execution outcomes back to mobile.

## Known Gaps

- Changing custom field type, removing custom field enum options or targets, durable thumbnail caching, production direct-upload provider adapters, model provider image use, and advanced search ranking/indexing are not implemented.
- Undo/redo is implemented only for the first asset slice. It is not yet available for hard delete, tenants, inventories, sharing, attachments, custom asset types, custom field definitions, search, or audit reads.
- Custom field definitions cannot yet perform destructive schema changes, be reordered, imported, exported, or managed through conversational flows.
- The first web inventory workspace direction is specified in `specs/platform/web-inventory-workspace.spec.md` and has been promoted into `apps/web` with frontend domain, port, API adapter, seeded adapter, and focused workspace components.
- The first SpiceDB search visibility adapter still evaluates candidate inventories one at a time behind the authorization visibility port; replace it with SpiceDB lookup APIs before large tenants are expected.
- Rate limiting is specified as required before public or multi-user deployment, but is not implemented.
- Invitation acceptance links exist for sharing, but they are not a primary authentication mechanism.
- The web UI still needs deeper media attachment management, production direct-upload UX, broader browser coverage against authenticated API/Dex flows, viewer-denied browser coverage, and component-level tests for the asset detail edit and move panels.
- `specs/platform/ui-design-workshop.spec.md` and `.codex/skills/stuffstash-ui-design` now codify the UI design workshop process, including product-owner decision gates, real SvelteKit candidates, responsive review, accessibility review, and adversarial critique lenses.
- API-key-backed speech synthesis adapters and the external MCP server are not yet complete.

## Next Work

1. Deepen the production mobile voice session surface.
   - Use `specs/agent-model/mobile-realtime-voice-query.spec.md` as the source of truth.
   - Show safe progress steps, full ephemeral transcript, final spoken response, cancellation, errors, and developer diagnostics without turning voice into a separate primary page.
2. Expand approval-backed write action plans beyond the first create slice.
   - Use `specs/agent-model/realtime-interaction.spec.md` and `specs/agent-model/mcp-agent-tools.spec.md` as the source of truth.
   - Continue adding additional command kinds only behind application services, tenant/inventory authorization, audit history, atomic execution, and explicit user confirmation.
3. Implement the external Stuff Stash MCP server.
   - Use `specs/agent-model/mcp-agent-tools.spec.md` as the source of truth.
   - Reuse the same application services, OIDC/auth middleware, authorization boundaries, and tool catalog used by the internal agent loop.
4. Resume promoted web workspace work after the mobile voice path is testable from the app.
   - Use `specs/platform/web-inventory-workspace.spec.md`, `specs/media/media-attachments.spec.md`, and `specs/identity-access/tenant-inventory-access.spec.md` as the source of truth.
   - Prioritize media attachment management, browser-level coverage, tenant-first switching, search, inventory settings, and sharing/access management.

## Later Work

- Google OIDC adapter end-to-end verification.
- Mobile inventory/auth tracer bullet after the Expo Go development loop is proven.
- Conversational inventory provider profiles (`specs/agent-model/provider-profiles.spec.md`), MCP read tools (`specs/agent-model/mcp-agent-tools.spec.md`), mobile realtime voice query (`specs/agent-model/mobile-realtime-voice-query.spec.md`), API-mediated realtime sessions, credential sealing, ports, and broader action plan execution. Public MCP write tools must wait for the external approval/action-plan contract.
- Import and export.
