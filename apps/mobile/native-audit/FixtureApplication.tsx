import { PhotoRecoveryFixture } from './PhotoRecoveryFixture';
export { InventorySwitcherFixture } from './InventorySwitcherFixture';
import { HomeReturnTaskProvider } from '../src/ui/navigation/HomeReturnTaskPresentation';
export { HomeReturnFixture } from './HomeReturnFixture';
export { default as HomeReturnDetailsRoute } from '../src/ui/screens/HomeReturnDetailsRouteScreen';
import { Host, TextField } from '@expo/ui/swift-ui';
import { accessibilityLabel, autocorrectionDisabled, keyboardType, textFieldStyle, textInputAutocapitalization } from '@expo/ui/swift-ui/modifiers';
import { VoicePlanPhotoDraftStrip } from '../src/ui/screens/VoicePlanPhotoDrafts';
import { createContext, useContext, useState, type ReactNode } from 'react';
import { Button, Image, ScrollView, Text, View } from 'react-native';
import { Stack, useRouter, type Href } from 'expo-router';
import { AppearancePreferenceController, type AppearancePreference } from '../src/application/settings/AppearancePreference';
import { AppearanceProvider, useAppearance } from '../src/ui/theme/AppearanceContext';
import { AppKeyboardProvider } from '../src/ui/components/AppKeyboardProvider';
import { AppKeyboardAccessory } from '../src/ui/components/AppKeyboardAccessory';
import { AppFeedbackProvider, useAppFeedback } from '../src/ui/feedback/AppFeedback';
import { BrowseFiltersScreen } from '../src/ui/screens/BrowseFiltersScreen';
import { ExpirationFiltersScreen } from '../src/ui/expiration/ExpirationFiltersScreen';
import { OnboardingScreen } from '../src/ui/screens/OnboardingScreen';
import { OnboardingCommand, type OnboardingStartState } from '../src/application/onboarding/OnboardingCommand';
import { onboardingFakes } from '../src/application/onboarding/OnboardingTestSupport';
import { ExpirationReminderEditor } from '../src/ui/components/ExpirationReminderEditor';
import type { ExpirationReminderPolicy } from '../src/domain/notifications/Notification';
import { AppearancePicker } from '../src/ui/components/AppearancePicker';
import { TagColorPicker } from '../src/ui/components/TagColorPicker';
import { ExpirationField } from '../src/ui/components/ExpirationField';
import { AppTextInput } from '../src/ui/components/AppTextInput';
import { CustomizationFieldControls } from '../src/ui/components/CustomizationEditorFields';
import { createAssetNativeSheetOptions } from '../src/ui/screens/AssetNativeSheetOptions';

export { SheetLayoutFixture } from './SheetLayoutFixture';

export { InventoryQueryFixture } from './InventoryQueryFixture';
export { AddAssetFixture } from './AddAssetFixture';
export { CheckoutHistoryFixture } from './CheckoutHistoryFixture';

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
  return <ResultContext.Provider value={{ result, setResult }}><AppFeedbackProvider><HomeReturnTaskProvider>
    <Stack screenOptions={{ headerTintColor: palette.action, contentStyle: { backgroundColor: palette.background } }}>
      <Stack.Screen name="audit-home-return" options={{ title: 'Home' }} />
      <Stack.Screen name="home-return-details" options={{ ...sheets.checkoutHistory, title: 'Return details', gestureEnabled: false }} />
      <Stack.Screen name="index" options={{ title: 'Native UI audit' }} />
      <Stack.Screen name="audit-sheet-diagnostic" options={{ presentation: 'formSheet', sheetAllowedDetents: [1], sheetGrabberVisible: true }} />
      <Stack.Screen name="audit-inventory-switcher" options={{ presentation: 'formSheet', headerShown: true, sheetAllowedDetents: [0.5, 1], sheetCornerRadius: 24, sheetGrabberVisible: true, contentStyle: { backgroundColor: palette.surface } }} />
      <Stack.Screen name="audit-inventory-query" options={{ title: 'Inventory query', presentation: 'formSheet', sheetAllowedDetents: [1] }} />
      <Stack.Screen name="audit-add-push" options={{ presentation: 'card', headerShown: false, contentStyle: { backgroundColor: palette.background } }} />
      <Stack.Screen name="audit-add-header" options={{ presentation: 'formSheet', sheetAllowedDetents: [1], sheetCornerRadius: 24, sheetGrabberVisible: true, headerShown: true, title: 'Add item', contentStyle: { backgroundColor: palette.background } }} />
      <Stack.Screen name="audit-add" options={{ presentation: 'formSheet', sheetAllowedDetents: [1], sheetCornerRadius: 24, sheetGrabberVisible: true, headerShown: false, contentStyle: { backgroundColor: palette.background } }} />
      <Stack.Screen name="audit-checkout-history" options={sheets.checkoutHistory} />
      <Stack.Screen name="audit-browse" options={sheets.filters} />
      <Stack.Screen name="audit-expiration-medium" options={sheets.filters} />
      <Stack.Screen name="audit-expiration" options={sheets.filters} />
    </Stack>
    <AppKeyboardAccessory />
  </HomeReturnTaskProvider></AppFeedbackProvider></ResultContext.Provider>;
}

