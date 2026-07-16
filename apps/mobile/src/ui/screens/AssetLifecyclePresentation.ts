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

export type AssetLifecycleOverflowPresentation = {
  readonly title: string;
  readonly message: string;
};

export type AssetLifecycleFailurePresentation = {
  readonly title: string;
  readonly message: string;
};

export type AssetDetailLoadErrorPresentation = {
  readonly title: string;
  readonly message: string;
  readonly canRetry: boolean;
};

export function assetDetailOverflowControlState(isActionPending: boolean): {
  readonly disabled: boolean;
  readonly accessibilityState: { readonly disabled: boolean };
} {
  return {
    disabled: isActionPending,
    accessibilityState: { disabled: isActionPending }
  };
}

export function assetDetailLoadErrorPresentation(error: unknown): AssetDetailLoadErrorPresentation {
  if (isUnavailableAssetError(error)) {
    return {
      title: 'Asset unavailable',
      message: 'This asset is not available in your current inventory.',
      canRetry: false
    };
  }

  return {
    title: 'Could not load asset',
    message: 'Check your connection and try again.',
    canRetry: true
  };
}

export type AssetLifecycleOverflowMenu = AssetLifecycleOverflowPresentation & {
  readonly options: readonly string[];
  readonly actionRows: readonly AssetLifecycleActionRow[];
  readonly checkoutHistoryIndex: number;
  readonly auditIndex: number;
  readonly lifecycleActionIndexes: readonly number[];
  readonly cancelIndex: number;
  readonly destructiveIndex?: number;
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

export function assetLifecycleOverflowPresentation(
  asset: Pick<AssetDetailViewModel, 'title' | 'lifecycleLabel'>
): AssetLifecycleOverflowPresentation {
  return {
    title: `${asset.title} actions`,
    message: 'History, lifecycle, and other actions for this asset.'
  };
}

export function assetLifecycleOverflowMenu(
  asset: Pick<AssetDetailViewModel, 'title' | 'lifecycleLabel' | 'canArchive' | 'canRestore' | 'canDeletePermanently'>
): AssetLifecycleOverflowMenu {
  const overflow = assetLifecycleOverflowPresentation(asset);
  const actionRows = assetLifecycleActionRows(asset);
  const checkoutHistoryIndex = 0;
  const auditIndex = 1;
  const lifecycleActionIndexes = actionRows.map((_, index) => index + 2);
  const cancelIndex = actionRows.length + 2;
  const destructiveActionIndex = actionRows.findIndex((action) => action.isDestructive);
  return {
    ...overflow,
    actionRows,
    options: ['Checkout history', 'History', ...actionRows.map((action) => action.label), 'Cancel'],
    checkoutHistoryIndex,
    auditIndex,
    lifecycleActionIndexes,
    cancelIndex,
    destructiveIndex: destructiveActionIndex >= 0 ? lifecycleActionIndexes[destructiveActionIndex] : undefined
  };
}

export function assetLifecycleFailurePresentation(
  action: AssetLifecycleActionKind,
  asset: Pick<AssetDetailViewModel, 'title' | 'canContainAssets'>,
  cause: string
): AssetLifecycleFailurePresentation {
  const validationKind = lifecycleValidationKind(cause);
  switch (action) {
    case 'archive':
      return {
        title: `Could not archive ${asset.title}`,
        message: validationKind === 'active_children' && asset.canContainAssets
          ? `${cause} Move or archive active things inside this asset, then try again.`
          : cause
      };
    case 'restore':
      return {
        title: `Could not restore ${asset.title}`,
        message: validationKind === 'archived_parent'
          ? `${cause} Check that its parent is active, then try again.`
          : cause
      };
    case 'delete':
      return {
        title: `Could not permanently delete ${asset.title}`,
        message: validationKind === 'active_children' && asset.canContainAssets
          ? `${cause} Permanent delete will not continue while active things are inside it.`
          : cause
      };
  }
}

function lifecycleValidationKind(cause: string): 'active_children' | 'archived_parent' | 'generic' {
  const normalized = cause.toLowerCase();
  if (normalized.includes('active child') || normalized.includes('active children') || normalized.includes('active things')) {
    return 'active_children';
  }
  if (normalized.includes('parent') && normalized.includes('archived')) {
    return 'archived_parent';
  }
  return 'generic';
}

function isUnavailableAssetError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null) {
    const status = 'status' in error ? error.status : 'statusCode' in error ? error.statusCode : undefined;
    if (status === 401 || status === 403 || status === 404) {
      return true;
    }
  }

  if (!(error instanceof Error)) {
    return false;
  }

  const message = error.message.toLowerCase();
  return message.includes('not found')
    || message.includes('forbidden')
    || message.includes('access denied')
    || message.includes('not available');
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
