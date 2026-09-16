import { useState } from 'react';
import { Button, ScrollView, Text } from 'react-native';
import { Stack } from 'expo-router';
import { NativeNavigationSearch } from '../src/ui/components/NativeNavigationSearch';

/** Isolate search installation from data loading and production header actions. */
export function ManagedSearchPlacementFixture() {
  const [enabled, setEnabled] = useState(false);
  const [reconfigured, setReconfigured] = useState(false);
  const [query, setQuery] = useState('');
  return <>
    <Stack.Screen options={{ title: reconfigured ? 'Search reconfigured' : 'Managed search' }} />
    <NativeNavigationSearch enabled={enabled} query={query} placeholder="Managed search probe"
      onChange={setQuery} onSubmit={setQuery} onClear={() => setQuery('')} />
    <ScrollView contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled">
      <Button title="Enable managed search" onPress={() => setEnabled(true)} />
      <Button title="Reconfigure search header" onPress={() => setReconfigured(true)} />
      <Text>{enabled ? 'Managed search enabled' : 'Managed search disabled'}</Text>
      <Text>Query: {query}</Text>
    </ScrollView>
  </>;
}
