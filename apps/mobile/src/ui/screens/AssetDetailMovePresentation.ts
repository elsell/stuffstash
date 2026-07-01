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
  return {
    id: asset.parentAssetId,
    title: asset.parentLocationTrailLabel,
    kind: 'container',
    subtitle: 'Current parent',
    pathLabel: asset.parentLocationTrailLabel,
    selectionHint: 'Current parent',
    willPromoteToContainer: false
  };
}

export type MovePlacementPreview = {
  readonly currentLocationLabel: string;
  readonly proposedLocationLabel: string;
  readonly hasChanged: boolean;
};

export function movePlacementPreview(
  asset: Pick<AssetDetailViewModel, 'parentAssetId' | 'parentLocationTrailLabel'>,
  selectedParent: ParentLookupResult | null
): MovePlacementPreview {
  const currentLocationLabel = asset.parentLocationTrailLabel;
  const proposedLocationLabel = selectedParent
    ? selectedParent.pathLabel
    : 'Inventory root';

  return {
    currentLocationLabel,
    proposedLocationLabel,
    hasChanged: canSaveMoveAsset(asset, selectedParent)
  };
}
