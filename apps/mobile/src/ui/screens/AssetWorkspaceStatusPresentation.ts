import { t } from '../../presentation/localization';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { AssetLifecycleActionKind } from './AssetLifecyclePresentation';

export type AssetWorkspaceStatusKind = 'success' | 'working';
export type AssetWorkspacePendingAction = 'archive' | 'restore' | 'delete' | 'edit' | 'move' | 'photos' | 'checkout' | 'return';

export type AssetWorkspaceStatus = {
  readonly kind: AssetWorkspaceStatusKind;
  readonly message: string;
};

export function assetWorkspaceWorkingStatus(action: Exclude<AssetWorkspacePendingAction, 'photos'>): AssetWorkspaceStatus {
  switch (action) {
    case 'archive':
      return { kind: 'working', message: t('mobile.AssetWorkspaceStatusPresentation.archivingAsset') };
    case 'delete':
      return { kind: 'working', message: t('mobile.AssetWorkspaceStatusPresentation.deletingAsset') };
    case 'edit':
      return { kind: 'working', message: t('mobile.AssetWorkspaceStatusPresentation.savingChanges') };
    case 'move':
      return { kind: 'working', message: t('mobile.AssetWorkspaceStatusPresentation.movingAsset') };
    case 'checkout':
      return { kind: 'working', message: t('mobile.AssetWorkspaceStatusPresentation.checkingOutAsset') };
    case 'return':
      return { kind: 'working', message: t('mobile.AssetWorkspaceStatusPresentation.returningAsset') };
    case 'restore':
      return { kind: 'working', message: t('mobile.AssetWorkspaceStatusPresentation.restoringAsset') };
  }
}

export function assetWorkspaceSuccessStatus(
  action: 'edit' | 'move',
  result: { readonly message: string }
): AssetWorkspaceStatus;
export function assetWorkspaceSuccessStatus(
  action: 'checkout' | 'return',
  asset: AssetDetailViewModel
): AssetWorkspaceStatus;
export function assetWorkspaceSuccessStatus(
  action: Exclude<AssetLifecycleActionKind, 'delete'>,
  asset: AssetDetailViewModel
): AssetWorkspaceStatus;
export function assetWorkspaceSuccessStatus(
  action: 'edit' | 'move' | 'checkout' | 'return' | Exclude<AssetLifecycleActionKind, 'delete'>,
  source: { readonly message: string } | AssetDetailViewModel
): AssetWorkspaceStatus {
  const title = 'title' in source ? source.title : 'asset';
  switch (action) {
    case 'edit':
    case 'move':
      return { kind: 'success', message: 'message' in source ? source.message : t('mobile.AssetWorkspaceStatusPresentation.updated', { title: String(source.title) }) };
    case 'checkout':
      return { kind: 'success', message: 'message' in source ? source.message : t('mobile.AssetWorkspaceStatusPresentation.checkedOut', { title: String(source.title) }) };
    case 'return':
      return { kind: 'success', message: 'message' in source ? source.message : t('mobile.AssetWorkspaceStatusPresentation.returned', { title: String(source.title) }) };
    case 'archive':
      return { kind: 'success', message: t('mobile.AssetWorkspaceStatusPresentation.archived', { title: String(title) }) };
    case 'restore':
      return { kind: 'success', message: t('mobile.AssetWorkspaceStatusPresentation.restored', { title: String(title) }) };
  }
}

export function visibleAssetWorkspaceStatus(
  pendingAction: AssetWorkspacePendingAction | undefined,
  currentStatus: AssetWorkspaceStatus | undefined
): AssetWorkspaceStatus | undefined {
  if (!pendingAction) {
    return currentStatus;
  }
  if (pendingAction === 'photos') {
    return currentStatus;
  }
  return assetWorkspaceWorkingStatus(pendingAction);
}
