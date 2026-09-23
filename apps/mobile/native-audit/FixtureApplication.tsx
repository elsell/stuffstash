import { BrowseFilterJourneyProvider } from './BrowseFilterJourneyFixture';
export { BrowseFilterJourneySearch, BrowseFilterJourneyFilters, BrowseFilterJourneyExpiration, BrowseFilterJourneyDetail } from './BrowseFilterJourneyFixture';
export { default as AddDestinationRoute } from '../src/ui/screens/AddDestinationRouteScreen';
import { AddDestinationTaskProvider } from '../src/ui/navigation/AddDestinationTask';
import { AssetTagSelectionTaskProvider } from '../src/ui/navigation/AssetTagSelectionTask';
export { default as AssetTagSelectionRoute } from '../src/ui/screens/AssetTagSelectionRouteScreen';
import { AssetEditJourneyProvider } from './AssetEditJourneyFixture';
export { AssetEditJourneyDetailFixture, AssetEditJourneyEditorFixture, AssetEditJourneyMoveFixture } from './AssetEditJourneyFixture';
export { BrowseJourneyFixture } from './BrowseJourneyFixture';
import { CustomFieldChoicesFixture } from './CustomFieldChoicesFixture';
export { NativeMenuOwnershipFixture } from './NativeMenuOwnershipFixture';
import { FilterGeometryProbe } from './FilterGeometryProbe';
import { useInputEventTrace } from './InputEventTrace';
export { AndroidHeaderCompositionFixture } from './AndroidHeaderCompositionFixture';
import { AndroidControlTargetsFixture } from './AndroidControlTargetsFixture';
import { inventorySwitcherNativeOptions } from '../src/ui/screens/InventorySwitcherNativeOptions';
export { ManagedSearchPlacementFixture } from './ManagedSearchPlacementFixture';
import { AppNoticeScreenLayout } from '../src/ui/feedback/AppNoticeScreenLayout';
import { returnToPreviousOrHome } from '../src/ui/navigation/returnToPreviousOrHome';
import { VoiceProposalFixtureProvider } from './VoiceProposalFixture';
import { voiceNativeSheetOptions } from '../src/ui/screens/VoiceNativeSheetOptions';
export { VoiceProposalFixture, VoicePlanLocationFixture } from './VoiceProposalFixture';
export { NoticePlacementFixture } from './NoticePlacementFixture';
export { NativeSearchPlacementFixture } from './NativeSearchPlacementFixture';
export { ProviderEditorFixture } from './ProviderEditorFixture';
export { AccountConnectionFixture } from './AccountConnectionFixture';
export { InvitationAcceptanceFixture } from './InvitationAcceptanceFixture';
export { NotificationInboxFixture, NotificationTargetFixture } from './NotificationInboxFixture';
export { InventorySharingFixture } from './InventorySharingFixture';
export { FooterAppearanceFixture } from './FooterAppearanceFixture';
export { MoveDestinationFixture } from './MoveDestinationFixture';
export { MoveHereRecoveryFixture } from './MoveHereRecoveryFixture';
export { CommandHeightFixture } from './CommandHeightFixture';
export { AssetRegionRecoveryFixture, AssetContentsSearchFixture, AssetDetailCommandsFixture } from './AssetRegionRecoveryFixture';
export { AssetEditRecoveryFixture, AssetEditTagsFixture } from './AssetEditRecoveryFixture';
import { PhotoRecoveryFixture } from './PhotoRecoveryFixture';
export { InventorySwitcherFixture } from './InventorySwitcherFixture';
import { HomeReturnTaskProvider } from '../src/ui/navigation/HomeReturnTaskPresentation';
export { HomeReturnFixture, HomeHeaderFixture, HomeTabShellFixture, TabShellBrowsePlaceholder, HomeAddProbeDestination, HomeProfileProbeDestination } from './HomeReturnFixture';
import { nativeTabHeaderOptions } from '../src/ui/navigation/NativeTabHeader';
export { default as HomeReturnDetailsRoute } from '../src/ui/screens/HomeReturnDetailsRouteScreen';
import { Host, TextField } from '@expo/ui/swift-ui';
import { accessibilityLabel, autocorrectionDisabled, keyboardType, textFieldStyle, textInputAutocapitalization } from '@expo/ui/swift-ui/modifiers';
import { VoicePlanPhotoDraftStrip } from '../src/ui/screens/VoicePlanPhotoDrafts';
import { createContext, useContext, useState, type ReactNode } from 'react';
import { Button, Image, Platform, ScrollView, Text, View } from 'react-native';
import { Stack, useRouter, type Href } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
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
export { AddAssetFixture, AddDestinationFixture } from './AddAssetFixture';
export { CheckoutHistoryFixture } from './CheckoutHistoryFixture';

