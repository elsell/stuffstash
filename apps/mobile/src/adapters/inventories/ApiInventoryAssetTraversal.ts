import { ReadPageGuard } from '../shared/ReadPageGuard';
import type {
  Asset
} from '@stuff-stash/api-client';

import type { InventoryApiClient } from './InventoryApiClient';
const inventoryAssetPageSize = 100;
export class ApiInventoryAssetTraversal {
  constructor(private readonly client: Pick<InventoryApiClient, 'getAsset' | 'listAssets'>) {}
  async loadAssetAncestors(asset: Asset, signal?: AbortSignal): Promise<readonly Asset[]> {
    const ancestors: Asset[] = [];
    const visited = new Set<string>([asset.id]);
    let parentId = asset.parentAssetId ?? undefined;
    while (parentId && !visited.has(parentId)) {
      visited.add(parentId);
      signal?.throwIfAborted();
      const parent = await this.client.getAsset(asset.tenantId, asset.inventoryId, parentId, signal);
      signal?.throwIfAborted();
      ancestors.unshift(parent);
      parentId = parent.parentAssetId ?? undefined;
    }
    return ancestors;
  }

  async loadAncestorsForAssets(
    assets: readonly Asset[],
    signal?: AbortSignal
  ): Promise<readonly Asset[]> {
    const knownAssets = new Map(assets.map((asset) => [asset.id, asset]));
    const pendingAssets = new Map<string, Promise<Asset>>();
    const loadParent = async (source: Asset, visited = new Set<string>()): Promise<void> => {
      signal?.throwIfAborted();
      if (visited.has(source.id)) {
        return;
      }
      const nextVisited = new Set(visited).add(source.id);
      const parentId = source.parentAssetId ?? undefined;
      if (!parentId || parentId === source.id) {
        return;
      }
      let parent = knownAssets.get(parentId);
      if (!parent) {
        let pending = pendingAssets.get(parentId);
        if (!pending) {
          pending = this.client.getAsset(source.tenantId, source.inventoryId, parentId, signal);
          pendingAssets.set(parentId, pending);
        }
        parent = await pending;
        knownAssets.set(parent.id, parent);
      }
      signal?.throwIfAborted();
      await loadParent(parent, nextVisited);
    };
    await Promise.all(assets.map((asset) => loadParent(asset)));
    const visibleIds = new Set(assets.map((asset) => asset.id));
    return [...knownAssets.values()].filter((asset) => !visibleIds.has(asset.id));
  }

  async listAllActiveInventoryAssets(
    tenantID: string,
    inventoryID: string,
    signal?: AbortSignal
  ): Promise<readonly Asset[]> {
    const assets: Asset[] = [];
    let cursor: string | undefined;
    const guard = new ReadPageGuard();

    do {
      signal?.throwIfAborted();
      const page = await this.client.listAssets(
        tenantID,
        inventoryID,
        inventoryAssetPageSize,
        cursor,
        'active',
        'id_asc',
        signal
      );
      signal?.throwIfAborted();
      assets.push(...page.items);
      cursor = guard.accept(page.pagination.nextCursor, page.pagination.hasMore);
    } while (cursor);

    return assets;
  }

}
