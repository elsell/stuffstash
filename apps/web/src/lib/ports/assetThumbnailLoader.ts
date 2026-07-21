import type { Asset, AssetAttachment } from '$lib/domain/inventory';

export interface AssetThumbnailLoader {
  loadAssetThumbnail(asset: Asset): Promise<Asset['photo'] | null>;
  loadAttachmentThumbnail?(asset: Asset, attachment: AssetAttachment): Promise<AssetAttachment>;
}

export interface AssetThumbnailLoaderLifecycle extends AssetThumbnailLoader {
  dispose(): void;
}

export const assetThumbnailLoaderContext = Symbol('stuffstash.asset-thumbnail-loader');
