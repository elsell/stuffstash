# Monorepo Layout Spec

## Purpose

Stuff Stash needs a stable monorepo layout before implementation grows.

The layout must keep deployable apps, generated clients, shared packages, specs, and documentation easy to find without weakening hexagonal architecture.

## Scope

This spec covers the initial repository layout and workspace commands.

This spec does not define every future package, CI job, or deployment manifest.

## Decisions

- The Go API service must live under `apps/api`.
- The SvelteKit web application must live under `apps/web` once created.
- The React Native and Expo mobile application must live under `apps/mobile` once created.
- The Astro and Starlight documentation site must live under the top-level `docs/` directory.
- Generated API clients must live under `packages/api-client` once generation exists.
- Client-side domain and adapter helpers may live under `packages/client-domain` once justified by client implementation.
- Product and platform specs remain in the top-level `specs/` directory.
- Project custom agents remain in `.codex/agents/`.

## Project Custom Agents

- Project-scoped Codex custom agents must live under `.codex/agents/`.
- Each custom agent must have a narrow responsibility that improves repository quality or process discipline.
- The documentation agent owns human-focused documentation review and synchronization.
- The code critic agent owns ruthless review feedback for code smells, repeated code, weak boundaries, hard-coded values, poor tests, and architectural drift.
- After each implementation pass, the main agent must run the code critic agent before finalizing the work.
- Code critic findings must be handled explicitly: fix confirmed issues, or explain why a finding is deferred or not applicable.
- Custom agents must not replace tests, hooks, specs, or human review.

## Initial Layout

```text
apps/
  api/
  mobile/
  web/
packages/
  api-client/
  client-domain/
specs/
docs/
```

Empty app or package directories may contain a `.gitkeep` file until implementation begins.

## Go Workspace

- The root must use a Go workspace when Go modules live below the repository root.
- The first Go workspace must include `./apps/api`.
- Root commands must delegate to the API module.
- API code must keep using hexagonal boundaries inside `apps/api`.
- Application service code must be organized by bounded context or domain use-case package under `apps/api/internal/app/` once a domain has more than trivial behavior.
- Domain application packages must use names such as `apps/api/internal/app/assets`, `apps/api/internal/app/inventories`, `apps/api/internal/app/access`, or other domain language established by spec.
- The root `apps/api/internal/app` package may expose a compatibility facade while migration is in progress, but domain behavior must move into domain-specific application packages rather than accumulating in root files.
- Shared application support packages are allowed only for cross-cutting concepts such as typed application errors, pagination cursors, audit record construction, or access guards. They must not become catch-all business-logic packages.
- Domain application packages may depend on domain packages and ports. They must not depend on HTTP, GORM, SpiceDB, OIDC, blob-storage SDKs, or other adapter implementation packages.
- A domain application package should split commands, queries, validation, cursor handling, audit helpers, and access helpers before any single file becomes a broad service file.
- Go command entrypoints must be thin. `main.go` should load configuration, build the top-level observer, handle process signals, dispatch command mode, and exit.
- API runtime construction, adapter wiring, migration command execution, startup checks, and background workers must live in a testable bootstrap package instead of accumulating in `main.go`.
- Bootstrap package files must be split by startup responsibility: runtime coordination, application dependency assembly, auth construction, repository/blob construction, migrations, SpiceDB schema bootstrap, background workers, and startup observability.

## Commands

- `make test` must run all currently implemented test suites.
- `make run` must run the local API service.
- `make compose-up` must start the local development topology.
- `make compose-down` must stop the local development topology.
- `make docs-install` should install documentation dependencies once the docs app exists.
- `make docs-dev` should run the documentation site once dependencies are installed.

## Testing

- Tests must pass after moving code into the monorepo layout.
- Docker builds must use the new API path.
- Lefthook must continue to run Go formatting and tests from the new API path.

## Inventory application ownership

The `internal/app/inventories` package owns tenant/inventory creation, querying,
lifecycle transitions, scoped access guards, and durable authorization-outbox
processing. Separate typed inputs, commands, queries/cursors, tenant lifecycle,
inventory lifecycle, access helpers, and outbox execution. Inject repository,
unit-of-work, authorizer, audit, observer, ID, clock and configured paging/lease
ports; do not import the root App or adapters. Shared audit construction and
opaque cursor encoding continue through appsupport.

The root App retains type aliases and forwarding methods for existing callers,
including temporary forwarding access/outbox helpers used by contexts still being
migrated. Construct the small inventory service from the App's already-normalized
dependencies; do not create providers or default clocks on each operation. Preserve
transactional writes, read auditing, scoped cursor validation, authorization error
propagation, outbox claim/lease/dead-letter behavior and event names exactly.

Existing creation/outbox, authorization-filtered pagination, lifecycle and HTTP
adversarial tests remain the critical behavioral contract across this relocation.
The refactor changes ownership only; it does not add an endpoint or permission.
CI must compile and run those suites before this slice is accepted. Other root
import/conversation/access behavior remains explicitly pending under G8.

Inventory membership, effective access summaries, current-user tenant discovery,
and invitation orchestration also belong to this package. Keep discovery, grants,
invitation creation/link validation, invitation acceptance, invitation queries and
invitation lifecycle in separate files. OIDC/session authentication remains outside
this inventory use-case package. Preserve exact role/permission enumeration,
fail-closed revocation, token comparison, email binding, expiration and invitation
URL restrictions. Root invitation error symbols alias package-owned errors so
errors.Is identity survives migration. Existing application and HTTP adversarial
access/invitation tests cover these unchanged boundaries; do not replace them with
structural or happy-path-only tests.

Import preview, job queries, durable execution/recovery, source validation,
credential handling, source links, progress and cleanup belong to an ImportService
in dataportability. Cross-context asset/tag/custom-field/media commands enter
through an ImportTargets port expressed in domain values, composed by the root
application from its existing services. Preparation still precedes the existing
atomic import unit-of-work writes; do not substitute non-atomic create calls.
Preserve import request fingerprints, idempotency/source-link deduplication,
bounded streaming, cancellation/discard semantics, vault lifetime, audit and safe
error projection. Import errors retain identity through root aliases. Existing
adversarial import HTTP and durability/recovery application tests remain required.

Attachment validation sentinels shared by media commands and import error projection live in `internal/app/apperrors`; root compatibility names alias the same values so `errors.Is` behavior remains unchanged.

Search orchestration belongs in `internal/app/search`: tenant visibility,
authorized inventory intersection, query/filter validation, scoped cursors,
repository queries, ancestor/photo projection, safe read audit and domain events.
The root facade only composes that authorized read model with existing expiration
and media services. Preserve the empty-authorized-scope short circuit; expose
that internal scope to composition without adding it to transport responses.
Shared lifecycle filter parsing belongs in application support, with existing
asset callers retaining compatibility. Existing scoped search and adversarial
HTTP tests remain the behavior contract for this extraction.
