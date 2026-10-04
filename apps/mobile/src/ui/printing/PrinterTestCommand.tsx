import { useRef, useState } from 'react';
import { Text } from 'react-native';
import type { PrintScope, PrintTemplate, PrintingWorkspace, RegisteredPrinter } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { SettingsActionRow } from '../screens/SettingsList';
import { useSettingsListStyles } from '../screens/SettingsList';

export function PrinterTestCommand({ workspace, scope, printer, template, lifetime, onQueued }: {
  readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly printer: RegisteredPrinter; readonly template: PrintTemplate;
  readonly lifetime: { current: AbortController | undefined }; readonly onQueued: (id: string) => void;
}) {
  const { styles } = useSettingsListStyles(); const request = useRef(workspace.requests.forTest(scope, printer.id));
  const [busy, setBusy] = useState(false); const [failed, setFailed] = useState(false); const running = useRef(false);
  const print = async () => {
    const owner = lifetime.current;
    if (!owner || owner.signal.aborted || running.current || (printer.retired && !request.current.locked)) return;
    running.current = true; setBusy(true); setFailed(false);
    try {
      const job = request.current.locked ? await request.current.retry() : await request.current.submit(scope, printer.id, { printerId: printer.id, mediaFingerprint: printer.mediaFingerprint, template, copies: 1 });
      if (!owner.signal.aborted) { workspace.requests.settled(scope, printer.id, 'test'); request.current = workspace.requests.forTest(scope, printer.id); onQueued(job.id); }
    } catch { if (!owner.signal.aborted) setFailed(true); }
    finally { running.current = false; setBusy(false); }
  };
  if (printer.retired && !request.current.locked) return null;
  return <>
    {failed ? <Text accessibilityRole="alert" style={styles.errorMessage}>{t(request.current.locked ? 'printing.mobile.unknownSubmission' : 'printing.mobile.unavailable')}</Text> : null}
    <SettingsActionRow label={t(request.current.locked ? 'printing.mobile.retryTest' : 'printing.mobile.testLabel')} disabled={busy} onPress={() => void print()} />
  </>;
}
