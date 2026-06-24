import type {
  AccessSummary as ApiAccessSummary,
  Asset as ApiAsset,
  AssetKind as ApiAssetKind,
  AssetSearchResult as ApiSearchResult,
  Inventory as ApiInventory,
  Principal as ApiPrincipal,
  Tenant as ApiTenant
} from '@stuff-stash/api-client';
import {
  canEditInventory,
  type AccessSummary,
  type Asset,
  type AssetKind,
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
    updatedAt: undefined
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

function mapAssetKind(kind: ApiAssetKind): AssetKind {
  return kind;
}

function mapAccess(access: ApiAccessSummary): AccessSummary {
  return {
    relationship: access.relationship,
    permissions: access.permissions ?? []
  };
}