export function FixtureMenu() {
  const router = useRouter();
  const { result, setResult } = useContext(ResultContext);
  const feedback = useAppFeedback();
  const [showDraftOptions, setShowDraftOptions] = useState(false);
  const [onboardingSubmission, setOnboardingSubmission] = useState(false);
  const [settingsControls, setSettingsControls] = useState(false);
  const [draftPhotos, setDraftPhotos] = useState(false);
  const [photoRecovery, setPhotoRecovery] = useState<'removal' | 'missing'>();
  const [inputMode, setInputMode] = useState<'controlled' | 'uncontrolled' | 'system' | 'plain' | 'multiline'>();
  if (photoRecovery) return <PhotoRecoveryFixture missingImage={photoRecovery === 'missing'} onBack={() => setPhotoRecovery(undefined)} />;
  if (onboardingSubmission) return <OnboardingSubmissionFixture />;
  if (draftPhotos) return <DraftPhotosFixture onBack={() => setDraftPhotos(false)} />;
  if (settingsControls) return <SettingsControlsFixture onBack={() => setSettingsControls(false)} />;
  return <FixturePage>
    <Button title="Audit inventory query" onPress={() => router.push('/audit-inventory-query' as Href)} />
    <Button title="Audit inventory switcher" onPress={() => router.push('/audit-inventory-switcher' as Href)} />
    <Button title="Audit Home Return" onPress={() => router.push('/audit-home-return' as Href)} />
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
    <Button title="Audit system input" onPress={() => setInputMode('system')} />
    {inputMode ? <InputFixture key={inputMode} mode={inputMode} /> : null}
    <Button title="Audit Add navigation draft" onPress={() => router.push('/audit-add-push' as Href)} />
    <Button title="Audit Add configured header" onPress={() => router.push('/audit-add-header' as Href)} />
    <Button title="Audit Add draft" onPress={() => router.push('/audit-add' as Href)} />
    <Button title="Audit onboarding submission" onPress={() => setOnboardingSubmission(true)} />
    <Button title="Audit settings controls" onPress={() => setSettingsControls(true)} />
    {['direct', 'nested', 'footer', 'direct-footer', 'scroll-footer'].map(variant => <Button key={variant} title={`Audit ${variant} sheet`}
      onPress={() => router.push({ pathname: '/audit-sheet-diagnostic', params: { variant } } as Href)} />)}
    <Button title="Audit Checkout history" onPress={() => router.push('/audit-checkout-history' as Href)} />
    <Button title="Audit draft photos" onPress={() => setDraftPhotos(true)} />
    <Button title="Audit plain input" onPress={() => setInputMode('plain')} />
    <Button title="Audit multiline input" onPress={() => setInputMode('multiline')} />
    <Button title="Audit photo removal recovery" onPress={() => setPhotoRecovery('removal')} />
    <Button title="Audit unavailable photo" onPress={() => setPhotoRecovery('missing')} />
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

function InputFixture({ mode }: { readonly mode: 'controlled' | 'uncontrolled' | 'system' | 'plain' | 'multiline' }) {
  const [value, setValue] = useState('');
  if (mode === 'plain' || mode === 'multiline') return <View>
    <AppTextInput accessibilityLabel={`Audit ${mode} text`} defaultValue="" multiline={mode === 'multiline'}
      onChangeText={setValue} style={{ minHeight: mode === 'multiline' ? 160 : 54, borderWidth: 1, padding: 12 }} />
    <Text>{`Observed ${mode} input: ${value}`}</Text>
  </View>;
  if (mode === 'system') return <View>
    <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 54 }}>
      <TextField defaultValue="" placeholder="https://example.invalid" onValueChange={setValue}
        modifiers={[accessibilityLabel('Audit system address'), keyboardType('url'), autocorrectionDisabled(), textInputAutocapitalization('never'), textFieldStyle('roundedBorder')]} />
    </Host>
    <Text>{`Observed system input: ${value}`}</Text>
  </View>;
  return <View><AppTextInput accessibilityLabel={`Audit ${mode} address`} keyboardType="url"
    autoCorrect={false} autoCapitalize="none" onChangeText={setValue}
    {...(mode === 'controlled' ? { value } : { defaultValue: '' })}
    style={{ minHeight: 54, borderWidth: 1, padding: 12 }} />
    <Text>{`Observed ${mode} input: ${value}`}</Text>
  </View>;
}


