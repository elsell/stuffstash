import { nativeHeaderActionOptions } from '../components/NativeHeaderActions.ios';
import type { StackScreenProps } from 'expo-router';
import { assetOverflowMenuGroups } from './AssetOverflowMenu';
import type { AssetHeaderOverflowProps } from './AssetHeaderOverflow.types';

type NonFunction<T> = T extends (...args: any[]) => unknown ? never : T;
type NativeStackScreenOptions = NonFunction<NonNullable<StackScreenProps['options']>>;
type NativeHeaderItem = ReturnType<NonNullable<NativeStackScreenOptions['unstable_headerRightItems']>>[number];
type NativeHeaderMenuItem = Extract<NativeHeaderItem, { type: 'menu' }>;
type NativeHeaderMenuEntry = NativeHeaderMenuItem['menu']['items'][number];
type NativeHeaderMenuAction = Extract<NativeHeaderMenuEntry, { type: 'action' }>;
type NativeHeaderMenuGroup = Extract<NativeHeaderMenuEntry, { type: 'submenu' }>;
type NativeSFSymbolIcon = Extract<NonNullable<NativeHeaderMenuAction['icon']>, { type: 'sfSymbol' }>;

/** iOS installs a real UIBarButtonItem/UIMenu directly in the native navigation header. */
export function assetHeaderOverflowScreenOptions({
  asset,
  disabled = false,
  onEdit,
  onCheckoutHistory,
  onHistory,
  onLifecycleAction, onMove, onAddPhotos, onCheckout, photosDisabled
}: AssetHeaderOverflowProps): NativeStackScreenOptions {
  const groups = assetOverflowMenuGroups({ asset, onCheckoutHistory, onHistory, onLifecycleAction, onMove, onAddPhotos, onCheckout, photosDisabled });
  return {
    headerShown: true as const,
    unstable_headerRightItems: (context): NativeHeaderItem[] => [
      ...(onEdit ? nativeHeaderActionOptions([{ kind: 'compose', label: 'Edit', disabled,
        onPress: onEdit }]).unstable_headerRightItems?.(context) ?? [] : []), {
      type: 'menu',
      label: '',
      accessibilityLabel: `More actions for ${asset.title}`,
      disabled,
      icon: { type: 'sfSymbol', name: 'ellipsis' },
      sharesBackground: true,
      menu: {
        multiselectable: true,
        items: groups.map((group): NativeHeaderMenuGroup => ({
          type: 'submenu',
          label: '',
          inline: true,
          multiselectable: true,
          items: group.items.map((item): NativeHeaderMenuAction => ({
            type: 'action',
            label: item.label,
            ...(item.systemImage ? {
              icon: { type: 'sfSymbol', name: item.systemImage } as NativeSFSymbolIcon
            } : {}),
            destructive: item.isDestructive ?? false,
            disabled: item.disabled ?? false,
            state: 'off',
            onPress: item.onPress
          }))
        }))
      }
    }]
  };
}
