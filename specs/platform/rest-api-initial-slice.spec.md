# REST API Initial Slice Spec

## Purpose

The first API slice should prove the security boundary, Huma/OpenAPI generation, response contracts, and tenant/inventory behavior before assets are added.

## Scope

This spec covers the first protected REST endpoints.

It does not cover asset, location, conversational, import/export, media, or search endpoints.

## Endpoints

The first REST slice includes:

- `GET /healthz`
- `GET /me`
- `POST /tenants`
- `POST /tenants/{tenantId}/inventories`
- `GET /tenants/{tenantId}/inventories`

## Authentication

- `GET /healthz` is public.
- Every other endpoint requires bearer authentication.
- The tracer bullet uses local development tokens defined in the authentication flow spec.

## Authorization

- `POST /tenants` requires authentication.
- `POST /tenants` grants the caller ownership of the new tenant.
- `POST /tenants/{tenantId}/inventories` requires `tenant.create_inventory`.
- `GET /tenants/{tenantId}/inventories` returns only inventories visible to the caller.
- Cross-tenant and hidden-resource access must return safe authorization failures.

## Responses

- Successful responses must use the API envelope from the API contract spec.
- Error responses must use the API error envelope from the API contract spec.
- Create endpoints must return `201 Created` when a resource is created.
- IDs must be ULIDs.
- JSON fields must use `camelCase`.

## OpenAPI

- Huma must generate OpenAPI for these endpoints.
- Generated docs must include bearer authentication for protected endpoints.
- Generated docs must include request and response models.

## Verification

- Tests must verify happy paths.
- Tests must verify missing token rejection.
- Tests must verify malformed token rejection.
- Tests must verify valid token with missing relationship rejection.
- Tests must verify cross-tenant attempts fail.
- Tests must verify the OpenAPI route is available.
