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
- `location`
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
- Location access inherited from inventory access.
- Attachment access inherited from attached resource access.
- Audit record access inherited from inventory access, with final read rules specified before audit API implementation.

## Permissions

The first model should express permissions for:

- View inventory.
- Configure inventory.
- Share inventory.
- Create asset.
- Edit asset.
- Move asset.
- Archive asset.
- View location.
- Create location.
- Move location.
- Upload attachment.
- View attachment.
- Export inventory.
- View audit history.
- Undo action.

Exact permission names must be defined before implementation.

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

- Should inventory viewer be allowed to view audit history?
- Which operations should inventory editor be allowed to undo?
- How should pending invitations be represented?
