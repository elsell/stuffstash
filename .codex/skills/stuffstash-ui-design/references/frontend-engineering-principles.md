# Frontend Engineering Principles

Use these principles for temporary SvelteKit candidates and promoted `apps/web` implementation. A candidate may be temporary, but it must reveal the same architectural tradeoffs the production app will face.

## Domain-Driven UI

Model UI state with Stuff Stash domain language:

- Tenant.
- Inventory.
- Asset.
- Item.
- Container.
- Location.
- Custom field.
- Custom asset type.
- Media attachment.
- Access grant.
- Invitation.
- Audit entry.
- Undoable operation.

Prefer typed frontend domain models when behavior depends on the concept. Do not let generated API DTOs become the product model for screens, forms, or interaction state.

## Hexagonal Frontend Boundaries

Preserve these boundaries:

- UI components render and collect user intent.
- Frontend domain models describe product concepts and presentation state.
- Application-facing helpers coordinate workflows.
- API adapters translate between generated clients, transport envelopes, errors, and frontend domain models.
- Auth helpers own OIDC/session behavior.
- Runtime configuration helpers own configurable URLs, issuer values, client IDs, redirect URIs, and feature flags.
- Observability helpers own product events and diagnostics.

Do not import generated DTOs directly into broad UI component trees when a domain model or adapter should sit between them.

## Typed Concepts Over Loose Strings

Use enums, literal unions, branded types, or small value objects for meaningful states:

- Asset lifecycle: active, archived.
- Asset kind: item, container, location.
- Access role: owner, editor, viewer.
- Invitation status.
- Form mode.
- Candidate route or panel state.
- Sort, filter, and lifecycle view values.

Raw strings are acceptable for user-entered values such as titles, descriptions, names, notes, and search text.

## DRY Without Abstraction Theater

Remove duplication when it protects behavior:

- Shared validation messages.
- Repeated state labels.
- Repeated adapter mapping.
- Repeated accessibility attributes for a reusable control.
- Repeated visual primitives already covered by shadcn-svelte components.

Do not introduce shared abstractions only because two screens currently look similar. Favor small product components until the repetition has a stable meaning.

## No God Files

Keep every file focused on one concern or a few small, highly related concerns. Split files by responsibility before they become hard to scan, test, or review.

Prefer these split points:

- Route files own route-level loading, page composition, and route-local state.
- Product components own one visible workflow surface or one reusable product composition.
- UI primitive files own one generic primitive or a tight primitive family.
- API adapter files own transport mapping and API error handling for a bounded workflow.
- Mapper files translate between generated DTOs and frontend domain models.
- Domain type files define product concepts and state shapes.
- Mock-data files provide fixtures only; they do not own UI behavior.
- Config files parse and expose runtime configuration.
- Observability files expose domain event helpers.

Flag and split files that mix unrelated concerns, such as a route file that also defines DTO mapping, mock data, observability events, validation rules, and generic UI primitives.

Large files are a smell even when they technically work. Split by domain workflow, component role, adapter responsibility, state machine, test helper, or fixture set before adding more behavior.

## Configuration Boundaries

Do not hard-code values that vary by environment, tenant, deployment, provider, auth setup, API host, feature flag, or operational mode.

Examples that must come from configuration or mock-data fixtures:

- API base URL.
- OIDC issuer.
- OIDC client ID.
- Redirect URI.
- Tenant IDs.
- Inventory IDs.
- External service URLs.
- Feature flags.

Mock data may use fixed IDs inside the mock fixture, but UI logic must not depend on those IDs as environment truth.

## Domain-Oriented Observability

Production paths must not use raw `console.log`, `print`, `println`, or ad hoc debugging output.

When observability is needed, express events in product terms through an explicit helper or port:

- `inventory.created`
- `asset.search_submitted`
- `asset.move_previewed`
- `asset.archive_requested`
- `sharing.invitation_copied`
- `undo.operation_requested`

Temporary candidates may expose visible debug panels only when they are part of the review artifact and clearly separate from product UI.

## Review Questions

Ask these during implementation feasibility review:

- Which domain concepts need explicit frontend types?
- Where does API transport mapping happen?
- What state belongs in a route, component, or adapter?
- What values must come from runtime configuration?
- What product events need observability?
- Which repeated patterns are real abstractions and which are coincidence?
- Does the design require backend behavior that specs do not define?
- Is any file becoming a catch-all for unrelated concerns?
