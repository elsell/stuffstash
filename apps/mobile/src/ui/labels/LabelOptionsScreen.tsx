import { useCallback, useEffect, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import { AppState, Image, ScrollView, Text, View } from 'react-native';
import type { LabelFile, LabelProfile, LabelScope, LabelSelection, LabelTemplate, LabelWorkspace } from '../../application/labels/LabelWorkspace';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { SettingsActionRow, SettingsLoadingRow, SettingsNavigationRow, SettingsSection, SettingsSeparator, SettingsSwitchRow, SettingsValueRow, useSettingsListStyles } from '../screens/SettingsList';
import { t } from '../../presentation/localization';

type Preview = { uri: string; release(): void; file: LabelFile };
export function LabelOptionsScreen({ workspace, assetId, scope, onPrintOptions }: { readonly workspace: LabelWorkspace; readonly assetId: string; readonly scope: LabelScope; readonly onPrintOptions?: () => void }) {
  const { styles } = useSettingsListStyles();
  const [previewWidth, setPreviewWidth] = useState(0);
  const [catalog, setCatalog] = useState<{ profiles: readonly LabelProfile[]; templates: readonly LabelTemplate[] }>();
  const [selection, setSelection] = useState<LabelSelection>();
  const [preview, setPreview] = useState<Preview>(); const previewRef = useRef<Preview | undefined>(undefined);
  const [busy, setBusy] = useState(false); const [error, setError] = useState(false);
  const [foreground, setForeground] = useState(AppState.currentState === 'active');
  const [reload, setReload] = useState(0);
  useEffect(() => { const listener = AppState.addEventListener('change', state => setForeground(state === 'active')); return () => listener.remove(); }, []);
  const lifetime = useRef<AbortController | undefined>(undefined); const running = useRef(false);
  const clearPreview = useCallback(() => { previewRef.current?.release(); previewRef.current = undefined; setPreview(undefined); }, []);
  useFocusEffect(useCallback(() => {
    if (!foreground) { setBusy(false); clearPreview(); return; }
    setCatalog(undefined); setSelection(undefined);
    const controller = new AbortController(); lifetime.current = controller;
    setBusy(true); running.current = true; setError(false);
    void workspace.repository.catalog(scope, controller.signal).then(result => {
      if (controller.signal.aborted) return;
      setCatalog(result);
      const template = result.templates.find(value => value.id === 'qr-title') ?? result.templates[0]; const media = result.profiles[0]?.media;
      setSelection(template && media ? { template, media, showReference: template.showReference } : undefined);
    }).catch(() => { if (!controller.signal.aborted) setError(true); })
      .finally(() => { if (!controller.signal.aborted) { running.current = false; setBusy(false); } });
    return () => { controller.abort(); running.current = false; clearPreview(); };
  }, [workspace, scope.tenantId, scope.inventoryId, assetId, clearPreview, foreground, reload]));
  const change = (next: LabelSelection) => { clearPreview(); setSelection(next); setError(false); };
  const run = async (action: 'preview' | 'png' | 'pdf' | 'print') => {
    const controller = lifetime.current;
    if (!controller || controller.signal.aborted || running.current || !selection || (error && action !== 'preview')) return;
    running.current = true; setBusy(true); setError(false);
    try {
      const file = await workspace.repository.render(scope, assetId, selection, action === 'pdf' || action === 'print' ? 'pdf' : 'png', controller.signal);
      if (controller.signal.aborted) return;
      if (action === 'preview') {
        const local = await workspace.files.preview(file, controller.signal);
        if (controller.signal.aborted) { local.release(); return; }
        clearPreview(); const next = { ...local, file }; previewRef.current = next; setPreview(next);
      } else await workspace.files.deliver(file, action === 'print' ? 'print' : 'share', controller.signal);
    } catch { if (!controller.signal.aborted) { clearPreview(); setError(true); } }
    finally { if (!controller.signal.aborted) { running.current = false; setBusy(false); } }
  };
  const width = previewWidth;
  const rotated = preview && Math.abs(preview.file.rotation) % 180 === 90;
  const aspectRatio = preview ? (rotated ? preview.file.height / preview.file.width : preview.file.width / preview.file.height) : 1;
  const height = width / aspectRatio;
  return <ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={styles.content} style={styles.shell}>
    {selection ? <SettingsSection>
      <SettingsValueRow label={t('labels.mobile.size')} value={catalog?.profiles.find(profile => profile.media === selection.media)?.name ?? t('labels.mobile.mediaUnavailable')} />
      <SettingsSeparator />
      <SettingsPickerRow label={t('labels.mobile.template')} accessibilityLabel={t('labels.mobile.template')} disabled={busy} value={`${selection.template.id}/${selection.template.version}`}
        options={(catalog?.templates ?? []).map(template => ({ value: `${template.id}/${template.version}`, label: template.name }))}
        onChange={value => { const template = catalog?.templates.find(item => `${item.id}/${item.version}` === value); if (template) change({ ...selection, template, showReference: template.showReference }); }} />
      {selection.template.supportsReference ? <><SettingsSeparator /><SettingsSwitchRow label={t('labels.mobile.reference')} value={selection.showReference} disabled={busy} onValueChange={showReference => change({ ...selection, showReference })} /></> : null}
    </SettingsSection> : null}
    <SettingsSection>
      {busy ? <SettingsLoadingRow label={t('labels.mobile.loading')} /> : null}
      {error ? <>
        <View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.dangerText}>{t('labels.mobile.unavailable')}</Text></View>
        <SettingsActionRow label={t('labels.mobile.retry')} disabled={busy} onPress={() => selection ? void run('preview') : setReload(value => value + 1)} />
      </> : null}
      {catalog && !selection ? <View style={styles.navigationRow}><Text style={styles.valueText}>{t('labels.mobile.mediaUnavailable')}</Text></View> : null}
      {selection && !error ? <SettingsActionRow label={t('labels.mobile.preview')} disabled={busy} onPress={() => void run('preview')} /> : null}
      {preview ? <View onLayout={event => setPreviewWidth(event.nativeEvent.layout.width)} style={{ alignSelf: 'center', width: '100%', maxWidth: 520, aspectRatio, overflow: 'hidden', backgroundColor: '#fff' }}>
        <Image accessible accessibilityRole="image" accessibilityLabel={t('labels.mobile.previewAlt')} source={{ uri: preview.uri }} resizeMode="contain"
          style={{ position: 'absolute', width: rotated ? height : width, height: rotated ? width : height,
            left: rotated ? (width - height) / 2 : 0, top: rotated ? (height - width) / 2 : 0, transform: [{ rotate: `${preview.file.rotation}deg` }] }} />
      </View> : null}
    </SettingsSection>
    {selection ? <SettingsSection footer={t('labels.mobile.actualSize')}>
      <SettingsActionRow label={t('labels.mobile.sharePNG')} disabled={busy || error} onPress={() => void run('png')} />
      <SettingsSeparator />
      <SettingsActionRow label={t('labels.mobile.sharePDF')} disabled={busy || error} onPress={() => void run('pdf')} />
      <SettingsSeparator />
      <SettingsActionRow label={t('labels.mobile.print')} disabled={busy || error} onPress={() => void run('print')} />
    </SettingsSection> : null}
    {onPrintOptions ? <SettingsSection><SettingsNavigationRow accessibilityLabel={t('printing.mobile.options')} label={t('printing.mobile.options')} disabled={busy} onPress={onPrintOptions} /></SettingsSection> : null}
  </ScrollView>;
}
