import { PrintResolutionControls, printOutcomeLabel, type PendingResolution } from './PrintResolutionControls';
import { useCallback, useRef, useState } from 'react';
import { ScrollView, Text } from 'react-native';
import type { PrintJob, PrintScope, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsLoadingRow, useSettingsListStyles } from '../screens/SettingsList';
import { printJobStatus } from './PrintingStatus';
import { usePrintingTask } from './usePrintingTask';
export function PrintJobScreen({ workspace, scope, jobId, canPrint }: { readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly jobId: string; readonly canPrint: boolean }) {
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
  const task = usePrintingTask(load, jobId, 5000); const job = task.data;
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
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content}>
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error ? <Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text> : null}
    {job ? <>
      <Text accessibilityRole="header" accessibilityLiveRegion="polite" style={styles.detailTitle}>{job.resolution ? t('printing.mobile.resolved') : printJobStatus(job.status)}</Text>
      <Text style={{ color: palette.text }}>{t('printing.mobile.copyProgress', { completed: job.completedCopies, total: job.copies })}</Text>
      {job.status === 'uncertain' ? <Text style={{ color: palette.text }}>{t('printing.mobile.uncertain')}</Text> : null}
      {job.resolution ? <><Text style={{ color: palette.text }}>{printOutcomeLabel(job.resolution.reportedOutcome)}</Text><Text style={{ color: palette.textMuted }}>{t('printing.mobile.resolvedDetail')}</Text></> : null}
      {canPrint && job.status === 'uncertain' ? <PrintResolutionControls key={resolutionGeneration.current} pending={pendingResolution} workspace={workspace} scope={scope} job={job} lifetime={task.lifetime} onResolved={resolved} /> : null}
      {canPrint && ['queued', 'claimed'].includes(job.status) ? <NativeCommandButton label={t('printing.mobile.cancel')} disabled={busy} role="destructive" onPress={() => void cancel()} /> : null}
    </> : null}
    {failed ? <Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text> : null}
    <NativeCommandButton label={t('printing.mobile.refresh')} disabled={busy || task.loading} onPress={refresh} />
  </ScrollView>;
}
