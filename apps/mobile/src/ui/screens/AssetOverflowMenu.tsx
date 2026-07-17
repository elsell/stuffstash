import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { NativeActionMenu, type NativeActionMenuGroup } from '../components/NativeActionMenu';
import {
  assetOverflowMenuActions,
  handleAssetOverflowAction,
  type AssetLifecycleActionKind
} from './AssetLifecyclePresentation';

type AssetOverflowMenuProps = {
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
  const callbacks = { onCheckoutHistory, onHistory, onLifecycleAction };
  const actions = assetOverflowMenuActions(asset);
  const groups = (['history', 'lifecycle', 'destructive'] as const)
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

  return (
    <NativeActionMenu
      accessibilityLabel={`More actions for ${asset.title}`}
      disabled={disabled}
      groups={groups}
      trigger={{ kind: 'ellipsis' }}
    />
  );
}
