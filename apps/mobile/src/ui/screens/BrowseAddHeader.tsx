import { Stack } from 'expo-router';
import { nativeHeaderActionOptions } from '../components/NativeHeaderActions';

export function BrowseAddHeader({ canAdd, onAdd }: { readonly canAdd: boolean; readonly onAdd: () => void }) {
  return <Stack.Screen options={{
    title: 'Browse',
    headerLeft: undefined,
    ...nativeHeaderActionOptions(canAdd ? [{ kind: 'add', label: 'Add an asset', onPress: onAdd }] : [])
  }} />;
}
