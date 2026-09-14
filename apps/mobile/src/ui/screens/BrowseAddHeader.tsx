import { Stack, router } from 'expo-router';
import { Pressable, Text, View, useWindowDimensions } from 'react-native';
import { ChevronDown } from 'lucide-react-native';
import { nativeHeaderActionOptions } from '../components/NativeHeaderActions';
import { useAppearanceAwarePalette } from '../theme/appearance';

export type BrowseHeaderContextProps = {
  readonly inventoryContext?: string;
  readonly inventoryContextStatus?: 'loading' | 'ready' | 'error';
  readonly onRetryInventoryContext?: () => void;
};

export function BrowseAddHeader({ canAdd, onAdd, inventoryContext, inventoryContextStatus = 'ready', onRetryInventoryContext }: BrowseHeaderContextProps & {
  readonly canAdd: boolean; readonly onAdd: () => void;
}) {
  const palette = useAppearanceAwarePalette();
  const { width } = useWindowDimensions();
  const failed = inventoryContextStatus === 'error';
  const name = inventoryContextStatus === 'loading' ? 'Loading inventory…' : failed ? 'Inventory unavailable' : inventoryContext ?? 'Browse';
  return <Stack.Screen options={{
    title: '',
    headerLeft: () => <Pressable accessibilityRole="button"
      accessibilityLabel={failed ? 'Retry inventory context' : `Current inventory ${name}. Switch inventory`}
      disabled={inventoryContextStatus === 'loading'}
      onPress={failed ? onRetryInventoryContext : () => router.push('/tenant-switcher')}
      style={{ flexDirection: 'row', alignItems: 'center', gap: 8, minHeight: 44, width: Math.max(100, Math.min(260, width - (canAdd ? 2 : 1) * 48 - 64)) }}>
      <View style={{ flex: 1, minWidth: 0 }}>
        <Text numberOfLines={1} style={{ color: palette.text, fontSize: 17, fontWeight: '600' }}>{name}</Text>
        <Text numberOfLines={1} style={{ color: palette.textMuted, fontSize: 12 }}>{failed ? 'Tap to retry' : 'Browse'}</Text>
      </View>
      <ChevronDown color={palette.textMuted} size={18} />
    </Pressable>,
    ...nativeHeaderActionOptions(canAdd ? [{ kind: 'add', label: 'Add an asset', onPress: onAdd }] : [])
  }} />;
}
