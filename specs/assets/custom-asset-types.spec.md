# Custom Asset Types Spec

## Purpose

Custom asset types let users describe reusable categories such as medicine, fertilizer, laptops, documents, or batteries without changing the base asset model.

An asset with a custom asset type is still a normal asset. Base `kind` controls containment behavior. Custom asset type controls classification, type-specific fields, display behavior, search/filter behavior, and future workflows.

## Scope

Cross-platform settings navigation and custom asset type management interaction behavior are defined by `specs/platform/client-settings-management.spec.md`. That client spec must preserve the additive scope, immutable key, compatibility, permission, and lifecycle rules defined here.

This spec covers the first custom asset type API and the changes needed for custom field definitions and asset custom field validation.

This spec does not implement multiple custom asset types per asset, custom type inheritance, icons, display ordering, import/export, search indexing, or UI editing flows. Restore and hard-delete behavior is defined by `specs/platform/resource-lifecycle.spec.md`.

## Model

Custom asset type definitions have:

- ID.
- Tenant ID.
- Optional inventory ID.
- Scope: `tenant` or `inventory`.
- Key.
- Display name.
- Description.
- Lifecycle state: `active` or `archived`.

Keys are stable API identifiers:

- Lowercase letters, numbers, and hyphens.
- Must start with a lowercase letter.
- Maximum length 80.
- Must be unique inside a tenant for tenant-scoped custom asset types.
- Must be unique inside one inventory for inventory-scoped custom asset types.
- An inventory-scoped key must not duplicate a tenant-scoped key available to that inventory.
- A tenant-scoped key must not be created when any existing inventory-scoped custom asset type in that tenant already uses the same key.

Display names are user-facing labels:

- Required.
- Maximum length 120.
- Trimmed before persistence.

Descriptions are optional:

- Empty string when not supplied.
- Maximum length 1000.
- Trimmed before persistence.

## Lifecycle And Mutability

Custom asset type definitions start as active records.

Active custom asset types may be assigned to new assets and targeted by new custom field definitions.

Archived custom asset types:

- Stay persisted.
- Preserve existing asset references.
- Preserve existing custom field definition target rows.
- Are hidden from normal custom asset type list endpoints.
- Must not be assignable to new assets.
- Must not be targetable by new custom field definitions.
- Must not be editable through metadata update endpoints.
- Must emit audit history when archived.

Archived custom asset types can be restored or hard-deleted only through the lifecycle endpoints defined by `specs/platform/resource-lifecycle.spec.md`.

The first update slice may change only human-facing metadata:

- Display name.
- Description.

These fields are immutable after creation:

- ID.
- Tenant ID.
- Inventory ID.
- Scope.
- Key.

Changing keys, scope, tenant, inventory, display ordering, and icon/media metadata are separate slices. They need their own compatibility rules because assets, field definitions, imports, exports, audit records, and generated clients may already refer to the existing custom asset type.

## Effective Scope

Tenant-scoped custom asset types flow down into all inventories in the tenant.

Inventory-scoped custom asset types apply only inside that inventory.

Effective custom asset type listing for an inventory returns tenant-scoped custom asset types first, then inventory-scoped custom asset types.

## REST Endpoints

The first custom asset type endpoints are:

