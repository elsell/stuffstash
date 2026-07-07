import { describe, expect, it } from 'vitest';
import type { UpdateInventoryAssetInput } from '../home/InventorySummaryRepository';
import type { AssetSummary } from '../../domain/assets/AssetSummary';
import { UpdateAssetCommand } from './UpdateAssetCommand';

class FakeAssetUpdateRepository {
  updateInput: UpdateInventoryAssetInput | undefined;
  createdTags: Array<{ readonly displayName: string; readonly color?: string }> = [];

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

  async createAssetTag(input: { readonly displayName: string; readonly color?: string }) {
    this.createdTags.push(input);
    return {
      id: `tag-created-${this.createdTags.length.toString()}`,
      key: input.displayName.toLowerCase().replaceAll(' ', '-'),
      displayName: input.displayName,
      color: input.color
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
      tagIds: [' tag-camping ', '', 'tag-kitchen'],
      newTags: [{ displayName: ' Travel ', color: ' #2f80ed ' }]
    })).resolves.toEqual({
      id: 'asset-water-bottle',
      title: 'Water bottle',
      message: 'Updated Water bottle.'
    });

    expect(repository.updateInput).toEqual({
      assetId: 'asset-water-bottle',
      title: 'Water bottle',
      description: 'Metal bottle',
      tagIds: ['tag-camping', 'tag-kitchen', 'tag-created-1']
    });
    expect(repository.createdTags).toEqual([{ displayName: 'Travel', color: '#2f80ed' }]);
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

  it('preserves an explicit empty tag list so edits can clear all tags', async () => {
    const repository = new FakeAssetUpdateRepository();
    const command = new UpdateAssetCommand(repository);

    await command.execute({
      assetId: 'asset-water-bottle',
      title: 'Water bottle',
      description: '',
      tagIds: []
    });

    expect(repository.updateInput?.tagIds).toEqual([]);
  });
});
