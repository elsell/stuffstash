import type { ReadRequest } from '../shared/ReadRequest';
import type { AssetCoreSnapshot } from './AssetCoreQuery';
import { toAssetPhotoViewModels, type AssetPhotoViewModel } from './AssetViewModels';

export type AssetPhotosRepository = {
  getAssetPhotos(core: AssetCoreSnapshot, request?: ReadRequest): Promise<AssetCoreSnapshot['asset']['photos']>;
};

export class AssetPhotosQuery {
  constructor(private readonly photos: AssetPhotosRepository) {}

  async execute(core: AssetCoreSnapshot, request: ReadRequest = {}): Promise<readonly AssetPhotoViewModel[]> {
    return toAssetPhotoViewModels(await this.photos.getAssetPhotos(core, request));
  }
}
