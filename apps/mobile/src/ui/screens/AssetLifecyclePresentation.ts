import { catalogRecoveryMessage } from '../../application/shared/CatalogRecoveryError';
import { t } from '../../presentation/localization';
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

export type AssetLifecycleFailurePresentation = {
  readonly title: string;
  readonly message: string;
};

export type AssetDetailLoadErrorPresentation = {
  readonly title: string;
  readonly message: string;
  readonly canRetry: boolean;
};

export function assetDetailLoadErrorPresentation(error: unknown): AssetDetailLoadErrorPresentation {
  if (isUnavailableAssetError(error)) {
    return {
      title: t('mobile.AssetLifecyclePresentation.assetUnavailable'),
      message: t('mobile.AssetLifecyclePresentation.thisAssetIsNotAvailableInYourCurrentInventory'),
      canRetry: false
    };
  }

  return {
    title: t('mobile.AssetLifecyclePresentation.couldNotLoadAsset'),
    message: t('mobile.AssetLifecyclePresentation.checkYourConnectionAndTryAgain'),
    canRetry: true
  };
}

export type AssetOverflowMenuActionId = 'checkout-history' | 'history' | AssetLifecycleActionKind;
export type AssetOverflowSystemImage =
  | 'clock.arrow.circlepath'
  | 'clock'
  | 'archivebox'
  | 'arrow.uturn.backward'
  | 'trash';

export type AssetOverflowMenuAction = {
  readonly id: AssetOverflowMenuActionId;
  readonly label: string;
  readonly section: 'history' | 'lifecycle' | 'destructive';
  readonly systemImage: AssetOverflowSystemImage;
  readonly isDestructive: boolean;
};

export type AssetOverflowMenuCallbacks = {
  readonly onCheckoutHistory: () => void;
  readonly onHistory: () => void;
  readonly onLifecycleAction: (action: AssetLifecycleActionKind) => void;
};

export function assetLifecycleActionRows(
  asset: Pick<AssetDetailViewModel, 'canArchive' | 'canRestore' | 'canDeletePermanently'>
): readonly AssetLifecycleActionRow[] {
  return [
    asset.canArchive ? {
      kind: 'archive' as const,
      label: t('mobile.AssetLifecyclePresentation.archive'),
      isDestructive: false
    } : undefined,
    asset.canRestore ? {
      kind: 'restore' as const,
      label: t('mobile.AssetLifecyclePresentation.restore'),
      isDestructive: false
    } : undefined,
    asset.canDeletePermanently ? {
      kind: 'delete' as const,
      label: t('mobile.AssetLifecyclePresentation.deletePermanently'),
      isDestructive: true
    } : undefined
  ].filter((action): action is AssetLifecycleActionRow => action !== undefined);
}

export function assetLifecycleConfirmation(
  action: Exclude<AssetLifecycleActionKind, 'restore'>,
  asset: Pick<AssetDetailViewModel, 'title' | 'photos' | 'containedAssetsLabel' | 'canContainAssets'>
): AssetLifecycleConfirmation {
  switch (action) {
    case 'archive':
      return {
        title: t('mobile.AssetLifecyclePresentation.archive2', { title: String(asset.title) }),
        message: t('mobile.AssetLifecyclePresentation.willBeHiddenFromNormalInventoryWorkYouCan', { title: String(asset.title) }),
        confirmLabel: t('mobile.AssetLifecyclePresentation.archive'),
        isDestructive: false
      };
    case 'delete':
      return {
        title: t('mobile.AssetLifecyclePresentation.deletePermanently2', { title: String(asset.title) }),
        message: permanentDeleteMessage(asset),
        confirmLabel: t('mobile.AssetLifecyclePresentation.deletePermanently'),
        isDestructive: true
      };
  }
}

export function assetOverflowMenuActions(
  asset: Pick<AssetDetailViewModel, 'canArchive' | 'canRestore' | 'canDeletePermanently'>
): readonly AssetOverflowMenuAction[] {
  return [
    {
      id: 'checkout-history',
      label: t('mobile.AssetLifecyclePresentation.checkoutHistory'),
      section: 'history',
      systemImage: 'clock.arrow.circlepath',
      isDestructive: false
    },
    {
      id: 'history',
      label: t('mobile.AssetLifecyclePresentation.history'),
      section: 'history',
      systemImage: 'clock',
      isDestructive: false
    },
    ...assetLifecycleActionRows(asset).map((action): AssetOverflowMenuAction => ({
      id: action.kind,
      label: action.label,
      section: action.isDestructive ? 'destructive' : 'lifecycle',
      systemImage: action.kind === 'archive'
        ? 'archivebox'
        : action.kind === 'restore'
          ? 'arrow.uturn.backward'
          : 'trash',
      isDestructive: action.isDestructive
    }))
  ];
}

export function handleAssetOverflowAction(
  action: AssetOverflowMenuActionId,
  callbacks: AssetOverflowMenuCallbacks
): void {
  if (action === 'checkout-history') {
    callbacks.onCheckoutHistory();
    return;
  }
  if (action === 'history') {
    callbacks.onHistory();
    return;
  }
  callbacks.onLifecycleAction(action);
}

export function assetLifecycleFailurePresentation(
  action: AssetLifecycleActionKind,
  asset: Pick<AssetDetailViewModel, 'title' | 'canContainAssets'>,
  cause: unknown
): AssetLifecycleFailurePresentation {
  const diagnostic = cause instanceof Error ? cause.message : typeof cause === 'string' ? cause : '';
  const validationKind = lifecycleValidationKind(diagnostic);
  const fallback = catalogRecoveryMessage(cause, t('mobile.AssetDetailRouteScreen.lifecycleActionFailed'));
  switch (action) {
    case 'archive':
      return {
        title: t('mobile.AssetLifecyclePresentation.couldNotArchive', { title: String(asset.title) }),
        message: validationKind === 'active_children' && asset.canContainAssets
          ? t('mobile.AssetLifecyclePresentation.moveOrArchiveActiveThingsInsideThisAssetThen')
          : fallback
      };
    case 'restore':
      return {
        title: t('mobile.AssetLifecyclePresentation.couldNotRestore', { title: String(asset.title) }),
        message: validationKind === 'archived_parent'
          ? t('mobile.AssetLifecyclePresentation.checkThatItsParentIsActiveThenTryAgain')
          : fallback
      };
    case 'delete':
      return {
        title: t('mobile.AssetLifecyclePresentation.couldNotPermanentlyDelete', { title: String(asset.title) }),
        message: validationKind === 'active_children' && asset.canContainAssets
          ? t('mobile.AssetLifecyclePresentation.permanentDeleteWillNotContinueWhileActiveThingsAre')
          : fallback
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
    ? t('photos.noneAttached')
    : t('photos.deleteCount', { count: asset.photos.length });
  const contentsCopy = asset.canContainAssets
    ? t('assets.deleteContents', { contents: asset.containedAssetsLabel })
    : '';

  return t('assets.deleteWarning', { title: asset.title, photos: photoCopy, contents: contentsCopy });
}
