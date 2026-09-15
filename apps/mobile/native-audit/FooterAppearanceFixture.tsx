import { useRef, useState } from 'react';
import { Button, ScrollView, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { useAppearance } from '../src/ui/theme/AppearanceContext';
import { NativeSheetActions } from '../src/ui/components/NativeSheetActions';

/** Isolates shipping footer appearance; no production session or mutation. */
export function FooterAppearanceFixture() {
  const { palette, preference, resolvedColorScheme, setPreference } = useAppearance();
  const initialPreference = useRef(preference);
  const router = useRouter();
  const [selected, setSelected] = useState(false);
  const [received, setReceived] = useState(false);
  return <View testID="footer-appearance-root" style={{ flex: 1, padding: 20, gap: 12, backgroundColor: palette.surface }}>
    <ScrollView contentContainerStyle={{ gap: 16 }}>
      <Text accessibilityRole="header" style={{ color: palette.text, fontSize: 20 }}>Footer appearance</Text>
      <Text style={{ color: palette.text }}>Appearance: {resolvedColorScheme}</Text>
      <Button title="Use light appearance" onPress={() => { void setPreference('light'); }} />
      <Button title="Use dark appearance" onPress={() => { void setPreference('dark'); }} />
      <Button title={selected ? 'Clear destination' : 'Select destination'} onPress={() => { setSelected(!selected); setReceived(false); }} />
      <Text style={{ color: palette.text }}>{received ? 'Move received' : 'Move idle'}</Text>
    </ScrollView>
    <NativeSheetActions primaryLabel="Move" secondaryLabel="Cancel" disabled={!selected}
      onApply={() => setReceived(true)}
      onBack={() => { void setPreference(initialPreference.current).then(() => router.back()); }} />
  </View>;
}
