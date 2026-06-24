# Tenant And Inventory Access Spec

## Purpose

Stuff Stash must support multi-tenant inventory sharing from the beginning.

The access model should feel familiar to users: tenants contain inventories, inventories can be shared, and users can have different relationships to different inventories.

## Scope

This spec covers the initial identity, tenant, inventory, authentication, and authorization model.

This spec does not define every SpiceDB relationship, UI screen, billing concept, groups, or ownership transfer.

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

It does not implement user search, tenant membership management, groups, or ownership transfer.

The first endpoints are:

- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-grants`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-grants`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-invitations`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/expiration`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/cancel`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}`

## My Access API

Web, mobile, and future agent clients need a first-class way to discover the authenticated user's accessible tenants and effective permissions without hard-coding authorization rules in the client.

`GET /me/tenants`:

- Requires authentication.
- Returns only active tenants the authenticated principal can view.
- Must use the standard response envelope.
- Must use cursor pagination with `limit` and `cursor`.
- Must preserve tenant isolation. Hidden tenants must not appear and must not affect response bodies except through pagination cursors.
- Must include effective tenant access metadata for each tenant:
  - `relationship`: `owner` when the principal can configure the tenant or create inventories, otherwise `viewer`.
  - `permissions`: stable permission strings granted to the caller for that tenant, initially `view`, `create_inventory`, and `configure`.
- Must produce safe tenant-scoped read audit history for each returned tenant using the `tenant.listed` audit action.
- Must emit domain-oriented observability.

Tenant detail responses and `GET /me/tenants` tenant entries must include the same effective tenant access metadata when returned to an authenticated caller.

Inventory create, detail, update, lifecycle, and list responses must include effective inventory access metadata:

- `relationship`: `owner` when the principal can share or configure the inventory, `editor` when the principal can create or edit assets, otherwise `viewer`.
- `permissions`: stable permission strings granted to the caller for that inventory, initially `view`, `create_asset`, `edit_asset`, `share`, and `configure`.

Effective access metadata is a client display and workflow affordance. It must be derived from authorization checks and must not replace server-side authorization checks for any state-changing operation.

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

`POST /access-invitations`:

- Requires authentication.
- Requires `inventory.share`.
- Accepts an invitee email address and a relationship.
- Supports only `viewer` and `editor` relationships in the first slice.
- Must create a pending invitation scoped to one tenant and one inventory.
- Must normalize invitation email addresses for matching.
- Must return one-time invite link material, including a raw acceptance token, to the caller.
- Must never store the raw acceptance token directly.
- Must store only a derived token verifier, such as a cryptographic hash.
- Must set an expiration timestamp for the invite link token.
- The invitation token TTL must come from environment-backed configuration.
- Must not require email delivery, SMTP, or a third-party messaging service.
- Must allow self-hosted deployments to copy and deliver invite links manually.
- Future email, chat, or app notification delivery must be adapters around the same invitation contract.
- Invite links are not a primary authentication mechanism.
- Invite links must be used only to prove possession of invitation acceptance material.
- Accepting an invite still requires an authenticated principal whose verified email matches the invitation.
- Must not create SpiceDB relationships or direct access grants before acceptance.
- Must produce audit history.
- Must reject duplicate pending invitations for the same tenant, inventory, email, and relationship instead of silently rotating acceptance tokens.

`GET /access-invitations`:

- Requires authentication.
- Requires `inventory.share`.
- Lists invitation metadata scoped to the tenant and inventory.
- Must never return raw acceptance token material or token verifiers.
- Must use the standard response envelope.
- Must use cursor pagination with `limit` and `cursor`.
- Must support a status filter with `all`, `pending`, `accepted`, `revoked`, `cancelled`, and `expired`.
- `all` is the default status filter.
- `expired` is a derived listing filter for pending invitations whose expiration timestamp is not in the future.
- `pending` must include only pending invitations whose expiration timestamp is still in the future.
- Invitation responses must include whether the invitation is currently expired so clients do not have to infer expiration from clocks alone.
- Must produce safe read audit history.

`POST /access-invitations/{invitationId}/accept`:

- Requires authentication.
- Requires the authenticated principal to have a verified email address matching the invitation email.
- Requires the raw acceptance token from the invite link.
- Must reject missing principal email, wrong email, missing token, wrong token, expired token, wrong tenant, wrong inventory, revoked invitations, and already accepted invitations.
- A valid token without a matching authenticated verified email must not grant access.
- A matching authenticated verified email without the valid token must not grant access.
- Must mark the invitation accepted, create the direct access grant for the accepting principal, enqueue the matching SpiceDB grant event, and write audit history in one transaction.
- Must drain the authorization outbox after commit and leave failed relationship writes retryable.

`GET /access-invitations/{invitationId}`:

- Requires authentication.
- Requires `inventory.share`.
- Returns invitation metadata without returning raw acceptance token material.
- Must produce safe read audit history.

`PATCH /access-invitations/{invitationId}/expiration`:

- Requires authentication.
- Requires `inventory.share`.
- Accepts an RFC3339 expiration timestamp.
- Must update only pending invitations.
- May set the expiration into the past to manually expire a pending invite.
- Must not rotate or return acceptance token material.
- Must reject accepted, revoked, cancelled, deleted, wrong-tenant, and wrong-inventory invitations.
- Must produce audit history.
- Must return the updated invitation metadata.

`PATCH /access-invitations/{invitationId}/cancel`:

- Requires authentication.
- Requires `inventory.share`.
- Must cancel only pending invitations.
- Must not remove existing direct access grants.
- Must produce audit history only when a pending invitation existed and was cancelled.
- Must use the standard no-content response.

`DELETE /access-invitations/{invitationId}`:

- Requires authentication.
- Requires `inventory.share`.
- Must hard-delete invitation metadata only.
- Must not remove existing direct access grants.
- Must preserve audit history.
- Must produce audit history only when invitation metadata existed and was removed.
- Must use the standard no-content response.

Granting `viewer` allows inventory viewing and asset listing.

Granting `editor` allows inventory viewing, asset listing, asset creation, asset update, and same-inventory asset movement.

Neither `viewer` nor `editor` allows sharing access onward.

## Repository Boundaries

- Inventory sharing persistence must live behind an explicit inventory access repository port.
- The inventory repository must own inventory aggregate persistence only.
- Direct access grants, invite-link invitations, invitation acceptance, invitation expiration management, and related authorization outbox writes must not remain on the inventory repository port.
- The inventory access repository must preserve transactional behavior for grant, revoke, invitation acceptance, expiration update, and audit/outbox writes.
- Persistence adapters may implement inventory and inventory access repositories on the same concrete store type, but the application layer must depend on the narrower port for each responsibility.

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
- Invitation API tests must prove owners can create, accept, and revoke pending invitations; viewers, editors, unrelated users, wrong-email users, missing-email users, missing-token users, wrong-token users, expired-token users, wrong-tenant callers, and wrong-inventory callers cannot exceed their permissions; revoked invitations cannot be accepted; accepted invitations create only the intended direct access relationship.
- Invitation listing and expiration tests must prove only users with `inventory.share` can list and update invitations, pagination preserves the standard contract, raw acceptance token material is not returned after creation, expired filtering works, manually expired invitations cannot be accepted, and non-pending invitations cannot have expiration changed.

## Open Questions

- What is the first mobile-friendly OIDC flow?
- How should users switch between tenants and inventories?
