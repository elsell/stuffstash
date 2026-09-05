import type { StuffStashClient } from '@stuff-stash/api-client';

export type InventoryApiClient = Pick<
  StuffStashClient,
  | 'listMyTenants'
  | 'listInventories'
  | 'listAssets'
  | 'getAsset'
  | 'listAssetTags'
  | 'createAsset'
  | 'createAssetTag'
  | 'updateAsset'
  | 'checkoutAsset'
  | 'returnAsset'
  | 'updateReturnedCheckoutDetails'
  | 'applyUndoableOperation'
  | 'archiveAsset'
  | 'restoreAsset'
  | 'deleteAsset'
  | 'createAssetAttachment'
  | 'initiateAssetAttachmentDirectUpload'
  | 'completeAssetAttachmentDirectUpload'
  | 'deleteAssetAttachment'
  | 'searchAssets'
  | 'listCheckedOutAssets'
  | 'listAssetAttachments'
  | 'assetAttachmentThumbnailReference'
>;

