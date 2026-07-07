import type { AssetTagViewModel } from '../../application/assets/AssetViewModels';

export type AssetTagChipPresentation = {
  readonly visibleTags: readonly AssetTagViewModel[];
  readonly hiddenCount: number;
  readonly shouldRender: boolean;
};

export type AssetTagChipLayoutPresentation = {
  readonly compactRow: boolean;
  readonly shrinkVisibleChips: boolean;
};

export function assetTagChipPresentation(
  tags: readonly AssetTagViewModel[] | undefined,
  overflowLimit?: number
): AssetTagChipPresentation {
  const allTags = tags ?? [];
  const visibleLimit = overflowLimit ?? allTags.length;
  const visibleTags = allTags.slice(0, visibleLimit);
  const hiddenCount = Math.max(0, allTags.length - visibleTags.length);
  return {
    visibleTags,
    hiddenCount,
    shouldRender: visibleTags.length > 0 || hiddenCount > 0
  };
}

export function assetTagChipLayoutPresentation(compact = false): AssetTagChipLayoutPresentation {
  return {
    compactRow: compact,
    shrinkVisibleChips: compact
  };
}
