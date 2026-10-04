import { canReprint } from '../../application/printing/PrintingWorkspace';
import { PrintResolutionControls, printOutcomeLabel, type PendingResolution } from './PrintResolutionControls';
import { useCallback, useRef, useState } from 'react';
import { ScrollView, Text, View } from 'react-native';
import type { PrintJob, PrintScope, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { SettingsActionRow, SettingsSection, SettingsLoadingRow, useSettingsListStyles } from '../screens/SettingsList';
import { printJobStatus } from './PrintingStatus';
import { usePrintingTask, type PrintingPolling } from './usePrintingTask';
const jobPolling: PrintingPolling<PrintJob> = {
  intervalMilliseconds: 5000, maximumDelayMilliseconds: 60000,
  shouldContinue: job => !canReprint(job),
};
export function PrintJobScreen({ workspace, scope, jobId, canPrint, onReprint }: { readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly jobId: string; readonly canPrint: boolean; readonly onReprint?: (id: string) => void }) {
  const { styles, palette } = useSettingsListStyles();
  const pendingResolution = useRef<PendingResolution | undefined>(undefined);
  const explicitRefresh = useRef(false); const resolutionGeneration = useRef(0);
  const load = useCallback(async (signal: AbortSignal) => {
    const explicitlyRequested = explicitRefresh.current; explicitRefresh.current = false;
    const next = await workspace.repository.job(scope, jobId, signal);
    const pending = pendingResolution.current;
    if (!signal.aborted && explicitlyRequested && pending && (next.revision !== pending.job.revision || next.latestAttemptId !== pending.job.latestAttemptId)) {
      pendingResolution.current = undefined; resolutionGeneration.current++;
    }
    return next;
  }, [workspace, scope.tenantId, scope.inventoryId, jobId]);
  const task = usePrintingTask(load, jobId, jobPolling); const job = task.data;
  const [busy, setBusy] = useState(false); const running = useRef(false); const [failed, setFailed] = useState(false);
  const resolved = (next: PrintJob) => { pendingResolution.current = undefined; task.setData(next); };
  const refresh = () => { explicitRefresh.current = true; task.reload(); };
  const cancel = async () => {
    const owner = task.lifetime.current;
    if (!owner || owner.signal.aborted || !job || !canPrint || running.current || !['queued', 'claimed'].includes(job.status)) return;
    running.current = true; setBusy(true); setFailed(false);
    try { const next = await workspace.repository.cancel(scope, job); if (!owner.signal.aborted) task.setData(next); }
    catch { if (!owner.signal.aborted) setFailed(true); }
    finally { running.current = false; setBusy(false); }
  };
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic">
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error ? <Text accessibilityRole="alert" style={[styles.sectionFooter, styles.dangerText]}>{t('printing.mobile.unavailable')}</Text> : null}
    {job ? <>
      <View style={styles.detailHeader}><Text accessibilityRole="header" accessibilityLiveRegion="polite" style={styles.detailTitle}>{job.resolution ? t('printing.mobile.resolved') : printJobStatus(job.status)}</Text>
      <Text style={styles.detailSubtitle}>{t('printing.mobile.copyProgress', { completed: job.completedCopies, total: job.copies })}</Text>
      {job.status === 'uncertain' ? <Text style={{ color: palette.text }}>{t('printing.mobile.uncertain')}</Text> : null}
      {job.resolution ? <><Text style={{ color: palette.text }}>{printOutcomeLabel(job.resolution.reportedOutcome)}</Text><Text style={{ color: palette.textMuted }}>{t('printing.mobile.resolvedDetail')}</Text></> : null}
      </View>{canPrint && job.status === 'uncertain' ? <PrintResolutionControls key={resolutionGeneration.current} pending={pendingResolution} workspace={workspace} scope={scope} job={job} lifetime={task.lifetime} onResolved={resolved} /> : null}
      {canPrint && ((onReprint && canReprint(job)) || ['queued', 'claimed'].includes(job.status)) ? <SettingsSection>{canPrint && onReprint && canReprint(job) ? <SettingsActionRow label={t('printing.mobile.reprint')} onPress={() => onReprint(job.id)} /> : null}
      {canPrint && ['queued', 'claimed'].includes(job.status) ? <SettingsActionRow label={t('printing.mobile.cancel')} disabled={busy} destructive onPress={() => void cancel()} /> : null}
      </SettingsSection> : null}
    </> : null}
    {failed ? <Text accessibilityRole="alert" style={[styles.sectionFooter, styles.dangerText]}>{t('printing.mobile.unavailable')}</Text> : null}
    <SettingsSection><SettingsActionRow label={t('printing.mobile.refresh')} disabled={busy || task.loading} onPress={refresh} /></SettingsSection>
  </ScrollView>;
}
