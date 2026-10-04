import { appKeyboardDismissMode } from '../components/AppTextInput';
import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Image, ScrollView, Text, View } from 'react-native';
import type { PrintScope, PrintTemplate, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import type { LabelWorkspace } from '../../application/labels/LabelWorkspace';
import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { PrintCopiesControl } from './PrintCopiesControl';
import { SettingsPickerRow } from '../components/SettingsPickerRow';

import { SettingsActionRow, SettingsLoadingRow, SettingsSection, SettingsSeparator, SettingsSwitchRow, useSettingsListStyles } from '../screens/SettingsList';
import { mediaSizeLabel, printerReadiness } from './PrintingStatus';
import { usePrintPreview } from './usePrintPreview';
import { usePrintingTask } from './usePrintingTask';
export interface AssetPrintDraft { printerId?: string; template?: PrintTemplate; copies?: string }
export function AssetPrintScreen({ workspace, scope, assetId, predecessorId, diagnostic = false, draftState, labelWorkspace, onQueued }: { readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly assetId: string; readonly predecessorId?: string; readonly diagnostic?: boolean; readonly draftState?: AssetPrintDraft; readonly labelWorkspace?: LabelWorkspace; readonly onQueued: (id: string) => void }) {
  const { palette, styles } = useSettingsListStyles(); const [previewWidth, setPreviewWidth] = useState(0);
  const load = useCallback((signal: AbortSignal) => workspace.repository.catalog(scope, signal), [workspace, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, predecessorId ?? assetId); const catalog = task.data;
  const [copyInput, setCopyInput] = useState(draftState?.copies ?? '1');
  const copies = Number(copyInput); const validCopies = /^\d+$/.test(copyInput) && Number.isSafeInteger(copies) && copies > 0;
  const [printerId, setPrinterId] = useState(draftState?.printerId ?? ''); const [template, setTemplate] = useState<PrintTemplate | undefined>(draftState?.template);
  const [busy, setBusy] = useState(false); const running = useRef(false); const [failed, setFailed] = useState(false);
  const submission = useRef(predecessorId ? workspace.requests.forReprint(scope, predecessorId) : workspace.requests.forAsset(scope, assetId));

  const initialized = useRef(!!draftState?.template);
  useLayoutEffect(() => { if (draftState && template) { draftState.printerId = printerId; draftState.template = template; draftState.copies = copyInput; } }, [draftState, printerId, template, copyInput]);
  useEffect(() => { if (catalog && !initialized.current) { initialized.current = true; const pending = submission.current.selection; setPrinterId(pending?.printerId ?? catalog.settings.defaultPrinterId ?? (catalog.printers.filter(value => !value.retired).length === 1 ? catalog.printers.find(value => !value.retired)!.id : '')); setTemplate(pending?.template ?? catalog.settings.template); if (pending) setCopyInput(String(pending.copies)); } }, [catalog]);
  const printer = catalog?.printers.find(item => item.id === printerId && !item.retired);
  const validTemplate = !!template && !!catalog?.templates.some(item => item.id === template.id && item.version === template.version);
  const locked = busy || submission.current.locked;
  const automatic = usePrintPreview(workspace, scope, assetId, printer, template, task.lifetime.current, !!catalog && validTemplate && !diagnostic && !submission.current.locked);
  const preview = automatic.preview;
  const [exportFailed, setExportFailed] = useState(false);
  const share = async (format: 'png' | 'pdf') => {
    const owner = task.lifetime.current;
    const fullTemplate = catalog?.templates.find(value => value.id === template?.id && value.version === template?.version);
    if (!owner || owner.signal.aborted || !labelWorkspace || !printer || !template || !fullTemplate || !preview || running.current || submission.current.locked) return;
    running.current = true; setBusy(true); setExportFailed(false);
    try {
      const file = await labelWorkspace.repository.render(scope, assetId, { media: printer.media, template: fullTemplate, showReference: template.showReference }, format, owner.signal);
      if (!owner.signal.aborted) await labelWorkspace.files.deliver(file, 'share', owner.signal);
    } catch { if (!owner.signal.aborted) setExportFailed(true); }
    finally { running.current = false; setBusy(false); }
  };
  const submit = async () => {
    const owner = task.lifetime.current;
    if (!owner || owner.signal.aborted || running.current || (!submission.current.locked && (!printer || !template || !validTemplate || !validCopies || (!diagnostic && !preview)))) return;
    running.current = true; setBusy(true); setFailed(false);
    try { const job = submission.current.locked ? await submission.current.retry() : await submission.current.submit(scope, predecessorId ?? assetId, { printerId, mediaFingerprint: printer!.mediaFingerprint, template: template!, copies: diagnostic ? 1 : copies, previewFingerprint: preview?.fingerprint }); if (!owner.signal.aborted) { workspace.requests.settled(scope, predecessorId ?? assetId, predecessorId ? 'reprint' : 'asset'); onQueued(job.id); } }
    catch { if (!owner.signal.aborted) setFailed(true); }
    finally { running.current = false; setBusy(false); }
  };
  const width = previewWidth; const rotated = preview && Math.abs(preview.file.rotation) % 180 === 90;
  const height = preview ? width * (rotated ? preview.file.width / preview.file.height : preview.file.height / preview.file.width) : 0;
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic" keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error ? <SettingsSection><View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.dangerText}>{t('printing.mobile.unavailable')}</Text></View><SettingsActionRow label={t('printing.mobile.retry')} onPress={task.reload} /></SettingsSection> : null}
    {catalog && !catalog.printers.some(item => !item.retired) ? <Text style={styles.sectionFooter}>{t('printing.mobile.empty')}</Text> : null}
    {catalog && template ? <>
      <SettingsSection footer={printer ? printerReadiness(printer.readiness) : undefined}><SettingsPickerRow label={t('printing.mobile.printer')} accessibilityLabel={t('printing.mobile.printer')} value={printerId} disabled={locked}
        options={[{ value: '', label: t('printing.mobile.none') }, ...catalog.printers.filter(item => !item.retired).map(item => ({ value: item.id, label: `${item.name} · ${mediaSizeLabel(item.mediaName, item.media)}` }))]}
        onChange={id => { if (!locked) { setFailed(false); setPrinterId(id); } }} />
      </SettingsSection><SettingsSection footer={catalog.templates.find(item => item.id === template.id)?.supportsReference ? t('labels.mobile.referenceHelp') : undefined}>
      <SettingsPickerRow label={t('printing.mobile.template')} accessibilityLabel={t('printing.mobile.template')} value={`${template.id}/${template.version}`} disabled={locked}
        options={catalog.templates.map(item => ({ value: `${item.id}/${item.version}`, label: item.name }))}
        onChange={value => { const item = catalog.templates.find(candidate => `${candidate.id}/${candidate.version}` === value); if (!locked && item) { setFailed(false); setTemplate({ id: item.id, version: item.version, showReference: item.showReference }); } }} />
      {catalog.templates.find(item => item.id === template.id)?.supportsReference ? <><SettingsSeparator /><SettingsSwitchRow label={t('printing.mobile.reference')} value={template.showReference} disabled={locked} onValueChange={showReference => { if (!locked) { setFailed(false); setTemplate({ ...template, showReference }); } }} /></> : null}
      {!diagnostic ? <><SettingsSeparator /><PrintCopiesControl label={t('printing.mobile.copies')} value={copyInput} disabled={locked} onChange={value => { if (!locked) { setFailed(false); setCopyInput(value); } }} />{!validCopies ? <Text accessibilityRole="alert" style={styles.sectionFooter}>{t('printing.mobile.invalidCopies')}</Text> : null}</> : null}
      </SettingsSection><SettingsSection>{diagnostic ? <View style={styles.navigationRow}><Text style={styles.valueText}>{t('printing.mobile.diagnostic')}</Text></View> : automatic.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
      {preview ? <View onLayout={event => setPreviewWidth(event.nativeEvent.layout.width)} style={{ alignSelf: 'center', width: '100%', maxWidth: 520, aspectRatio: rotated ? preview.file.height / preview.file.width : preview.file.width / preview.file.height, overflow: 'hidden', backgroundColor: '#fff' }}><Image accessibilityRole="image" accessibilityLabel={t('printing.mobile.previewAlt')} accessible source={{ uri: preview.uri }} resizeMode="contain"
        style={{ position: 'absolute', width: rotated ? height : width, height: rotated ? width : height, left: rotated ? (width - height) / 2 : 0, top: rotated ? (height - width) / 2 : 0, transform: [{ rotate: `${preview.file.rotation}deg` }] }} /></View> : null}
      {failed || automatic.failed ? <View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.dangerText}>{t(submission.current.locked ? 'printing.mobile.unknownSubmission' : 'printing.mobile.unavailable')}</Text></View> : null}
      {automatic.failed ? <SettingsActionRow label={t('printing.mobile.retry')} disabled={busy} onPress={automatic.retry} /> : null}
      </SettingsSection><View style={[styles.contentBlock, { marginTop: 24 }]}><NativeCommandButton label={t(submission.current.locked ? 'printing.mobile.retry' : predecessorId ? 'printing.mobile.reprint' : 'printing.mobile.print')} prominence="primary" disabled={busy || (!submission.current.locked && ((!diagnostic && !preview) || !printer || !validTemplate || !validCopies))} onPress={() => void submit()} /></View>
      {labelWorkspace && !diagnostic ? <SettingsSection>
        <SettingsActionRow label={t('labels.mobile.sharePNG')} disabled={locked || !preview} onPress={() => void share('png')} />
        <SettingsSeparator /><SettingsActionRow label={t('labels.mobile.sharePDF')} disabled={locked || !preview} onPress={() => void share('pdf')} />
        {exportFailed ? <View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.dangerText}>{t('labels.mobile.unavailable')}</Text></View> : null}
      </SettingsSection> : null}
    </> : null}
  </ScrollView>;
}
