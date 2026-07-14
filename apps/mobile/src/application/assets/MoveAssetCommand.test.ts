import { describe, expect, it } from 'vitest';
import type { UpdateInventoryAssetInput } from '../home/InventorySummaryRepository';
import type { AssetSummary } from '../../domain/assets/AssetSummary';
import { MoveAssetCommand } from './MoveAssetCommand';

class FakeAssetUpdateRepository {
  updateInput: UpdateInventoryAssetInput | undefined;

  async updateAsset(input: UpdateInventoryAssetInput): Promise<AssetSummary> {
    this.updateInput = input;
    return {
      id: input.assetId,
      title: 'Water bottle',
      kind: 'item',
      lifecycleState: 'active',
      parentAssetId: input.parentAssetId ?? undefined,
      description: '',
      locationLabel: input.parentAssetId ? 'Kitchen' : 'Inventory root',
      locationTrail: input.parentAssetId ? ['Home', 'Kitchen', 'Water bottle'] : ['Home', 'Water bottle'],
      parentLocationTrail: [],
      updatedAtLabel: 'Updated now',
      hasPhoto: false
    };
  }
}

describe('MoveAssetCommand', () => {
  it('moves an asset to a selected parent', async () => {
    const repository = new FakeAssetUpdateRepository();
    const command = new MoveAssetCommand(repository);

    await expect(command.execute({
      assetId: 'asset-water-bottle',
      parentAssetId: 'asset-kitchen'
    })).resolves.toEqual({
      id: 'asset-water-bottle',
      title: 'Water bottle',
      message: 'Moved Water bottle.'
    });

    expect(repository.updateInput).toEqual({
      assetId: 'asset-water-bottle',
      parentAssetId: 'asset-kitchen'
    });
  });

  it('moves an asset to no parent', async () => {
    const repository = new FakeAssetUpdateRepository();
    const command = new MoveAssetCommand(repository);

    await expect(command.execute({ assetId: 'asset-water-bottle' })).resolves.toMatchObject({
      message: 'Moved Water bottle to No parent.'
    });
    expect(repository.updateInput).toEqual({
      assetId: 'asset-water-bottle',
      parentAssetId: null
    });
  });

  it('rejects self parent moves before calling the repository', async () => {
    const repository = new FakeAssetUpdateRepository();
    const command = new MoveAssetCommand(repository);

    await expect(command.execute({
      assetId: 'asset-water-bottle',
      parentAssetId: 'asset-water-bottle'
    })).rejects.toThrow('An asset cannot be moved into itself.');
    expect(repository.updateInput).toBeUndefined();
  });
});
