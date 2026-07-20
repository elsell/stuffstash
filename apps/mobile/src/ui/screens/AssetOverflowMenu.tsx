import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { NativeActionMenu, type NativeActionMenuGroup } from '../components/NativeActionMenu';
import {
  assetOverflowMenuActions,
  handleAssetOverflowAction,
  type AssetLifecycleActionKind
} from './AssetLifecyclePresentation';

export type AssetOverflowMenuProps = {
  readonly asset: Pick<
    AssetDetailViewModel,
    'title' | 'canArchive' | 'canRestore' | 'canDeletePermanently'
  >;
  readonly disabled?: boolean;
  readonly onCheckoutHistory: () => void;
  readonly onHistory: () => void;
  readonly onLifecycleAction: (action: AssetLifecycleActionKind) => void;
};

export function AssetOverflowMenu({
  asset,
  disabled = false,
  onCheckoutHistory,
  onHistory,
  onLifecycleAction
}: AssetOverflowMenuProps) {
  const groups = assetOverflowMenuGroups({ asset, onCheckoutHistory, onHistory, onLifecycleAction });

  return (
    <NativeActionMenu
      accessibilityLabel={`More actions for ${asset.title}`}
      disabled={disabled}
      groups={groups}
      trigger={{ kind: 'ellipsis' }}
    />
  );
}

export function assetOverflowMenuGroups({
  asset,
  onCheckoutHistory,
  onHistory,
  onLifecycleAction
}: Omit<AssetOverflowMenuProps, 'disabled'>): readonly NativeActionMenuGroup[] {
  const callbacks = { onCheckoutHistory, onHistory, onLifecycleAction };
  const actions = assetOverflowMenuActions(asset);
  return (['history', 'lifecycle', 'destructive'] as const)
    .map((section): NativeActionMenuGroup => ({
      id: section,
      items: actions
        .filter((action) => action.section === section)
        .map((action) => ({
          id: action.id,
          label: action.label,
          systemImage: action.systemImage,
          isDestructive: action.isDestructive,
          onPress: () => handleAssetOverflowAction(action.id, callbacks)
        }))
    }))
    .filter((group) => group.items.length > 0);
}