// Runner-only composition. No production session, service, or credentials are loaded.
const ResultContext = createContext({ result: '', setResult: (_value: string) => {}, keyboardAccessoryEnabled: true, setKeyboardAccessoryEnabled: (_value: boolean) => {} });

const AppearanceWriteCount = createContext(0);

export function FixtureLayout() {
  return <FixtureRoot keyboardProviderEnabled />;
}

export function FixtureLayoutWithoutKeyboardProvider() {
  return <FixtureRoot keyboardProviderEnabled={false} />;
}

function FixtureRoot({ keyboardProviderEnabled }: { readonly keyboardProviderEnabled: boolean }) {
  const [appearanceWrites, setAppearanceWrites] = useState(0);
  const [controller] = useState(() => {
    let preference: AppearancePreference = 'system';
    return new AppearancePreferenceController({
      load: async () => preference,
      save: async value => { preference = value; setAppearanceWrites(count => count + 1); }
    });
  });
  const body = <AppearanceProvider controller={controller}>
    <AppearanceWriteCount.Provider value={appearanceWrites}><VoiceProposalFixtureProvider><FixtureNavigation keyboardProviderEnabled={keyboardProviderEnabled} /></VoiceProposalFixtureProvider></AppearanceWriteCount.Provider>
  </AppearanceProvider>;
  return keyboardProviderEnabled ? <AppKeyboardProvider>{body}</AppKeyboardProvider>
    : <View testID="audit-keyboard-provider-omitted" collapsable={false} style={{ flex: 1 }}>{body}</View>;
}

