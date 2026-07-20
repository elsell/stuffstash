# Flexible Asset Fields Spec

## Purpose

Stuff Stash must support useful inventory details without forcing every household, hobby, tool, pantry item, document, or garage object into one rigid schema.

The asset model needs a stable core plus custom asset types and a custom-field storage path from day one.

## Scope

This spec covers flexible asset metadata requirements, custom asset types, and how custom fields attach to assets.

This spec does not define the full asset aggregate, search model, UI editor, or import/export behavior. Resource lifecycle behavior for assets, custom fields, and custom asset types is defined by `specs/platform/resource-lifecycle.spec.md`.

The detailed implementation contract for custom asset type APIs, field applicability, and asset assignment is defined in `specs/assets/custom-asset-types.spec.md`.

Cross-platform settings navigation and custom field management interaction behavior are defined by `specs/platform/client-settings-management.spec.md`. That client spec must preserve the scope, inheritance, compatibility, and lifecycle rules defined here.

The first asset slice stores validated custom field values once definitions exist for the target inventory.

Custom asset types are user-defined classifications layered onto normal base assets. They do not replace the base asset model.

## Requirements

- Assets must have a stable domain core for fields that are truly common to inventory items.
- Assets must support optional custom asset types.
- Assets must support custom fields.
- The initial asset schema must include a place to store custom field values.
- Non-empty custom field values must be rejected unless matching custom field definitions exist and validation passes.
- Custom field definitions may be tenant-scoped or inventory-scoped.
- Tenant-scoped custom field definitions must flow down into inventories.
- Inventory-scoped custom field definitions must apply only inside that inventory.
- Custom field definitions must declare whether they apply to all assets or only to assets with one or more specific custom asset types.
- A custom field definition scoped to custom asset types must not apply to assets without one of those custom asset types.
- A custom field definition scoped to all assets must apply regardless of custom asset type.
- Custom fields must be typed.
- Custom field types must be represented with enumerations or typed value objects, not magic strings.
- Custom field definitions must be separate from custom field values.
- Custom field values must be validated against their definitions.
- Custom fields must support future use in search, filtering, sorting, display, import, export, and conversational updates.
- The system must avoid overfitting persistence schemas to the first set of asset examples.
- In the PostgreSQL adapter, asset custom field values should be stored in a JSONB column.
- Custom field definitions must live in separate persistence structures from asset custom field values.
- Custom asset type definitions must live in separate persistence structures from assets and custom field definitions.
- The domain must not expose raw database JSON structures as the asset model.
- Persistence choices must remain behind repositories and adapters.

## Custom Asset Types

Custom asset types let a tenant or inventory define reusable asset categories without creating new base asset classes.

Examples:

- `medicine`
- `fertilizer`
- `laptop`
- `document`
- `battery`

A medicine asset is still a normal asset. It still has a base asset kind such as `item`, `container`, or `location`. The custom asset type only controls user-facing classification, custom fields, display behavior, search/filter behavior, and future workflows.

Custom asset type definitions must have:

- ID.
- Tenant ID.
- Optional inventory ID.
- Scope: `tenant` or `inventory`.
- Key.
- Display name.
- Optional description.

Custom asset type keys must follow the same stable key format as custom field keys.

Custom asset type scoping follows custom field scoping:

- Tenant-scoped custom asset types flow down into inventories.
- Inventory-scoped custom asset types apply only inside that inventory.
- Inventory-scoped custom asset type keys must not conflict with tenant-scoped custom asset type keys available to that inventory.

Assets may have zero or one custom asset type in the initial model.

Future specs may allow multiple custom asset types, type inheritance, templates, icons, display ordering, or richer type-specific UI. Those features are explicitly out of scope until specified.

## Field Applicability

Custom field definitions must include an applicability target:

- `all_assets`: applies to every asset in the definition's effective scope.
- `custom_asset_types`: applies only to assets whose custom asset type is one of the field definition's targeted custom asset types.

For example, a tenant or inventory can define `medicine` and `fertilizer` custom asset types and then define one `expiration-date` custom field that applies to both.

Field applicability must be validated before a definition is persisted:

- A custom-type-scoped field must target one or more existing custom asset types available in the same effective tenant/inventory scope.
- A tenant-scoped field may target tenant-scoped custom asset types.
- An inventory-scoped field may target tenant-scoped custom asset types available to that inventory or inventory-scoped custom asset types in that inventory.
- A tenant-scoped field must not target inventory-scoped custom asset types.
- An `all_assets` field must not have custom asset type targets.

Asset value validation must use the asset's custom asset type when resolving effective field definitions.

When an asset has no custom asset type, only `all_assets` field definitions apply.

When an asset has a custom asset type, both `all_assets` definitions and definitions targeting that custom asset type apply.

## Initial Field Types

The initial custom field type set includes:

- `text`
- `number`
- `boolean`
- `date`
- `url`
- `enum`

Future specs may add richer field types such as money, quantity with units, expiration date, warranty details, file attachment, image reference, serial number, barcode, or model number.

## First Definition API Slice

The first custom-field definition slice supports durable tenant-scoped and inventory-scoped definitions.

It does not implement deletion, enum option removal, enum option renaming, enum option reordering, applicability narrowing, target removal, display ordering, search, filtering, import/export, conversational definition creation, or per-field permissions.

Definitions must have:

- ID.
- Tenant ID.
- Optional inventory ID.
- Scope: `tenant` or `inventory`.
- Key.
- Display name.
- Type.
- Enum options when type is `enum`.
- Applicability target.
- Custom asset type targets when applicability target is `custom_asset_types`.

Field keys must be stable API keys:

- Lowercase letters, numbers, and hyphens.
- Must start with a lowercase letter.
- Maximum length 80.
- Must be unique inside a tenant for tenant-scoped definitions.
- Must be unique inside one inventory for inventory-scoped definitions.
- An inventory-scoped key must not duplicate a tenant-scoped key available to that inventory.
- A tenant-scoped key must not be created when any existing inventory-scoped definition in that tenant already uses the same key.

Display names are user-facing labels:

- Required.
- Maximum length 120.
- Trimmed before persistence.

## Definition Lifecycle And Mutability

Custom field definitions start as active records.

The first update slice may change the human-facing display name.

The first schema evolution slice may also make compatibility-preserving changes:

- Enum fields may add new enum options.
- Enum options must never be removed, renamed, or reordered in this slice.
- A field targeted to custom asset types may add new active custom asset type targets in the same effective scope.
- A field targeted to custom asset types may expand to `all_assets`.
- A field that applies to `all_assets` must not be narrowed to custom asset types.
- Field type remains immutable.
- Field key, scope, tenant, and inventory remain immutable.

These rules preserve existing asset custom field values. They allow future data to become valid without making previously valid data invalid.

These fields are immutable after creation:

- ID.
- Tenant ID.
- Inventory ID.
- Scope.
- Key.
- Field type.

Changing field type, removing enum options, reordering enum options, narrowing applicability, or removing target custom asset types is not allowed in the first schema evolution slice because assets may already store values validated against the current definition.

Enum options:

- Are required for `enum` fields.
- Must use the same key format as field keys.
- Must be unique inside the definition.
- Are not allowed for non-enum field types in the first slice.

The first custom field endpoints are:

