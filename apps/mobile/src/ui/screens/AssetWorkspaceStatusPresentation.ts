import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { AssetLifecycleActionKind } from './AssetLifecyclePresentation';

export type AssetWorkspaceStatusKind = 'success' | 'working';
export type AssetWorkspacePendingAction = 'archive' | 'restore' | 'delete' | 'edit' | 'move' | 'photos';

export type AssetWorkspaceStatus = {
  readonly kind: AssetWorkspaceStatusKind;
  readonly message: string;
};

export function assetWorkspaceWorkingStatus(action: Exclude<AssetWorkspacePendingAction, 'photos'>): AssetWorkspaceStatus {
  switch (action) {
    case 'archive':
      return { kind: 'working', message: 'Archiving asset...' };
    case 'delete':
      return { kind: 'working', message: 'Deleting asset...' };
    case 'edit':
      return { kind: 'working', message: 'Saving changes...' };
    case 'move':
      return { kind: 'working', message: 'Moving asset...' };
    case 'restore':
      return { kind: 'working', message: 'Restoring asset...' };
  }
}

export function assetWorkspaceSuccessStatus(
  action: 'edit' | 'move',
  result: { readonly message: string }
): AssetWorkspaceStatus;
export function assetWorkspaceSuccessStatus(
  action: Exclude<AssetLifecycleActionKind, 'delete'>,
  asset: AssetDetailViewModel
): AssetWorkspaceStatus;
export function assetWorkspaceSuccessStatus(
  action: 'edit' | 'move' | Exclude<AssetLifecycleActionKind, 'delete'>,
  source: { readonly message: string } | AssetDetailViewModel
): AssetWorkspaceStatus {
  const title = 'title' in source ? source.title : 'asset';
  switch (action) {
    case 'edit':
    case 'move':
      return { kind: 'success', message: 'message' in source ? source.message : `Updated ${source.title}.` };
    case 'archive':
      return { kind: 'success', message: `Archived ${title}.` };
    case 'restore':
      return { kind: 'success', message: `Restored ${title}.` };
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
