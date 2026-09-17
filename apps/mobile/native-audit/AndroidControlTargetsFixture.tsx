import { useState } from 'react';
import { Text, View } from 'react-native';
import { NativeRefinementButton } from '../src/ui/components/NativeRefinementButton';
import { NativeActionMenu } from '../src/ui/components/NativeActionMenu';

/** Real adapters; counters expose activation without changing inventory data. */
export function AndroidControlTargetsFixture() {
  const [filters, setFilters] = useState(0);
  const [commands, setCommands] = useState(0);
  const groups = [{ id: 'probe', items: [{ id: 'run', label: 'Run target probe', onPress: () => setCommands(value => value + 1) }] }];
  return <View style={{ gap: 16 }}>
    <NativeRefinementButton iconOnly label="Filters" accessibilityLabel="Probe filter target" onPress={() => setFilters(value => value + 1)} />
    <NativeActionMenu accessibilityLabel="Probe sort target" trigger={{ kind: 'icon', androidIcon: 'sort', systemImage: 'arrow.up.arrow.down' }} groups={groups} />
    <NativeActionMenu accessibilityLabel="Probe more target" groups={groups} />
    <Text>Filter activations: {filters}</Text>
    <Text>Menu activations: {commands}</Text>
  </View>;
}
