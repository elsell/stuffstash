import { useCallback } from 'react';
import { ScrollView, Text } from 'react-native';
import { canReprint, type PrintScope, type PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { SettingsActionRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { AssetPrintScreen, type AssetPrintDraft } from './AssetPrintScreen';
import { usePrintingTask } from './usePrintingTask';

export function ReprintLabelScreen({ workspace, scope, jobId, draftState, onQueued }: {
  readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly jobId: string; readonly draftState?: AssetPrintDraft; readonly onQueued: (id: string) => void;
}) {
  const { styles } = useSettingsListStyles();
  const load = useCallback((signal: AbortSignal) => workspace.repository.job(scope, jobId, signal), [workspace, scope.tenantId, scope.inventoryId, jobId]);
  const task = usePrintingTask(load, jobId); const job = task.data;
  // A retained request can be recovered after predecessor retention expires.
  const retained = workspace.requests.forReprint(scope, jobId).locked;
  if (retained || (job && canReprint(job) && (job.assetId || job.kind === 'printer_test'))) {
    return <AssetPrintScreen draftState={draftState} workspace={workspace} scope={scope} assetId={job?.assetId ?? ''} predecessorId={jobId} diagnostic={job?.kind === 'printer_test'} onQueued={onQueued} />;
  }
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic"><SettingsSection><Text style={styles.sectionFooter} accessibilityRole={task.error ? 'alert' : undefined}>{t(task.loading ? 'printing.mobile.loading' : job ? 'printing.mobile.reprintUnavailable' : 'printing.mobile.unavailable')}</Text>
    <SettingsActionRow label={t('printing.mobile.refresh')} disabled={task.loading} onPress={task.reload} /></SettingsSection></ScrollView>;
}