function FixtureNavigation({ keyboardProviderEnabled }: { readonly keyboardProviderEnabled: boolean }) {
  const { palette, isHydrated, resolvedColorScheme } = useAppearance();
  const [result, setResult] = useState('');
  const [keyboardAccessoryEnabled, setKeyboardAccessoryEnabled] = useState(true);
  const sheets = createAssetNativeSheetOptions(palette);
  if (!isHydrated) return <View />;
  return <ResultContext.Provider value={{ result, setResult, keyboardAccessoryEnabled, setKeyboardAccessoryEnabled }}><AppFeedbackProvider noticePlacement="screen"><HomeReturnTaskProvider><AssetTagSelectionTaskProvider><AddDestinationTaskProvider><AssetEditJourneyProvider><BrowseFilterJourneyProvider>
    <StatusBar style={resolvedColorScheme === 'dark' ? 'light' : 'dark'} />
    <Stack screenLayout={AppNoticeScreenLayout} screenOptions={{
      headerBackTitle: 'Back', headerTintColor: palette.action,
      contentStyle: { backgroundColor: palette.background },
      headerStyle: { backgroundColor: palette.surface },
      headerTitleStyle: { color: palette.text, fontWeight: '700' }
    }}>
      <Stack.Screen name="voice" options={voiceNativeSheetOptions(palette)} />
      <Stack.Screen name="voice-plan-location" options={{ title: 'Containing location' }} />
      <Stack.Screen name="search" options={{ ...nativeTabHeaderOptions(palette, Platform.OS, Platform.Version), headerBackVisible: false }} />
      <Stack.Screen name="browse-filters" options={sheets.filters} />
      <Stack.Screen name="expiration" options={{ title: 'Expiration' }} />
      <Stack.Screen name="assets/[assetId]/index" options={{ title: 'Details' }} />
      <Stack.Screen name="audit-browse-journey" options={{ ...nativeTabHeaderOptions(palette, Platform.OS, Platform.Version), headerBackVisible: false }} />
      <Stack.Screen name="audit-home-return" options={{ title: 'Home' }} />
      <Stack.Screen name="audit-home-header" options={{ ...nativeTabHeaderOptions(palette, Platform.OS, Platform.Version, palette.background), headerBackVisible: false }} />
      <Stack.Screen name="add-destination" options={{ ...sheets.selection, title: 'Put in' }} />
      <Stack.Screen name="asset-tag-selection" options={{ ...sheets.selection, title: 'Tags' }} />
      <Stack.Screen name="home-return-details" options={{ ...sheets.checkoutHistory, title: 'Return details', gestureEnabled: false }} />
      <Stack.Screen name="index" options={{ title: 'Native UI audit' }} />
      <Stack.Screen name="audit-tabs" options={{ headerShown: false }} />
      <Stack.Screen name="audit-sheet-diagnostic" options={{ presentation: 'formSheet', sheetAllowedDetents: [1], sheetGrabberVisible: true }} />
      <Stack.Screen name="audit-inventory-switcher" options={inventorySwitcherNativeOptions(palette)} />
      <Stack.Screen name="audit-inventory-query" options={{ title: 'Inventory query', presentation: 'formSheet', sheetAllowedDetents: [1] }} />
      <Stack.Screen name="audit-add-push" options={{ presentation: 'card', headerShown: false, contentStyle: { backgroundColor: palette.background } }} />
      <Stack.Screen name="audit-add-header" options={sheets.add} />
      <Stack.Screen name="audit-add" options={{ presentation: 'formSheet', sheetAllowedDetents: [1], sheetCornerRadius: 24, sheetGrabberVisible: true, headerShown: false, contentStyle: { backgroundColor: palette.background } }} />
      <Stack.Screen name="audit-detail-commands" options={{ title: 'Details' }} />
      <Stack.Screen name="audit-contents-search" options={{ title: 'Place' }} />
      <Stack.Screen name="audit-contents-search-preconfigured" options={{ title: 'Place', headerSearchBarOptions: {
        placeholder: 'Search this place', placement: 'integratedButton', allowToolbarIntegration: false,
        hideWhenScrolling: false, hideNavigationBar: false, obscureBackground: false, autoCapitalize: 'none'
      } }} />
      <Stack.Screen name="audit-managed-search" options={{ title: 'Managed search' }} />
      <Stack.Screen name="audit-native-search-placement" options={{ title: 'Search placement', headerSearchBarOptions: {
        placeholder: 'Search placement probe', placement: 'integratedButton', allowToolbarIntegration: false,
        hideWhenScrolling: false, hideNavigationBar: false, obscureBackground: false, autoCapitalize: 'none'
      } }} />
      <Stack.Screen name="audit-region-recovery" options={{ title: 'Place' }} />
      <Stack.Screen name="audit-notice" options={{ title: 'Notice placement' }} />
      <Stack.Screen name="audit-notice-sheet" options={{ title: 'Notice placement', presentation: 'formSheet', headerShown: true, sheetAllowedDetents: [1], sheetGrabberVisible: true }} />
      <Stack.Screen name="audit-provider-editor" options={{ title: 'Provider editor' }} />
      <Stack.Screen name="audit-account" options={{ title: 'Account' }} />
      <Stack.Screen name="audit-invitation" options={{ title: 'Invitation' }} />
      <Stack.Screen name="audit-connection" options={{ title: 'Connection' }} />
      <Stack.Screen name="audit-notifications" options={{ title: 'Notifications' }} />
      <Stack.Screen name="audit-notification-target" options={{ title: 'Notification destination' }} />
      <Stack.Screen name="audit-customization" options={{ title: 'Tags' }} />
      <Stack.Screen name="audit-sharing" options={{ title: 'Sharing' }} />
      <Stack.Screen name="audit-footer-appearance" options={{ ...sheets.move, sheetInitialDetentIndex: 1 }} />
      <Stack.Screen name="audit-menu-ownership" options={{ title: 'Menu ownership' }} />
      <Stack.Screen name="audit-command-height" options={{ title: 'Command sizing' }} />
      <Stack.Screen name="audit-move-destination" options={sheets.move} />
      <Stack.Screen name="audit-move-here-recovery" options={sheets.moveHere} />
      <Stack.Screen name="audit-edit-journey" options={{ title: 'Details' }} />
      <Stack.Screen name="assets/[assetId]/edit" options={sheets.edit} />
      <Stack.Screen name="assets/[assetId]/move" options={sheets.move} />
      <Stack.Screen name="audit-edit-tags" options={sheets.edit} />
      <Stack.Screen name="audit-edit-recovery" options={sheets.edit} />
      <Stack.Screen name="audit-checkout-history" options={sheets.checkoutHistory} />
      <Stack.Screen name="audit-browse" options={sheets.filters} />
      <Stack.Screen name="audit-expiration-medium" options={sheets.filters} />
      <Stack.Screen name="audit-expiration" options={sheets.filters} />
    </Stack>
    {keyboardProviderEnabled ? <AppKeyboardAccessory enabled={keyboardAccessoryEnabled} /> : null}
  </BrowseFilterJourneyProvider></AssetEditJourneyProvider></AddDestinationTaskProvider></AssetTagSelectionTaskProvider></HomeReturnTaskProvider></AppFeedbackProvider></ResultContext.Provider>;
}

