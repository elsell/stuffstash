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
