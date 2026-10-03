import { useCallback, useEffect, useRef, useState } from 'react';
import { Image, ScrollView, Text, View, useWindowDimensions } from 'react-native';
import type { PrintScope, PrintTemplate, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import type { LabelFile } from '../../application/labels/LabelWorkspace';
import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { AppSwitchField } from '../components/AppSwitchField';
import { SettingsLoadingRow, useSettingsListStyles } from '../screens/SettingsList';
import { printerReadiness } from './PrintingStatus';
import { usePrintingTask } from './usePrintingTask';
type Preview = { uri: string; release(): void; file: LabelFile; fingerprint: string };
export function AssetPrintScreen({ workspace, scope, assetId, onQueued }: { readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly assetId: string; readonly onQueued: (id: string) => void }) {
  const { palette, styles } = useSettingsListStyles(); const dimensions = useWindowDimensions();
  const load = useCallback((signal: AbortSignal) => workspace.repository.catalog(scope, signal), [workspace, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, assetId); const catalog = task.data;
  const [printerId, setPrinterId] = useState(''); const [template, setTemplate] = useState<PrintTemplate>();
  const [preview, setPreview] = useState<Preview>(); const heldPreview = useRef<Preview | undefined>(undefined);
  const [busy, setBusy] = useState(false); const running = useRef(false); const [failed, setFailed] = useState(false);
  const submission = useRef(workspace.requests.forAsset(scope, assetId));
  const clearPreview = useCallback(() => { heldPreview.current?.release(); heldPreview.current = undefined; setPreview(undefined); }, []);
  useEffect(() => { if (catalog) { const pending = submission.current.selection; setPrinterId(pending?.printerId ?? catalog.settings.defaultPrinterId ?? ''); setTemplate(pending?.template ?? catalog.settings.template); } }, [catalog]);
  useEffect(() => () => heldPreview.current?.release(), []);
  // A retained preview is private content: release it as soon as the task read owner changes.
  useEffect(() => { if (!catalog) clearPreview(); }, [catalog, clearPreview]);
  const printer = catalog?.printers.find(item => item.id === printerId && !item.retired);
  const locked = busy || submission.current.locked;
  const render = async () => {
    const owner = task.lifetime.current;
    if (!owner || owner.signal.aborted || !printer || !template || running.current || submission.current.locked) return;
    running.current = true; setBusy(true); setFailed(false); clearPreview();
    try {
      const result = await workspace.repository.preview(scope, assetId, printer, template, owner.signal);
      if (owner.signal.aborted) return;
      const local = await workspace.files.preview(result.file, owner.signal);
      if (owner.signal.aborted) { local.release(); return; }
      const value = { ...local, ...result }; heldPreview.current = value; setPreview(value);
    } catch { if (!owner.signal.aborted) setFailed(true); }
    finally { running.current = false; setBusy(false); }
  };
  const submit = async () => {
    const owner = task.lifetime.current;
    if (!owner || owner.signal.aborted || running.current || (!submission.current.locked && (!printer || !template || !preview))) return;
    running.current = true; setBusy(true); setFailed(false);
    try { const job = submission.current.locked ? await submission.current.retry() : await submission.current.submit(scope, assetId, { printerId, mediaFingerprint: printer!.mediaFingerprint, template: template!, copies: 1, previewFingerprint: preview!.fingerprint }); if (!owner.signal.aborted) { workspace.requests.settled(scope, assetId); onQueued(job.id); } }
    catch { if (!owner.signal.aborted) setFailed(true); }
    finally { running.current = false; setBusy(false); }
  };
  const width = Math.max(1, Math.min(dimensions.width - 48, 520)); const rotated = preview && Math.abs(preview.file.rotation) % 180 === 90;
  const height = preview ? width * (rotated ? preview.file.width / preview.file.height : preview.file.height / preview.file.width) : 0;
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content}>
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error ? <><Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text><NativeCommandButton label={t('printing.mobile.retry')} onPress={task.reload} /></> : null}
    {catalog && !catalog.printers.some(item => !item.retired) ? <Text style={{ color: palette.text }}>{t('printing.mobile.empty')}</Text> : null}
    {catalog && template ? <>
      <SettingsPickerRow label={t('printing.mobile.printer')} accessibilityLabel={t('printing.mobile.printer')} value={printerId} disabled={locked}
        options={[{ value: '', label: t('printing.mobile.none') }, ...catalog.printers.filter(item => !item.retired).map(item => ({ value: item.id, label: `${item.name} · ${item.mediaName}` }))]}
        onChange={id => { if (!locked) { clearPreview(); setPrinterId(id); } }} />
      {printer ? <Text style={{ color: palette.textMuted }}>{printerReadiness(printer.readiness)}</Text> : null}
      <SettingsPickerRow label={t('printing.mobile.template')} accessibilityLabel={t('printing.mobile.template')} value={`${template.id}/${template.version}`} disabled={locked}
        options={catalog.templates.map(item => ({ value: `${item.id}/${item.version}`, label: item.name }))}
        onChange={value => { const item = catalog.templates.find(candidate => `${candidate.id}/${candidate.version}` === value); if (!locked && item) { clearPreview(); setTemplate({ id: item.id, version: item.version, showReference: item.showReference }); } }} />
      {catalog.templates.find(item => item.id === template.id)?.supportsReference ? <AppSwitchField label={t('printing.mobile.reference')} value={template.showReference} disabled={locked} onValueChange={showReference => { if (!locked) { clearPreview(); setTemplate({ ...template, showReference }); } }} /> : null}
      <NativeCommandButton label={t('printing.mobile.preview')} disabled={locked || !printer} onPress={() => void render()} />
      {preview ? <View style={{ alignSelf: 'center', width, height, overflow: 'hidden', backgroundColor: '#fff' }}><Image accessibilityLabel={t('printing.mobile.previewAlt')} accessible source={{ uri: preview.uri }} resizeMode="contain"
        style={{ position: 'absolute', width: rotated ? height : width, height: rotated ? width : height, left: rotated ? (width - height) / 2 : 0, top: rotated ? (height - width) / 2 : 0, transform: [{ rotate: `${preview.file.rotation}deg` }] }} /></View> : null}
      {failed ? <Text accessibilityRole="alert" style={styles.errorMessage}>{t(submission.current.locked ? 'printing.mobile.unknownSubmission' : 'printing.mobile.unavailable')}</Text> : null}
      {failed && !submission.current.locked ? <NativeCommandButton label={t('printing.mobile.refresh')} disabled={busy} onPress={task.reload} /> : null}
      <NativeCommandButton label={t(submission.current.locked ? 'printing.mobile.retry' : 'printing.mobile.print')} prominence="primary" disabled={busy || (!submission.current.locked && (!preview || !printer))} onPress={() => void submit()} />
    </> : null}
  </ScrollView>;
}