type InputFixtureMode = 'controlled' | 'uncontrolled' | 'system' | 'plain' | 'multiline' | 'native-default'
  | 'plain-controlled' | 'plain-no-assistance' | 'plain-no-accessory'
  | 'plain-controlled-no-assistance' | 'plain-controlled-no-accessory';

export { CustomizationCollectionFixture } from './CustomizationCollectionFixture';
export { CustomizationEditorFixture } from './CustomizationEditorFixture';

export function FixtureMenu() {
  const router = useRouter();
  const { result, setResult, setKeyboardAccessoryEnabled } = useContext(ResultContext);
  const feedback = useAppFeedback();
  const [fieldChoices, setFieldChoices] = useState(false);
  const [showDraftOptions, setShowDraftOptions] = useState(false);
  const [onboardingSubmission, setOnboardingSubmission] = useState(false);
  const [settingsControls, setSettingsControls] = useState<'scroll' | 'fixed'>();
  const [draftPhotos, setDraftPhotos] = useState(false);
  const [photoRecovery, setPhotoRecovery] = useState<'removal' | 'missing'>();
  const [inputMode, setInputMode] = useState<InputFixtureMode>();
  if (inputMode) return <FixturePage key={`input-${inputMode}`} persistHandledTaps>
    <InputFixture mode={inputMode} />
    <Button title="Back to audit menu" onPress={() => { setInputMode(undefined); setKeyboardAccessoryEnabled(true); }} />
  </FixturePage>;
  if (fieldChoices) return <CustomFieldChoicesFixture onBack={() => setFieldChoices(false)} />;
  if (photoRecovery) return <PhotoRecoveryFixture missingImage={photoRecovery === 'missing'} onBack={() => setPhotoRecovery(undefined)} />;
  if (onboardingSubmission) return <OnboardingSubmissionFixture />;
  if (draftPhotos) return <DraftPhotosFixture onBack={() => setDraftPhotos(false)} />;
  if (settingsControls) return <SettingsControlsFixture scrollEnabled={settingsControls === 'scroll'} onBack={() => setSettingsControls(undefined)} />;
  return <FixturePage>
    <Button title="Audit voice proposal" onPress={() => router.push('/voice' as Href)} />
    <Button title="Audit Notice push" onPress={() => router.push('/audit-notice' as Href)} />
    <Button title="Audit Notice sheet" onPress={() => router.push('/audit-notice-sheet' as Href)} />
    <Button title="Audit Provider credential" onPress={() => router.push('/audit-provider-editor?kind=credential' as Href)} />
    <Button title="Audit Provider prompt" onPress={() => router.push('/audit-provider-editor?kind=prompt' as Href)} />
    <Button title="Audit Account" onPress={() => router.push('/audit-account' as Href)} />
    <Button title="Audit Home in tabs" onPress={() => router.push('/audit-tabs/(home)' as Href)} />
    <Button title="Audit invitation acceptance" onPress={() => router.push('/audit-invitation' as Href)} />
    <Button title="Audit Connection" onPress={() => router.push('/audit-connection' as Href)} />
    <Button title="Audit static search placement" onPress={() => router.push('/audit-native-search-placement' as Href)} />
    <Button title="Audit preconfigured place search" onPress={() => router.push('/audit-contents-search-preconfigured' as Href)} />
    <Button title="Audit Notifications" onPress={() => router.push('/audit-notifications' as Href)} />
    <Button title="Audit Sharing" onPress={() => router.push('/audit-sharing' as Href)} />
    <Button title="Audit inventory query" onPress={() => router.push('/audit-inventory-query' as Href)} />
    <Button title="Audit inventory switcher" onPress={() => router.push('/audit-inventory-switcher' as Href)} />
    <Button title="Audit Home Return" onPress={() => router.push('/audit-home-return' as Href)} />
    <Button title="Audit Home header" onPress={() => router.push('/audit-home-header' as Href)} />
    <Button title="Audit Browse filters" onPress={() => router.push('/audit-browse' as Href)} />
    <Button title="Audit Expiration filters" onPress={() => router.push('/audit-expiration' as Href)} />
    <Button title="Audit medium expiration filters" onPress={() => router.push('/audit-expiration-medium' as Href)} />
    <Button title="Audit feedback" onPress={() => feedback.showNotice({
      tone: 'error', title: 'Audit action needs attention', message: 'This is synthetic audit data.',
      action: { label: 'Retry audit action', onPress: () => setResult('Audit retry completed') }
    })} />
    <Button title="Audit field choices" onPress={() => setFieldChoices(true)} />
    <Button title="Audit draft options" onPress={() => setShowDraftOptions(true)} />
    {showDraftOptions ? <DraftOptionsFixture /> : null}
    <Button title="Audit controlled input" onPress={() => setInputMode('controlled')} />
    <Button title="Audit uncontrolled input" onPress={() => setInputMode('uncontrolled')} />
    <Button title="Audit input without accessory" onPress={() => { setKeyboardAccessoryEnabled(false); setInputMode('uncontrolled'); }} />
    <Button title="Audit system input" onPress={() => setInputMode('system')} />
    <Button title="Audit Add navigation draft" onPress={() => router.push('/audit-add-push' as Href)} />
    <Button title="Audit Add configured header" onPress={() => router.push('/audit-add-header' as Href)} />
    <Button title="Audit Add destination" onPress={() => router.push('/audit-add-destination' as Href)} />
    <Button title="Audit Add draft" onPress={() => router.push('/audit-add' as Href)} />
    <Button title="Audit onboarding submission" onPress={() => setOnboardingSubmission(true)} />
    <Button title="Audit settings controls" onPress={() => setSettingsControls('scroll')} />
    <Button title="Audit settings controls without scrolling" onPress={() => setSettingsControls('fixed')} />
    <Button title="Audit settings collection" onPress={() => router.push('/audit-customization' as Href)} />
    <Button title="Audit settings editor" onPress={() => router.push('/audit-customization-editor' as Href)} />
    <Button title="Audit settings save recovery" onPress={() => router.push('/audit-customization-editor?recovery=save' as Href)} />
    {['direct', 'nested', 'footer', 'direct-footer', 'scroll-footer'].map(variant => <Button key={variant} title={`Audit ${variant} sheet`}
      onPress={() => router.push({ pathname: '/audit-sheet-diagnostic', params: { variant } } as Href)} />)}
    <Button title="Audit Checkout history" onPress={() => router.push('/audit-checkout-history' as Href)} />
    <Button title="Audit draft photos" onPress={() => setDraftPhotos(true)} />
    <Button title="Audit plain input" onPress={() => setInputMode('plain')} />
    <Button title="Audit native-default input" onPress={() => setInputMode('native-default')} />
    <Button title="Audit plain-controlled input" onPress={() => setInputMode('plain-controlled')} />
    <Button title="Audit plain-controlled-no-assistance input" onPress={() => setInputMode('plain-controlled-no-assistance')} />
    <Button title="Audit plain-controlled-no-accessory input" onPress={() => { setKeyboardAccessoryEnabled(false); setInputMode('plain-controlled-no-accessory'); }} />
    <Button title="Audit plain-no-assistance input" onPress={() => setInputMode('plain-no-assistance')} />
    <Button title="Audit plain-no-accessory input" onPress={() => { setKeyboardAccessoryEnabled(false); setInputMode('plain-no-accessory'); }} />
    <Button title="Audit multiline input" onPress={() => setInputMode('multiline')} />
    <Button title="Audit photo removal recovery" onPress={() => setPhotoRecovery('removal')} />
    <Button title="Audit unavailable photo" onPress={() => setPhotoRecovery('missing')} />
    <Button title="Audit footer appearance" onPress={() => router.push('/audit-footer-appearance' as Href)} />
    <Button title="Audit menu ownership" onPress={() => router.push('/audit-menu-ownership' as Href)} />
    <Button title="Audit command height" onPress={() => router.push('/audit-command-height' as Href)} />
    <Button title="Audit Move destination" onPress={() => router.push('/audit-move-destination' as Href)} />
    <Button title="Audit Move here recovery" onPress={() => router.push('/audit-move-here-recovery' as Href)} />
    <Button title="Audit asset Edit journey" onPress={() => router.push('/audit-edit-journey' as Href)} />
    <Button title="Audit Edit tags" onPress={() => router.push('/audit-edit-tags' as Href)} />
    <Button title="Audit Edit recovery" onPress={() => router.push('/audit-edit-recovery' as Href)} />
    <Button title="Audit contents recovery" onPress={() => router.push('/audit-region-recovery' as Href)} />
    <Button title="Audit place search" onPress={() => router.push('/audit-contents-search' as Href)} />
    <Button title="Audit detail commands" onPress={() => router.push('/audit-detail-commands' as Href)} />
    <Text>{result}</Text>
  </FixturePage>;
}

