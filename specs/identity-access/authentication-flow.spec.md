# Authentication Flow Spec

## Purpose

Stuff Stash must authenticate users from the beginning while keeping provider details outside the domain core.

## Scope

This spec covers the first authentication boundary and local development authentication.

It does not define the final Google OIDC browser or mobile flow.

## Decisions

- Authentication must be behind ports and adapters.
- Google OIDC is still the first real external provider.
- The tracer bullet must use a deterministic local development authentication adapter.
- Local development authentication exists only to prove the boundary and support adversarial tests.
- Domain and application services must receive an authenticated principal, not provider-specific claims.
- Authentication failures must return bland error responses.

## Local Development Tokens

- The local development adapter accepts bearer tokens with the shape `dev:<user-id>`.
- The token identifies one user principal.
- The adapter must reject missing, malformed, and empty user IDs.
- Local development tokens must not be accepted by a production authentication adapter.

## Principal

The authenticated principal must include:

- Stable user ID.

Future provider metadata may be carried at the adapter edge, but domain behavior must not depend on it.

## Verification

- Protected endpoints must reject missing tokens.
- Protected endpoints must reject malformed tokens.
- Protected endpoints must accept valid local development tokens in local/test mode.
- Authentication tests must use fakes or local adapters, not mocks.
