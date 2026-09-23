import { useEffect, useRef, useState } from 'react';
import { Text, View } from 'react-native';
import { NativeActionMenu } from '../src/ui/components/NativeActionMenu';
import { NativeCommandButton } from '../src/ui/components/NativeCommandButton';

/** Runner-only delayed parent lock; production menu adapter and no inventory writes. */
export function NativeMenuOwnershipFixture() {
  const [locked, setLocked] = useState(false);
  const [armed, setArmed] = useState(false);
  const [activations, setActivations] = useState(0);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => () => clearTimeout(timer.current), []);
  return <View style={{ padding: 24, gap: 24 }}>
    <Text>{locked ? 'Menu locked' : 'Menu ready'}</Text>
    <Text>Menu activations: {activations}</Text>
    <NativeActionMenu accessibilityLabel="Menu ownership actions" disabled={locked}
      groups={[{ id: 'commands', items: [{ id: 'run', label: 'Run menu command',
        onPress: () => setActivations(value => value + 1) }] }]} />
    <NativeCommandButton label="Lock menu shortly" disabled={armed || locked} onPress={() => {
      clearTimeout(timer.current);
      setArmed(true);
      timer.current = setTimeout(() => { setLocked(true); setArmed(false); }, 5000);
    }} />
    <NativeCommandButton label="Unlock menu" disabled={!locked} onPress={() => setLocked(false)} />
  </View>;
}
