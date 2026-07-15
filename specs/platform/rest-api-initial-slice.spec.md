# REST API Initial Slice Spec

## Purpose

The first API slice should prove the security boundary, Huma/OpenAPI generation, response contracts, tenant/inventory behavior, asset operations, and asset-scoped media attachment operations.

## Scope

This spec covers the first protected REST endpoints.

It does not cover cross-inventory asset movement, location-specific APIs, conversational, or import/export endpoints.

The current REST surface extends this initial slice with resource detail, archive, restore, cancel, and hard-delete endpoints defined by `specs/platform/resource-lifecycle.spec.md`. When this initial slice and the lifecycle spec differ about the current lifecycle surface, the lifecycle spec is authoritative for lifecycle behavior.

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
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-grants`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-grants`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/preview`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}`
- `POST /tenants/{tenantId}/custom-field-definitions`
- `GET /tenants/{tenantId}/custom-field-definitions`
- `PATCH /tenants/{tenantId}/custom-field-definitions/{definitionId}`
- `POST /tenants/{tenantId}/custom-asset-types`
- `GET /tenants/{tenantId}/custom-asset-types`
- `PATCH /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}`
- `GET /tenants/{tenantId}/audit-records`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}`
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
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive` requires `inventory.edit_asset`.
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore` requires `inventory.edit_asset`.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments` requires `inventory.edit_asset`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments` requires `inventory.view`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content` requires `inventory.view`.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-grants` requires `inventory.share`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-grants` requires `inventory.share`.
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}` requires `inventory.share`.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations` requires `inventory.share` and returns time-limited one-time invite link material for delivery outside the core service.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/preview` requires an authenticated principal whose verified email matches the invitation plus the valid one-time acceptance token; it returns only safe presentation metadata and creates no grant.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept` requires a matching authenticated principal email and an unexpired invite acceptance token, then creates the accepted direct grant.
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}` requires `inventory.share`. In the lifecycle slice this endpoint is reserved for deliberate hard delete, and `PATCH /cancel` is the normal pending-invitation cancellation endpoint.
- `POST /tenants/{tenantId}/custom-field-definitions` requires `tenant.configure`.
- `GET /tenants/{tenantId}/custom-field-definitions` requires `tenant.configure`.
- `PATCH /tenants/{tenantId}/custom-field-definitions/{definitionId}` requires `tenant.configure`.
- `POST /tenants/{tenantId}/custom-asset-types` requires `tenant.configure`.
- `GET /tenants/{tenantId}/custom-asset-types` requires `tenant.configure`.
- `PATCH /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}` requires `tenant.configure`.
- `GET /tenants/{tenantId}/audit-records` requires `tenant.configure`.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions` requires `inventory.configure`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions` requires `inventory.view`.
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}` requires `inventory.configure`.
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types` requires `inventory.configure`.
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types` requires `inventory.view`.
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}` requires `inventory.configure`.
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
- Asset listing defaults to active assets and may request `lifecycleState=active`, `lifecycleState=archived`, or `lifecycleState=all`.
- Asset listing defaults to stable ID-ascending order and may request `sort=updated_desc` for most recently changed assets.
- Asset listing cursors must include lifecycle and sort scope validation.
- Asset list responses, asset detail responses, and asset mutation responses must include `createdAt` and `updatedAt` timestamps in RFC 3339 format.
- Asset update must support replacing title, description, parent asset reference, and custom field values.
- Asset update must not support kind changes, lifecycle changes, tenant changes, or inventory changes.
- Asset movement is represented by `parentAssetId` updates.
- Asset update must allow moving an asset to the inventory root by sending `parentAssetId: null`.
- Asset update must prevent self-parenting, containment cycles, item parents, archived parents, cross-tenant parents, and cross-inventory parents.
- Asset update must validate custom field values against effective custom field definitions.
- Asset archive must change lifecycle state from `active` to `archived`, reject assets with active children, emit audit history, and return the archived asset.
- Asset restore must change lifecycle state from `archived` to `active`, reject assets whose parent is archived, emit audit history, and return the restored asset.
- Custom asset type definitions and custom-field applicability to custom asset types are part of the current REST surface. Their detailed implementation contract is specified separately in `specs/assets/custom-asset-types.spec.md`.

## First Media REST Slice

- Asset attachments must support JSON base64 upload for the first slice.
- Attachment upload must store metadata separately from blob content.
- Attachment upload must enforce a JSON HTTP body limit derived from the configured application attachment size limit.
- Attachment upload must reject unsupported content types, unsafe file names, empty content, oversized content, and invalid base64.
- Attachment list returns attachments scoped to one asset.
- Attachment list must support cursor pagination with `limit` and `cursor` query parameters.
- Attachment list must include pagination metadata in the response envelope.
- Attachment content download must return stored bytes with the stored content type.
- Attachment routes must not expose storage keys, bucket names, local filesystem paths, or provider internals.
- Attachment behavior is specified in `specs/media/media-attachments.spec.md`.

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
- HTTP adapter tests must be split by API/domain surface. `server_test.go` is reserved for platform-level behavior such as health, the local API index, unknown routes, and generated OpenAPI. Domain endpoint tests must live in focused files named for the relevant surface. Cross-cutting test setup and request helpers live in `helpers_test.go`; domain wire helpers live in focused `*_helpers_test.go` files.
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
- Tests must verify attachment upload, list, download, pagination, authentication, authorization, tenant isolation, inventory isolation, invalid base64 rejection, unsupported content type rejection, unsafe file name rejection, and oversize rejection.
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
