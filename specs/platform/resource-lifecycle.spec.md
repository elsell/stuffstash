# Resource Lifecycle Spec

## Purpose

Stuff Stash needs complete, domain-oriented REST lifecycle coverage before the web UI grows around incomplete API behavior.

The API must support normal user deletion through reversible archive/cancel behavior, deliberate hard delete where the domain allows it, resource detail reads, and audit history for both reads and writes.

## Scope

This spec covers lifecycle behavior for the current REST resources:

- Tenants.
- Inventories.
- Assets.
- Attachments.
- Custom field definitions.
- Custom asset types.
- Inventory access grants.
- Inventory access invitations.

It does not define `PUT` replacement endpoints, undo, export-before-delete workflows, retention policies, background compaction, or final user-facing delete copy.

## Cross-Cutting Rules

- Specs must be updated before lifecycle code changes.
- No resource gets verbs mechanically. Each endpoint must map to real domain behavior.
- `PUT` is not implemented in this lifecycle slice.
- Normal delete UX must prefer soft delete:
  - `archive` for resources that remain restorable.
  - `cancel` for pending invitations.
  - `revoke` for access grants.
- Hard delete must exist where specified, but it must be explicit and separate from archive/cancel/revoke.
- Hard delete must still emit audit history before the resource is removed.
- Hard delete must not remove audit records.
- Hard delete must not bypass tenant isolation, authorization, validation, observability, or outbox consistency.
- Detail reads must exist for all current REST resources.
- List and detail reads must write audit history.
- State-changing endpoints must write audit history.
- Read audit records may use compact safe metadata so normal read traffic does not store sensitive payloads.
- All lifecycle endpoints must use the standard success envelope and safe error envelope.
- Hidden, unauthorized, cross-tenant, archived-where-not-requested, and hard-deleted resources must use safe not-found or safe authorization failures according to the existing API security contract.

## Lifecycle States

Resources with reversible lifecycle support use:

- `active`: visible in normal workflows.
- `archived`: hidden from normal workflows, still restorable, still available to audit history.

Invitations use:

- `pending`.
- `accepted`.
- `revoked`.
- `cancelled`.

Access grants do not use archive. Revocation removes the active relationship and writes audit/outbox records. A hard-delete operation is not exposed for grants because the direct relationship row is the resource.

## Tenant Lifecycle

Tenant endpoints:

- `GET /tenants/{tenantId}`
- `PATCH /tenants/{tenantId}`
- `PATCH /tenants/{tenantId}/archive`
- `PATCH /tenants/{tenantId}/restore`
- `DELETE /tenants/{tenantId}`

Rules:

- Tenant detail requires `tenant.view`.
- Tenant update initially supports renaming only and requires `tenant.configure`.
- Tenant archive requires `tenant.configure`.
- Archived tenants are hidden from normal tenant-scoped workflows. Restore is the explicit endpoint for returning an archived tenant to normal use; detail reads return not found while the tenant is archived.
- Tenant restore requires `tenant.configure`.
- Tenant hard delete requires `tenant.configure`.
- Tenant hard delete must be blocked while any inventories exist, including archived inventories.
- Tenant hard delete must preserve tenant audit records.
- Tenant archive/restore/delete must emit audit records.
- Tenant detail and tenant update must emit audit records.

## Inventory Lifecycle

Inventory endpoints:

- `GET /tenants/{tenantId}/inventories/{inventoryId}`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/archive`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/restore`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}`

Rules:

- Inventory detail requires `inventory.view`.
- Inventory update initially supports renaming only and requires `inventory.configure`.
- Inventory archive requires `inventory.configure`.
- Archived inventories are hidden from normal inventory lists.
- Archived inventories remain restorable.
- Inventory restore requires `inventory.configure` and an active tenant.
- Inventory hard delete requires `inventory.configure`.
- Inventory hard delete must be blocked while any assets exist, including archived assets.
- Inventory hard delete must preserve audit records.
- Inventory lifecycle changes and detail reads must emit audit records.

## Asset Lifecycle

Asset endpoints:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}`
- existing `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}`
- existing `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive`
- existing `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}`

Rules:

- Asset detail requires `inventory.view`.
- Asset detail must not return archived assets unless the request explicitly allows archived detail through a lifecycle-aware endpoint or query parameter. The first implementation may allow direct detail for archived assets to support restore UI when the caller has `inventory.view`.
- Asset hard delete requires `inventory.edit_asset`.
- Asset hard delete must be blocked while active children exist.
- Asset hard delete must preserve audit records.
- Asset hard delete must emit audit history before removing the asset.
- Existing archive/restore behavior remains the default reversible delete path.

## Attachment Lifecycle

