import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import type { AssetContainedItemViewModel } from '../../application/assets/AssetViewModels';

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

export type ContainedAssetsSectionHeading = {
  readonly title: string;
  readonly summary: string;
};

export type ContainedAssetRowViewModel = {
  readonly id: string;
  readonly title: string;
  readonly eyebrowLabel: string;
  readonly supportingLabel: string;
  readonly imagePlaceholderLabel: string;
  readonly photo?: AssetCardViewModel['photo'];
};

export function containedAssetActions(
  asset: Pick<AssetDetailViewModel, 'canAddContainedAssets' | 'canContainAssets'>
): readonly ContainedAssetAction[] {
  if (!asset.canContainAssets || !asset.canAddContainedAssets) {
    return [];
  }
  return [
    { kind: 'add_here', label: 'Add item here', isPrimary: true },
    { kind: 'move_here', label: 'Move items here', isPrimary: false }
  ];
}

export function containedAssetsEmptyState(
  asset: Pick<AssetDetailViewModel, 'canAddContainedAssets'>
): ContainedAssetsEmptyState {
  return {
    title: 'Nothing inside yet',
    message: asset.canAddContainedAssets
      ? 'Add an item here or move items into this space.'
      : 'This space is empty.'
  };
}

export function containedAssetsSectionHeading(
  asset: Pick<AssetDetailViewModel, 'title' | 'containedAssetsLabel'>
): ContainedAssetsSectionHeading {
  return {
    title: `Inside ${asset.title}`,
    summary: asset.containedAssetsLabel
  };
}

export function containedSpacesSectionHeading(
  asset: Pick<AssetDetailViewModel, 'title' | 'containedSpacesLabel'>,
  counts?: ContainedAssetsFilteredCount
): ContainedAssetsSectionHeading {
  return {
    title: `Spaces in ${asset.title}`,
    summary: filteredCountLabel(asset.containedSpacesLabel, counts)
  };
}

export function containedItemsSectionHeading(
  asset: Pick<AssetDetailViewModel, 'title' | 'containedItemsLabel'>,
  counts?: ContainedAssetsFilteredCount
): ContainedAssetsSectionHeading {
  return {
    title: `Items in ${asset.title}`,
    summary: filteredCountLabel(asset.containedItemsLabel, counts)
  };
}

export type ContainedAssetsFilteredCount = {
  readonly visibleCount: number;
  readonly totalCount: number;
};

function filteredCountLabel(
  totalLabel: string,
  counts: ContainedAssetsFilteredCount | undefined
): string {
  if (!counts || counts.visibleCount === counts.totalCount) {
    return totalLabel;
  }
  return `${counts.visibleCount.toString()} of ${totalLabel}`;
}

export function containedSpacesEmptyState(): ContainedAssetsEmptyState {
  return {
    title: 'No spaces here yet',
    message: 'Containers and nested places will appear here.'
  };
}

export function containedItemsEmptyState(
  asset: Pick<AssetDetailViewModel, 'canAddContainedAssets'>
): ContainedAssetsEmptyState {
  return {
    title: 'Nothing here yet',
    message: asset.canAddContainedAssets
      ? 'Add an item here or move items into this place.'
      : 'There are no items in this place.'
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

export function containedAssetRows(
  assets: readonly (AssetCardViewModel | AssetContainedItemViewModel)[]
): readonly ContainedAssetRowViewModel[] {
  return assets.map((asset) => ({
    id: asset.id,
    title: asset.title,
    eyebrowLabel: [asset.kindLabel, asset.customTypeLabel]
      .filter((value): value is string => value !== undefined && value.trim().length > 0)
      .join(' · '),
    supportingLabel: containedAssetSupportingLabel(asset),
    imagePlaceholderLabel: asset.imagePlaceholderLabel,
    photo: asset.photo
  }));
}

function containedAssetSupportingLabel(asset: AssetCardViewModel): string {
  const relativePathLabel = (asset as Partial<AssetContainedItemViewModel>).relativePathLabel;
  const context = [asset.checkedOutLabel, relativePathLabel]
    .filter((value): value is string => value !== undefined && value.trim().length > 0)
    .join(' · ');
  if (context.length > 0) {
    return context;
  }
  return asset.description.trim() || asset.updatedAtLabel;
}