function FixturePage({ children, persistHandledTaps = false, scrollEnabled = true }: { readonly children: ReactNode; readonly persistHandledTaps?: boolean; readonly scrollEnabled?: boolean }) {
  return <ScrollView scrollEnabled={scrollEnabled} keyboardShouldPersistTaps={persistHandledTaps ? 'handled' : 'never'} contentInsetAdjustmentBehavior="automatic" contentContainerStyle={{ padding: 20, gap: 20 }}>{children}</ScrollView>;
}

export function BrowseFilterGeometryFixture() { return <BrowseFilterFixture geometry />; }

export function BrowseFilterFixture({ geometry = false }: { readonly geometry?: boolean } = {}) {
  const router = useRouter();
  const { setResult } = useContext(ResultContext);
  return <><BrowseFiltersScreen initial={{ scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [], sort: 'updated_desc' }}
    query="" tags={[{ id: 'audit-tools', key: 'tools', label: 'Tools' }, { id: 'audit-holiday', key: 'holiday', label: 'Holiday supplies' }, ...Array.from({ length: 30 }, (_, index) => ({ id: `audit-tag-${index}`, key: `audit-tag-${index}`, label: `Long list tag ${String(index + 1).padStart(2, '0')}` })), { id: 'audit-last', key: 'audit-last', label: 'ZZ final tag' }]}
    onApply={draft => { setResult(draft.tagIds.length ? `Browse selected tags: ${draft.tagIds.join(',')}` : `Browse availability: ${draft.checkoutState}`); router.back(); }}
    onCancel={() => returnToPreviousOrHome(router)}
    onExpiration={mode => { setResult(`Expiration mode: ${mode}`); router.back(); }} />{geometry ? <FilterGeometryProbe /> : null}</>;
}

