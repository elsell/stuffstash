# Flexible Asset Fields Spec

## Purpose

Stuff Stash must support useful inventory details without forcing every household, hobby, tool, pantry item, document, or garage object into one rigid schema.

The asset model needs a stable core plus a custom-field storage path from day one.

## Scope

This spec covers flexible asset metadata requirements.

This spec does not define the full asset aggregate, asset lifecycle, search model, UI editor, or import/export behavior.

The first asset slice stores validated custom field values once definitions exist for the target inventory.

## Requirements

- Assets must have a stable domain core for fields that are truly common to inventory items.
- Assets must support custom fields.
- The initial asset schema must include a place to store custom field values.
- Non-empty custom field values must be rejected unless matching custom field definitions exist and validation passes.
- Custom field definitions may be tenant-scoped or inventory-scoped.
- Tenant-scoped custom field definitions must flow down into inventories.
- Inventory-scoped custom field definitions must apply only inside that inventory.
- Custom fields must be typed.
- Custom field types must be represented with enumerations or typed value objects, not magic strings.
- Custom field definitions must be separate from custom field values.
- Custom field values must be validated against their definitions.
- Custom fields must support future use in search, filtering, sorting, display, import, export, and conversational updates.
- The system must avoid overfitting persistence schemas to the first set of asset examples.
- In the PostgreSQL adapter, asset custom field values should be stored in a JSONB column.
- Custom field definitions must live in separate persistence structures from asset custom field values.
- The domain must not expose raw database JSON structures as the asset model.
- Persistence choices must remain behind repositories and adapters.

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

It does not implement definition update, deletion, enum option editing, display ordering, search, filtering, import/export, conversational definition creation, or per-field permissions.

Definitions must have:

- ID.
- Tenant ID.
- Optional inventory ID.
- Scope: `tenant` or `inventory`.
- Key.
- Display name.
- Type.
- Enum options when type is `enum`.

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

Enum options:

- Are required for `enum` fields.
- Must use the same key format as field keys.
- Must be unique inside the definition.
- Are not allowed for non-enum field types in the first slice.

The first endpoints are:

- `POST /tenants/{tenantId}/custom-field-definitions`
- `GET /tenants/{tenantId}/custom-field-definitions`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions`

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
- Inventory-scoped list returns the effective definitions available to that inventory: tenant-scoped definitions first, then inventory-scoped definitions.
- Must use cursor pagination with `limit` and `cursor`.
- Must use the standard response envelope.

## Asset Value Validation

Asset create may accept non-empty custom field values only when every provided key resolves to an effective definition for the target inventory.

Value validation rules:

- `text`: string.
- `number`: JSON number.
- `boolean`: JSON boolean.
- `date`: string in `YYYY-MM-DD` format.
- `url`: string with `http` or `https` scheme.
- `enum`: string matching one configured enum option key.

Unknown field keys, wrong value types, invalid enum values, and malformed date or URL values must return a safe invalid-request error.

## Conversational Use

- Conversational inventory flows must be able to read and update custom fields through application services.
- A model may propose custom field updates, but domain services must validate field definitions, field types, tenant scope, and authorization before changes are saved.
- The system must ask for clarification when a spoken or typed command could refer to multiple custom fields.
- The system must ask for confirmation before creating a new custom field definition from conversational input.

## Security And Tenancy

- Custom field definitions and values must be tenant-isolated.
- Inventory-scoped custom field definitions and values must be inventory-isolated.
- Users must not infer another tenant's custom field names, definitions, or values.
- Users must not infer custom field names, definitions, or values from inventories they cannot access.
- Authorization checks must apply to creating, reading, updating, deleting, and using custom fields.
- Custom field definitions do not need separate per-field permissions at first.
- Error responses must not leak hidden field definitions or values.

## Testing

- Tests must verify custom field validation using fakes, not mocks.
- Tests must cover tenant isolation, authorization, type validation, unknown fields, conflicting field names, and conversational update proposals.
- Security-sensitive custom field behavior must have adversarial end-to-end tests before public interaction points expose it.
- The first API slice must include adversarial tests for unauthenticated requests, unrelated users, viewers attempting definition creation, duplicate keys, wrong tenant, wrong inventory, wrong-scope cursors, and asset values referencing hidden or unknown definitions.
