import { useCallback } from 'react';
import { ScrollView, Text } from 'react-native';
import type { PrintScope, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { SettingsActionRow, SettingsLoadingRow, SettingsNavigationRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { printJobStatus } from './PrintingStatus';
import { usePrintingTask } from './usePrintingTask';

export function PrintHistoryScreen({ workspace, scope, onJob }: { readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly onJob: (id: string) => void }) {
  const { styles } = useSettingsListStyles();
  const load = useCallback(async (signal: AbortSignal) => ({ catalog: await workspace.repository.catalog(scope, signal), jobs: await workspace.repository.jobs(scope, signal) }), [workspace, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, `${scope.tenantId}/${scope.inventoryId}`);
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic">
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error ? <SettingsSection><Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text><SettingsActionRow label={t('printing.mobile.retry')} onPress={task.reload} /></SettingsSection> : null}
    {task.data ? <><SettingsSection>
      {task.data.jobs.length ? task.data.jobs.map(job => <SettingsNavigationRow key={job.id} label={printJobStatus(job.status)} accessibilityLabel={`${printJobStatus(job.status)} · ${job.id}`} context={task.data!.catalog.printers.find(printer => printer.id === job.printerId)?.name} onPress={() => onJob(job.id)} />)
        : <Text style={[styles.navigationRow, styles.secondaryText]}>{t('printing.mobile.noJobs')}</Text>}
    </SettingsSection><SettingsSection><SettingsActionRow label={t('printing.mobile.refresh')} onPress={task.reload} /></SettingsSection></> : null}
  </ScrollView>;
}
