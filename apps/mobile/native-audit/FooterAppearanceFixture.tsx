import { useRef, useState } from 'react';
import { Button, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { useAppearance } from '../src/ui/theme/AppearanceContext';
import { NativeFilterSheet } from '../src/ui/components/NativeFilterSheet';

/** Isolates shipping footer appearance; no production session or mutation. */
export function FooterAppearanceFixture() {
  const { palette, preference, resolvedColorScheme, setPreference } = useAppearance();
  const initialPreference = useRef(preference);
  const router = useRouter();
  const [selected, setSelected] = useState(false);
  const [received, setReceived] = useState(false);
  return <NativeFilterSheet title="Footer appearance" footerTestID="footer-appearance-actions" actions={{
    primaryLabel: 'Move', secondaryLabel: 'Cancel', disabled: !selected,
    onApply: () => setReceived(true),
    onBack: () => { void setPreference(initialPreference.current).then(() => router.back()); }
  }}>
    <View style={{ padding: 20, gap: 16 }}>
      <Text accessibilityRole="header" style={{ color: palette.text, fontSize: 20 }}>Footer appearance</Text>
      <Text style={{ color: palette.text }}>Appearance: {resolvedColorScheme}</Text>
      <Button title="Use light appearance" onPress={() => { void setPreference('light'); }} />
      <Button title="Use dark appearance" onPress={() => { void setPreference('dark'); }} />
      <Button title={selected ? 'Clear destination' : 'Select destination'} onPress={() => { setSelected(!selected); setReceived(false); }} />
      <Text style={{ color: palette.text }}>{received ? 'Move received' : 'Move idle'}</Text>
    </View>
  </NativeFilterSheet>;
}
