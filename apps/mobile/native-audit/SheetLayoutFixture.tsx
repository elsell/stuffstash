import { ScrollView, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import { NativeSheetActions } from '../src/ui/components/NativeSheetActions';
import { SettingsNavigationRow, SettingsSection } from '../src/ui/screens/SettingsList';
import { useAppearancePalette } from '../src/ui/theme/AppearanceContext';

// Isolates container structure without changing the real filter implementation.
export function SheetLayoutFixture() {
  const { variant } = useLocalSearchParams<{ variant: string }>();
  const palette = useAppearancePalette();
  const router = useRouter();
  const body = <ScrollView style={{ flex: 1 }} contentInsetAdjustmentBehavior="automatic">
    <SettingsSection>
      {['Type', 'Availability', 'Tags', 'Location'].map(label =>
        <SettingsNavigationRow key={label} label={label} accessibilityLabel={`Diagnostic ${label}`} onPress={() => {}} />)}
    </SettingsSection>
    <Text>Sheet layout diagnostic: {variant}</Text>
  </ScrollView>;
  return <>
    <Stack.Screen options={{ title: 'Sheet diagnostic' }} />
    {variant === 'direct' ? body : <View style={{ flex: 1, backgroundColor: palette.background }}>
      {body}
      {variant === 'footer' ? <SafeAreaView edges={['bottom']} style={{ paddingHorizontal: 20, paddingVertical: 12 }}>
        <NativeSheetActions primaryLabel="Finish diagnostic" secondaryLabel="Cancel diagnostic" disabled={false}
          onApply={() => router.back()} onBack={() => router.back()} />
      </SafeAreaView> : null}
    </View>}
  </>;
}
