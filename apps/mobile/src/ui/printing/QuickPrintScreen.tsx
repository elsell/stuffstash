import { useCallback, useEffect, useRef, useState } from 'react';
import { ScrollView, Text } from 'react-native';
import type { PrintScope, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsLoadingRow, useSettingsListStyles } from '../screens/SettingsList';
import { AssetPrintScreen } from './AssetPrintScreen';
import { usePrintingTask } from './usePrintingTask';
/** Mounted only by an explicit Print label command, never by status refresh. */
export function QuickPrintScreen({ workspace, scope, assetId, onQueued }: {
  readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly assetId: string; readonly onQueued: (id: string) => void;
}) {
  const { styles } = useSettingsListStyles();
  const load = useCallback((signal: AbortSignal) => workspace.repository.catalog(scope, signal), [workspace, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, assetId); const catalog = task.data;
  const request = useRef(workspace.requests.forAsset(scope, assetId));
  const attempted = useRef(request.current.locked); const running = useRef(false);
  const [busy, setBusy] = useState(false); const [failed, setFailed] = useState(false); const [options, setOptions] = useState(false);
  const run = async () => {
    const owner = task.lifetime.current;
    if (!catalog || !owner || owner.signal.aborted || running.current) return;
    running.current = true; setBusy(true); setFailed(false);
    try {
      let job;
      if (request.current.locked) job = await request.current.retry();
      else {
        const printer = catalog.printers.find(value => value.id === catalog.settings.defaultPrinterId && !value.retired);
        const template = catalog.settings.template;
        if (!printer || !catalog.templates.some(value => value.id === template.id && value.version === template.version)) { setOptions(true); return; }
        let preview;
        try { preview = await workspace.repository.preview(scope, assetId, printer, template, owner.signal); }
        catch { if (!owner.signal.aborted) setOptions(true); return; }
        if (owner.signal.aborted) return;
        job = await request.current.submit(scope, assetId, { printerId: printer.id, mediaFingerprint: printer.mediaFingerprint, template, copies: 1, previewFingerprint: preview.fingerprint });
      }
      if (!owner.signal.aborted) { workspace.requests.settled(scope, assetId); onQueued(job.id); }
    } catch { if (!owner.signal.aborted) setFailed(true); }
    finally { running.current = false; setBusy(false); }
  };
  useEffect(() => { if (catalog && !attempted.current) { attempted.current = true; void run(); } }, [catalog]);
  if (options) return <AssetPrintScreen workspace={workspace} scope={scope} assetId={assetId} onQueued={onQueued} />;
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content}>
    {task.loading || busy ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error || failed ? <Text accessibilityRole="alert" style={styles.errorMessage}>{t(request.current.locked ? 'printing.mobile.unknownSubmission' : 'printing.mobile.unavailable')}</Text> : null}
    {task.error ? <NativeCommandButton label={t('printing.mobile.retry')} onPress={task.reload} /> : null}
    {catalog && !busy ? <NativeCommandButton label={t('printing.mobile.retry')} onPress={() => void run()} /> : null}
    {catalog && !busy && !request.current.locked ? <NativeCommandButton label={t('printing.mobile.options')} onPress={() => setOptions(true)} /> : null}
  </ScrollView>;
}
