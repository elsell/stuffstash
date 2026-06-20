# Authentication Flow Spec

## Purpose

Stuff Stash must authenticate users from the beginning while keeping provider details outside the domain core.

## Scope

This spec covers authentication boundaries, local development authentication, the first local OIDC verification fixture, and the first production-shaped OIDC token verification adapter.

It does not define the final browser redirect UX, mobile sign-in UX, refresh-token storage, logout, account-linking flows, final external identity provider rollout, or invitation delivery adapters.

## Decisions

- Authentication must be behind ports and adapters.
- Dex is the local OIDC verification fixture.
- Google OIDC remains the first planned external provider profile.
- The tracer bullet must use a deterministic local development authentication adapter.
- Local development authentication exists only to prove the boundary and support adversarial tests.
- Production-shaped authentication must use an OIDC adapter selected through environment configuration.
- OIDC ID token verification must use a maintained OIDC library; JWT cryptography must not be hand-rolled.
- Domain and application services must receive an authenticated principal, not provider-specific claims.
- Authentication failures must return bland error responses.
- Local development authentication must be explicitly selected and must not be the implicit production fallback.
- Local Dex must prove real OIDC discovery, JWKS signature verification, issuer checks, and audience checks before Google OIDC is wired.
- Dex static users and password grants are local verification fixtures only. They must not be used as the production authentication strategy.

## Authentication Modes

The API supports these authentication modes:

- `local-dev`: accepts deterministic local bearer tokens for local development and tests.
- `oidc`: verifies bearer ID tokens against a configured OIDC issuer and client ID.

Any unknown mode must fail startup.

## Local Development Tokens

- The local development adapter accepts bearer tokens with the shape `dev:<user-id>`.
- The token identifies one user principal.
- The adapter must reject missing, malformed, and empty user IDs.
- Local development tokens must not be accepted by a production authentication adapter.

## OIDC Bearer Tokens

- OIDC requests use `Authorization: Bearer <id-token>`.
- The adapter must verify issuer, audience/client ID, signature, expiry, and token shape.
- The principal ID must be derived from the OIDC issuer and `sub` claim.
- OIDC provider subject values must be normalized into safe internal principal IDs before they reach application services.
- Provider-specific claims must not be exposed to the domain core.
- The first supported provider is Google, but the adapter must support any standards-compliant OIDC issuer.
- OIDC configuration must come from environment variables.
- Missing issuer, missing client ID, provider discovery failure, verification failure, or malformed claims must fail closed.

## Local Dex Fixture

Local Compose provides Dex for realistic OIDC verification without requiring a Google client during early development.

- Dex must run behind the same `oidc` authentication mode used for any other OIDC issuer.
- Dex must use deterministic local users so API user-flow tests can grant and verify access across two principals.
- Dex credentials are local-only fixtures and must be documented as unsafe for production.
- The API must receive only bearer ID tokens from the verifier script. It must not call Dex-specific APIs from application code.
- The verifier may use the OAuth password grant against Dex because it is a test harness. User-facing clients must use a browser or native-app OIDC flow when those clients are built.

## Temporary Kubernetes Dex Fixture

The local Kubernetes deployment may temporarily run Dex beside Stuff Stash to prove production-shaped OIDC discovery, JWKS verification, issuer checks, and audience checks before Google OIDC is configured.

- The temporary Kubernetes Dex fixture must be scoped to the local development cluster.
- The fixture must use the same `oidc` authentication mode and environment configuration used for any OIDC provider.
- The fixture may use Dex's mock connector for interactive proof-only sign-in so no real password fixture is committed.
- The fixture issuer must be externally reachable at the same issuer URL the API verifies.
- The fixture client ID must be explicitly configured as a development client and must not be treated as a production Google client ID.
- The fixture must be removed or replaced with Google OIDC before a production rollout.

## Principal

The authenticated principal must include:

- Stable user ID.
- Optional verified email address when the authentication adapter can provide one.

Provider-specific claims must be normalized at the adapter edge. Domain behavior may use project-owned principal fields such as verified email, but must not depend on provider-specific claim names or token objects.

Local development authentication may accept an optional email fixture in addition to the stable user ID so invitation flows can be tested without OIDC. The email fixture exists only in explicit local/test mode.

## Verification

- Protected endpoints must reject missing tokens.
- Protected endpoints must reject malformed tokens.
- Protected endpoints must accept valid local development tokens in local/test mode.
- Protected endpoints must accept valid Dex ID tokens in OIDC mode.
- The OIDC verifier must fail closed for missing, malformed, wrong-issuer, wrong-audience, expired, or unsigned OIDC tokens.
- The local Dex verification path must reject missing, malformed, unsigned, and wrong-audience OIDC tokens at the HTTP boundary.
- The local Dex verification path must run the same happy-path API user flow and authorization adversary checks as the local-dev verifier.
- The OIDC adapter must have fake-backed tests for valid tokens, verifier failures, empty issuer or subject claims, provider-specific subject characters, malformed authorization headers, and unsupported schemes.
- Authentication tests must use fakes or local adapters, not mocks.
