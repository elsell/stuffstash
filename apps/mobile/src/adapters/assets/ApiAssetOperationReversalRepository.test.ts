import { describe, expect, it } from 'vitest';
import { ApiAssetOperationReversalRepository } from './ApiAssetOperationReversalRepository';

describe('ApiAssetOperationReversalRepository', () => {
  it('maps a selected asset-operation reversal to the scoped Undo transport command', async () => {
    const requests: unknown[] = [];
    const impacts: unknown[] = [];
    const repository = new ApiAssetOperationReversalRepository({
      async applyUndoableOperation(tenantId, inventoryId, operationId, direction) {
        requests.push({ tenantId, inventoryId, operationId, direction });
        return { id: 'asset', parentAssetId: 'destination' } as never;
      }
    }, { onInventoryMutation: (impact) => impacts.push(impact) });

    await repository.reverseAssetOperation({ tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one' });

    expect(impacts).toEqual([{ kind: 'operation_reversed', tenantId: 'tenant-home', inventoryId: 'inventory-garage', assetId: 'asset', relatedAssetIds: ['destination'] }]);
    expect(requests).toEqual([{ tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one', direction: 'undo' }]);
  });
  it('does not invalidate data when the reversal is rejected', async () => {
    const impacts: unknown[] = [];
    const repository = new ApiAssetOperationReversalRepository({
      applyUndoableOperation: async () => { throw new Error('denied'); }
    }, { onInventoryMutation: (impact) => impacts.push(impact) });
    await expect(repository.reverseAssetOperation({ tenantId: 'tenant', inventoryId: 'inventory', operationId: 'operation' })).rejects.toThrow('denied');
    expect(impacts).toEqual([]);
  });

});
