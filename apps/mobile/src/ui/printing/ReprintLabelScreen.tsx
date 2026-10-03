import { useCallback } from 'react';
import { Text } from 'react-native';
import { canReprint, type PrintScope, type PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { AssetPrintScreen } from './AssetPrintScreen';
import { usePrintingTask } from './usePrintingTask';

export function ReprintLabelScreen({ workspace, scope, jobId, onQueued }: {
  readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly jobId: string; readonly onQueued: (id: string) => void;
}) {
  const load = useCallback((signal: AbortSignal) => workspace.repository.job(scope, jobId, signal), [workspace, scope.tenantId, scope.inventoryId, jobId]);
  const task = usePrintingTask(load, jobId); const job = task.data;
  // A retained request can be recovered after predecessor retention expires.
  const retained = workspace.requests.forReprint(scope, jobId).locked;
  if (retained || (job && canReprint(job) && (job.assetId || job.kind === 'printer_test'))) {
    return <AssetPrintScreen workspace={workspace} scope={scope} assetId={job?.assetId ?? ''} predecessorId={jobId} diagnostic={job?.kind === 'printer_test'} onQueued={onQueued} />;
  }
  return <><Text accessibilityRole={task.error ? 'alert' : undefined}>{t(task.loading ? 'printing.mobile.loading' : job ? 'printing.mobile.reprintUnavailable' : 'printing.mobile.unavailable')}</Text>
    <NativeCommandButton label={t('printing.mobile.refresh')} disabled={task.loading} onPress={task.reload} /></>;
}
