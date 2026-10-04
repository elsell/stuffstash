import { createContext, useContext, useState } from 'react';
import { Button, ScrollView, Text } from 'react-native';
import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import { parseMobileLabelLink } from '../src/adapters/labels/LabelLinkParser';
import { OpenLabel } from '../src/application/labels/OpenLabel';
import { createScannerRepository, scannerLink, scannerTarget } from './LabelScannerRepository';
import { AppearancePreferenceController } from '../src/application/settings/AppearancePreference';
import { AppKeyboardProvider } from '../src/ui/components/AppKeyboardProvider';
import { AppFeedbackProvider } from '../src/ui/feedback/AppFeedback';
import { AppNoticeScreenLayout } from '../src/ui/feedback/AppNoticeScreenLayout';
import { LabelLinkProvider, usePendingLabel } from '../src/ui/labels/LabelLinkContext';
import { PendingLabelNavigation } from '../src/ui/labels/PendingLabelNavigation';
import { createAssetNativeSheetOptions } from '../src/ui/screens/AssetNativeSheetOptions';
import { AppearanceProvider, useAppearancePalette } from '../src/ui/theme/AppearanceContext';

function createServices(changed: () => void) {
  const repository = createScannerRepository(changed);
  return { serviceScopeId: 'native-scanner-fixture', labels: { parse: parseMobileLabelLink }, repository,
    openLabel: new OpenLabel(repository, async () => repository.selected()) };
}
const Context = createContext<ReturnType<typeof createServices> | undefined>(undefined);
export function useAppServices() { const value = useContext(Context); if (!value) throw new Error('Scanner fixture service missing'); return value; }
export function useAppConnectionActions() { return { changeServer: async () => {}, signOut: async () => {} }; }
const ReadyContext = createContext({ ready: true, setReady: (_ready: boolean) => {} });
const appearance = new AppearancePreferenceController({ load: async () => 'system', save: async () => {} });
export function LabelScannerLayout() {
  const [, redraw] = useState(0); const [services] = useState(() => createServices(() => redraw(value => value + 1)));
  const [ready, setReady] = useState(true);
  return <AppKeyboardProvider><AppearanceProvider controller={appearance}><LabelLinkProvider>
    <Context.Provider value={services}><ReadyContext.Provider value={{ ready, setReady }}><ScannerNavigation /></ReadyContext.Provider></Context.Provider>
  </LabelLinkProvider></AppearanceProvider></AppKeyboardProvider>;
}
function ScannerNavigation() {
  const palette = useAppearancePalette(); const sheets = createAssetNativeSheetOptions(palette); const { ready } = useContext(ReadyContext);
  return <AppFeedbackProvider noticePlacement="screen"><>
    {ready ? <PendingLabelNavigation /> : null}
    <Stack screenLayout={AppNoticeScreenLayout} screenOptions={{ headerTintColor: palette.action, contentStyle: { backgroundColor: palette.background } }}>
      <Stack.Screen name="index" options={{ title: 'Scanner acceptance' }} />
      <Stack.Screen name="scan-label" options={{ ...sheets.selection, title: 'Scan label' }} />
      <Stack.Screen name="assets/[assetId]/index" options={{ title: 'Resolved asset' }} />
    </Stack>
  </></AppFeedbackProvider>;
}
export function LabelScannerHome() {
  const router = useRouter(); const services = useAppServices(); const pending = usePendingLabel(); const { ready, setReady } = useContext(ReadyContext);
  return <ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={{ padding: 20, gap: 16 }}>
    <Text>Scanner acceptance home</Text>
    <Button title="Open scanner" onPress={() => router.push('/scan-label')} />
    <Button title="Complete delayed lookup" onPress={() => services.repository.complete()} />
    <Text testID="scanner-evidence">{JSON.stringify(services.repository.evidence)}</Text>
    <Button title="Capture label before simulated sign-in" onPress={() => { setReady(false); pending.capture(scannerLink); }} />
    {!ready ? <><Text>Simulated sign-in required</Text><Button title="Complete simulated sign-in" onPress={() => setReady(true)} /></> : null}
    <Text>{pending.reference ? 'Pending label retained' : 'No pending label'}</Text>
  </ScrollView>;
}
export function LabelScannerDestination() {
  const { assetId } = useLocalSearchParams<{ assetId: string }>(); const { repository } = useAppServices();
  return <ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={{ padding: 20 }}>
    <Text>{assetId === scannerTarget.assetId ? 'Cordless drill resolved' : 'Unexpected asset'}</Text>
    <Text testID="scanner-selected">Selected scope count: {repository.evidence.selected}</Text>
  </ScrollView>;
}
