# REST API Initial Slice Spec

## Purpose

The first API slice should prove the security boundary, Huma/OpenAPI generation, response contracts, tenant/inventory behavior, and the first asset create/list/update flow.

## Scope

This spec covers the first protected REST endpoints.

It does not cover cross-inventory asset movement, location-specific APIs, conversational, import/export, media, or search endpoints.

## Local Convenience Endpoint

- `GET /` is public and returns a small JSON API index for people who visit the base URL during local development.
- The API index must use the project success envelope.
- The API index includes links for health, generated OpenAPI JSON, and interactive local API docs.
- The API index is a human convenience endpoint and must be omitted from generated OpenAPI and SDK contracts.

## Endpoints

The first protected REST slice includes:

- `GET /healthz`
- `GET /me`
- `POST /tenants`
- `POST /tenants/{tenantId}/inventories`
- `GET /tenants/{tenantId}/inventories`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-grants`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-grants`
- `POST /tenants/{tenantId}/custom-field-definitions`
- `GET /tenants/{tenantId}/custom-field-definitions`
- `GET /tenants/{tenantId}/audit-records`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/audit-records`

## Authentication

- `GET /healthz` is public.
- Every other endpoint requires bearer authentication.
- The tracer bullet uses local development tokens defined in the authentication flow spec.

## Authorization

- `POST /tenants` requires authentication.
- `POST /tenants` grants the caller ownership of the new tenant.
- `POST /tenants/{tenantId}/inventories` requires `tenant.create_inventory`.
- `GET /tenants/{tenantId}/inventories` returns only inventories visible to the caller.
- `GET /tenants/{tenantId}/inventories` must support cursor pagination with `limit` and `cursor` query parameters.
- `GET /tenants/{tenantId}/inventories` must include pagination metadata in the response envelope.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets` requires `inventory.create_asset`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets` requires `inventory.view`.
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}` requires `inventory.edit_asset`.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-grants` requires `inventory.share`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-grants` requires `inventory.share`.
- `POST /tenants/{tenantId}/custom-field-definitions` requires `tenant.configure`.
- `GET /tenants/{tenantId}/custom-field-definitions` requires `tenant.configure`.
- `GET /tenants/{tenantId}/audit-records` requires `tenant.configure`.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions` requires `inventory.configure`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions` requires `inventory.view`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/audit-records` requires `inventory.view`.
- Cross-tenant and hidden-resource access must return safe authorization failures.

## First Asset REST Slice

- Asset creation must support `item`, `container`, and `location` asset kinds.
- Asset creation must support root assets and child assets through `parentAssetId`.
- Parent assets must be in the same tenant and inventory as the child.
- Parent assets must have kind `container` or `location`.
- Non-empty custom field values must be validated against effective custom field definitions before asset creation.
- New assets must be created with lifecycle state `active`.
- Asset listing returns assets scoped to one inventory.
- Asset listing must support cursor pagination with `limit` and `cursor` query parameters.
- Asset listing must include pagination metadata in the response envelope.
- Asset update must support replacing title, description, parent asset reference, and custom field values.
- Asset update must not support kind changes, lifecycle changes, tenant changes, or inventory changes in the first update slice.
- Asset movement is represented by `parentAssetId` updates.
- Asset update must allow moving an asset to the inventory root by sending `parentAssetId: null`.
- Asset update must prevent self-parenting, containment cycles, item parents, archived parents, cross-tenant parents, and cross-inventory parents.
- Asset update must validate custom field values against effective custom field definitions.

## Responses

- Successful responses must use the API envelope from the API contract spec.
- Error responses must use the API error envelope from the API contract spec.
- Create endpoints must return `201 Created` when a resource is created.
- IDs must be ULIDs.
- JSON fields must use `camelCase`.

## First Audit REST Slice

- Audit listing must be read-only.
- Tenant audit listing returns all records in the tenant, including inventory-scoped records.
- Inventory audit listing returns records scoped to the inventory.
- Audit listing must support cursor pagination with `limit` and `cursor` query parameters.
- Audit listing cursors must preserve `(occurredAt, id)` ordering.
- Audit listing must include pagination metadata in the response envelope.
- Audit records must include typed action, source, target type, target ID, principal ID, occurred timestamp, and safe metadata.
- State-changing REST endpoints must accept `X-Request-ID` and pass it through to emitted audit records when supplied.
- Inventory viewers may read inventory-scoped audit records.
- Inventory viewers must not read tenant-wide audit records unless they also have tenant configuration permission.

## OpenAPI

- Huma must generate OpenAPI for protected REST endpoints.
- Generated docs must include bearer authentication for protected endpoints.
- Generated docs must include request and response models.

## HTTP Adapter Organization

- REST adapter implementation must be organized domain-first under `apps/api/internal/adapters/httpserver/`.
- Each HTTP domain owns separate `routes/`, `dto/`, and `mapper/` directories when it has non-trivial behavior.
- Route files must register operations and call application services only.
- DTO files must define transport request and response types only.
- Mapper files must convert application or domain values to transport DTOs only.
- Cross-domain HTTP primitives must live under `httpserver/shared/`.
- `server.go` and `api.go` must compose the HTTP server and route registrations without accumulating domain route logic.
- The Go structural pre-commit hook must mechanically guard this organization where practical, including route registration in composition files, DTO or interface definitions in route files, application/domain/port imports in DTO files, and route registration in mapper files.

## Verification

- Tests must verify happy paths.
- Tests must verify missing token rejection.
- Tests must verify malformed token rejection.
- Tests must verify valid token with missing relationship rejection.
- Tests must verify cross-tenant attempts fail.
- Tests must verify tenant owners can list all inventories in their tenant.
- Tests must verify inventory owners who are not tenant owners see only the inventories they can view.
- Tests must verify asset creation and listing.
- Tests must verify asset update and movement.
- Tests must verify asset endpoint authentication and authorization failures.
- Tests must verify cross-tenant and cross-inventory asset containment attempts fail.
- Tests must verify item assets cannot be parents.
- Tests must verify movement to root, moving containers or locations with descendants, self-parent rejection, and cycle rejection.
- Tests must verify non-empty custom field values are accepted only when definitions exist and values validate.
- Tests must verify custom field definitions are authenticated, authorized, cursor-paginated, tenant/inventory isolated, and used for asset value validation.
- Tests must verify audit records are written for state-changing operations.
- Tests must verify tenant and inventory audit record listing is authenticated, authorized, cursor-paginated, and tenant/inventory isolated.
- Tests must verify audit records capture `X-Request-ID` when supplied.
- Tests must verify unauthorized errors use the safe error envelope.
- Tests must verify the API index route is available.
- Tests must verify the OpenAPI route is available.