function SettingsControlsFixture({ onBack }: { readonly onBack: () => void }) {
  const { preference } = useAppearance();
  const [color, setColor] = useState('');
  const [expiration, setExpiration] = useState('No expiration');
  const [reminder, setReminder] = useState<ExpirationReminderPolicy | null>(null);
  return <FixturePage>
    <Button title="Back to audit menu" onPress={onBack} />
    <AppearancePicker />
    <Text>{`Appearance value: ${preference}`}</Text>
    <TagColorPicker value={color} onChange={setColor} />
    <Text>{`Color value: ${color || 'none'}`}</Text>
    <ExpirationField initialPickerDate={new Date(2026, 8, 14, 12)} onChange={value => setExpiration(value?.date ?? 'No expiration')} />
    <Text>{`Expiration value: ${expiration}`}</Text>
    <ExpirationReminderEditor initialPolicy={reminder}
      inheritedPolicy={{ enabled: true, upcoming: true, expired: true, advanceDays: 7 }}
      onSave={async value => setReminder(value)} onEditDays={() => {}} />
    <Text>{`Reminder mode: ${reminder === null ? 'defaults' : reminder.enabled ? 'custom' : 'off'}`}</Text>
  </FixturePage>;
}


function OnboardingSubmissionFixture() {
  const [fakes] = useState(onboardingFakes);
  const [command] = useState(() => new OnboardingCommand(fakes.profiles, () => fakes.api, fakes.auth));
  const [state, setState] = useState<OnboardingStartState>({ step: 'instance' });
  return <View style={{ flex: 1 }}>
    <OnboardingScreen command={command} initialState={state} onStateChange={setState} onComplete={() => {}} />
    <Text>{`Submitted address: ${fakes.auth.signIns.at(-1) ?? 'none'}`}</Text>
  </View>;
}

function DraftPhotosFixture({ onBack }: { readonly onBack: () => void }) {
  const [photos, setPhotos] = useState(() => [1, 2, 3, 4].map(index => ({
    id: `photo-${index}`, uri: Image.resolveAssetSource(require('../assets/brand/stuff-stash-glyph.png')).uri,
    fileName: `photo-${index}.png`, contentType: 'image/png' as const, sizeBytes: 1
  })));
  const [removed, setRemoved] = useState('none');
  const [added, setAdded] = useState(0);
  const [readOnly, setReadOnly] = useState(false);
  return <FixturePage>
    <Button title="Back to audit menu" onPress={onBack} />
    <VoicePlanPhotoDraftStrip commandKey="audit-command" photos={photos} readOnly={readOnly}
      onAddPhotos={() => setAdded(value => value + 1)}
      onRemovePhoto={(_, id) => { setRemoved(id); setPhotos(value => value.filter(photo => photo.id !== id)); }} />
    <Text>{`Removed photo: ${removed}`}</Text>
    <Text>{`Photos remaining: ${photos.length}`}</Text>
    <Text>{`Photo add requests: ${added}`}</Text>
    <Button title={readOnly ? 'Enable photo editing' : 'Make photos read only'} onPress={() => setReadOnly(value => !value)} />
  </FixturePage>;
}
