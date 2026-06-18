# Tenant And Inventory Access Spec

## Purpose

Stuff Stash must support multi-tenant inventory sharing from the beginning.

The access model should feel familiar to users: tenants contain inventories, inventories can be shared, and users can have different relationships to different inventories.

## Scope

This spec covers the initial identity, tenant, inventory, authentication, and authorization model.

This spec does not define every SpiceDB relationship, REST endpoint, invitation flow, UI screen, or billing concept.

## Model

- A tenant is the top-level security boundary.
- A tenant may contain multiple inventories.
- An inventory belongs to exactly one tenant.
- A user may belong to multiple tenants.
- A user may have access to one or more inventories within a tenant.
- Access should be modeled as relationships, not hard-coded application roles.
- Sharing should be inspired by Google Drive-style access.
- Guests, family members, or other users may be granted limited access to specific inventories.

## Authentication

- OIDC and SSO are core capabilities from the beginning.
- Authentication must be behind ports and adapters.
- The first concrete authentication provider must be Google Single Sign-On.
- The architecture must support arbitrary OIDC providers.
- OIDC provider details must not leak into domain logic.
- Authentication configuration must come from environment-backed configuration.
- Auth and session endpoints must comply with the relevant OIDC flow chosen for web and mobile clients.

## Authorization

- Authorization must use SpiceDB.
- Authorization must be relationship-based.
- Authorization must be behind ports and adapters.
- Authorization checks must be performed for every security-sensitive operation.
- Application services must authorize operations using the current principal, tenant, inventory, and target resource.
- Conversational actions must use the permissions of the authenticated user who initiated the action.
- A language model, agent loop, MCP tool, or generated SDK must not receive elevated permissions.
- Authorization checks must apply equally to REST, MCP, web, mobile, background jobs, imports, and conversational flows.

## Initial Relationships

The first relationship model should include these concepts:

- Tenant owner.
- Tenant admin.
- Inventory owner.
- Inventory editor.
- Inventory viewer.
- Direct inventory sharing with another user.
- Inherited access from tenant to inventory where appropriate.

Exact SpiceDB schema names and inheritance rules must be specified before implementation.

## Custom Fields

- Tenant-scoped custom field definitions are controlled by tenant-level relationships.
- Inventory-scoped custom field definitions are controlled by inventory-level relationships.
- Custom field definitions do not need separate per-field permissions at first.
- A user who can configure an inventory may configure that inventory's custom field definitions.
- A user who can edit assets in an inventory may set values for custom fields available to that inventory.

## Security Tests

- Every authenticated and authorized interaction point must have adversarial end-to-end tests.
- Tests must cover unauthenticated access, wrong-tenant access, wrong-inventory access, viewer attempting edits, editor attempting admin operations, malformed tokens, expired tokens, and privilege escalation attempts.
- Tests must verify that conversational actions cannot exceed the initiating user's permissions.

## Open Questions

- What exact relationship graph should be used in SpiceDB?
- Which tenant relationships inherit to inventories?
- How are invitations created, accepted, revoked, and audited?
- What is the first mobile-friendly OIDC flow?
- How should users switch between tenants and inventories?
