# Stuff Stash Repository Guidance

This repository is for Stuff Stash, a home inventory management system. Stuff Stash is intended to provide a core domain service with pluggable ports and adapters for experiences such as a REST API, an MCP server for agents, and a mobile application.

These instructions are binding for all agents and contributors working in this repository.

## Source Of Truth

- This is a spec-driven project.
- The top-level `specs/` directory contains product and engineering specifications.
- The `specs/` directory must be organized by domain or bounded context.
- Product domain specs must live under their domain directory, such as `specs/assets/`, `specs/locations/`, `specs/identity-access/`, or other domain folders established by spec.
- Cross-cutting engineering specs must live under a clearly named non-product area, such as `specs/platform/`.
- Specification files must be Markdown files ending in `.spec.md`.
- The spec must be updated before any coding begins.
- Code must follow the spec. If code and spec disagree, update the spec first, then update the code.
- Do not introduce behavior, architecture, dependencies, or domain concepts that are not represented in the relevant spec.

## Documentation

- Human-focused documentation is a primary project concern.
- Documentation must live in the top-level `docs/` directory.
- The documentation site must use Astro and Starlight.
- Documentation should help a technically adept newcomer understand what Stuff Stash is, how it is structured, and how to run it locally.
- Prefer concise, high-signal writing. If there were more time, write less.
- Documentation should sound human, plain, and direct. Aim for roughly an eighth-grade reading level unless the topic requires precise technical language.
- Do not write documentation for documentation's sake.
- Document concepts, workflows, setup steps, operating procedures, and architectural decisions that are useful to humans.
- Do not duplicate information that is self-evident from code, tests, or specs unless the human explanation adds meaningful context.
- Setup and local development instructions must be verified by running them when possible. If they cannot be verified, mark that explicitly.
- Documentation must stay consistent with specs. If documentation and specs disagree, update the spec first, then update the documentation.
- Documentation must stay consistent with code. The documentation agent is responsible for finding and correcting drift between documentation, specs, code, and repository history.

## Custom Agents

- Project-scoped Codex custom agents live under `.codex/agents/`.
- Custom agents should be narrow, opinionated, and useful enough to justify their own identity.
- The documentation agent owns human-facing documentation quality, structure, and synchronization with the codebase.
- Add or update custom agents when a durable role would improve review quality, implementation discipline, or project process.

## Architecture

- Always use hexagonal architecture, also known as ports and adapters.
- Never bypass ports or adapters for convenience.
- Keep domain logic independent of transport, persistence, authentication providers, authorization providers, observability implementations, and framework details.
- Design for flexibility, pluggability, and replacement of infrastructure concerns.
- Prefer dependency injection over global state or hard-wired implementations.
- The core service must remain usable behind multiple adapters, including REST, MCP, mobile-facing APIs, workers, CLIs, or future interfaces.

## Security

- Security is a primary concern for this project.
- Supply-chain security is part of the security boundary.
- Every dependency, base image, toolchain, action, plugin, and generated artifact source must be pinned to a known, reviewed version wherever the ecosystem supports it.
- Container base images must be pinned by immutable digest, not only by tag.
- Floating versions such as `latest`, broad version ranges, unreviewed transitive dependency changes, and unpinned remote install scripts are not allowed unless a spec explicitly justifies the exception.
- Dependency and base image updates must be intentional, reviewed, and tested.
- Authentication and authorization boundaries are non-negotiable and must be verified with end-to-end tests.
- Every endpoint, adapter, command, worker, MCP tool, mobile-facing interaction, or other interaction point that supports authentication or authorization must have security-focused end-to-end test coverage.
- Security tests must be adversarial. They must verify that unauthorized, unauthenticated, cross-tenant, wrong-role, expired-session, malformed-token, and privilege-escalation attempts fail correctly where applicable.
- Security tests must also verify that legitimate principals can perform the actions they are authorized to perform.
- Authorization and authentication behavior must be tested at the real boundary where callers interact with the system, not only through isolated domain or application-service tests.
- Multi-tenant isolation is part of the security boundary and must be tested accordingly.
- Do not add or modify a security-sensitive endpoint or interaction point without updating or adding the corresponding adversarial end-to-end tests first.

