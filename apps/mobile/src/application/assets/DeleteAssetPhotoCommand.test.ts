import { describe, expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { DeleteAssetPhotoCommand } from './DeleteAssetPhotoCommand';

class FakeAssetPhotoDeletionRepository {
  deletedPhoto:
    | {
        readonly assetId: string;
        readonly photoId: string;
      }
    | undefined;

  async deleteAssetPhoto(assetIdValue: string, photoId: string): Promise<void> {
    this.deletedPhoto = { assetId: assetIdValue, photoId };
  }
}

describe('DeleteAssetPhotoCommand', () => {
  it('removes an asset photo through the inventory repository', async () => {
    const repository = new FakeAssetPhotoDeletionRepository();
    const command = new DeleteAssetPhotoCommand(repository);

    await expect(command.execute({
      assetId: 'asset-mug',
      photoId: 'attachment-one'
    })).resolves.toEqual({
      message: 'Photo removed.'
    });

    expect(repository.deletedPhoto).toEqual({
      assetId: assetId('asset-mug'),
      photoId: 'attachment-one'
    });
  });

  it('rejects empty photo IDs before calling the repository', async () => {
    const repository = new FakeAssetPhotoDeletionRepository();
    const command = new DeleteAssetPhotoCommand(repository);

    await expect(command.execute({ assetId: 'asset-mug', photoId: ' ' })).rejects.toThrow(
      'Photo ID is required.'
    );
    expect(repository.deletedPhoto).toBeUndefined();
  });
});
