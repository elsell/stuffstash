import { describe, expect, it } from 'vitest';
import type { AssetOperationReversalRepository } from './AssetOperationReversalRepository';
import { UndoAssetEditCommand } from './UndoAssetEditCommand';

class FakeUndoRepository implements AssetOperationReversalRepository {
  inputs: Parameters<AssetOperationReversalRepository['reverseAssetOperation']>[0][] = [];
  async reverseAssetOperation(input: Parameters<AssetOperationReversalRepository['reverseAssetOperation']>[0]): Promise<void> { this.inputs.push(input); }
}

describe('UndoAssetEditCommand', () => {
  it('preserves the explicit tenant and non-default inventory scope', async () => {
    const repository = new FakeUndoRepository();
    const command = new UndoAssetEditCommand(repository);
    await command.execute({ tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one' });
    expect(repository.inputs).toEqual([{ tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one' }]);
  });
});
