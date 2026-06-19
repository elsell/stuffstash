# Flexible Asset Fields Spec

## Purpose

Stuff Stash must support useful inventory details without forcing every household, hobby, tool, pantry item, document, or garage object into one rigid schema.

The asset model needs a stable core plus a custom-field storage path from day one.

## Scope

This spec covers flexible asset metadata requirements.

This spec does not define the full asset aggregate, asset lifecycle, search model, UI editor, or import/export behavior.

The first asset slice may store only an empty custom field value object until custom field definition persistence and validation are implemented.

## Requirements

- Assets must have a stable domain core for fields that are truly common to inventory items.
- Assets must support custom fields.
- The initial asset schema must include a place to store custom field values.
- Non-empty custom field values must be rejected until custom field definition persistence and validation are implemented.
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

The initial custom field type set should include:

- Text.
- Number.
- Boolean.
- Date.
- URL.
- Enumeration.

Future specs may add richer field types such as money, quantity with units, expiration date, warranty details, file attachment, image reference, serial number, barcode, or model number.

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
