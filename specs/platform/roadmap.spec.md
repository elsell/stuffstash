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

The current focus is the UI spec and design workshop before further frontend expansion.

The goal is to prove a production-shaped path through:

- user-centered inventory creation, asset creation, browsing, and sharing workflows,
- mobile-first interaction patterns that can later support conversational inventory,
- a web visual system based on SvelteKit and Svelte-compatible shadcn primitives,
- clear separation between generated API DTOs and frontend domain models,
- performance-conscious frontend choices,
- generated OpenAPI/client integration without hand-written API clients,
- a frontend direction that does not overfit the disposable tracer-bullet screens.

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
- Asset media attachments now support JSON base64 upload, cursor-paginated listing, raw content download, local filesystem blob storage, Garage S3-compatible blob storage, audit history, generated OpenAPI, and adversarial API tests.
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
- Sharing and user-management backend hardening now includes an explicit inventory access repository port, paginated invitation listing with status filters, pending-invitation expiration management, generated OpenAPI/client updates, documentation updates, and adversarial API tests for token redaction, tenant/inventory boundaries, and role denial.
- The first audit-backed undo/redo slice now supports asset create, update, move, archive, and restore through operation-scoped compensating commands, dedicated undoable-operation persistence, generated OpenAPI/client updates, and adversarial API coverage.

## Known Gaps

- Changing custom field type, removing custom field enum options or targets, media direct upload, thumbnails, and advanced search ranking/indexing are not implemented.
- Undo/redo is implemented only for the first asset slice. It is not yet available for hard delete, tenants, inventories, sharing, attachments, custom asset types, custom field definitions, search, or audit reads.
- Custom field definitions cannot yet perform destructive schema changes, be reordered, imported, exported, or managed through conversational flows.
- The real product UI is intentionally underspecified and should be redesigned before further frontend feature investment.
- Search authorization filtering currently enumerates tenant inventories and checks each one; a future authorization lookup port should replace that before large tenants are expected.
- Invitation acceptance links exist for sharing, but they are not a primary authentication mechanism.
- The web UI can create inventories, create assets, browse active and archived assets, archive assets, restore assets, and hard-delete assets, but it does not yet expose sharing, search, audit history, custom fields, custom asset types, or media.
- The web UI still uses hand-written tracer-bullet components; introduce the Svelte-compatible shadcn component foundation before broad UI expansion.

## Next Work

1. Run a UI spec and design workshop before expanding the web frontend.
   - Treat the current screens as disposable tracer-bullet UI.
   - Keep the shadcn foundation, but do not assume the current layout, copy, or workflow survives.
2. Decide whether the undoable-operation creation boundary should remain on asset repository methods or move to a dedicated transactional command/unit-of-work port before expanding undo/redo beyond assets.

## Later Work

- Google OIDC adapter end-to-end verification.
- Mobile app scaffold with React Native and Expo.
- Conversational inventory ports and action plan execution.
- Import and export.
