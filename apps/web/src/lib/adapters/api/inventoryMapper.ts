import type {
  AccessSummary as ApiAccessSummary,
  AuditRecord as ApiAuditRecord,
  Asset as ApiAsset,
  AssetCheckout as ApiAssetCheckout,
  Attachment as ApiAttachment,
  CheckedOutAsset as ApiCheckedOutAsset,
  CurrentCheckout as ApiCurrentCheckout,
  AssetKind as ApiAssetKind,
  CustomAssetType as ApiCustomAssetType,
  CustomFieldDefinition as ApiCustomFieldDefinition,
  AssetSearchResult as ApiSearchResult,
  InventoryAccessGrant as ApiInventoryAccessGrant,
  InventoryAccessInvitation as ApiInventoryAccessInvitation,
  Inventory as ApiInventory,
  Principal as ApiPrincipal,
  Tenant as ApiTenant
} from '@stuff-stash/api-client';
import {
  canEditInventory,
  type AccessSummary,
  type Asset,
  type AssetAttachment,
  type AssetCheckout,
  type AssetCheckoutState,
  type AuditRecord,
  type AttachmentContentType,
  type AssetKind,
  type CheckedOutAsset,
  type CurrentCheckout,
  type CreatedInventoryAccessInvitation,
  type CustomAssetType,
  type CustomFieldDefinition,
  type InventoryAccessGrant,
  type InventoryAccessInvitation,
  type Capability,
  type Inventory,
  type Principal,
  type SearchResult,
  type Tenant
} from '$lib/domain/inventory';

export function mapPrincipal(principal: ApiPrincipal): Principal {
  return {
    id: principal.id,
    email: principal.email
  };
}

export function mapTenant(tenant: ApiTenant): Tenant {
  return {
    id: tenant.id,
    name: tenant.name,
    access: mapAccess(tenant.access)
  };
}

export function mapInventory(inventory: ApiInventory): Inventory {
  return {
    id: inventory.id,
    tenantId: inventory.tenantId,
    name: inventory.name,
    access: mapAccess(inventory.access)
  };
}

export function mapCapability(inventory: Inventory | null | undefined): Capability {
  if (canEditInventory(inventory)) {
    return 'editor';
  }
  return 'viewer';
}

export function mapAsset(asset: ApiAsset): Asset {
  return {
    id: asset.id,
    tenantId: asset.tenantId,
    inventoryId: asset.inventoryId,
    kind: mapAssetKind(asset.kind),
    title: asset.title,
    description: asset.description,
    parentAssetId: asset.parentAssetId,
    lifecycleState: asset.lifecycleState,
    customAssetTypeId: asset.customAssetTypeId,
    customFields: asset.customFields,
    currentCheckout: mapCurrentCheckout(asset.currentCheckout),
    updatedAt: undefined
  };
}

export function mapCurrentCheckout(checkout: ApiCurrentCheckout | undefined): CurrentCheckout | undefined {
  if (!checkout) {
    return undefined;
  }
  return {
    id: checkout.id,
    state: checkout.state as AssetCheckoutState,
    checkedOutAt: checkout.checkedOutAt,
    checkedOutByPrincipalId: checkout.checkedOutByPrincipalId
  };
}

export function mapAssetCheckout(checkout: ApiAssetCheckout): AssetCheckout {
  return {
    id: checkout.id,
    tenantId: checkout.tenantId,
    inventoryId: checkout.inventoryId,
    assetId: checkout.assetId,
    state: checkout.state as AssetCheckoutState,
    checkedOutAt: checkout.checkedOutAt,
    checkedOutByPrincipalId: checkout.checkedOutByPrincipalId,
    checkoutDetails: checkout.checkoutDetails,
    returnedAt: checkout.returnedAt,
    returnedByPrincipalId: checkout.returnedByPrincipalId,
    returnDetails: checkout.returnDetails,
    createdAt: checkout.createdAt,
    updatedAt: checkout.updatedAt
  };
}

