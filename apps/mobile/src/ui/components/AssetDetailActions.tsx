import { StyleSheet, View } from 'react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { NativeCommandButton } from './NativeCommandButton';
import { assetDetailAvailabilityAction } from './AssetDetailPresentation';
import { spacing } from '../theme/tokens';

type ActionAsset = Pick<AssetDetailViewModel, 'kind' | 'canEdit' | 'canMove' | 'canAddPhotos' | 'canCheckout' | 'canReturn'>;
export function AssetDetailActions({ asset, isActionPending, isPhotosLoading = false, showEditAction = false,
  onAddPhotos, onMove, onCheckout, onReturn, onEdit }: {
  readonly asset: ActionAsset;
  readonly isActionPending: boolean;
  readonly isPhotosLoading?: boolean;
  readonly showEditAction?: boolean;
  readonly onAddPhotos?: () => void;
  readonly onMove?: () => void;
  readonly onCheckout?: () => void;
  readonly onReturn?: () => void;
  readonly onEdit?: () => void;
}) {
  const availability = asset.kind === 'location' ? undefined : assetDetailAvailabilityAction(asset);
  const actions: readonly { label: string; handler?: () => void; pending?: boolean }[] = [
    ...(showEditAction && asset.canEdit ? [{ label: 'Edit', handler: onEdit }] : []),
    ...(asset.canAddPhotos ? [{ label: 'Add photos', handler: onAddPhotos, pending: isPhotosLoading }] : []),
    ...(asset.canMove ? [{ label: asset.kind === 'location' ? 'Move place' : 'Move', handler: onMove }] : []),
    ...(availability ? [{ label: availability.label, handler: availability.id === 'return' ? onReturn : onCheckout }] : [])
  ];
  if (!actions.length) return null;
  return <View accessibilityLabel="Asset actions" style={styles.actions}>
    {actions.map(action => <View key={action.label} style={styles.action}>
      <NativeCommandButton label={action.label} disabled={isActionPending || action.pending || !action.handler}
        onPress={() => { if (!isActionPending && !action.pending) action.handler?.(); }} />
    </View>)}
  </View>;
}
const styles = StyleSheet.create({
  actions: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
  action: { flex: 1, minWidth: 100 }
});
