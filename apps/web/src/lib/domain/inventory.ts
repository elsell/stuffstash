export type AssetKind = 'item' | 'container' | 'location';
export type AssetLifecycleState = 'active' | 'archived';
export type AssetLifecycleFilter = AssetLifecycleState;
export type SearchLifecycleFilter = AssetLifecycleFilter | 'all';
export type AttachmentContentType = 'image/jpeg' | 'image/png' | 'image/webp' | 'application/pdf';
export type SearchMode = 'fuzzy' | 'exact';
export type WorkspaceMode = 'home' | 'location' | 'asset' | 'search' | 'settings';
export type Capability = 'editor' | 'viewer';
export type InventoryAccessRelationship = 'viewer' | 'editor';
export type InvitationStatus = 'pending' | 'accepted' | 'revoked' | 'cancelled' | 'expired';
export type InvitationStatusFilter = InvitationStatus | 'all';
export type CustomDefinitionScope = 'tenant' | 'inventory';
export type CustomDefinitionLifecycleState = 'active' | 'archived';
export type CustomFieldType = 'text' | 'number' | 'boolean' | 'date' | 'url' | 'enum';
export type CustomFieldApplicability = 'all_assets' | 'custom_asset_types';

export interface Principal {
  id: string;
  email?: string;
}

export interface AccessSummary {
  relationship: string;
  permissions: string[];
}

export interface Tenant {
  id: string;
  name: string;
  access: AccessSummary;
}

export interface Inventory {
  id: string;
  tenantId: string;
  name: string;
  access: AccessSummary;
}

export interface AssetPhoto {
  id: string;
  url: string;
  alt: string;
}

export interface AssetAttachment {
  id: string;
  tenantId: string;
  inventoryId: string;
  assetId: string;
  fileName: string;
  contentType: AttachmentContentType;
  sizeBytes: number;
  lifecycleState: AssetLifecycleState;
  thumbnailUrl?: string;
  thumbnailHeaders?: Record<string, string>;
}

export interface InventoryAccessGrant {
  tenantId: string;
  inventoryId: string;
  principalId: string;
  relationship: InventoryAccessRelationship;
}

export interface InventoryAccessInvitation {
  id: string;
  tenantId: string;
  inventoryId: string;
  email: string;
  relationship: InventoryAccessRelationship;
  status: InvitationStatus;
  isExpired: boolean;
  expiresAt: string;
  inviterPrincipalId: string;
  acceptedPrincipalId?: string;
}

export interface CreatedInventoryAccessInvitation {
  invitation: InventoryAccessInvitation;
  acceptanceToken?: string;
}

export type AuditScope = 'inventory' | 'tenant';

export interface AuditRecord {
  id: string;
  tenantId: string;
  inventoryId: string | null;
  principalId: string;
  action: string;
  source: string;
  targetType: string;
  targetId: string;
  occurredAt: string;
  requestId?: string;
  metadata: Record<string, string>;
}

export interface CustomAssetType {
  id: string;
  tenantId: string;
  inventoryId: string | null;
  scope: CustomDefinitionScope;
  key: string;
  displayName: string;
  description: string;
  lifecycleState: CustomDefinitionLifecycleState;
}

export interface CustomFieldDefinition {
  id: string;
  tenantId: string;
  inventoryId: string | null;
  scope: CustomDefinitionScope;
  key: string;
  displayName: string;
  type: CustomFieldType;
  enumOptions: string[];
  applicability: CustomFieldApplicability;
  customAssetTypeIds: string[];
  lifecycleState: CustomDefinitionLifecycleState;
}

export interface MediaUploadPolicy {
  supportedContentTypes: AttachmentContentType[];
  maxBytes: number;
}

export interface Asset {
  id: string;
  tenantId: string;
  inventoryId: string;
  kind: AssetKind;
  title: string;
  description: string;
  parentAssetId: string | null;
  lifecycleState: AssetLifecycleState;
  customAssetTypeId?: string;
  customFields?: Record<string, unknown>;
  customAssetTypeLabel?: string;
  photo?: AssetPhoto;
  updatedAt?: string;
}

export interface SearchResult {
  type: 'asset';
  asset: Asset;
  inventory: Pick<Inventory, 'id' | 'name'>;
  matches: Array<{
    field: string;
    value: string;
  }>;
}

export interface SearchRequest {
  tenantId: string;
  inventoryId: string;
  query: string;
  lifecycleState: SearchLifecycleFilter;
  mode: SearchMode;
}

export interface AddAssetDraft {
  kind: AssetKind;
  title: string;
  description: string;
  parentAssetId: string | null;
  customAssetTypeId?: string;
  customFields?: Record<string, unknown>;
  photos: SelectedPhoto[];
}

export interface UpdateAssetDraft {
  title: string;
  description: string;
  parentAssetId: string | null;
  customFields?: Record<string, unknown>;
}

export interface SelectedPhoto {
  id: string;
  name: string;
  sizeBytes: number;
  contentType: AttachmentContentType;
  previewUrl: string;
  file: File;
}

export interface WorkspaceContext {
  principal: Principal;
  tenants: Tenant[];
  inventories: Inventory[];
  selectedTenantId: string;
  selectedInventoryId: string;
  assetLifecycleState: AssetLifecycleFilter;
  mediaUploadPolicy: MediaUploadPolicy;
  customAssetTypes: CustomAssetType[];
  customFieldDefinitions: CustomFieldDefinition[];
  capability: Capability;
}

export interface WorkspaceData {
  context: WorkspaceContext;
  assets: Asset[];
}

export interface LocationSummary {
  location: Asset;
  assetCount: number;
}

export interface AssetViewModel extends Asset {
  containmentTrail: string;
}

export const assetKinds: AssetKind[] = ['item', 'container', 'location'];
export const defaultMediaUploadPolicy: MediaUploadPolicy = {
  supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'],
  maxBytes: 5 * 1024 * 1024
};

export function assetKindLabel(kind: AssetKind): string {
  switch (kind) {
    case 'item':
      return 'Item';
    case 'container':
      return 'Container';
    case 'location':
      return 'Location';
  }
}

export function hasAccessPermission(access: AccessSummary | null | undefined, permission: string): boolean {
  return access?.permissions.includes(permission) ?? false;
}

export function canCreateInventory(tenant: Tenant | null | undefined): boolean {
  return hasAccessPermission(tenant?.access, 'create_inventory');
}

export function canCreateAsset(inventory: Inventory | null | undefined): boolean {
  return hasAccessPermission(inventory?.access, 'create_asset');
}

export function canEditInventory(inventory: Inventory | null | undefined): boolean {
  return canCreateAsset(inventory) || hasAccessPermission(inventory?.access, 'edit_asset');
}

export function canEditAsset(inventory: Inventory | null | undefined): boolean {
  return hasAccessPermission(inventory?.access, 'edit_asset');
}

export function applicableCustomFieldDefinitions(
  definitions: CustomFieldDefinition[],
  customAssetTypeId: string | undefined
): CustomFieldDefinition[] {
  return definitions.filter(
    (definition) =>
      definition.lifecycleState === 'active' &&
      (definition.applicability === 'all_assets' ||
        (!!customAssetTypeId && definition.customAssetTypeIds.includes(customAssetTypeId)))
  );
}
