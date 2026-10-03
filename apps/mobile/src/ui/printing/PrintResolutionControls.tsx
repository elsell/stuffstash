import { useRef, useState } from 'react';
import { Text } from 'react-native';
import type { PrintJob, PrintOutcome, PrintScope, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { AppSwitchField } from '../components/AppSwitchField';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { useSettingsListStyles } from '../screens/SettingsList';

export type PendingResolution = { job: PrintJob; outcome: PrintOutcome };

export function PrintResolutionControls({ workspace, scope, job, lifetime, pending, onResolved }: {
  readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly job: PrintJob;
  readonly pending: { current: PendingResolution | undefined };
  readonly lifetime: { current: AbortController | undefined }; readonly onResolved: (job: PrintJob) => void;
}) {
  const { styles, palette } = useSettingsListStyles();
  const [outcome, setOutcome] = useState<PrintOutcome | ''>(pending.current?.outcome ?? ''); const [acknowledged, setAcknowledged] = useState(!!pending.current);
  const [busy, setBusy] = useState(false); const [failed, setFailed] = useState(false); const running = useRef(false);
  const locked = busy || !!pending.current;
  const resolve = async () => {
    const owner = lifetime.current;
    if (!owner || owner.signal.aborted || running.current || !job.idleConfirmed || job.status !== 'uncertain' || !outcome || !acknowledged) return;
    pending.current ??= { job, outcome };
    running.current = true; setBusy(true); setFailed(false);
    try {
      const next = await workspace.repository.resolve(scope, pending.current.job, pending.current.outcome);
      if (!owner.signal.aborted) onResolved(next);
    } catch { if (!owner.signal.aborted) setFailed(true); }
    finally { running.current = false; setBusy(false); }
  };
  return <>
    {!job.idleConfirmed ? <Text style={{ color: palette.text }}>{t('printing.mobile.waitingForIdle')}</Text> : null}
    <SettingsPickerRow label={t('printing.mobile.outcome')} accessibilityLabel={t('printing.mobile.outcome')} value={outcome} disabled={locked || !job.idleConfirmed}
      options={[{ value: '', label: t('printing.mobile.outcome.choose') }, { value: 'printed', label: t('printing.mobile.outcome.printed') }, { value: 'not_printed', label: t('printing.mobile.outcome.not_printed') }, { value: 'unknown', label: t('printing.mobile.outcome.unknown') }]}
      onChange={setOutcome} />
    <AppSwitchField label={t('printing.mobile.acknowledge')} value={acknowledged} disabled={locked || !job.idleConfirmed} onValueChange={setAcknowledged} />
    {failed ? <Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text> : null}
    <NativeCommandButton label={t(pending.current ? 'printing.mobile.resolveRetry' : 'printing.mobile.resolve')} disabled={busy || !job.idleConfirmed || !outcome || !acknowledged} onPress={() => void resolve()} />
  </>;
}

export function printOutcomeLabel(outcome: PrintOutcome) {
  return t(outcome === 'printed' ? 'printing.mobile.outcome.printed' : outcome === 'not_printed' ? 'printing.mobile.outcome.not_printed' : 'printing.mobile.outcome.unknown');
}