Attachment endpoints:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}`
- existing `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/archive`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/restore`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}`

Rules:

- Attachment detail requires `inventory.view`.
- Attachment archive/restore/hard delete requires `inventory.edit_asset`.
- Attachment create, list, detail, content download, archive, restore, and hard delete require the parent asset to be active.
- Archived attachments are hidden from normal attachment lists.
- Attachment hard delete must delete blob content through the blob storage port before removing attachment metadata.
- Blob deletion failure must stop metadata removal so the system does not create orphaned blob content.
- The attachment delete audit record must be written with the metadata removal so durable history is preserved when the delete completes.

## Custom Field Definition Lifecycle

Tenant-scoped endpoints:

- `GET /tenants/{tenantId}/custom-field-definitions/{definitionId}`
- existing `PATCH /tenants/{tenantId}/custom-field-definitions/{definitionId}`
- `PATCH /tenants/{tenantId}/custom-field-definitions/{definitionId}/archive`
- `PATCH /tenants/{tenantId}/custom-field-definitions/{definitionId}/restore`
- `DELETE /tenants/{tenantId}/custom-field-definitions/{definitionId}`

Inventory-scoped endpoints:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}`
- existing `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/archive`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/restore`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}`

Rules:

- Detail requires the same permission as list for that scope.
- Update/archive/restore/hard delete require configure permission for that scope.
- Archived field definitions are hidden from normal lists and ignored for new asset validation.
- Restore must revalidate custom asset type targets.
- Archive does not remove stored asset values and must not be blocked only because active assets have values for that field.
- Hard delete is blocked while any active asset stores a non-empty value for the field key.
- Hard delete must preserve audit records.

## Custom Asset Type Lifecycle

Tenant-scoped endpoints:

- `GET /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}`
- existing `PATCH /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}`
- existing `PATCH /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/archive`
- `PATCH /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/restore`
- `DELETE /tenants/{tenantId}/custom-asset-types/{customAssetTypeId}`

Inventory-scoped endpoints:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}`
- existing `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}`
- existing `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/archive`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/restore`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}`

Rules:

- Detail requires the same permission as list for that scope.
- Update/archive/restore/hard delete require configure permission for that scope.
- Archived custom asset types are hidden from normal lists and unavailable for new assets or new field targets.
- Restore makes the type available again.
- Hard delete is blocked while active assets or custom field definition targets reference the type.
- Hard delete must preserve audit records.

## Inventory Access Grant Lifecycle

Grant endpoints:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}`
- existing `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}`

Rules:

- Grant detail requires `inventory.share`.
- Grant revoke uses `DELETE` because the active relationship row is removed.
- Revocation must remain outbox-backed so SpiceDB consistency is preserved.
- Grant reads and revokes must emit audit records.

## Inventory Access Invitation Lifecycle

Invitation endpoints:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}`
- existing `POST /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/cancel`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}`

Rules:

- Invitation detail requires `inventory.share`, except acceptance may load only the minimal invitation needed for token verification.
- Pending invitations can be cancelled.
- Accepted, revoked, cancelled, and expired invitations cannot be accepted.
- `PATCH /cancel` is the normal user-visible lifecycle endpoint.
- `DELETE` is reserved for deliberate hard delete of invitation metadata and requires `inventory.share`.
- Invitation hard delete must preserve audit records.

## Read Audit

Read audit actions must be explicit action names ending in `.viewed` or `.listed`.

Required first actions:

- `tenant.viewed`.
- `inventory.viewed`.
- `inventory.listed`.
- `asset.viewed`.
- `asset.listed`.
- `attachment.viewed`.
- `attachment.listed`.
- `attachment.content_downloaded`.
- `custom_field_definition.viewed`.
- `custom_field_definition.listed`.
- `custom_asset_type.viewed`.
- `custom_asset_type.listed`.
- `inventory_access_grant.viewed`.
- `inventory_access_grant.listed`.
- `inventory_invitation.viewed`.
- `audit_record.listed`.

Read audit records must not contain response bodies, secrets, invite acceptance tokens, blob contents, or authorization internals.

## Write Audit

Lifecycle write audit actions must be explicit action names ending in `.created`, `.updated`, `.archived`, `.restored`, `.deleted`, `.cancelled`, `.revoked`, `.accepted`, or another domain-specific past-tense verb approved by the relevant spec.

Required lifecycle write actions include:

- `tenant.updated`, `tenant.archived`, `tenant.restored`, and `tenant.deleted`.
- `inventory.updated`, `inventory.archived`, `inventory.restored`, and `inventory.deleted`.
- `asset.deleted`, in addition to the existing asset create/update/archive/restore actions.
- `attachment.archived`, `attachment.restored`, and `attachment.deleted`.
- `custom_field_definition.archived`, `custom_field_definition.restored`, and `custom_field_definition.deleted`.
- `custom_asset_type.restored` and `custom_asset_type.deleted`, in addition to existing custom asset type create/update/archive actions.
- `inventory_invitation.cancelled` and `inventory_invitation.deleted`, in addition to existing invitation create/accept/revoke actions.

## Verification

- Every lifecycle endpoint must have happy-path tests.
- Every lifecycle endpoint must have adversarial auth/authz tests.
- Archive endpoints must verify hidden-from-normal-list behavior.
- Restore endpoints must verify the resource returns to normal lists.
- Hard delete endpoints must verify detail/list disappearance and preserved audit history.
- Read endpoints must verify audit records are written.
- OpenAPI tests must include all lifecycle endpoints.
- Generated TypeScript client artifacts must be regenerated after endpoint changes.
