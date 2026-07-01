import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';

export type ContainedAssetActionKind = 'add_here' | 'move_here';

export type ContainedAssetAction = {
  readonly kind: ContainedAssetActionKind;
  readonly label: string;
  readonly isPrimary: boolean;
};

export type ContainedAssetsEmptyState = {
  readonly title: string;
  readonly message: string;
};

export function containedAssetActions(
  asset: Pick<AssetDetailViewModel, 'canAddContainedAssets' | 'canContainAssets'>
): readonly ContainedAssetAction[] {
  if (!asset.canContainAssets || !asset.canAddContainedAssets) {
    return [];
  }
  return [
    { kind: 'add_here', label: 'Add item here', isPrimary: true },
    { kind: 'move_here', label: 'Move things here', isPrimary: false }
  ];
}

export function containedAssetsEmptyState(
  asset: Pick<AssetDetailViewModel, 'canAddContainedAssets'>
): ContainedAssetsEmptyState {
  return {
    title: 'Nothing inside yet',
    message: asset.canAddContainedAssets
      ? 'Add something here or move existing things into this place.'
      : 'This space is empty.'
  };
}

export function canUseContainedAssetAction({
  isActionPending,
  onPress
}: {
  readonly isActionPending: boolean;
  readonly onPress?: () => void;
}): boolean {
  return !isActionPending && onPress !== undefined;
}
