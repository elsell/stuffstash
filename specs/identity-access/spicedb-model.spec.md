# SpiceDB Model Spec

## Purpose

Stuff Stash needs an initial relationship-based authorization model before protected endpoints and application services are implemented.

## Scope

This spec defines the first SpiceDB relationship model direction.

This spec does not define every schema line, invitation workflow, or migration process for authorization data.

## Decisions

- Authorization must use SpiceDB.
- Relationships must follow a Google Drive-style sharing model.
- Tenants are the top-level security boundary.
- Inventories are shareable organizational units inside tenants.
- Assets, locations, custom fields, audit records, search results, and media access inherit from inventory access unless a future spec defines narrower permissions.

## Initial Object Types

The first SpiceDB schema should include:

- `user`
- `tenant`
- `inventory`
- `asset`
- `attachment`
- `audit_record`

## Initial Relationships

The first model should support:

- Tenant owner.
- Tenant admin.
- Inventory owner.
- Inventory editor.
- Inventory viewer.
- Inventory access inherited from tenant owner or tenant admin.
- Asset access inherited from inventory access.
- Location-like access represented through asset access for assets with kind `location`.
- Attachment access inherited from attached resource access.
- Inventory audit record access inherited from inventory view access.
- Tenant-wide audit record access limited to tenant configuration permission.

## Permissions

The first model should express permissions for:

- View inventory.
- Configure inventory.
- Share inventory.
- Create asset.
- Edit asset.
- Move asset through asset edit permission in the first REST slice.
- View location as asset view on a `location` asset.
- Create location as asset create with kind `location`.
- Move location as asset move on a `location` asset.
- Upload attachment.
- View attachment.
- Export inventory.
- View audit history.
- Undo action.

The first Go port permission name for asset update and same-inventory movement is `edit_asset`, mapped to the SpiceDB inventory `edit` permission.
The first audit read slice uses existing `inventory.view` for inventory-scoped audit records and existing `tenant.configure` for tenant-wide audit records.
Exact permission names for future per-asset permissions must be defined before implementation.

Archive and unarchive permissions are future permissions and must not be added to the first asset slice until archive behavior is specified and exposed.

## Conversational Actions

- Conversational actions must check permissions using the initiating user.
- Agents, model providers, and MCP tools must not receive independent authorization grants.
- Every command produced by an action plan must be authorized before execution.

## Testing

- Tests must verify inherited tenant admin access.
- Tests must verify direct inventory sharing.
- Tests must verify viewer, editor, owner, and admin behavior.
- Tests must verify denied cross-tenant access.
- Tests must verify that conversational actions cannot exceed the user's relationships.

## Open Questions

- Which operations should inventory editor be allowed to undo?
- How should pending invitations be represented?
