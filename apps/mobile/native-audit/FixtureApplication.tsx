import { createContext, useContext, useState, type ReactNode } from 'react';
import { Button, ScrollView, Text, View } from 'react-native';
import { Stack, useRouter, type Href } from 'expo-router';
import { AppearancePreferenceController, type AppearancePreference } from '../src/application/settings/AppearancePreference';
import { AppearanceProvider, useAppearance } from '../src/ui/theme/AppearanceContext';
import { AppKeyboardProvider } from '../src/ui/components/AppKeyboardProvider';
import { AppKeyboardAccessory } from '../src/ui/components/AppKeyboardAccessory';
import { AppFeedbackProvider, useAppFeedback } from '../src/ui/feedback/AppFeedback';
import { BrowseFiltersScreen } from '../src/ui/screens/BrowseFiltersScreen';
import { ExpirationFiltersScreen } from '../src/ui/expiration/ExpirationFiltersScreen';
import { AppTextInput } from '../src/ui/components/AppTextInput';
import { CustomizationFieldControls } from '../src/ui/components/CustomizationEditorFields';
import { createAssetNativeSheetOptions } from '../src/ui/screens/AssetNativeSheetOptions';

// Runner-only composition. No production session, service, or credentials are loaded.
const ResultContext = createContext({ result: '', setResult: (_value: string) => {} });

export function FixtureLayout() {
  const [controller] = useState(() => {
    let preference: AppearancePreference = 'system';
    return new AppearancePreferenceController({
      load: async () => preference,
      save: async value => { preference = value; }
    });
  });
  return <AppKeyboardProvider><AppearanceProvider controller={controller}>
    <FixtureNavigation />
  </AppearanceProvider></AppKeyboardProvider>;
}

function FixtureNavigation() {
  const { palette, isHydrated } = useAppearance();
  const [result, setResult] = useState('');
  const sheets = createAssetNativeSheetOptions(palette);
  if (!isHydrated) return <View />;
  return <ResultContext.Provider value={{ result, setResult }}><AppFeedbackProvider>
    <Stack screenOptions={{ headerTintColor: palette.action, contentStyle: { backgroundColor: palette.background } }}>
      <Stack.Screen name="index" options={{ title: 'Native UI audit' }} />
      <Stack.Screen name="audit-browse" options={sheets.filters} />
      <Stack.Screen name="audit-expiration-medium" options={sheets.filters} />
      <Stack.Screen name="audit-expiration" options={sheets.filters} />
    </Stack>
    <AppKeyboardAccessory />
  </AppFeedbackProvider></ResultContext.Provider>;
}

export function FixtureMenu() {
  const router = useRouter();
  const { result, setResult } = useContext(ResultContext);
  const feedback = useAppFeedback();
  const [showDraftOptions, setShowDraftOptions] = useState(false);
  const [inputMode, setInputMode] = useState<'controlled' | 'uncontrolled'>();
  return <FixturePage>
    <Button title="Audit Browse filters" onPress={() => router.push('/audit-browse' as Href)} />
    <Button title="Audit Expiration filters" onPress={() => router.push('/audit-expiration' as Href)} />
    <Button title="Audit medium expiration filters" onPress={() => router.push('/audit-expiration-medium' as Href)} />
    <Button title="Audit feedback" onPress={() => feedback.showNotice({
      tone: 'error', title: 'Audit action needs attention', message: 'This is synthetic audit data.',
      action: { label: 'Retry audit action', onPress: () => setResult('Audit retry completed') }
    })} />
    <Button title="Audit draft options" onPress={() => setShowDraftOptions(true)} />
    {showDraftOptions ? <DraftOptionsFixture /> : null}
    <Button title="Audit controlled input" onPress={() => setInputMode('controlled')} />
    <Button title="Audit uncontrolled input" onPress={() => setInputMode('uncontrolled')} />
    {inputMode ? <InputFixture key={inputMode} mode={inputMode} /> : null}
    <Text>{result}</Text>
  </FixturePage>;
}

function FixturePage({ children }: { readonly children: ReactNode }) {
  return <ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={{ padding: 20, gap: 20 }}>{children}</ScrollView>;
}

export function BrowseFilterFixture() {
  const router = useRouter();
  const { setResult } = useContext(ResultContext);
  return <BrowseFiltersScreen initial={{ scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [], sort: 'updated_desc' }}
    query="" tags={[{ id: 'audit-tools', key: 'tools', label: 'Tools' }, { id: 'audit-holiday', key: 'holiday', label: 'Holiday supplies' }]}
    onApply={draft => { setResult(`Browse availability: ${draft.checkoutState}`); router.back(); }}
    onCancel={() => router.back()}
    onExpiration={mode => { setResult(`Expiration mode: ${mode}`); router.back(); }} />;
}

export function ExpirationFilterFixture() {
  const router = useRouter();
  const { setResult } = useContext(ResultContext);
  return <ExpirationFiltersScreen initial={{ mode: 'all', throughDate: '2026-10-15' }} choices={{
    types: [{ id: 'audit-food', label: 'Food' }],
    tags: [{ id: 'audit-tools', label: 'Tools' }, { id: 'audit-holiday', label: 'Holiday supplies' }],
    locations: [{ id: 'audit-kitchen', label: 'Kitchen / Cabinet' }, { id: 'audit-garage', label: 'Garage / Cabinet' }]
  }} onApply={filter => { setResult(`Expiration mode: ${filter.mode}`); router.back(); }} onCancel={() => router.back()} />;
}

function DraftOptionsFixture() {
  const [options, setOptions] = useState<readonly string[]>(['saved', 'draft']);
  const [newOption, setNewOption] = useState('');
  return <CustomizationFieldControls applicability="all_assets" canMutate eligibleTypes={[]}
    enumOptions={options} persistedEnumOptions={['saved']} fieldType="enum" mode="edit"
    newOption={newOption} onNewOption={setNewOption} onEnumOptions={setOptions}
    persistedTargetIds={[]} targetIds={[]} onTargets={() => {}} onApplicability={() => {}} onFieldType={() => {}} />;
}

function InputFixture({ mode }: { readonly mode: 'controlled' | 'uncontrolled' }) {
  const [value, setValue] = useState('');
  return <AppTextInput accessibilityLabel={`Audit ${mode} address`} keyboardType="url"
    autoCorrect={false} autoCapitalize="none" onChangeText={setValue}
    {...(mode === 'controlled' ? { value } : { defaultValue: '' })}
    style={{ minHeight: 54, borderWidth: 1, padding: 12 }} />;
}
