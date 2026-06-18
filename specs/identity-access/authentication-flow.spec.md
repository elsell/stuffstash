# Authentication Flow Spec

## Purpose

Stuff Stash must authenticate users from the beginning while keeping provider details outside the domain core.

## Scope

This spec covers authentication boundaries, local development authentication, and the first production-shaped OIDC token verification adapter.

It does not define the final browser redirect UX, mobile sign-in UX, refresh-token storage, logout, or account-linking flows.

## Decisions

- Authentication must be behind ports and adapters.
- Google OIDC is still the first real external provider.
- The tracer bullet must use a deterministic local development authentication adapter.
- Local development authentication exists only to prove the boundary and support adversarial tests.
- Production-shaped authentication must use an OIDC adapter selected through environment configuration.
- OIDC ID token verification must use a maintained OIDC library; JWT cryptography must not be hand-rolled.
- Domain and application services must receive an authenticated principal, not provider-specific claims.
- Authentication failures must return bland error responses.
- Local development authentication must be explicitly selected and must not be the implicit production fallback.

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

## Principal

The authenticated principal must include:

- Stable user ID.

Future provider metadata may be carried at the adapter edge, but domain behavior must not depend on it.

## Verification

- Protected endpoints must reject missing tokens.
- Protected endpoints must reject malformed tokens.
- Protected endpoints must accept valid local development tokens in local/test mode.
- The OIDC adapter must have fake-backed tests for valid tokens, verifier failures, empty issuer or subject claims, provider-specific subject characters, malformed authorization headers, and unsupported schemes.
- Authentication tests must use fakes or local adapters, not mocks.
