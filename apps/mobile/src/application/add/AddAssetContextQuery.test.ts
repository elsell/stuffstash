import { describe, expect, it } from 'vitest';
import { AddAssetContextQuery } from './AddAssetContextQuery';

describe('AddAssetContextQuery', () => {
  it('loads only the selected identity, permission, and tag choices needed by Add', async () => {
    const query = new AddAssetContextQuery({
      getAddAssetContext: async () => ({
        tenantId: 'tenant-home',
        tenantName: 'Home',
        inventoryId: 'inventory-home',
        inventoryName: 'Home Inventory',
        canAdd: true,
        assetTags: [{ id: 'tag-tools', key: 'tools', displayName: 'Tools' }]
      })
    });

    await expect(query.execute()).resolves.toMatchObject({
      inventoryId: 'inventory-home',
      canAdd: true,
      assetTags: [{ id: 'tag-tools' }]
    });
  });
});
