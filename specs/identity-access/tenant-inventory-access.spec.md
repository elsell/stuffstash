# Tenant And Inventory Access Spec

## Purpose

Stuff Stash must support multi-tenant inventory sharing from the beginning.

The access model should feel familiar to users: tenants contain inventories, inventories can be shared, and users can have different relationships to different inventories.

## Scope

This spec covers the initial identity, tenant, inventory, authentication, and authorization model.

This spec does not define every SpiceDB relationship, email invitation acceptance flow, UI screen, or billing concept.

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
- Inherited access from tenant to inventory as specified in the SpiceDB schema spec.

## First Sharing API Slice

The first user-management slice supports direct inventory sharing with known principal IDs.

It does not implement email invitations, invite acceptance, user search, tenant membership management, groups, or ownership transfer.

The first endpoints are:

- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-grants`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-grants`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}`

`POST /access-grants`:

- Requires authentication.
- Requires `inventory.share`.
- Accepts a target principal ID and a relationship.
- Supports only `viewer` and `editor` relationships in the first slice.
- Must reject granting access to the caller's own principal when that would be a no-op.
- Must persist the grant intent durably before SpiceDB is updated.
- Must use the authorization outbox for the SpiceDB relationship write.
- Must be idempotent for the same tenant, inventory, principal, and relationship.
- Repeating the same grant must not create another direct grant row or another authorization outbox event.

`GET /access-grants`:

- Requires authentication.
- Requires `inventory.share`.
- Lists direct inventory grants known to Stuff Stash persistence.
- Must use the standard response envelope.
- Must use cursor pagination with `limit` and `cursor`.
- Must not list inherited tenant-owner or tenant-admin access as direct grants.

`DELETE /access-grants/{principalId}/{relationship}`:

- Requires authentication.
- Requires `inventory.share`.
- Supports only direct `viewer` and `editor` relationships in the first slice.
- Must remove only the requested direct grant row.
- Must not remove inherited tenant-owner, tenant-admin, or inventory-owner access.
- Must persist the grant removal, matching authorization outbox event, and request-owned outbox claim in the same database transaction.
- Must be idempotent for a missing direct grant.
- Must still enqueue a matching authorization outbox event for a missing direct grant so stale direct SpiceDB relationships can self-heal.
- Must produce audit history only when a direct grant actually existed and was removed.
- Must process the claimed revoke authorization outbox event after the durable transaction commits.
- Must not return the standard successful deletion response when the immediate revoke drain fails, because a caller must not receive success while the direct authorization relationship may still be active.
- Must use the standard no-content response for successful deletion.

Granting direct inventory access must also grant tenant `viewer` to the target principal so the user can resolve the containing tenant without receiving tenant configuration or sibling-inventory access.

Revoking direct inventory access removes the inventory relationship. Tenant viewer cleanup is deferred until tenant membership and invitation semantics are specified, because a user may still need tenant view through another inventory.

Granting `viewer` allows inventory viewing and asset listing.

Granting `editor` allows inventory viewing, asset listing, asset creation, asset update, and same-inventory asset movement.

Neither `viewer` nor `editor` allows sharing access onward.

## Custom Fields

- Tenant-scoped custom asset type definitions are controlled by tenant-level relationships.
- Inventory-scoped custom asset type definitions are controlled by inventory-level relationships.
- Custom asset type definitions do not need separate per-type permissions at first.
- A user who can configure an inventory may configure that inventory's custom asset type definitions.
- Tenant-scoped custom field definitions are controlled by tenant-level relationships.
- Inventory-scoped custom field definitions are controlled by inventory-level relationships.
- Custom field definitions do not need separate per-field permissions at first.
- A user who can configure an inventory may configure that inventory's custom field definitions.
- A user who can edit assets in an inventory may set values for custom fields available to that inventory.
- A user who can edit assets in an inventory may assign available custom asset types to assets in that inventory.

## Security Tests

- Every authenticated and authorized interaction point must have adversarial end-to-end tests.
- Tests must cover unauthenticated access, wrong-tenant access, wrong-inventory access, viewer attempting edits, editor attempting admin operations, malformed tokens, expired tokens, and privilege escalation attempts.
- Tests must verify that conversational actions cannot exceed the initiating user's permissions.
- Sharing API tests must prove owners can grant and revoke direct access, unrelated users cannot grant, revoke, or list grants, viewers cannot share or revoke, editors cannot share or revoke, granted viewers/editors receive only their intended permissions, and revoked users lose the revoked inventory permissions.

## Open Questions

- How are email invitations created, accepted, revoked, and audited?
- What is the first mobile-friendly OIDC flow?
- How should users switch between tenants and inventories?
