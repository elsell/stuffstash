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

The current focus is the secure inventory tracer bullet.

The goal is to prove a production-shaped path through:

- authentication,
- relationship-based authorization,
- tenant and inventory persistence,
- safe REST responses,
- generated OpenAPI,
- adversarial security tests,
- domain-oriented observability,
- local Docker verification.

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
- Custom field definitions can now be renamed without changing their keys, types, enum options, applicability, or targets.
- Asset lifecycle now supports archive and restore operations with audit history, active-only default listing, and authorization checks.
- Asset media attachments now support JSON base64 upload, cursor-paginated listing, raw content download, local filesystem blob storage, audit history, generated OpenAPI, and adversarial API tests.

## Known Gaps

- User invitation flows, access revocation, custom asset type delete/archive, search, changing custom field type or targets, custom field deletion APIs, permanent asset deletion, media deletion, media direct upload, thumbnails, and S3-compatible Garage storage are not implemented.
- Undo is not yet implemented for audit history.
- Custom field definitions cannot yet change schema shape, be deleted, reordered, searched, imported, exported, or managed through conversational flows.

## Next Work

1. Specify and implement search with authorization-aware filtering.
   - Start with Postgres-backed exact and fuzzy search scoped by tenant and SpiceDB-visible inventories.
   - Search assets, custom field values, and attachment metadata where useful.
   - Preserve pagination and safe response contracts.

## Later Work

- Google OIDC adapter end-to-end verification.
- Generated API client workflow from OpenAPI.
- Web app scaffold with SvelteKit.
- Mobile app scaffold with React Native and Expo.
- Audit history and undo.
- Conversational inventory ports and action plan execution.
- Import and export.
