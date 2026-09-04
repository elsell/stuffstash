import { describe, expect, it } from 'vitest';
import { createMobileQueryClient, mobileQueryKeys } from './MobileQueryClient';
import { QueryClientInventoryMutationObserver } from './QueryClientInventoryMutationObserver';

describe('QueryClientInventoryMutationObserver', () => {
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
      mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a'),
      mobileQueryKeys.inventoryAssets('scope-a', 'tenant-a', 'inventory-a'),
      mobileQueryKeys.locationAssets('scope-a', 'tenant-a', 'inventory-a', 'location-a'),
      mobileQueryKeys.browse('scope-a', 'tenant-a', 'inventory-a', {}),
      mobileQueryKeys.asset('scope-a', 'tenant-a', 'inventory-a', 'asset-a')
    ];
    const unaffected = [
      mobileQueryKeys.assetTags('scope-a', 'tenant-a', 'inventory-a'),
      mobileQueryKeys.inventoryAssets('scope-a', 'tenant-a', 'inventory-b'),
      mobileQueryKeys.inventoryAssets('scope-b', 'tenant-a', 'inventory-a')
    ];
    for (const key of [...affected, ...unaffected]) {
      client.setQueryData(key, 'cached');
    }

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
});
