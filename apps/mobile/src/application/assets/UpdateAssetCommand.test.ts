import { describe, expect, it } from 'vitest';
import type { UpdateInventoryAssetInput } from '../home/InventorySummaryRepository';
import type { AssetSummary } from '../../domain/assets/AssetSummary';
import { UpdateAssetCommand } from './UpdateAssetCommand';

class FakeAssetUpdateRepository {
  updateInput: UpdateInventoryAssetInput | undefined;

  async updateAsset(input: UpdateInventoryAssetInput): Promise<AssetSummary> {
    this.updateInput = input;
    return {
      id: input.assetId,
      title: input.title ?? 'Water bottle',
      kind: 'item',
      lifecycleState: 'active',
      description: input.description ?? '',
      locationLabel: 'Kitchen',
      locationTrail: ['Home', 'Kitchen', input.title ?? 'Water bottle'],
      updatedAtLabel: 'Updated now',
      hasPhoto: false
    };
  }
}

describe('UpdateAssetCommand', () => {
  it('trims editable fields and updates through the repository', async () => {
    const repository = new FakeAssetUpdateRepository();
    const command = new UpdateAssetCommand(repository);

    await expect(command.execute({
      assetId: 'asset-water-bottle',
      title: '  Water bottle  ',
      description: '  Metal bottle  ',
      tagIds: [' tag-camping ', '', 'tag-kitchen']
    })).resolves.toEqual({
      id: 'asset-water-bottle',
      title: 'Water bottle',
      message: 'Updated Water bottle.'
    });

    expect(repository.updateInput).toEqual({
      assetId: 'asset-water-bottle',
      title: 'Water bottle',
      description: 'Metal bottle',
      tagIds: ['tag-camping', 'tag-kitchen']
    });
  });

  it('rejects blank names before updating', async () => {
    const repository = new FakeAssetUpdateRepository();
    const command = new UpdateAssetCommand(repository);

    await expect(command.execute({
      assetId: 'asset-water-bottle',
      title: '   ',
      description: ''
    })).rejects.toThrow('Name is required.');
    expect(repository.updateInput).toBeUndefined();
  });
});