- `POST /tenants/{tenantId}/custom-field-definitions`
- `GET /tenants/{tenantId}/custom-field-definitions`
- `PATCH /tenants/{tenantId}/custom-field-definitions/{definitionId}`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}`

The first custom asset type endpoints are specified in `specs/assets/custom-asset-types.spec.md`.

Create endpoints:

- Require authentication.
- Tenant-scoped create requires `tenant.configure`.
- Inventory-scoped create requires `inventory.configure`.
- Must reject duplicate keys in the effective scope.
- Must use the standard response envelope.

List endpoints:

- Require authentication.
- Tenant-scoped list requires `tenant.configure`.
- Inventory-scoped list requires `inventory.view`.
- Default to `lifecycleState=active`.
- Accept `lifecycleState=active`, `lifecycleState=archived`, or `lifecycleState=all`; any other value must return the standard invalid-input response.
- Inventory-scoped list returns the effective definitions for the requested lifecycle view: matching tenant-scoped definitions first, then matching inventory-scoped definitions. Archived tenant definitions in an inventory result describe definitions inherited from that tenant that would become effective again if restored; inventory callers must not mutate them through inventory endpoints.
- Must use cursor pagination with `limit` and `cursor`.
- Must bind the requested lifecycle state, tenant ID, inventory ID when present, and effective-scope mode into cursor validation. A cursor from another lifecycle view, tenant, inventory, or scope must fail safely rather than skip, duplicate, or reveal records.
- Must preserve the same authentication, authorization, tenant-isolation, and inventory-isolation behavior for active, archived, and all lifecycle views. Requesting archived records must not broaden visibility.
- Must use the standard response envelope.

Update endpoints:

- Require authentication.
- Tenant-scoped update requires `tenant.configure`.
- Inventory-scoped update requires `inventory.configure`.
- May update `displayName`.
- May add enum options to enum fields.
- May add custom asset type targets to `custom_asset_types` definitions.
- May expand a `custom_asset_types` definition to `all_assets`.
- Must reject empty update bodies.
- Must return the updated definition in the standard response envelope.
- Tenant-scoped update must not update inventory-scoped definitions.
- Inventory-scoped update must not update inherited tenant-scoped definitions.
- Must reject field type changes.
- Must reject enum option removal, rename, or reorder.
- Must reject enum options on non-enum fields.
- Must reject narrowing `all_assets` to `custom_asset_types`.
- Must reject removing existing custom asset type targets.
- Must reject new custom asset type targets that are unknown, archived, hidden, wrong-scope, wrong-inventory, or cross-tenant.
- Must preserve existing asset custom field values.

## Asset Value Validation

Asset create may accept non-empty custom field values only when every provided key resolves to an effective definition for the target inventory and the asset's custom asset type.

Asset update may change custom field values only when every provided key remains valid for the asset's current custom asset type.

Changing an asset's custom asset type is out of scope for the first implementation. Before it is implemented, a future spec must define what happens to existing custom field values that no longer apply.

Value validation rules:

- `text`: string.
- `number`: JSON number.
- `boolean`: JSON boolean.
- `date`: string in `YYYY-MM-DD` format.
- `url`: string with `http` or `https` scheme.
- `enum`: string matching one configured enum option key.

Unknown field keys, fields that do not apply to the asset's custom asset type, wrong value types, invalid enum values, and malformed date or URL values must return a safe invalid-request error.

## Conversational Use

- Conversational inventory flows must be able to read and update custom fields through application services.
- A model may propose custom asset types or custom field updates, but domain services must validate type definitions, field definitions, field applicability, field types, tenant scope, and authorization before changes are saved.
- The system must ask for clarification when a spoken or typed command could refer to multiple custom fields.
- The system must ask for confirmation before creating a new custom asset type or custom field definition from conversational input.

## Security And Tenancy

- Custom field definitions and values must be tenant-isolated.
- Custom asset type definitions must be tenant-isolated.
- Inventory-scoped custom field definitions and values must be inventory-isolated.
- Inventory-scoped custom asset type definitions must be inventory-isolated.
- Users must not infer another tenant's custom field names, definitions, or values.
- Users must not infer another tenant's custom asset type names or definitions.
- Users must not infer custom field names, definitions, or values from inventories they cannot access.
- Users must not infer custom asset type names or definitions from inventories they cannot access.
- Authorization checks must apply to creating, reading, updating, deleting, and using custom fields.
- Authorization checks must apply to creating, reading, updating, deleting, and using custom asset types.
- Custom field definitions do not need separate per-field permissions at first.
- Error responses must not leak hidden field definitions or values.
- Error responses must not leak hidden custom asset type definitions.

## Testing

- Tests must verify custom field validation using fakes, not mocks.
- Tests must cover tenant isolation, authorization, custom asset type applicability, type validation, unknown fields, fields that do not apply to the asset's custom asset type, conflicting field names, and conversational update proposals.
- Tests must cover active-by-default and explicit active, archived, and all custom field definition listing for tenant and effective-inventory scopes.
- Tests must cover effective-inventory archived ordering with inherited tenant definitions before inventory-owned definitions, without granting inventory mutation authority over inherited records.
- Tests must cover unknown lifecycle filters and cursors reused across lifecycle, tenant, inventory, or effective-scope boundaries.
- Security-sensitive custom field behavior must have adversarial end-to-end tests before public interaction points expose it.
- The first API slice must include adversarial tests for unauthenticated requests, unrelated users, viewers attempting definition creation, duplicate keys, wrong tenant, wrong inventory, wrong-scope cursors, and asset values referencing hidden or unknown definitions.
