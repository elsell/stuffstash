import { t } from '../../presentation/localization';
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
    { kind: 'add_here', label: t('mobile.ContainedAssetsPresentation.addItemHere'), isPrimary: true },
    { kind: 'move_here', label: t('mobile.ContainedAssetsPresentation.moveItemsHere'), isPrimary: false }
  ];
}

export function containedAssetsEmptyState(
  asset: Pick<AssetDetailViewModel, 'canAddContainedAssets'>
): ContainedAssetsEmptyState {
  return {
    title: t('mobile.ContainedAssetsPresentation.nothingInsideYet'),
    message: asset.canAddContainedAssets
      ? t('mobile.ContainedAssetsPresentation.addAnItemHereOrMoveItemsIntoThis')
      : t('mobile.ContainedAssetsPresentation.thisSpaceIsEmpty')
  };
}

export function containedAssetsSectionHeading(
  asset: Pick<AssetDetailViewModel, 'title' | 'containedAssetsLabel'>
): ContainedAssetsSectionHeading {
  return {
    title: t('mobile.ContainedAssetsPresentation.contents'),
    summary: asset.containedAssetsLabel
  };
}

export function containedSpacesSectionHeading(
  asset: Pick<AssetDetailViewModel, 'title' | 'containedSpacesLabel'>,
  counts?: ContainedAssetsFilteredCount
): ContainedAssetsSectionHeading {
  return {
    title: t('mobile.ContainedAssetsPresentation.spacesIn', { title: String(asset.title) }),
    summary: filteredCountLabel(asset.containedSpacesLabel, counts)
  };
}

export function containedItemsSectionHeading(
  asset: Pick<AssetDetailViewModel, 'title' | 'containedItemsLabel'>,
  counts?: ContainedAssetsFilteredCount
): ContainedAssetsSectionHeading {
  return {
    title: t('mobile.ContainedAssetsPresentation.itemsIn', { title: String(asset.title) }),
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
  return t('mobile.ContainedAssetsPresentation.of', { value: String(counts.visibleCount.toString()), totalLabel: String(totalLabel) });
}

export function containedSpacesEmptyState(): ContainedAssetsEmptyState {
  return {
    title: t('mobile.ContainedAssetsPresentation.noSpacesHereYet'),
    message: t('mobile.ContainedAssetsPresentation.containersAndNestedPlacesWillAppearHere')
  };
}

export function containedItemsEmptyState(
  asset: Pick<AssetDetailViewModel, 'canAddContainedAssets'>
): ContainedAssetsEmptyState {
  return {
    title: t('mobile.ContainedAssetsPresentation.nothingHereYet'),
    message: asset.canAddContainedAssets
      ? t('mobile.ContainedAssetsPresentation.addAnItemHereOrMoveItemsIntoThis2')
      : t('mobile.ContainedAssetsPresentation.thereAreNoItemsInThisPlace')
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