export function mapCheckedOutAsset(checkedOut: ApiCheckedOutAsset): CheckedOutAsset {
  const checkout = mapCurrentCheckout(checkedOut.checkout);
  if (!checkout) {
    throw new Error('Checked-out asset is missing checkout state.');
  }
  return {
    asset: mapAsset(checkedOut.asset),
    checkout
  };
}

export function mapAttachment(attachment: ApiAttachment, thumbnailUrl?: string, thumbnailHeaders?: Record<string, string>): AssetAttachment {
  return {
    id: attachment.id,
    tenantId: attachment.tenantId,
    inventoryId: attachment.inventoryId,
    assetId: attachment.assetId,
    fileName: attachment.fileName,
    contentType: mapAttachmentContentType(attachment.contentType),
    sizeBytes: attachment.sizeBytes,
    lifecycleState: attachment.lifecycleState,
    thumbnailUrl,
    thumbnailHeaders
  };
}

export function mapSearchResult(result: ApiSearchResult): SearchResult {
  return {
    type: 'asset',
    asset: mapAsset(result.asset),
    inventory: result.inventory,
    matches: result.matches
  };
}

export function mapInventoryAccessGrant(grant: ApiInventoryAccessGrant): InventoryAccessGrant {
  return {
    tenantId: grant.tenantId,
    inventoryId: grant.inventoryId,
    principalId: grant.principalId,
    relationship: grant.relationship
  };
}

export function mapInventoryAccessInvitation(invitation: ApiInventoryAccessInvitation): InventoryAccessInvitation {
  return {
    id: invitation.id,
    tenantId: invitation.tenantId,
    inventoryId: invitation.inventoryId,
    email: invitation.email,
    relationship: invitation.relationship,
    status: invitation.status,
    isExpired: invitation.isExpired,
    expiresAt: invitation.expiresAt,
    inviterPrincipalId: invitation.inviterPrincipalId,
    acceptedPrincipalId: invitation.acceptedPrincipalId
  };
}

export function mapCreatedInventoryAccessInvitation(
  invitation: ApiInventoryAccessInvitation
): CreatedInventoryAccessInvitation {
  return {
    invitation: mapInventoryAccessInvitation(invitation),
    acceptanceToken: invitation.acceptanceToken
  };
}

export function mapAuditRecord(record: ApiAuditRecord): AuditRecord {
  return {
    id: record.id,
    tenantId: record.tenantId,
    inventoryId: record.inventoryId,
    principalId: record.principalId,
    action: record.action,
    source: record.source,
    targetType: record.targetType,
    targetId: record.targetId,
    occurredAt: record.occurredAt,
    requestId: record.requestId,
    metadata: record.metadata
  };
}

export function mapCustomAssetType(assetType: ApiCustomAssetType): CustomAssetType {
  return {
    id: assetType.id,
    tenantId: assetType.tenantId,
    inventoryId: assetType.inventoryId,
    scope: assetType.scope,
    key: assetType.key,
    displayName: assetType.displayName,
    description: assetType.description,
    lifecycleState: assetType.lifecycleState
  };
}

export function mapCustomFieldDefinition(definition: ApiCustomFieldDefinition): CustomFieldDefinition {
  return {
    id: definition.id,
    tenantId: definition.tenantId,
    inventoryId: definition.inventoryId,
    scope: definition.scope,
    key: definition.key,
    displayName: definition.displayName,
    type: definition.type,
    enumOptions: definition.enumOptions,
    applicability: definition.applicability,
    customAssetTypeIds: definition.customAssetTypeIds,
    lifecycleState: definition.lifecycleState
  };
}

function mapAttachmentContentType(contentType: string): AttachmentContentType {
  if (contentType === 'image/jpeg' || contentType === 'image/png' || contentType === 'image/webp' || contentType === 'application/pdf') {
    return contentType;
  }
  throw new Error(`Unsupported attachment content type: ${contentType}`);
}

function mapAssetKind(kind: ApiAssetKind): AssetKind {
  return kind;
}

function mapAccess(access: ApiAccessSummary): AccessSummary {
  return {
    relationship: access.relationship,
    permissions: access.permissions ?? []
  };
}
