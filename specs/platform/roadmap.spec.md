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

## Known Gaps

- Unrecoverable authorization outbox events do not yet have a dead-letter state.
- Asset and containment behavior is specified but not implemented.

## Next Work

1. Add dead-letter handling for unrecoverable authorization outbox events.
   - Add an explicit terminal state for malformed or permanently invalid events.
   - Keep retryable SpiceDB outages separate from unrecoverable event data problems.
   - Preserve observability for operators to inspect and recover dead-lettered events.

2. Start the first asset and containment implementation slice.
   - Update the asset and containment specs first.
   - Implement the smallest useful asset model inside an inventory.
   - Preserve tenant and inventory isolation.
   - Keep containment behavior explicit and testable.

## Later Work

- Google OIDC adapter end-to-end verification.
- Generated API client workflow from OpenAPI.
- Web app scaffold with SvelteKit.
- Mobile app scaffold with React Native and Expo.
- Audit history and undo.
- Media attachments.
- Conversational inventory ports and action plan execution.
- Search with authorization-aware filtering.
- Import and export.
