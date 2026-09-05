import { expect, it } from 'vitest';
import { ApiInventoryAssetTraversal } from './ApiInventoryAssetTraversal';
it('stops map traversal between pages even when a transport resolves after cancellation', async () => {
  const controller = new AbortController(); let reads = 0;
  const traversal = new ApiInventoryAssetTraversal({ getAsset: async () => { throw new Error('unused'); }, listAssets: async () => { reads++; if (reads === 1) controller.abort(); return { items: [], pagination: { limit: 100, hasMore: reads === 1, nextCursor: reads === 1 ? 'next' : null } }; } });
  await expect(traversal.listAllActiveInventoryAssets('tenant', 'inventory', controller.signal)).rejects.toThrow();
  expect(reads).toBe(1);
});
it('rejects a repeated traversal cursor', async () => {
  let reads = 0;
  const traversal = new ApiInventoryAssetTraversal({ getAsset: async () => { throw new Error('unused'); }, listAssets: async () => { reads++; if (reads > 3) throw new Error('too many reads'); return { items: [], pagination: { limit: 100, hasMore: true, nextCursor: 'repeat' } }; } });
  await expect(traversal.listAllActiveInventoryAssets('tenant', 'inventory')).rejects.toThrow('Invalid read continuation cursor');
  expect(reads).toBe(2);
});
