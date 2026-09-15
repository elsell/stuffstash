import { useMemo } from 'react';
import { Stack } from 'expo-router';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';

export function BrowseAddHeader({ canAdd, onAdd }: { readonly canAdd: boolean; readonly onAdd: () => void }) {
  const actions = useNativeHeaderActionOptions(canAdd ? [{ kind: 'add', label: 'Add an asset', onPress: onAdd }] : []);
  const options = useMemo(() => ({ title: 'Browse', headerLeft: undefined, ...actions }), [actions]);
  return <Stack.Screen options={options} />;
}
