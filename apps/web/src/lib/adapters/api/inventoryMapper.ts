import type {
  Asset as ApiAsset,
  AssetKind as ApiAssetKind,
  AssetSearchResult as ApiSearchResult,
  Inventory as ApiInventory,
  Principal as ApiPrincipal,
  Tenant as ApiTenant
} from '@stuff-stash/api-client';
import type { Asset, AssetKind, Inventory, Principal, SearchResult, Tenant } from '$lib/domain/inventory';

export function mapPrincipal(principal: ApiPrincipal): Principal {
  return {
    id: principal.id,
    email: principal.email
  };
}

export function mapTenant(tenant: ApiTenant): Tenant {
  return {
    id: tenant.id,
    name: tenant.name
  };
}

export function mapInventory(inventory: ApiInventory): Inventory {
  return {
    id: inventory.id,
    tenantId: inventory.tenantId,
    name: inventory.name
  };
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
