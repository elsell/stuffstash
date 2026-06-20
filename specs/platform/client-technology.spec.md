# Client Technology Spec

## Purpose

Stuff Stash needs client applications that are fast, maintainable, and able to share the same domain language without forcing one UI technology to fit every platform.

## Scope

This spec covers the technology direction for:

- The web application.
- The native mobile applications for iOS and Android.
- Shared client contracts and generated API clients.

This spec does not define user-facing screens, navigation, visual design, offline behavior, push notifications, or mobile release workflows.

## Decisions

- The web application must use SvelteKit.
- Native mobile applications must use React Native with Expo.
- Mobile targets must include iOS and Android.
- The web and mobile applications must be separate clients in the monorepo.
- The web application must be a separate deployable frontend from the beginning.
- The Go API must not embed or serve the production web application bundle as the primary deployment model.
- Local development may provide a convenience command that starts the API and web frontend together, but the services remain separate processes.
- A backend-for-frontend layer is not part of the initial architecture.
- The web and mobile clients must consume the same public API contract unless a future spec justifies a BFF for a concrete product or security need.
- The clients should share generated API contracts, domain vocabulary, test scenarios, and design tokens where useful.
- The clients must not share UI code at the cost of weaker platform behavior, worse performance, or unclear ownership.

## Requirements

- Client code must be organized so the web app, mobile app, and generated API clients can evolve independently.
- Shared packages must be introduced only when they remove meaningful duplication without hiding platform-specific concerns.
- API access from clients must use generated SDKs from the backend OpenAPI contract unless a spec explicitly justifies an exception.
- Generated SDKs must be treated as infrastructure adapters, not as the client domain model.
- Web and mobile clients must maintain separate client-side domain models where they need product behavior, validation, presentation state, or offline behavior.
- Generated DTOs must be mapped into client domain models at adapter boundaries.
- Client UI, state management, and product logic must not depend directly on generated DTOs.
- Client applications must treat the backend API contract as the source of truth for endpoint shapes, response bodies, error bodies, pagination, and authentication behavior.
- The web frontend must use runtime configuration for the API base URL and auth settings rather than assuming same-origin deployment.
- Client applications must not hard-code environment-specific service URLs, tenant identifiers, OAuth/OIDC settings, or other deployment configuration.
- Runtime or build-time configuration must come from environment-backed configuration appropriate for each platform.
- Performance must be a design constraint from the start.
- The web application must prefer server rendering, prerendering, minimal client JavaScript, and route-level loading boundaries where they fit the user experience.
- The mobile applications must preserve native performance expectations for common flows such as scanning, searching, editing, image capture, and inventory browsing.
- Large lists must use virtualization or another measured strategy that avoids loading or rendering unbounded client data.
- Image upload, preview, and metadata flows must be designed with memory, network, and latency limits in mind.
- Client dependencies must be pinned to known reviewed versions.
- Client dependency updates must be intentional, reviewed, and tested.

## Performance Guardrails

- Each client surface must have explicit performance budgets before broad feature implementation begins.
- Web builds must track bundle size and route-level JavaScript cost.
- Web flows must be tested with Web Vitals or an equivalent performance signal once the web app exists.
- Mobile flows must be tested on realistic devices or emulators before release.
- Third-party UI libraries must not be added unless a spec explains the need, expected cost, and replacement strategy.
- Accessibility and keyboard or assistive-technology responsiveness are part of client performance, not separate polish work.

## Monorepo Expectations

The exact monorepo layout remains open, but it must support these logical areas:

- Go backend service.
- SvelteKit web application.
- Independent frontend build and deployment path.
- React Native and Expo mobile application.
- Generated API clients or contract packages.
- Client-side adapter packages that map generated DTOs to web and mobile domain models.
- Shared design tokens or small shared client utilities, if justified.
- Astro and Starlight documentation site.

## Verification

- Client setup instructions must be documented once the client projects exist.
- The web app must have build, test, lint, and performance checks before user-facing web features are merged.
- The mobile app must have build, test, lint, and platform smoke checks before user-facing mobile features are merged.
- Generated API client code must be reproducible from pinned tools and the checked-in OpenAPI contract.
- Tests must verify DTO-to-domain mapping at client adapter boundaries.
