import { describe, expect, it } from 'vitest';
import type { CreateInventoryAssetPhotoInput } from '../home/InventorySummaryRepository';
import { AddAssetPhotosCommand } from './AddAssetPhotosCommand';

class FakeAssetPhotoAddRepository {
  addedPhotos: Array<{ readonly assetId: string; readonly photo: CreateInventoryAssetPhotoInput }> = [];
  failUploads = 0;

  async addAssetPhoto(assetIdValue: string, input: CreateInventoryAssetPhotoInput): Promise<void> {
    if (this.failUploads > 0) {
      this.failUploads -= 1;
      throw new Error('Upload failed.');
    }
    this.addedPhotos.push({ assetId: assetIdValue, photo: input });
  }
}

describe('AddAssetPhotosCommand', () => {
  it('uploads selected photos through the repository', async () => {
    const repository = new FakeAssetPhotoAddRepository();
    const command = new AddAssetPhotosCommand(repository);

    await expect(command.execute({
      assetId: 'asset-water-bottle',
      photos: [{ fileName: 'one.jpg', contentType: 'image/jpeg', contentBase64: 'MQ==' }]
    })).resolves.toEqual({
      attachedCount: 1,
      failedCount: 0,
      failedPhotos: [],
      message: '1 photo added.',
      canRetry: false
    });
    expect(repository.addedPhotos).toEqual([{
      assetId: 'asset-water-bottle',
      photo: { fileName: 'one.jpg', contentType: 'image/jpeg', contentBase64: 'MQ==' }
    }]);
  });

  it('reports partial photo failures without failing the whole command', async () => {
    const repository = new FakeAssetPhotoAddRepository();
    repository.failUploads = 1;
    const command = new AddAssetPhotosCommand(repository);

    const result = await command.execute({
      assetId: 'asset-water-bottle',
      photos: [
        { fileName: 'one.jpg', contentType: 'image/jpeg', contentBase64: 'MQ==' },
        { fileName: 'two.jpg', contentType: 'image/jpeg', contentBase64: 'Mg==' }
      ]
    });
    expect(result).toEqual({
      attachedCount: 1,
      failedCount: 1,
      failedPhotos: [{ fileName: 'one.jpg', contentType: 'image/jpeg', contentBase64: 'MQ==' }],
      message: '1 of 2 photos added.',
      canRetry: true,
      failureMessage: 'Upload failed.'
    });
  });

  it('preserves the first safe failure reason when every selected photo fails', async () => {
    const repository = new FakeAssetPhotoAddRepository();
    repository.failUploads = 1;
    const command = new AddAssetPhotosCommand(repository);

    await expect(command.execute({
      assetId: 'asset-water-bottle',
      photos: [{ fileName: 'one.jpg', contentType: 'image/jpeg', contentBase64: 'MQ==' }]
    })).resolves.toMatchObject({
      attachedCount: 0,
      failedCount: 1,
      message: 'Photos could not be uploaded: Upload failed.',
      failureMessage: 'Upload failed.'
    });
  });

  it('reports per-photo upload progress in selected order', async () => {
    const repository = new FakeAssetPhotoAddRepository();
    repository.failUploads = 1;
    const command = new AddAssetPhotosCommand(repository);
    const progress: Array<{ readonly index: number; readonly fileName: string; readonly status: string }> = [];

    await command.execute({
      assetId: 'asset-water-bottle',
      photos: [
        { fileName: 'one.jpg', contentType: 'image/jpeg', contentBase64: 'MQ==' },
        { fileName: 'two.jpg', contentType: 'image/jpeg', contentBase64: 'Mg==' }
      ],
      onPhotoProgress: (event) => progress.push(event)
    });

    expect(progress).toEqual([
      { index: 0, fileName: 'one.jpg', status: 'uploading' },
      { index: 0, fileName: 'one.jpg', status: 'failed' },
      { index: 1, fileName: 'two.jpg', status: 'uploading' },
      { index: 1, fileName: 'two.jpg', status: 'attached' }
    ]);
  });

  it('requires at least one selected photo', async () => {
    const command = new AddAssetPhotosCommand(new FakeAssetPhotoAddRepository());

    await expect(command.execute({ assetId: 'asset-water-bottle', photos: [] })).rejects.toThrow(
      'Choose at least one photo.'
    );
  });
});