## Domain-Driven Design

- Follow a strict domain-driven design approach.
- Model the domain explicitly with entities, value objects, aggregates, repositories, services, policies, and domain events where appropriate.
- Use enumerations and typed domain concepts instead of hard-coded strings, magic numbers, or loosely defined values.
- Keep domain language consistent across specs, code, tests, and observability.
- Initial domain candidates for discussion:
  - Asset domain
  - Location domain
  - Agent/model domain
  - Expiration domain
  - Identity and access management domain
- These domain boundaries are not final until specified. Discuss and capture changes in specs before implementation.

## Observability

- Use domain-oriented observability at all times.
- Do not use individual `print`, `println`, or ad hoc logging statements.
- Observability must be pluggable, injectable, and expressed through ports.
- Observability must support fan-out so multiple implementations can be active at once, such as console logging, OpenTelemetry, metrics, tracing, audit events, or other future sinks.
- Observability events should describe meaningful domain and application behavior, not incidental implementation details.

## Testing

- Always use test-driven development.
- Write the failing test first, then implement the smallest correct behavior, then refactor.
- Tests must verify real functionality and meaningful behavior.
- Do not write arbitrary tests that only exercise implementation details.
- Do not use mocks.
- Use fakes instead of mocks.
- Fakes must behave like real, in-memory or controlled implementations of the relevant port.
- Tests should preserve hexagonal boundaries and validate domain behavior through appropriate ports and application services.
- Security-sensitive behavior requires adversarial end-to-end tests in addition to lower-level tests.

## Pre-Commit Hooks

- This repository must maintain a pre-commit hook configuration.
- The preferred hook runner is Lefthook, configured through `lefthook.yml`, unless a future spec chooses a different Go-native or Go-friendlier standard.
- Pre-commit hooks must run Go formatting and relevant tests.
- Hooks should be fast enough for routine local use while preserving the project rules that matter most.
- Add new hooks whenever a structural mistake can be detected automatically.
- Examples of structural mistakes that should become hooks when encountered:
  - Raw SQL usage.
  - Ad hoc print statements.
  - Hard-coded environment-specific configuration.
  - Use of mocks in tests.
  - Missing spec updates for code changes where this can be detected reliably.
- Hook behavior must be documented in the relevant spec or repository guidance when introduced.
- Do not rely on hooks as the only enforcement mechanism. CI and tests should also enforce important project guarantees.

## Continuous Improvement

- When a recurring or structural mistake is identified, improve the system so the mistake is harder to repeat.
- If the mistake can be detected with a pre-commit hook, add or update the hook.
- If the mistake can be prevented by a fake, test helper, typed API, enumeration, domain abstraction, or port, prefer that stronger design fix.
- Capture new project rules in `AGENTS.md` and, when they affect product or architecture behavior, in the relevant `specs/*.spec.md` file.
- Treat contributor feedback as a signal to improve the repository workflow, not only the immediate code.

## Configuration

- Use environment variables for runtime configuration.
- Do not hard-code configuration values that may vary by environment, deployment, tenant, provider, datastore, credentials, endpoints, feature flags, or operational settings.
- Configuration must be parsed and validated at application boundaries.
- Domain code must not read environment variables directly.

## Technology Decisions

- Primary language: Go.
- Monorepo: this project will use a monorepo structure.
- Minimize external dependencies where practical.
- Persistence:
  - Use PostgreSQL as the production backend.
  - SQLite may be used locally.
  - Use GORM as the ORM.
  - Do not use direct SQL in application code.
  - Direct SQL is a code smell and requires spec-level justification if ever considered.
- Authorization:
  - Use SpiceDB.
  - Use relationship-based authorization.
  - Authorization must be behind ports and adapters.
- Authentication:
  - Use single sign-on through OIDC.
  - Authentication must be behind ports and adapters.
  - Initially support Google.
  - Support arbitrary OIDC providers by design.
