import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';

export type AssetLifecycleActionKind = 'archive' | 'restore' | 'delete';

export type AssetLifecycleActionRow = {
  readonly kind: AssetLifecycleActionKind;
  readonly label: string;
  readonly isDestructive: boolean;
};

export type AssetLifecycleConfirmation = {
  readonly title: string;
  readonly message: string;
  readonly confirmLabel: string;
  readonly isDestructive: boolean;
};

export function assetLifecycleActionRows(
  asset: Pick<AssetDetailViewModel, 'canArchive' | 'canRestore' | 'canDeletePermanently'>
): readonly AssetLifecycleActionRow[] {
  return [
    asset.canArchive ? {
      kind: 'archive' as const,
      label: 'Archive',
      isDestructive: false
    } : undefined,
    asset.canRestore ? {
      kind: 'restore' as const,
      label: 'Restore',
      isDestructive: false
    } : undefined,
    asset.canDeletePermanently ? {
      kind: 'delete' as const,
      label: 'Delete permanently',
      isDestructive: true
    } : undefined
  ].filter((action): action is AssetLifecycleActionRow => action !== undefined);
}

export function assetLifecycleConfirmation(
  action: AssetLifecycleActionKind,
  asset: Pick<AssetDetailViewModel, 'title' | 'photos' | 'containedAssetsLabel' | 'canContainAssets'>
): AssetLifecycleConfirmation {
  switch (action) {
    case 'archive':
      return {
        title: `Archive ${asset.title}?`,
        message: `${asset.title} will be hidden from normal inventory work. You can restore it later from archived asset views.`,
        confirmLabel: 'Archive',
        isDestructive: false
      };
    case 'restore':
      return {
        title: `Restore ${asset.title}?`,
        message: `${asset.title} will return to active inventory work.`,
        confirmLabel: 'Restore',
        isDestructive: false
      };
    case 'delete':
      return {
        title: `Delete ${asset.title} permanently?`,
        message: permanentDeleteMessage(asset),
        confirmLabel: 'Delete permanently',
        isDestructive: true
      };
  }
}

function permanentDeleteMessage(
  asset: Pick<AssetDetailViewModel, 'title' | 'photos' | 'containedAssetsLabel' | 'canContainAssets'>
): string {
  const photoCopy = asset.photos.length === 0
    ? 'No photos are attached.'
    : `${asset.photos.length.toString()} ${asset.photos.length === 1 ? 'photo' : 'photos'} will be removed with it.`;
  const contentsCopy = asset.canContainAssets
    ? ` Current contents: ${asset.containedAssetsLabel}. Deletion will not continue while active things are inside it.`
    : '';

  return `This permanently removes ${asset.title}. ${photoCopy}${contentsCopy} Audit history remains, but the asset itself cannot be restored.`;
}