export function ExpirationFilterGeometryFixture() { return <ExpirationFilterFixture geometry />; }

export function ExpirationFilterFixture({ geometry = false }: { readonly geometry?: boolean } = {}) {
  const router = useRouter();
  const { setResult } = useContext(ResultContext);
  return <><ExpirationFiltersScreen initial={{ mode: 'all', throughDate: '2026-10-15' }} choices={{
    types: [{ id: 'audit-food', label: 'Food' }],
    tags: [{ id: 'audit-tools', label: 'Tools' }, { id: 'audit-holiday', label: 'Holiday supplies' }],
    locations: [{ id: 'audit-kitchen', label: 'Kitchen / Cabinet' }, { id: 'audit-garage', label: 'Garage / Cabinet' }]
  }} onApply={filter => { setResult(`Expiration mode: ${filter.mode}`); router.back(); }} onCancel={() => returnToPreviousOrHome(router)} />{geometry ? <FilterGeometryProbe /> : null}</>;
}

function DraftOptionsFixture() {
  const [options, setOptions] = useState<readonly string[]>(['saved', 'draft']);
  const [newOption, setNewOption] = useState('');
  return <CustomizationFieldControls applicability="all_assets" canMutate eligibleTypes={[]}
    enumOptions={options} persistedEnumOptions={['saved']} fieldType="enum" mode="edit"
    newOption={newOption} onNewOption={setNewOption} onEnumOptions={setOptions}
    persistedTargetIds={[]} targetIds={[]} onTargets={() => {}} onApplicability={() => {}} onFieldType={() => {}} />;
}

