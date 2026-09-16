import { useState } from 'react';
import { Button, ScrollView, Text } from 'react-native';
import { Stack } from 'expo-router';
import { useNativeHeaderActionOptions } from '../src/ui/components/useNativeHeaderActionOptions';
import { NativeNavigationSearch } from '../src/ui/components/NativeNavigationSearch';

/** Isolate search installation from data loading and production header actions. */
export function ManagedSearchPlacementFixture() {
  const [enabled, setEnabled] = useState(false);
  const [reconfigured, setReconfigured] = useState(false);
  const [showAction, setShowAction] = useState(false);
  const [activations, setActivations] = useState(0);
  const actionOptions = useNativeHeaderActionOptions([{ kind: 'add', label: 'Probe action', onPress: () => setActivations(value => value + 1) }]);
  const [query, setQuery] = useState('');
  return <>
    <Stack.Screen options={{ title: reconfigured ? 'Search reconfigured' : 'Managed search', ...(showAction ? actionOptions : {}) }} />
    <NativeNavigationSearch enabled={enabled} query={query} placeholder="Managed search probe"
      onChange={setQuery} onSubmit={setQuery} onClear={() => setQuery('')} />
    <ScrollView contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled">
      <Button title="Enable managed search" onPress={() => setEnabled(true)} />
      <Button title="Reconfigure search header" onPress={() => setReconfigured(true)} />
      <Button title="Add native header action" onPress={() => setShowAction(true)} />
      <Text>Action activations: {activations}</Text>
      <Text>{enabled ? 'Managed search enabled' : 'Managed search disabled'}</Text>
      <Text>Query: {query}</Text>
    </ScrollView>
  </>;
}
