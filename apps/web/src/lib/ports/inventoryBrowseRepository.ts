import type {
  Asset, BrowseScope, BrowseSort, SearchCheckoutFilter, SearchLifecycleFilter, SearchMode, SearchResult
} from '$lib/domain/inventory';

export interface BrowseAssetsRequest {
  tenantId: string;
  inventoryId: string;
  query: string;
  tagIds: string[];
  lifecycleState: SearchLifecycleFilter;
  checkoutState: SearchCheckoutFilter;
  scope: BrowseScope;
  sort: BrowseSort;
  mode: SearchMode;
  limit: number;
  cursor?: string;
}

export interface BrowseAssetsPage {
  assets: Asset[];
  searchResults: SearchResult[];
  nextCursor: string | null;
  hasMore: boolean;
}

export interface InventoryBrowseRepository {
  browseAssets(request: BrowseAssetsRequest): Promise<BrowseAssetsPage>;
  hasAnyAssets(tenantId: string, inventoryId: string): Promise<boolean>;
  loadActiveContainmentMap(tenantId: string, inventoryId: string): Promise<Asset[]>;
}
