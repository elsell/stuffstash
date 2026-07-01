import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';

export function canSaveMoveAsset(
  asset: Pick<AssetDetailViewModel, 'parentAssetId'>,
  selectedParent: ParentLookupResult | null
): boolean {
  return (selectedParent?.id ?? null) !== (asset.parentAssetId ?? null);
}

export function parentFromCurrentAssetPath(asset: AssetDetailViewModel): ParentLookupResult | null {
  if (!asset.parentAssetId) {
    return null;
  }
  const parts = asset.locationTrailLabel.split('/').map((part) => part.trim()).filter(Boolean);
  const title = parts.length > 1 ? parts[parts.length - 2] : asset.locationTrailLabel;
  return {
    id: asset.parentAssetId,
    title,
    kind: 'container',
    subtitle: 'Current parent',
    selectionHint: 'Current parent',
    willPromoteToContainer: false
  };
}
