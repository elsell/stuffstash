import { mediaSizeLabel } from './PrintingStatus';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Text } from 'react-native';
import type { PrintScope, PrintingWorkspace, RegisteredPrinter } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { SettingsLoadingRow, useSettingsListStyles } from '../screens/SettingsList';
import { usePrintingTask } from './usePrintingTask';

export function PrinterMediaSettings({ workspace, scope, printer, onSaved, onReload }: {
  readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly printer: RegisteredPrinter;
  readonly onSaved: (printer: RegisteredPrinter) => void; readonly onReload: () => void;
}) {
  const { palette, styles } = useSettingsListStyles();
  const load = useCallback((signal: AbortSignal) => workspace.repository.mediaPresets(scope, printer, signal), [workspace, scope.tenantId, scope.inventoryId, printer.id, printer.adapterId]);
  const task = usePrintingTask(load, printer.id);
  const [selected, setSelected] = useState(`${printer.media.presetId}/${printer.media.version}`);
  const [busy, setBusy] = useState(false); const running = useRef(false); const [status, setStatus] = useState<'idle' | 'saved' | 'failed'>('idle');
  useEffect(() => setSelected(`${printer.media.presetId}/${printer.media.version}`), [printer.revision, printer.media.presetId, printer.media.version]);
  const preset = task.data?.find(value => `${value.id}/${value.version}` === selected);
  const save = async () => {
    const owner = task.lifetime.current;
    if (!owner || owner.signal.aborted || running.current || !preset) return;
    running.current = true; setBusy(true); setStatus('idle');
    try { const next = await workspace.repository.configurePrinter(scope, printer, preset); if (!owner.signal.aborted) { setStatus('saved'); onSaved(next); } }
    catch { if (!owner.signal.aborted) setStatus('failed'); }
    finally { running.current = false; setBusy(false); }
  };
  return <>
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.data?.length ? <SettingsPickerRow label={t('labels.mobile.size')} accessibilityLabel={t('labels.mobile.size')} value={selected} disabled={busy}
      options={task.data.map(value => ({ value: `${value.id}/${value.version}`, label: mediaSizeLabel(value.name, value) }))} onChange={value => { setSelected(value); setStatus('idle'); }} /> : null}
    {task.error ? <><Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text><NativeCommandButton label={t('printing.mobile.reloadSizes')} onPress={task.reload} /></> : null}
    {!task.loading && task.data?.length === 0 ? <Text style={{ color: palette.textMuted }}>{t('printing.mobile.noSizes')}</Text> : null}
    <NativeCommandButton label={t('printing.mobile.saveSize')} disabled={busy || !preset} onPress={() => void save()} />
    {status !== 'idle' ? <Text accessibilityLiveRegion="polite" style={{ color: status === 'failed' ? palette.danger : palette.text }}>{t(status === 'failed' ? 'printing.mobile.sizeFailed' : 'printing.mobile.sizeSaved')}</Text> : null}
    {status === 'failed' ? <NativeCommandButton label={t('printing.mobile.reloadPrinter')} disabled={busy} onPress={onReload} /> : null}
  </>;
}
