# Web Frontend Tracer Bullet Spec

## Purpose

Stuff Stash needs a real web frontend early so the API contract, auth flow, and product ergonomics are tested through the same kind of browser workflow a user will actually run.

## Scope

This spec covers the first separate SvelteKit web application under `apps/web`.

It defines:

- Local Dex OIDC sign-in for the web app.
- Runtime configuration.
- API access boundaries.
- The first inventory and asset user flow.
- Basic visual direction for the first product surface.
- Verification expectations for the tracer bullet.

It does not define mobile UI, conversational inventory, production Google OIDC rollout, final deployment manifests, offline behavior, image upload UX, or a complete design system.

## Decisions

- The web app must live in `apps/web`.
- The web app must be independently buildable and deployable from the Go API.
- The Go API must not embed or serve the production web bundle as the primary deployment model.
- The first web flow must support creating an inventory, creating an asset in that inventory, and browsing assets in that inventory.
- The web tracer bullet must use Dex OIDC from the start.
- The web app must use Authorization Code with PKCE for local Dex sign-in.
- The web app must not use the Dex password grant.
- The web app must not store or require a static OIDC client secret.
- API calls from the web app must send the OIDC ID token as `Authorization: Bearer <id-token>`.
- The API must allow explicitly configured browser origins for local web development.
- CORS must be deny-by-default and configured through environment-backed settings.
- The web app must use runtime configuration for API base URL, OIDC issuer, OIDC client ID, and redirect URI.
- There is no backend-for-frontend layer in this slice.
- The web app must use the public REST API contract directly through a frontend adapter boundary.
- Generated OpenAPI types or client code must be used for the API adapter boundary as soon as the web package exists.
- Generated DTOs must not become frontend domain models.
- The web app must move toward shadcn-style reusable components through the Svelte-compatible shadcn implementation before broad UI expansion.

## Local OIDC Shape

Local web development uses Dex as a real OIDC issuer.

The first local web client must be a public OIDC client:

- Client ID: `stuff-stash-web-local`.
- Redirect URI: `http://localhost:5173/callback`.
- No client secret in browser code.
- PKCE required.

For local browser development, the API must be configured to trust the same issuer and client ID as the browser flow. The local development topology may run infrastructure in Docker while running the API and web dev server as host processes when that is the simplest way to make issuer discovery work for both browser and API verifier.

The first web tracer bullet may use direct Dex authorization and token endpoints derived from the configured issuer. General OIDC discovery for arbitrary providers must be added before this helper is treated as a provider-neutral OIDC client.

## Runtime Configuration

The web app must load runtime configuration before calling the API or starting OIDC sign-in.

The first runtime config shape is:

```json
{
  "apiBaseUrl": "http://localhost:8080",
  "oidcIssuer": "http://localhost:5556/dex",
  "oidcClientId": "stuff-stash-web-local",
  "oidcRedirectUri": "http://localhost:5173/callback"
}
```

The runtime config may be served as a static file by the web app in local development. It must not be compiled into frontend source code as the only configuration mechanism.

## API Boundary

The first web app must keep these layers distinct:

- Generated OpenAPI types or generated client code.
- A small API adapter that knows HTTP, tokens, response envelopes, errors, and pagination.
- Frontend domain models for inventory and asset screens.
- Svelte components and routes.

The UI must not depend directly on raw generated DTOs.

The first adapter may be intentionally small, but it must expose domain-oriented operations such as:

- Get current identity.
- Create inventory.
- List inventories.
- Create asset.
- List assets.

## First User Flow

The first web user flow is:

1. Start the web dev server.
2. Sign in with local Dex.
3. See the authenticated identity.
4. Create an inventory.
5. Select or see that inventory.
6. Create an asset inside the inventory.
7. Browse the inventory's assets.
8. Refresh the page and keep the signed-in browser session if the token is still valid.
9. Sign out locally.

Inventory creation and asset creation are inseparable for this tracer bullet. If one is present, the other must be present enough for a user to prove the loop.

The first asset create form may support only the base asset fields needed by the API:

- Kind.
- Title.
- Optional description.

It does not need custom fields, media, search, asset movement, sharing, or conversational actions in the first web slice.

## Lifecycle Web Extension

The next web pass must expose the lifecycle API enough for a user to manage test data without leaving the browser.

The asset browser must support:

- Active asset browsing by default.
- Archived asset browsing through an explicit lifecycle view.
- Asset archive from the active view.
- Asset restore from the archived view.
- Asset hard delete from either view.
- Refreshing the selected lifecycle view after each lifecycle action.

The UI must keep this simple:

- Do not add a separate route yet.
- Do not introduce frontend state management beyond the current page/component boundary.
- Keep generated DTOs behind the API adapter.
- Use small text actions or existing button styles; do not create a new design system for this pass.

The API adapter must expose domain-oriented operations for:

- List assets by lifecycle state.
- Archive asset.
- Restore asset.
- Delete asset.

Lifecycle actions must surface safe API errors to the user through the existing page message/error pattern.

## Visual Direction

The first web UI must follow `specs/platform/brand-guidelines.spec.md`.

The web UI component foundation must follow `specs/platform/client-technology.spec.md`.

The initial tracer bullet may use hand-written markup and CSS while the surface is small. Before adding substantial sharing, custom fields, media, search, or conversational screens, introduce the Svelte-compatible shadcn component foundation and migrate common primitives such as buttons, inputs, tabs, dialogs, badges, forms, and toasts toward it.

The tracer bullet should:

- Feel like a clean personal tool, not warehouse software.
- Use mostly system grays and whites.
- Use the brand glyph palette sparingly through semantic tokens.
- Keep primary actions blue/system-like.
- Avoid green SaaS styling, purple-blue AI gradients, beige/rustic palettes, decorative blobs, and oversized marketing composition.
- Keep the first screen as the usable product surface, not a landing page.

The provisional logo-direction colors are:

- Charcoal frame: `#303A41`.
- Dusty blue contained shape: `#6B90AA`.
- Amber placed-item accent: `#F5AB4B`.

These raw values must live behind CSS variables or future design tokens.

## Verification

The web tracer bullet must include:

- Type checking.
- Build verification.
- Unit tests for runtime config parsing, OIDC state helpers, and API adapter behavior with fakes.
- A browser-level smoke test for the route shell and unauthenticated sign-in state.
- A documented manual or automated local Dex flow.

Before the slice is considered complete:

- `make web-build` must pass.
- `make web-test` must pass.
- `make docs-build` must pass.
- Existing Go tests and structural hooks must still pass.

## Open Questions

- Should the first full browser e2e flow run against host-run API plus Docker infrastructure, or a dedicated Compose profile that makes Dex issuer URLs valid for both browser and API?
- Which generated OpenAPI client tool should become the long-term standard for TypeScript clients?
- Should design tokens live in `packages/client-domain`, a future `packages/design-tokens`, or inside `apps/web` until mobile exists?
- Which exact pinned shadcn-svelte setup command, component registry, and dependency set should be used for the first migration pass?