- Web application:
  - Use SvelteKit.
  - Treat performance as a primary technology-selection and implementation concern.
- Mobile applications:
  - Use React Native with Expo.
  - Target iOS and Android.
- API contracts:
  - REST endpoints must follow standard REST conventions.
  - REST endpoints must use consistent response envelopes, error envelopes, and pagination behavior.
  - REST API documentation must be generated through OpenAPI tooling, not manually maintained Swagger files.
  - Client API code should be generated from the OpenAPI contract where practical.
- Conversational inventory:
  - Mobile and web clients must support low-friction natural-language inventory interactions.
  - Speech-to-text, language model, and text-to-speech integrations must be behind ports and adapters.
  - Model providers must be pluggable, including remote providers and local models where practical.
  - Model output must never bypass domain services, authorization, tenancy, validation, or audit behavior.

## Multi-Tenancy

- Multi-tenancy is a native capability, not an afterthought.
- Tenant identity and tenant boundaries must be represented explicitly in the domain and application layers where relevant.
- Persistence, authorization, authentication, observability, and APIs must preserve tenant isolation.
- Tenant-specific behavior must be specified before implementation.

## Dependency And Integration Rules

- Prefer standard library and well-established Go ecosystem packages.
- Add dependencies only when they are justified by the spec and reduce meaningful risk or complexity.
- Pin dependencies, toolchains, container images, CI actions, plugins, and external runtime components to known versions.
- Pin container images by immutable digest.
- Do not use floating tags such as `latest` in committed runtime, build, CI, or local development configuration.
- Dependency updates must be explicit changes with tests and any relevant documentation or spec updates.
- Infrastructure integrations must live behind ports and adapters.
- Do not leak GORM, SpiceDB, OIDC provider SDKs, HTTP framework types, or other infrastructure concerns into domain logic.

## Git Workflow

- Use atomic commits.
- Use Conventional Commits for every commit message.
- Each commit must represent one coherent change and include the related spec, tests, code, docs, and configuration updates needed for that change.
- Do not mix unrelated changes in the same commit.
- Prefer small, reviewable commits over broad checkpoint commits.

## Coding Workflow

Before implementing any change:

1. Identify or create the relevant `specs/*.spec.md` file.
2. Update the spec with the intended behavior, domain language, interfaces, and constraints.
3. Write real tests first, using fakes rather than mocks.
4. Implement through ports and adapters.
5. Ensure observability is domain-oriented and injected.
6. Validate configuration comes from environment-backed configuration objects.
7. Add or update adversarial end-to-end security tests for every authentication or authorization boundary touched by the change.
8. Run the relevant tests.
9. Run the relevant pre-commit hooks, or explain why they could not be run.
10. Commit using an atomic Conventional Commit message when asked to commit changes.

## Code Smells

Treat the following as code smells that require correction or explicit spec-level discussion:

- Coding before updating the relevant spec.
- Direct SQL in application code.
- Domain logic depending on framework, persistence, transport, auth, or observability details.
- Hard-coded environment-specific configuration.
- Magic strings or numbers where enumerations or typed values belong.
- Ad hoc print statements or unstructured logging.
- Mocks in tests.
- Tests that do not validate real behavior.
- Security-sensitive changes without adversarial end-to-end tests.
- Unpinned dependencies, base images, CI actions, toolchains, plugins, or external runtime components.
- Floating image tags such as `latest`.
- Tenant handling added only in infrastructure or transport layers.
- Authorization checks scattered outside explicit authorization ports or policies.
- Missing or bypassed pre-commit hook coverage for automatable structural rules.

## Open Discussion Points

The following are known open questions and should be resolved through specs and design discussion:

- Final domain boundaries and bounded contexts.
- Monorepo layout.
- Initial REST API shape.
- MCP server capabilities and agent interaction model.
- Local development topology for PostgreSQL, SQLite, SpiceDB, and OIDC.
- Observability event taxonomy and fan-out adapter design.
