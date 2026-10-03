import { useCallback, useEffect, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import { ActivityIndicator, AppState, Image, ScrollView, Text, View, useWindowDimensions } from 'react-native';
import type { LabelFile, LabelProfile, LabelScope, LabelSelection, LabelTemplate, LabelWorkspace } from '../../application/labels/LabelWorkspace';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { NativeChoicePicker } from '../components/NativeChoicePicker';
import { AppSwitchField } from '../components/AppSwitchField';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { t } from '../../presentation/localization';

type Preview = { uri: string; release(): void; file: LabelFile };
export function LabelOptionsScreen({ workspace, assetId, scope }: { readonly workspace: LabelWorkspace; readonly assetId: string; readonly scope: LabelScope }) {
  const colors = useAppearancePalette(); const dimensions = useWindowDimensions();
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
    if (!controller || controller.signal.aborted || running.current || !selection) return;
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
  const width = Math.min(dimensions.width - 48, 520);
  const rotated = preview && Math.abs(preview.file.rotation) % 180 === 90;
  const height = preview ? width * (rotated ? preview.file.width / preview.file.height : preview.file.height / preview.file.width) : 0;
  return <ScrollView contentContainerStyle={{ padding: 24, gap: 16, paddingBottom: 48 }} style={{ backgroundColor: colors.background }}>
    {busy ? <View accessibilityLiveRegion="polite"><ActivityIndicator /><Text style={{ color: colors.text }}>{t('labels.mobile.loading')}</Text></View> : null}
    {error ? <><Text accessibilityRole="alert" style={{ color: colors.danger }}>{t('labels.mobile.unavailable')}</Text><NativeCommandButton label={t('labels.mobile.retry')} disabled={busy} onPress={() => setReload(value => value + 1)} /></> : null}
    {catalog && !selection ? <Text style={{ color: colors.text }}>{t('labels.mobile.mediaUnavailable')}</Text> : null}
    {selection ? <>
      <Text style={{ color: colors.text }}>{t('labels.mobile.size')}: {catalog?.profiles.find(profile => profile.media === selection.media)?.name}</Text>
      <NativeChoicePicker label={t('labels.mobile.template')} disabled={busy} value={`${selection.template.id}/${selection.template.version}`}
        options={(catalog?.templates ?? []).map(template => ({ value: `${template.id}/${template.version}`, label: template.name }))}
        onChange={value => { const template = catalog?.templates.find(item => `${item.id}/${item.version}` === value); if (template) change({ ...selection, template, showReference: template.showReference }); }} />
      {selection.template.supportsReference ? <AppSwitchField label={t('labels.mobile.reference')} value={selection.showReference} disabled={busy} onValueChange={showReference => change({ ...selection, showReference })} /> : null}
      <NativeCommandButton label={t('labels.mobile.preview')} disabled={busy} onPress={() => void run('preview')} />
      {preview ? <View style={{ alignSelf: 'center', width, height, overflow: 'hidden', backgroundColor: '#fff' }}>
        <Image accessible accessibilityLabel={t('labels.mobile.previewAlt')} source={{ uri: preview.uri }} resizeMode="contain"
          style={{ position: 'absolute', width: rotated ? height : width, height: rotated ? width : height,
            left: rotated ? (width - height) / 2 : 0, top: rotated ? (height - width) / 2 : 0, transform: [{ rotate: `${preview.file.rotation}deg` }] }} />
      </View> : null}
      <NativeCommandButton label={t('labels.mobile.sharePNG')} disabled={busy} onPress={() => void run('png')} />
      <NativeCommandButton label={t('labels.mobile.sharePDF')} disabled={busy} onPress={() => void run('pdf')} />
      <Text style={{ color: colors.textMuted }}>{t('labels.mobile.actualSize')}</Text>
      <NativeCommandButton label={t('labels.mobile.print')} prominence="primary" disabled={busy} onPress={() => void run('print')} />
    </> : null}
  </ScrollView>;
}