- `POST /tenants/{tenantId}/custom-asset-types`
- `GET /tenants/{tenantId}/custom-asset-types`
- `PATCH /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}`
- `PATCH /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/archive`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/archive`

All endpoints require bearer authentication.

Tenant-scoped create and list require `tenant.configure`.

Tenant-scoped update requires `tenant.configure`.

Tenant-scoped archive requires `tenant.configure`.

Inventory-scoped create requires `inventory.configure`.

Inventory-scoped list requires `inventory.view` and returns the effective custom asset types available to that inventory.

Tenant and inventory list endpoints default to `lifecycleState=active` and accept `lifecycleState=active`, `lifecycleState=archived`, or `lifecycleState=all`. Any other value must return the standard invalid-input response.

For every lifecycle view, effective inventory listing returns matching tenant-scoped custom asset types first, then matching inventory-scoped custom asset types. An archived inherited type is visible for lifecycle management context but remains tenant-owned and cannot be mutated through an inventory endpoint.

Inventory-scoped update requires `inventory.configure` and may update only custom asset types owned by that inventory. Tenant-scoped types inherited by an inventory must be updated through the tenant endpoint.

Inventory-scoped archive requires `inventory.configure` and may archive only custom asset types owned by that inventory. Tenant-scoped types inherited by an inventory must be archived through the tenant endpoint.

Collection endpoints must use cursor pagination with `limit` and `cursor`.

Collection cursors must bind lifecycle state, tenant ID, inventory ID when present, and effective-scope mode. Cursors reused across lifecycle views, tenants, inventories, or scope modes must fail safely. Active, archived, and all views must preserve the same authentication, authorization, tenant isolation, and inventory isolation; lifecycle filtering must never broaden visibility.

Successful responses must use the standard success envelope.

Error responses must use the standard safe error envelope.

Create endpoints must return `201 Created`.

Update and archive endpoints must return the updated custom asset type using the standard success envelope.

## API Shapes

Create request:

```json
{
  "key": "medicine",
  "displayName": "Medicine",
  "description": "Medication, vitamins, and related supplies"
}
```

Response item:

```json
{
  "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
  "tenantId": "01ARZ3NDEKTSV4RRFFQ69G5FAW",
  "scope": "tenant",
  "key": "medicine",
  "displayName": "Medicine",
  "description": "Medication, vitamins, and related supplies",
  "lifecycleState": "active"
}
```

Inventory-scoped response items include the inventory ID.

`lifecycleState` is server-owned. Create and update requests must not set it.

Update request:

```json
{
  "displayName": "Medicine and Vitamins",
  "description": "Medication, vitamins, and supplement supplies"
}
```

Both fields are optional in the request, but at least one field must be present. Present fields must pass the same validation as create.

## Custom Field Definition Changes

Custom field definitions must include an applicability value:

- `all_assets`
- `custom_asset_types`

When `applicability` is `all_assets`:

- The request must omit `customAssetTypeIds` or send an empty list.
- The persisted field definition must have no custom asset type target rows.
- The field applies to every asset in the definition's effective scope.

When `applicability` is `custom_asset_types`:

- The request must include one or more `customAssetTypeIds`.
- The persisted field definition must have one join row for each target custom asset type.
- The field applies only to assets whose custom asset type is one of those targets.

Create custom field request shape:

```json
{
  "key": "expiration-date",
  "displayName": "Expiration Date",
  "type": "date",
  "applicability": "custom_asset_types",
  "customAssetTypeIds": [
    "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "01ARZ3NDEKTSV4RRFFQ69G5FAW"
  ]
}
```

Existing custom field definitions default to `all_assets` when this slice is introduced.

## Custom Field Target Validation

Custom field definition create must validate target custom asset types before persistence.

Archived custom asset types must be treated as unavailable for new custom field definition targets. The API should use the same safe not-found behavior used for unknown or unauthorized targets so archived type names are not exposed through validation errors.

Tenant-scoped field definitions:

- May target tenant-scoped custom asset types in the same tenant.
- Must not target inventory-scoped custom asset types.
- Must reject target custom asset types from another tenant.

Inventory-scoped field definitions:

- May target tenant-scoped custom asset types available to the inventory.
- May target inventory-scoped custom asset types in the same inventory.
- Must reject inventory-scoped custom asset types from another inventory.
- Must reject custom asset types from another tenant.

Duplicate target IDs in one request must be rejected as invalid input.

Unknown or unauthorized target custom asset types must return safe not-found or forbidden behavior consistent with the existing authorization boundary. Responses must not reveal hidden custom asset type names.

## Asset API Changes

Asset create must support optional `customAssetTypeId` once this slice is implemented.

Asset response items must include `customAssetTypeId` when set.

Asset create:

- May omit `customAssetTypeId`.
- Must reject a custom asset type that is not available to the asset's tenant and inventory.
- Must reject archived custom asset types with the same safe not-found behavior used for unavailable custom asset types.
- Must validate `customFields` using `all_assets` field definitions plus field definitions targeted to the selected custom asset type.

Future asset update behavior:

- May update `customAssetTypeId` only if the implementation also validates all resulting custom field values.
- Before implementation, tests must define whether changing `customAssetTypeId` drops no-longer-applicable values, rejects the change, or requires an explicit replacement `customFields` object.

Until that update behavior is specified, the first implementation should allow assigning `customAssetTypeId` at create time and keep changing it out of scope.

## Persistence

The durable schema must include:

- `custom_asset_types`
- `custom_asset_types.lifecycle_state`
- `custom_field_definition_asset_types`
- `assets.custom_asset_type_id`
- `custom_field_definitions.applicability`

`custom_field_definitions` must not store a single custom asset type foreign key. The join table is required so one field can target more than one custom asset type.

The join table must preserve tenant and inventory scope defensively enough that repository adapters can validate target rows without relying only on application checks.

PostgreSQL migrations must include database-level enforcement for custom asset type effective-key uniqueness, using the same advisory-lock pattern already used for custom field definitions.

## Authorization

Custom asset type definition create/list uses the same permissions as custom field definition create/list:

- Tenant-scoped create/list: `tenant.configure`.
- Inventory-scoped create: `inventory.configure`.
- Inventory-scoped effective list: `inventory.view`.

Assigning a custom asset type to an asset uses the same permission as creating or updating that asset:

- Create-time assignment requires `inventory.create_asset`.
- Update-time assignment requires `inventory.edit_asset` once update behavior is specified.

## Audit

Creating a custom asset type must emit an audit record.

Updating a custom asset type must emit an audit record with safe metadata for changed fields only.

Archiving a custom asset type must emit an audit record with safe metadata containing the type key and scope.

Creating a custom field definition targeted to custom asset types must emit an audit record that includes safe metadata about applicability and target count.

Assigning a custom asset type to an asset must be included in the asset create audit metadata.

## OpenAPI

The generated OpenAPI contract must include:

- Custom asset type create/list endpoints.
- Custom asset type update endpoints.
- Custom asset type archive endpoints.
- Custom asset type DTOs.
- Custom asset type lifecycle state response field.
- Custom field applicability and `customAssetTypeIds` request fields.
- Asset `customAssetTypeId` request/response fields.

## Testing

Tests must be written before implementation.

Domain and application tests must cover:

- Custom asset type key validation.
- Tenant-scoped custom asset type creation.
- Inventory-scoped custom asset type creation.
- Tenant-scoped custom asset type metadata update.
- Inventory-scoped custom asset type metadata update.
- Tenant-scoped custom asset type archive.
- Inventory-scoped custom asset type archive.
- Archived custom asset types disappearing from normal list and lookup behavior for new assignments.
- Existing asset references to archived custom asset types staying intact.
- Existing custom field definition targets for archived custom asset types staying intact.
- Rejection when an inventory update attempts to update an inherited tenant-scoped custom asset type.
- Effective inventory listing.
- Active-by-default and explicit active, archived, and all tenant and effective-inventory list behavior.
- Effective-inventory archived ordering and inherited read-only ownership semantics.
- Duplicate effective keys.
- Field applicability `all_assets`.
- Field applicability `custom_asset_types`.
- Multiple target custom asset types for one field definition.
- Asset custom field validation with no custom asset type.
- Asset custom field validation with a matching custom asset type.
- Asset custom field rejection when a field targets a different custom asset type.

Adversarial API tests must cover:

- Missing and malformed authentication.
- Unrelated users.
- Viewers attempting custom asset type creation.
- Viewers attempting custom asset type updates.
- Viewers attempting custom asset type archive.
- Wrong tenant.
- Wrong inventory.
- Archiving through the wrong scope route.
- Reusing archived custom asset types for new asset assignment.
- Reusing archived custom asset types for new custom field definition targets.
- Target custom asset types from another tenant.
- Target inventory-scoped custom asset types from another inventory.
- Hidden custom asset type targets.
- Duplicate target IDs.
- Wrong-scope cursors.
- Unknown lifecycle filters and cursors reused across lifecycle views.

PostgreSQL tests must cover:

- Effective-key uniqueness for custom asset types under concurrency-safe database constraints.
- Archive persistence without deleting asset references or custom field target rows.
- Join table persistence for multiple field targets.
- Rejection of invalid custom field applicability rows.
- Asset `custom_asset_type_id` round trip.
