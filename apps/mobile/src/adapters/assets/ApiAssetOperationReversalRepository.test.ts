import { describe, expect, it } from 'vitest';
import { ApiAssetOperationReversalRepository } from './ApiAssetOperationReversalRepository';

describe('ApiAssetOperationReversalRepository', () => {
  it('maps a selected asset-operation reversal to the scoped Undo transport command', async () => {
    const requests: unknown[] = [];
    const repository = new ApiAssetOperationReversalRepository({
      async applyUndoableOperation(tenantId, inventoryId, operationId, direction) {
        requests.push({ tenantId, inventoryId, operationId, direction });
        return {} as never;
      }
    });

    await repository.reverseAssetOperation({ tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one' });

    expect(requests).toEqual([{ tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one', direction: 'undo' }]);
  });
});
