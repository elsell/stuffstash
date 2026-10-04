import { createContext, useContext, useMemo, useState } from 'react';
import { ScrollView } from 'react-native';
import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import { createCompletePrintingWorkspace, createCompleteLabelWorkspace, completePrintScope as scope } from './CompletePrintingRepository';
import type { PrintingWorkspace } from '../src/application/printing/PrintingWorkspace';
import { AppearancePreferenceController } from '../src/application/settings/AppearancePreference';
import { AppearanceProvider, useAppearancePalette } from '../src/ui/theme/AppearanceContext';
import { AppKeyboardProvider } from '../src/ui/components/AppKeyboardProvider';
import { AppFeedbackProvider } from '../src/ui/feedback/AppFeedback';
import { AppNoticeScreenLayout } from '../src/ui/feedback/AppNoticeScreenLayout';
import { createAssetNativeSheetOptions } from '../src/ui/screens/AssetNativeSheetOptions';
import { SettingsSection, SettingsActionRow, useSettingsListStyles } from '../src/ui/screens/SettingsList';
import { LabelTaskHeader } from '../src/ui/labels/LabelTaskHeader';
import { AssetLabelTask } from '../src/ui/labels/AssetLabelTask';
import { PrintJobScreen } from '../src/ui/printing/PrintJobScreen';
import { ReprintLabelScreen } from '../src/ui/printing/ReprintLabelScreen';

const Workspace = createContext<PrintingWorkspace | undefined>(undefined);
const appearance = new AppearancePreferenceController({ load: async () => 'dark', save: async () => {} });
function useWorkspace() { const value = useContext(Workspace); if (!value) throw new Error('Printing audit services missing'); return value; }
export function CompletePrintingLayout() {
  const [workspace] = useState(createCompletePrintingWorkspace);
  return <AppKeyboardProvider><AppearanceProvider controller={appearance}><Workspace.Provider value={workspace}><CompletePrintingNavigation /></Workspace.Provider></AppearanceProvider></AppKeyboardProvider>;
}
function CompletePrintingNavigation() {
  const palette = useAppearancePalette(); const sheets = createAssetNativeSheetOptions(palette);
  return <AppFeedbackProvider noticePlacement="screen"><Stack screenLayout={AppNoticeScreenLayout} screenOptions={{ headerTintColor: palette.action, headerStyle: { backgroundColor: palette.surface }, headerTitleStyle: { color: palette.text, fontWeight: '700' }, contentStyle: { backgroundColor: palette.background } }}>
    <Stack.Screen name="index" options={{ title: 'Printing acceptance' }} />
    <Stack.Screen name="assets/[assetId]/print" options={{ ...sheets.add, title: 'Print label' }} />
    <Stack.Screen name="print-jobs/[jobId]" options={{ title: 'Print job' }} />
  </Stack></AppFeedbackProvider>;
}
export function CompletePrintingHome() {
  const router = useRouter(); const { styles } = useSettingsListStyles();
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic"><SettingsSection>
    <SettingsActionRow label="Open label" onPress={() => router.push('/assets/printing-item/print' as never)} />
    <SettingsActionRow label="Inspect uncertain job" onPress={() => router.push('/print-jobs/uncertain' as never)} />
  </SettingsSection></ScrollView>;
}
export function CompletePrintingTask() {
  const router = useRouter(); const workspace = useWorkspace();
  const labels = useMemo(() => createCompleteLabelWorkspace(workspace), [workspace]);
  return <><LabelTaskHeader /><AssetLabelTask printing={workspace} labels={labels} scope={scope} assetId="printing-item" onQueued={jobId => router.replace({ pathname: '/print-jobs/[jobId]', params: { jobId } } as never)} /></>;
}
export function CompletePrintingJob() {
  const router = useRouter(); const workspace = useWorkspace(); const { jobId, action } = useLocalSearchParams<{ jobId: string; action?: string }>();
  const queued = (id: string) => router.replace({ pathname: '/print-jobs/[jobId]', params: { jobId: id } } as never);
  return action === 'reprint' ? <><Stack.Screen options={{ title: 'Reprint label' }} /><LabelTaskHeader /><ReprintLabelScreen workspace={workspace} scope={scope} jobId={jobId} onQueued={queued} /></>
    : <><Stack.Screen options={{ title: 'Print job' }} /><PrintJobScreen workspace={workspace} scope={scope} jobId={jobId} canPrint onReprint={id => router.push({ pathname: '/print-jobs/[jobId]', params: { jobId: id, action: 'reprint' } } as never)} /></>;
}
