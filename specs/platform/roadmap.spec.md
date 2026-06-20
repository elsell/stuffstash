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

The current focus is the web lifecycle tracer bullet.

The goal is to prove a production-shaped path through:

- authentication,
- relationship-based authorization,
- tenant, inventory, and asset lifecycle persistence,
- safe REST responses,
- generated OpenAPI,
- adversarial security tests,
- domain-oriented observability,
- local Docker verification,
- a usable SvelteKit browser flow.

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
- The separate SvelteKit web app exists under `apps/web`, uses Dex OIDC with PKCE, uses runtime configuration, calls the API through the generated OpenAPI client boundary, and proves inventory creation, asset creation, and active asset browsing.

## Known Gaps

- Changing custom field type, removing custom field enum options or targets, media direct upload, thumbnails, and advanced search ranking/indexing are not implemented.
- Undo is not yet implemented for audit history.
- Custom field definitions cannot yet perform destructive schema changes, be reordered, imported, exported, or managed through conversational flows.
- Inventory access behavior still shares the broad inventory repository port; split an inventory access repository before adding invitation listing, resend, expiration management, membership management, or richer sharing UX.
- Search authorization filtering currently enumerates tenant inventories and checks each one; a future authorization lookup port should replace that before large tenants are expected.
- Invitation acceptance links exist for sharing, but they are not a primary authentication mechanism.
- The web UI can create and browse active assets, but it does not yet expose archive, restore, hard delete, sharing, search, audit history, custom fields, custom asset types, or media.

## Next Work

1. Extend the SvelteKit web tracer bullet with asset lifecycle management.
   - Browse active and archived assets.
   - Archive, restore, and hard-delete assets through the generated API client adapter.
   - Add focused frontend tests for lifecycle UI behavior and adapter calls.
2. Add sharing and user-management UI.
   - Show direct grants.
   - Create invite-link tokens.
   - Cancel invitations and revoke direct grants.

## Later Work

- Google OIDC adapter end-to-end verification.
- Mobile app scaffold with React Native and Expo.
- Audit history and undo.
- Conversational inventory ports and action plan execution.
- Import and export.
