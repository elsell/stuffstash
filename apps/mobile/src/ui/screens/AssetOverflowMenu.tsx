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
  readonly onMove?: () => void;
  readonly onAddPhotos?: () => void;
  readonly onCheckout?: () => void;
  readonly photosDisabled?: boolean;
  readonly onCheckoutHistory: () => void;
  readonly onHistory: () => void;
  readonly onLifecycleAction: (action: AssetLifecycleActionKind) => void;
};

export function AssetOverflowMenu({
  asset,
  disabled = false,
  onCheckoutHistory,
  onHistory,
  onLifecycleAction, onMove, onAddPhotos, onCheckout, photosDisabled
}: AssetOverflowMenuProps) {
  const groups = assetOverflowMenuGroups({ asset, onCheckoutHistory, onHistory, onLifecycleAction, onMove, onAddPhotos, onCheckout, photosDisabled });

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
  onLifecycleAction, onMove, onAddPhotos, onCheckout, photosDisabled
}: Omit<AssetOverflowMenuProps, 'disabled'>): readonly NativeActionMenuGroup[] {
  const callbacks = { onCheckoutHistory, onHistory, onLifecycleAction };
  const actions = assetOverflowMenuActions(asset);
  const commands: NativeActionMenuGroup = { id: 'commands', items: [
    ...(onAddPhotos ? [{ id: 'add_photos', label: 'Add photos', systemImage: 'photo.badge.plus', disabled: photosDisabled, onPress: () => { if (!photosDisabled) onAddPhotos(); } }] : []),
    ...(onMove ? [{ id: 'move', label: 'Move', systemImage: 'folder', onPress: onMove }] : []),
    ...(onCheckout ? [{ id: 'checkout', label: 'Check out', systemImage: 'arrow.up.right', onPress: onCheckout }] : [])
  ] };
  return [commands, ...(['history', 'lifecycle', 'destructive'] as const)
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
    ].filter((group) => group.items.length > 0);
}
