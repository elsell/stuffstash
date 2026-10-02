import { expect, it } from 'vitest';
import { FakeInventoryApiClient } from './testing/InventoryApiClient';

class BrokenPagesClient extends FakeInventoryApiClient {
  reads = 0;
  constructor(private readonly surface: 'map' | 'tags' | 'checkout') { super(); }
  private brokenPage() {
    this.reads++;
    if (this.reads > 2) throw new Error('Pagination did not stop');
    return { items: [], pagination: { limit: 100, hasMore: true, nextCursor: 'repeat' } };
  }
  override async listAssets(...args: Parameters<FakeInventoryApiClient['listAssets']>) {
    return this.surface === 'map' ? this.brokenPage() : super.listAssets(...args);
  }
  override async listAssetTags(...args: Parameters<FakeInventoryApiClient['listAssetTags']>) {
    return this.surface === 'tags' ? this.brokenPage() : super.listAssetTags(...args);
  }
  override async listCheckedOutAssets(...args: Parameters<FakeInventoryApiClient['listCheckedOutAssets']>) {
    return this.surface === 'checkout' ? this.brokenPage() : super.listCheckedOutAssets(...args);
  }
}

it('catalogs bounded Map, tag selection and Home checkout recovery', async () => {
  const previous = process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
  process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = 'en-XA';
  try {
    const { t } = await import('../../presentation/localization');
    const recovery = t('recovery.pagedRead');
    expect(recovery).toMatch(/^\[/);
    const { ApiInventorySummaryRepository } = await import('./ApiInventorySummaryRepository');
    const { ApiInventoryAssetTraversal } = await import('./ApiInventoryAssetTraversal');
    for (const surface of ['map', 'tags', 'checkout'] as const) {
      const client = new BrokenPagesClient(surface);
      const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
      const operation = surface === 'map'
        ? new ApiInventoryAssetTraversal(client).listAllActiveInventoryAssets('tenant-home', 'inventory-home')
        : surface === 'tags' ? repository.getInventoryAssetTags() : repository.getHomeDashboardSnapshot();
      await expect(operation).rejects.toThrow(recovery);
      expect(client.reads).toBe(2);
    }
    const { ReadPageGuard } = await import('../shared/ReadPageGuard');
    expect(() => new ReadPageGuard().accept(null, true)).toThrow(recovery);
    expect(() => new ReadPageGuard('initial').accept('initial', true)).toThrow(recovery);
    const bounded = new ReadPageGuard(undefined, 2);
    expect(bounded.accept('first', true)).toBe('first');
    expect(() => bounded.accept('second', true)).toThrow(recovery);
    expect(new ReadPageGuard().accept(null, false)).toBeUndefined();
  } finally {
    if (previous === undefined) delete process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
    else process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = previous;
  }
});