function InputFixture({ mode }: { readonly mode: InputFixtureMode }) {
  const [value, setValue] = useState('');
  const trace = useInputEventTrace(value);
  if (mode === 'native-default') return <View>
    <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 54 }}>
      <TextField defaultValue="" onValueChange={setValue}
        modifiers={[accessibilityLabel('Audit native-default text'), textFieldStyle('roundedBorder')]} />
    </Host>
    <Text>{`Observed native-default input: ${value}`}</Text>
  </View>;
  if (mode.startsWith('plain') || mode === 'multiline') return <View>
    <AppTextInput accessibilityLabel={`Audit ${mode} text`} multiline={mode === 'multiline'}
      {...(mode.startsWith('plain-controlled') ? { value } : { defaultValue: '' })}
      {...(mode.endsWith('no-assistance') ? { autoCorrect: false, spellCheck: false, smartInsertDelete: false } : {})}
      onChangeText={setValue} onChange={trace.onChange} onKeyPress={trace.onKeyPress} onSelectionChange={trace.onSelectionChange} style={{ minHeight: mode === 'multiline' ? 160 : 54, borderWidth: 1, padding: 12 }} />
    <Text>{`Observed ${mode} input: ${value}`}</Text>
    {trace.controls}
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
    onChange={trace.onChange} onKeyPress={trace.onKeyPress} onSelectionChange={trace.onSelectionChange}
    {...(mode === 'controlled' ? { value } : { defaultValue: '' })}
    style={{ minHeight: 54, borderWidth: 1, padding: 12 }} />
    <Text>{`Observed ${mode} input: ${value}`}</Text>
    {trace.controls}
  </View>;
}


function SettingsControlsFixture({ onBack, scrollEnabled }: { readonly onBack: () => void; readonly scrollEnabled: boolean }) {
  const { preference } = useAppearance();
  const [color, setColor] = useState('');
  const [colorLocked, setColorLocked] = useState(false);
  const [expiration, setExpiration] = useState('No expiration');
  const [reminder, setReminder] = useState<ExpirationReminderPolicy | null>(null);
  return <FixturePage scrollEnabled={scrollEnabled}>
    <Button title="Back to audit menu" onPress={onBack} />
    <AppearancePicker />
    <Text>{`Appearance value: ${preference}`}</Text>
    <Text>{`Appearance writes: ${useContext(AppearanceWriteCount)}`}</Text>
    <TagColorPicker disabled={colorLocked} value={color} onChange={setColor} />
    <Text>{`Color value: ${color || 'none'}`}</Text>
    <Button title={colorLocked ? 'Unlock color editing' : 'Lock color editing'} onPress={() => setColorLocked(locked => !locked)} />
    <ExpirationField initialPickerDate={new Date(2026, 8, 14, 12)} onChange={value => setExpiration(value?.date ?? 'No expiration')} />
    <Text>{`Expiration value: ${expiration}`}</Text>
    <ExpirationReminderEditor initialPolicy={reminder}
      inheritedPolicy={{ enabled: true, upcoming: true, expired: true, advanceDays: 7 }}
      onSave={async value => setReminder(value)} onEditDays={() => {}} />
    <Text>{`Reminder mode: ${reminder === null ? 'defaults' : reminder.enabled ? 'custom' : 'off'}`}</Text>
    {Platform.OS === 'android' ? <AndroidControlTargetsFixture /> : null}
  </FixturePage>;
}


function OnboardingSubmissionFixture() {
  const [fakes] = useState(onboardingFakes);
  const [command] = useState(() => new OnboardingCommand(fakes.profiles, () => fakes.api, fakes.auth));
  const [state, setState] = useState<OnboardingStartState>({ step: 'instance' });
  return <View style={{ flex: 1 }}>
    <Stack.Screen options={{ headerShown: false }} />
    <OnboardingScreen command={command} initialState={state} onStateChange={setState} onComplete={() => {}} />
    <Text pointerEvents="none" style={{ position: 'absolute', bottom: 0 }}>{`Submitted address: ${fakes.auth.signIns.at(-1) ?? 'none'}`}</Text>
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
