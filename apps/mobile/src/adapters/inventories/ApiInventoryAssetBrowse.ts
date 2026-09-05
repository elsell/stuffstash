import { assertReadActive } from '../../application/shared/ReadRequest';
import type {
  Asset,
  AssetSearchResult
} from '@stuff-stash/api-client';
import {
  AssetBrowsePage,
  AssetBrowsePageInput
} from '../../application/home/InventorySummaryRepository';
import { assetId } from '../../domain/assets/AssetSummary';
import {
  InventorySummary
} from '../../domain/inventories/InventorySummary';
import { ApiInventoryDirectory } from './ApiInventoryDirectory';

import { ApiInventoryAssetPhotos } from './ApiInventoryAssetPhotos';
import { ApiInventoryAssetTraversal } from './ApiInventoryAssetTraversal';
import type { InventoryApiClient } from './InventoryApiClient';
import { assetMatchesKind, emptyInventorySummary, filterAssetsByCheckoutState, filterAssetsByKind, searchMatchLabels } from './InventoryAssetMapping';
const maxBrowseScanPages = 5;

export class ApiInventoryAssetBrowse {
  constructor(private readonly client: Pick<InventoryApiClient, 'listAssets' | 'searchAssets' | 'listCheckedOutAssets'>, private readonly directory: ApiInventoryDirectory, private readonly traversal: ApiInventoryAssetTraversal, private readonly photos: ApiInventoryAssetPhotos) {}
  async browseAssets(input: AssetBrowsePageInput): Promise<AssetBrowsePage> {
    const selected = await this.directory.selected(input.signal);
    const inventory = emptyInventorySummary(selected.tenant, selected.inventory);
    const hasTagFilters = (input.tagIds?.length ?? 0) > 0;
    if (!hasTagFilters && input.query.trim().length === 0 && input.checkoutState === 'checked_out') {
      return await this.listCheckedOutInventoryAssetPage(inventory, input);
    }
    return input.query.trim().length > 0 || hasTagFilters
      ? await this.searchInventoryAssetPage(inventory, input)
      : await this.listInventoryAssetPage(inventory, input);
  }

  private async listInventoryAssetPage(
    inventory: InventorySummary,
    input: AssetBrowsePageInput
  ): Promise<AssetBrowsePage> {
    const desiredMatches = input.limit ?? 20;
    const selectedAssets: Asset[] = [];
    let cursor = input.cursor;
    let nextCursor: string | undefined;
    let hasMore = false;
    let scannedPages = 0;
    const seenCursors = new Set<string>(input.cursor ? [input.cursor] : []);

    do {
      assertReadActive(input.signal);
      scannedPages++;
      const pageSize = desiredMatches - selectedAssets.length;
      const page = await this.client.listAssets(
        inventory.tenantId,
        inventory.id,
        pageSize,
        cursor,
        input.lifecycleState,
        input.sort,
        input.signal
      );
      selectedAssets.push(...filterAssetsByCheckoutState(
        filterAssetsByKind(page.items, input.kind),
        input.checkoutState
      ));
      assertReadActive(input.signal);
      nextCursor = page.pagination.nextCursor ?? undefined;
      hasMore = page.pagination.hasMore;
      if (hasMore && (!nextCursor || seenCursors.has(nextCursor))) throw new Error('Invalid Browse continuation cursor.');
      if (nextCursor) seenCursors.add(nextCursor);
      cursor = nextCursor;
    } while (selectedAssets.length < desiredMatches && hasMore && scannedPages < maxBrowseScanPages);

    const knownAssets = [...selectedAssets, ...await this.traversal.loadAncestorsForAssets(selectedAssets, input.signal)];
    const assets = await Promise.all(
      selectedAssets
        .slice(0, desiredMatches)
        .map((asset) => this.photos.mapAssetWithPrimaryPhoto(inventory.name, asset, knownAssets))
    );

    return {
      assets,
      nextCursor,
      hasMore
    };
  }

  private async searchInventoryAssetPage(
    inventory: InventorySummary,
    input: AssetBrowsePageInput
  ): Promise<AssetBrowsePage> {
    const desiredMatches = input.limit ?? 20;
    const selectedResults: AssetSearchResult[] = [];
    let cursor = input.cursor;
    let nextCursor: string | undefined;
    let hasMore = false;
    let scannedPages = 0;
    const seenCursors = new Set<string>(input.cursor ? [input.cursor] : []);

    do {
      assertReadActive(input.signal);
      scannedPages++;
      const pageSize = desiredMatches - selectedResults.length;
      const page = await this.client.searchAssets(inventory.tenantId, input.query, {
        limit: pageSize,
        cursor,
        inventoryId: inventory.id,
        tagIds: input.tagIds,
        lifecycleState: input.lifecycleState,
        checkoutState: input.checkoutState,
        signal: input.signal
      });
      selectedResults.push(
        ...page.items
          .filter((item) => item.inventory.id === inventory.id)
          .filter((item) => assetMatchesKind(item.asset, input.kind))
      );
      assertReadActive(input.signal);
      nextCursor = page.pagination.nextCursor ?? undefined;
      hasMore = page.pagination.hasMore;
      if (hasMore && (!nextCursor || seenCursors.has(nextCursor))) throw new Error('Invalid Browse continuation cursor.');
      if (nextCursor) seenCursors.add(nextCursor);
      cursor = nextCursor;
    } while (selectedResults.length < desiredMatches && hasMore && scannedPages < maxBrowseScanPages);

    const pageResults = selectedResults.slice(0, desiredMatches);
    const selectedAssets = pageResults.map((item) => item.asset);
    const knownAssets = [...selectedAssets, ...await this.traversal.loadAncestorsForAssets(selectedAssets, input.signal)];
    const assets = await Promise.all(
      pageResults.map((item) =>
        this.photos.mapAssetWithPrimaryPhoto(
          inventory.name,
          item.asset,
          knownAssets
        )
      )
    );
    const searchMatches = pageResults
      .map((item) => ({
        assetId: assetId(item.asset.id),
        labels: searchMatchLabels(item.matches)
      }))
      .filter((item) => item.labels.length > 0);

    return {
      assets,
      searchMatches,
      nextCursor,
      hasMore
    };
  }

  private async listCheckedOutInventoryAssetPage(
    inventory: InventorySummary,
    input: AssetBrowsePageInput
  ): Promise<AssetBrowsePage> {
    const page = await this.client.listCheckedOutAssets(
      inventory.tenantId,
      inventory.id,
      input.limit ?? 20,
      input.cursor,
      input.signal
    );
    const selectedAssets = page.items
      .map((item) => item.asset)
      .filter((asset) => input.lifecycleState === 'all' || asset.lifecycleState === input.lifecycleState);
    const visibleAssets = filterAssetsByKind(selectedAssets, input.kind);
    const knownAssets = [...visibleAssets, ...await this.traversal.loadAncestorsForAssets(visibleAssets, input.signal)];
    const assets = await Promise.all(
      visibleAssets.map((asset) =>
        this.photos.mapAssetWithPrimaryPhoto(inventory.name, asset, knownAssets)
      )
    );

    return {
      assets,
      nextCursor: page.pagination.nextCursor ?? undefined,
      hasMore: page.pagination.hasMore
    };
  }
}
