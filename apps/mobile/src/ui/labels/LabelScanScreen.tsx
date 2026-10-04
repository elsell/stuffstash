import { useCallback, useEffect, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import { AppState, ScrollView, Text, View } from 'react-native';
import { ExpoLabelCamera } from '../../adapters/labels/ExpoLabelCamera';
import type { LabelReference } from '../../application/labels/LabelWorkspace';
import { LabelFailure } from '../../application/labels/LabelWorkspace';
import type { OpenLabel } from '../../application/labels/OpenLabel';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { AppTextInput } from '../components/AppTextInput';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { t } from '../../presentation/localization';

type Props = { readonly open: Pick<OpenLabel, 'execute'>; readonly parse: (source: string) => LabelReference; readonly pending?: LabelReference; readonly invalid: boolean;
  readonly onResolved: (assetId: string) => void; readonly onServer: () => void; readonly onAccount: () => void };
export function LabelScanScreen({ open, parse, pending, invalid, onResolved, onServer, onAccount }: Props) {
  const colors = useAppearancePalette();
  const [source, setSource] = useState(''); const [camera, setCamera] = useState(!pending);
  const [active, setActive] = useState(AppState.currentState === 'active'); const [focused, setFocused] = useState(false);
  const [busy, setBusy] = useState(false); const [error, setError] = useState<'wrong_instance' | 'invalid_label' | 'unavailable' | 'camera'>();
  const operationReference = useRef<LabelReference | undefined>(undefined);
  const operation = useRef<AbortController | undefined>(undefined); const lifetime = useRef(false); const failedFrame = useRef<string | undefined>(undefined);
  useFocusEffect(useCallback(() => {
    lifetime.current = true; setFocused(true); setBusy(false);
    return () => { lifetime.current = false; setFocused(false); operation.current?.abort(); operation.current = undefined; operationReference.current = undefined; };
  }, [open]));
  useEffect(() => {
    const subscription = AppState.addEventListener('change', state => {
      setActive(state === 'active');
      if (state !== 'active') { operation.current?.abort(); operation.current = undefined; operationReference.current = undefined; setBusy(false); }
    });
    return () => subscription.remove();
  }, []);
  useEffect(() => {
    const current = operationReference.current;
    if (invalid || current && pending && (current.instanceId !== pending.instanceId || current.labelId !== pending.labelId)) {
      operation.current?.abort(); operation.current = undefined; operationReference.current = undefined; setBusy(false);
    }
  }, [pending?.instanceId, pending?.labelId, invalid]);
  const resolve = async (value: LabelReference | string, frame = false) => {
    if (!lifetime.current || !active || operation.current || frame && failedFrame.current === value) return;
    if (frame && typeof value === 'string') failedFrame.current = value;
    const controller = new AbortController(); operation.current = controller; setBusy(true); setError(undefined); setCamera(false);
    try {
      const reference = typeof value === 'string' ? parse(value) : value;
      operationReference.current = reference;
      const target = await open.execute(reference, controller.signal);
      if (!controller.signal.aborted && lifetime.current) onResolved(target.assetId);
    } catch (failure) {
      if (!controller.signal.aborted && lifetime.current) setError(failure instanceof LabelFailure ? failure.code : 'unavailable');
    } finally { if (operation.current === controller) { operation.current = undefined; operationReference.current = undefined; if (lifetime.current) setBusy(false); } }
  };
  const message = error === 'wrong_instance' ? 'labels.mobile.wrongInstance' : error === 'invalid_label' || invalid ? 'labels.mobile.unsupported' : error === 'camera' ? 'labels.mobile.cameraDenied' : error ? 'labels.mobile.unavailable' : undefined;
  return <ScrollView keyboardShouldPersistTaps="handled" contentContainerStyle={{ padding: 24, gap: 16, paddingBottom: 48 }} style={{ backgroundColor: colors.background }}>
    {busy ? <Text accessibilityLiveRegion="polite" style={{ color: colors.text }}>{t('labels.mobile.resolving')}</Text> : null}
    {message ? <Text accessibilityRole="alert" style={{ color: colors.danger }}>{t(message)}</Text> : null}
    {pending ? <NativeCommandButton label={t('labels.mobile.open')} disabled={busy} prominence="primary" onPress={() => void resolve(pending)} /> : null}
    {camera && focused ? <View style={{ height: 280 }}>
      <ExpoLabelCamera active={active} onCode={value => void resolve(value, true)} onUnavailable={() => { setCamera(false); setError('camera'); }} />
    </View> : null}
    <NativeCommandButton label={t('labels.mobile.camera')} disabled={busy} onPress={() => { failedFrame.current = undefined; setCamera(true); setError(undefined); }} />
    <Text style={{ color: colors.text }}>{t('labels.mobile.paste')}</Text>
    <AppTextInput accessibilityLabel={t('labels.mobile.paste')} value={source} onChangeText={setSource} autoCapitalize="none" autoCorrect={false} keyboardType="url" editable={!busy}
      style={{ color: colors.text, borderColor: colors.controlBorder, borderWidth: 1, borderRadius: 8, padding: 12 }} />
    <NativeCommandButton label={t('labels.mobile.open')} disabled={busy || !source.trim()} onPress={() => void resolve(source.trim())} />
    {error ? <>
      <NativeCommandButton label={t('labels.mobile.server')} disabled={busy} onPress={onServer} />
      <NativeCommandButton label={t('labels.mobile.account')} disabled={busy} onPress={onAccount} />
    </> : null}
  </ScrollView>;
}
