import type { Asset } from '$lib/domain/inventory';

export interface AssetThumbnailLoader {
  loadAssetThumbnail(asset: Asset): Promise<Asset['photo'] | null>;
}

export interface AssetThumbnailLoaderLifecycle extends AssetThumbnailLoader {
  dispose(): void;
}

export const assetThumbnailLoaderContext = Symbol('stuffstash.asset-thumbnail-loader');
