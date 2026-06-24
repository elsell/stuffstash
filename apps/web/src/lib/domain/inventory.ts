export type AssetKind = 'item' | 'container' | 'location';
export type AssetLifecycleState = 'active' | 'archived';
export type WorkspaceMode = 'home' | 'location' | 'asset' | 'search' | 'settings';
export type Capability = 'editor' | 'viewer';

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

export interface Asset {
  id: string;
  tenantId: string;
  inventoryId: string;
  kind: AssetKind;
  title: string;
  description: string;
  parentAssetId: string | null;
  lifecycleState: AssetLifecycleState;
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

export interface AddAssetDraft {
  kind: AssetKind;
  title: string;
  description: string;
  parentAssetId: string | null;
  photos: SelectedPhoto[];
}

export interface SelectedPhoto {
  id: string;
  name: string;
  sizeBytes: number;
  contentType: string;
  previewUrl: string;
}

export interface WorkspaceContext {
  principal: Principal;
  tenants: Tenant[];
  inventories: Inventory[];
  selectedTenantId: string;
  selectedInventoryId: string;
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
