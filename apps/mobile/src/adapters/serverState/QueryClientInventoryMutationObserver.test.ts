import { describe, expect, it } from 'vitest';
import { createMobileQueryClient, mobileQueryKeys } from './MobileQueryClient';
import { QueryClientInventoryMutationObserver } from './QueryClientInventoryMutationObserver';

describe('QueryClientInventoryMutationObserver', () => {
  it('refreshes descendant breadcrumbs after an ancestor rename', () => {
    const client = createMobileQueryClient();
    const contents = mobileQueryKeys.assetContents('scope-a', 'tenant-a', 'inventory-a', 'child', 'item:active:parent');
    client.setQueryData(contents, { parentLocationTrail: [{ id: 'parent', title: 'Old name' }] });
    new QueryClientInventoryMutationObserver(client, 'scope-a').onInventoryMutation({
      kind: 'asset_updated', tenantId: 'tenant-a', inventoryId: 'inventory-a', assetId: 'parent'
    });
    expect(client.getQueryState(contents)?.isInvalidated).toBe(true);
    client.clear();
  });

  it('refreshes recursive ancestor contents for a new child in a nested destination', () => {
    const client = createMobileQueryClient();
    const ancestor = mobileQueryKeys.assetContents('scope-a', 'tenant-a', 'inventory-a', 'garage', 'location:active:root');
    const parent = mobileQueryKeys.assetContents('scope-a', 'tenant-a', 'inventory-a', 'bin', 'container:active:garage');
    client.setQueryData(ancestor, { containedItems: [], containedSpaces: [{ id: 'bin' }] });
    client.setQueryData(parent, { parentLocationTrail: [{ id: 'garage' }], containedAssets: [] });
    new QueryClientInventoryMutationObserver(client, 'scope-a').onInventoryMutation({
      kind: 'asset_created', tenantId: 'tenant-a', inventoryId: 'inventory-a', assetId: 'new-child', relatedAssetIds: ['bin']
    });
    expect(client.getQueryState(ancestor)?.isInvalidated).toBe(true);
    expect(client.getQueryState(parent)?.isInvalidated).toBe(true);
    client.clear();
  });
  it('invalidates affected selected-inventory reads without touching another scope', async () => {
    const client = createMobileQueryClient();
    const observer = new QueryClientInventoryMutationObserver(client, 'scope-a');
    const home = mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a');
    const assets = mobileQueryKeys.inventoryAssets('scope-a', 'tenant-a', 'inventory-a');
    const other = mobileQueryKeys.home('scope-b', 'tenant-a', 'inventory-a');
    client.setQueryData(home, 'home');
    client.setQueryData(assets, 'assets');
    client.setQueryData(other, 'other');

    observer.onInventoryMutation({
      kind: 'asset_created',
      tenantId: 'tenant-a',
      inventoryId: 'inventory-a',
      assetId: 'asset-new'
    });
    await Promise.resolve();

    expect(client.getQueryState(home)?.isInvalidated).toBe(true);
    expect(client.getQueryState(assets)?.isInvalidated).toBe(true);
    expect(client.getQueryState(other)?.isInvalidated).toBe(false);
  });

  it('limits a new tag mutation to tag and Browse caches', async () => {
    const client = createMobileQueryClient();
    const observer = new QueryClientInventoryMutationObserver(client, 'scope-a');
    const tags = mobileQueryKeys.assetTags('scope-a', 'tenant-a', 'inventory-a');
    const home = mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a');
    const browse = mobileQueryKeys.browse('scope-a', 'tenant-a', 'inventory-a', {});
    client.setQueryData(tags, 'tags');
    client.setQueryData(home, 'home');
    client.setQueryData(browse, 'browse');

    observer.onInventoryMutation({
      kind: 'asset_tag_created',
      tenantId: 'tenant-a',
      inventoryId: 'inventory-a'
    });
    await Promise.resolve();

    expect(client.getQueryState(tags)?.isInvalidated).toBe(true);
    expect(client.getQueryState(browse)?.isInvalidated).toBe(true);
    expect(client.getQueryState(home)?.isInvalidated).toBe(false);
  });

  it('invalidates every checkout-bearing selected-inventory view only', async () => {
    const client = createMobileQueryClient();
    const observer = new QueryClientInventoryMutationObserver(client, 'scope-a');
    const affected = [
      mobileQueryKeys.inventoryMap('scope-a', 'tenant-a', 'inventory-a'),
      mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a'),
      mobileQueryKeys.inventoryAssets('scope-a', 'tenant-a', 'inventory-a'),
      mobileQueryKeys.locationAssets('scope-a', 'tenant-a', 'inventory-a', 'location-a'),
      mobileQueryKeys.browse('scope-a', 'tenant-a', 'inventory-a', {}),
      mobileQueryKeys.asset('scope-a', 'tenant-a', 'inventory-a', 'asset-a'),
      mobileQueryKeys.assetCore('scope-a', 'tenant-a', 'inventory-a', 'asset-a'),
      mobileQueryKeys.assetContents('scope-a', 'tenant-a', 'inventory-a', 'parent-a', 'container:active:root')
    ];
    const unaffected = [
      mobileQueryKeys.assetTags('scope-a', 'tenant-a', 'inventory-a'),
      mobileQueryKeys.assetContents('scope-a', 'tenant-a', 'inventory-a', 'asset-a', 'item:active:parent-a'),
      mobileQueryKeys.assetPhotos('scope-a', 'tenant-a', 'inventory-a', 'asset-a'),
      mobileQueryKeys.assetCore('scope-a', 'tenant-a', 'inventory-a', 'asset-b'),
      mobileQueryKeys.inventoryAssets('scope-a', 'tenant-a', 'inventory-b'),
      mobileQueryKeys.inventoryAssets('scope-b', 'tenant-a', 'inventory-a')
    ];
    for (const key of [...affected, ...unaffected]) {
      client.setQueryData(key, 'cached');
    }
    client.setQueryData(
      mobileQueryKeys.assetCore('scope-a', 'tenant-a', 'inventory-a', 'asset-a'),
      { snapshot: { asset: { parentAssetId: 'parent-a' } } }
    );

    observer.onInventoryMutation({
      kind: 'asset_checkout_changed',
      tenantId: 'tenant-a',
      inventoryId: 'inventory-a',
      assetId: 'asset-a'
    });
    await Promise.resolve();

    for (const key of affected) {
      expect(client.getQueryState(key)?.isInvalidated).toBe(true);
    }
    for (const key of unaffected) {
      expect(client.getQueryState(key)?.isInvalidated).toBe(false);
    }
  });

  it('keeps unrelated and photo-only asset regions fresh after an ordinary edit', async () => {
    const client = createMobileQueryClient();
    const observer = new QueryClientInventoryMutationObserver(client, 'scope-a');
    const core = mobileQueryKeys.assetCore('scope-a', 'tenant-a', 'inventory-a', 'asset-a');
    const contents = mobileQueryKeys.assetContents(
      'scope-a', 'tenant-a', 'inventory-a', 'asset-a', 'item:active:parent-a'
    );
    const photos = mobileQueryKeys.assetPhotos('scope-a', 'tenant-a', 'inventory-a', 'asset-a');
    const unrelated = mobileQueryKeys.assetCore('scope-a', 'tenant-a', 'inventory-a', 'asset-b');
    for (const key of [core, contents, photos, unrelated]) client.setQueryData(key, 'cached');

    observer.onInventoryMutation({
      kind: 'asset_updated',
      tenantId: 'tenant-a',
      inventoryId: 'inventory-a',
      assetId: 'asset-a'
    });
    await Promise.resolve();

    expect(client.getQueryState(core)?.isInvalidated).toBe(true);
    expect(client.getQueryState(contents)?.isInvalidated).toBe(true);
    expect(client.getQueryState(photos)?.isInvalidated).toBe(false);
    expect(client.getQueryState(unrelated)?.isInvalidated).toBe(false);
  });

  it('invalidates a cached containing asset even when the child core was never loaded', async () => {
    const client = createMobileQueryClient();
    const observer = new QueryClientInventoryMutationObserver(client, 'scope-a');
    const parentContents = mobileQueryKeys.assetContents(
      'scope-a', 'tenant-a', 'inventory-a', 'parent-a', 'container:active:root'
    );
    client.setQueryData(parentContents, {
      containedAssets: [{ id: 'asset-a' }],
      containedSpaces: [],
      containedItems: []
    });

    observer.onInventoryMutation({
      kind: 'asset_photo_changed',
      tenantId: 'tenant-a',
      inventoryId: 'inventory-a',
      assetId: 'asset-a'
    });
    await Promise.resolve();

    expect(client.getQueryState(parentContents)?.isInvalidated).toBe(true);
  });
});
