import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Stack, useFocusEffect } from 'expo-router';
import { ActivityIndicator, AppState, Linking, ScrollView, StyleSheet, Text, View } from 'react-native';
import { ExpoLabelCamera } from '../../adapters/labels/ExpoLabelCamera';
import type { LabelReference } from '../../application/labels/LabelWorkspace';
import { LabelFailure } from '../../application/labels/LabelWorkspace';
import type { OpenLabel } from '../../application/labels/OpenLabel';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { AppTextInput } from '../components/AppTextInput';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import type { NativeHeaderAction } from '../components/NativeHeaderActions.types';
import { radius, spacing } from '../theme/tokens';
import { t } from '../../presentation/localization';

type Props = { readonly open: Pick<OpenLabel, 'execute'>; readonly parse: (source: string) => LabelReference; readonly pending?: LabelReference; readonly revision?: number; readonly retain?: (reference: LabelReference) => number; readonly invalid: boolean;
  readonly onResolved: (assetId: string) => void; readonly onServer: (reference?: LabelReference) => void; readonly onAccount: (reference?: LabelReference) => void };
export function LabelScanScreen({ open, parse, pending, revision = 0, retain, invalid, onResolved, onServer, onAccount }: Props) {
  const colors = useAppearancePalette();
  const [source, setSource] = useState(''); const [camera, setCamera] = useState(!pending && !invalid);
  const [active, setActive] = useState(AppState.currentState === 'active'); const [focused, setFocused] = useState(false);
  const [manual, setManual] = useState(false); const [torch, setTorch] = useState(false);
  const attempted = useRef<string | undefined>(undefined);
  const lastReference = useRef<LabelReference | undefined>(pending);
  const [busy, setBusy] = useState(false); const [error, setError] = useState<'wrong_instance' | 'invalid_label' | 'unavailable' | 'camera'>();
  const operationReference = useRef<LabelReference | undefined>(undefined);
  const operation = useRef<AbortController | undefined>(undefined); const lifetime = useRef(false); const failedFrame = useRef<string | undefined>(undefined);
  useFocusEffect(useCallback(() => {
    lifetime.current = true; setFocused(true); setBusy(false);
    return () => { if (operation.current) setError('unavailable'); lifetime.current = false; setFocused(false); operation.current?.abort(); operation.current = undefined; operationReference.current = undefined; };
  }, [open]));
  useEffect(() => {
    const subscription = AppState.addEventListener('change', state => {
      setActive(state === 'active');
      if (state !== 'active') { if (operation.current) setError('unavailable'); operation.current?.abort(); operation.current = undefined; operationReference.current = undefined; setBusy(false); }
    });
    return () => subscription.remove();
  }, []);
  useEffect(() => {
    const current = operationReference.current;
    if (invalid) { setError(undefined); setCamera(false); setManual(false); }
    if (invalid || current && pending && attempted.current !== `${revision}:${pending.instanceId}/${pending.labelId}`) {
      operation.current?.abort(); operation.current = undefined; operationReference.current = undefined; setBusy(false);
    }
  }, [pending?.instanceId, pending?.labelId, revision, invalid]);
  const resolve = async (value: LabelReference | string, frame = false) => {
    if (!lifetime.current || !active || operation.current || frame && failedFrame.current === value) return;
    if (frame && typeof value === 'string') failedFrame.current = value;
    const controller = new AbortController(); operation.current = controller; setBusy(true); setError(undefined); setCamera(false);
    try {
      const reference = typeof value === 'string' ? parse(value) : value;
      operationReference.current = reference; lastReference.current = reference;
      if (typeof value === 'string' && retain) attempted.current = `${retain(reference)}:${reference.instanceId}/${reference.labelId}`;
      const target = await open.execute(reference, controller.signal);
      if (!controller.signal.aborted && lifetime.current) onResolved(target.assetId);
    } catch (failure) {
      if (!controller.signal.aborted && lifetime.current) {
        const code = failure instanceof LabelFailure ? failure.code : 'unavailable';
        setError(code); if (code === 'wrong_instance') setManual(false);
      }
    } finally { if (operation.current === controller) { operation.current = undefined; operationReference.current = undefined; if (lifetime.current) setBusy(false); } }
  };
  useEffect(() => {
    if (!pending || invalid || !focused || !active) return;
    const key = `${revision}:${pending.instanceId}/${pending.labelId}`;
    if (attempted.current !== key) { attempted.current = key; void resolve(pending); }
  });
  const message = error === 'wrong_instance' ? 'labels.mobile.wrongInstance' : error === 'invalid_label' || invalid ? 'labels.mobile.unsupported' : error === 'camera' ? 'labels.mobile.cameraDenied' : error ? 'labels.mobile.unavailable' : undefined;
  const startCamera = () => { failedFrame.current = undefined; setCamera(true); setManual(false); setTorch(false); setError(undefined); };
  const headerActions: NativeHeaderAction[] = busy ? [] : camera ? [
    { kind: 'flashlight', label: t(torch ? 'labels.mobile.lightOff' : 'labels.mobile.lightOn'), tintColor: torch ? colors.action : colors.text, onPress: () => setTorch(value => !value) },
    { kind: 'link', label: t('labels.mobile.enterLink'), onPress: () => { setCamera(false); setManual(true); setError(undefined); } }
  ] : [
    { kind: 'scan', label: t('labels.mobile.camera'), onPress: startCamera },
    ...(!manual ? [{ kind: 'link' as const, label: t('labels.mobile.enterLink'), onPress: () => setManual(true) }] : []),
    ...(error === 'camera' ? [{ kind: 'settings' as const, label: t('labels.mobile.settings'), onPress: () => { void Linking.openSettings().catch(() => undefined); } }] : []),
    ...(error === 'unavailable' ? [{ kind: 'account' as const, label: t('labels.mobile.account'), onPress: () => onAccount(lastReference.current) }] : [])
  ];
  const header = useNativeHeaderActionOptions(headerActions);
  const title = t(busy ? 'labels.mobile.open' : manual ? 'labels.mobile.enterLink' : message ? 'labels.mobile.open' : 'labels.mobile.scan');
  const headerOptions = useMemo(() => ({ ...header, title }), [header, title]);
  return <><Stack.Screen options={headerOptions} /><ScrollView keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag" automaticallyAdjustKeyboardInsets
    contentInsetAdjustmentBehavior="automatic" contentContainerStyle={[styles.content, !camera && styles.readableContent]} style={{ backgroundColor: colors.background }}>
    {busy ? <View style={styles.status}>
      <ActivityIndicator color={colors.action} />
      <Text accessibilityLiveRegion="polite" style={[styles.body, { color: colors.text }]}>{t('labels.mobile.resolving')}</Text>
    </View> : <>
      {message ? <View style={styles.status}>
        <Text accessibilityRole="alert" style={[styles.body, { color: colors.text }]}>{t(message)}</Text>
        {error === 'wrong_instance' && !manual ? <NativeCommandButton label={t('labels.mobile.server')} prominence="primary" onPress={() => onServer(lastReference.current)} /> : null}
        {error === 'unavailable' && lastReference.current && !manual ? <NativeCommandButton label={t('labels.mobile.retry')} prominence="primary" onPress={() => void resolve(lastReference.current!)} /> : null}
      </View> : null}
      {camera && focused ? <>
        <Text style={[styles.body, { color: colors.textMuted }]}>{t('labels.mobile.aim')}</Text>
        <View style={[styles.preview, { backgroundColor: colors.scrim }]}>
          <ExpoLabelCamera active={active} torch={torch} onCode={value => void resolve(value, true)} onUnavailable={() => { setCamera(false); setManual(true); setError('camera'); }} />
        </View>
      </> : manual ? <View style={styles.form}>
        <Text style={[styles.body, { color: colors.text }]}>{t('labels.mobile.paste')}</Text>
        <AppTextInput accessibilityLabel={t('labels.mobile.paste')} value={source} onChangeText={setSource} autoCapitalize="none" autoCorrect={false} keyboardType="url" returnKeyType="go"
          onSubmitEditing={() => { if (source.trim()) void resolve(source.trim()); }}
          style={[styles.input, { color: colors.text, backgroundColor: colors.surface, borderColor: colors.controlBorder }]} />
        <NativeCommandButton label={t('labels.mobile.open')} prominence="primary" disabled={!source.trim()} onPress={() => void resolve(source.trim())} />
      </View> : null}
    </>}
  </ScrollView></>;
}
const styles = StyleSheet.create({
  content: { flexGrow: 1, padding: spacing.lg, gap: spacing.lg, paddingBottom: spacing.xl },
  readableContent: { width: '100%', maxWidth: 640, alignSelf: 'center' },
  preview: { flexGrow: 1, minHeight: 280, borderRadius: radius.lg, overflow: 'hidden' },
  status: { gap: spacing.md, paddingVertical: spacing.md },
  body: { fontSize: 17 },
  form: { gap: spacing.md },
  input: { minHeight: 48, borderWidth: 1, borderRadius: radius.md, padding: spacing.md, fontSize: 17 }
});
