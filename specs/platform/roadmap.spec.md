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

## Known Gaps

- The SpiceDB adapter has fake-backed unit coverage and local Compose verification, but does not yet have a dedicated real-SpiceDB adversarial test suite.
- `golang-migrate/migrate` is specified in `tooling-versions.spec.md`, but the CLI is not wired into root commands yet.
- The app still relies on GORM schema migration for the local tracer bullet. Production migration execution must use reviewed migration files.
- Asset and containment behavior is specified but not implemented.

## Next Work

1. Add real SpiceDB-backed adversarial authorization verification.
   - Keep memory authorization as a fake.
   - Add tests or local verification that prove the SpiceDB adapter enforces the same relationship model as the fake.
   - Treat blocked real-SpiceDB verification as a blocker to record here, not as optional work.
   - Keep SpiceDB behind the authorization port.

2. Wire the pinned migration CLI.
   - Use the `golang-migrate/migrate` version already specified in `tooling-versions.spec.md`.
   - Add root migration commands.
   - Verify migrations against Postgres.
   - Stop relying on GORM schema migration beyond the local tracer bullet allowance.

3. Start the first asset and containment implementation slice.
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
