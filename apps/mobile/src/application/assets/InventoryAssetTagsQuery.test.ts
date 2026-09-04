import { describe, expect, it } from 'vitest';
import { InventoryAssetTagsQuery } from './InventoryAssetTagsQuery';

describe('InventoryAssetTagsQuery', () => {
  it('maps focused inventory tags into mobile edit options and forwards cancellation', async () => {
    const controller = new AbortController();
    let receivedSignal: AbortSignal | undefined;
    const query = new InventoryAssetTagsQuery({
      getInventoryAssetTags: async (request) => {
        receivedSignal = request?.signal;
        return [{
          id: 'tag-workshop',
          key: 'workshop',
          displayName: 'Workshop',
          color: '#2F80ED'
        }];
      }
    });

    await expect(query.execute({ signal: controller.signal })).resolves.toEqual([{
      id: 'tag-workshop',
      key: 'workshop',
      label: 'Workshop',
      color: '#2F80ED'
    }]);
    expect(receivedSignal).toBe(controller.signal);
  });
});
