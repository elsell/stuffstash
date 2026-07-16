import { describe, expect, it } from 'vitest';
import type { AssetOperationReversalRepository } from './AssetOperationReversalRepository';
import { RevertAssetChangeCommand } from './RevertAssetChangeCommand';

class FakeChangeReversalRepository implements AssetOperationReversalRepository {
  inputs: Parameters<AssetOperationReversalRepository['reverseAssetOperation']>[0][] = [];
  completion: Promise<void> = Promise.resolve();

  async reverseAssetOperation(input: Parameters<AssetOperationReversalRepository['reverseAssetOperation']>[0]): Promise<void> {
    this.inputs.push(input);
    await this.completion;
  }
}

describe('RevertAssetChangeCommand', () => {
  it('reverses the explicitly selected historical operation in its original scope', async () => {
    const repository = new FakeChangeReversalRepository();
    const command = new RevertAssetChangeCommand(repository);

    await command.execute({ tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one' });

    expect(repository.inputs).toEqual([{ tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one' }]);
  });

  it('rejects incomplete historical operation scope', async () => {
    const command = new RevertAssetChangeCommand(new FakeChangeReversalRepository());

    await expect(command.execute({ tenantId: 'tenant-home', inventoryId: '', operationId: 'operation-one' }))
      .rejects.toThrow('This change can’t be reverted.');
  });

  it('suppresses repeated activation while the same historical reversal is in progress', async () => {
    const repository = new FakeChangeReversalRepository();
    let complete!: () => void;
    repository.completion = new Promise<void>((resolve) => { complete = resolve; });
    const command = new RevertAssetChangeCommand(repository);
    const input = { tenantId: 'tenant-home', inventoryId: 'inventory-garage', operationId: 'operation-one' };

    const first = command.execute(input);
    await expect(command.execute(input)).resolves.toBe(false);
    expect(repository.inputs).toEqual([input]);

    complete();
    await expect(first).resolves.toBe(true);
  });
});
