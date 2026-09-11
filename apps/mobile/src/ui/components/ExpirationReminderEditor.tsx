import { useEffect, useRef, useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import type { ExpirationReminderPolicy } from '../../domain/notifications/Notification';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing } from '../theme/tokens';
import { AppSwitchField } from './AppSwitchField';
import { AppTextInput } from './AppTextInput';

export function ExpirationReminderEditor({ initialPolicy, inheritedPolicy, disabled = false, onSave }: {
  readonly initialPolicy: ExpirationReminderPolicy | null;
  readonly inheritedPolicy?: ExpirationReminderPolicy;
  readonly disabled?: boolean;
  readonly onSave: (value: ExpirationReminderPolicy | null) => Promise<void>;
}) {
  const colors = useAppearancePalette();
  const initial = initialPolicy ?? inheritedPolicy;
  if (!initial) throw new Error('A reminder policy is required.');
  const [inherit, setInherit] = useState(initialPolicy === null);
  const [draft, setDraft] = useState({ ...initial });
  const [days, setDays] = useState(String(initial.advanceDays));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);
  const pending = useRef(false);
  const mounted = useRef(true);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  const displayed = inherit && inheritedPolicy ? inheritedPolicy : draft;
  const valid = /^\d+$/.test(days) && Number(days) <= 3650;
  const locked = disabled || saving || inherit;
  function changed() { setSaved(false); setError(''); }
  function toggle(field: 'enabled' | 'upcoming' | 'expired', value: boolean) {
    if (locked) return;
    setDraft((previous) => ({ ...previous, [field]: value })); changed();
  }
  async function save() {
    if (disabled || pending.current || (!inherit && !valid)) return;
    pending.current = true; setSaving(true); setError(''); setSaved(false);
    try {
      await onSave(inherit ? null : { ...draft, advanceDays: Number(days) });
      if (mounted.current) setSaved(true);
    } catch (caught) {
      if (mounted.current) setError(caught instanceof NotificationFailure ? caught.message : 'Reminders could not be saved. Your changes are still here. Try again.');
    } finally { pending.current = false; if (mounted.current) setSaving(false); }
  }
  return <View style={styles.form}>
    {inheritedPolicy ? <>
      <AppSwitchField label="Use inventory defaults" value={inherit} disabled={disabled || saving} onValueChange={(value) => { if (disabled || saving) return; setInherit(value); changed(); }} />
      <Text style={{ color: colors.textMuted }}>{inherit ? 'Inherited from your inventory settings.' : 'Custom settings for this asset type.'}</Text>
    </> : <Text style={{ color: colors.textMuted }}>Your inventory defaults. Asset types with custom settings can still send reminders when these defaults are off.</Text>}
    <AppSwitchField label="Enable expiration reminders" value={displayed.enabled} disabled={locked} onValueChange={(value) => toggle('enabled', value)} />
    <AppSwitchField label="Notify before expiration" value={displayed.upcoming} disabled={locked} onValueChange={(value) => toggle('upcoming', value)} />
    <Text style={{ color: colors.text }}>Days before expiration</Text>
    <AppTextInput accessibilityLabel="Days before expiration" keyboardType="number-pad" editable={!locked} value={inherit ? String(displayed.advanceDays) : days} onChangeText={(value) => { if (locked) return; setDays(value); changed(); }} style={[styles.input, { color: colors.text, borderColor: colors.controlBorder }]} />
    {!inherit && !valid ? <Text accessibilityRole="alert" style={{ color: colors.danger }}>Enter a whole number from 0 to 3650.</Text> : null}
    <AppSwitchField label="Notify when expired" value={displayed.expired} disabled={locked} onValueChange={(value) => toggle('expired', value)} />
    {error ? <Text accessibilityRole="alert" style={{ color: colors.danger }}>{error}</Text> : null}
    {saved ? <Text accessibilityLiveRegion="polite" style={{ color: colors.textMuted }}>Reminders saved.</Text> : null}
    <Pressable accessibilityRole="button" accessibilityLabel="Save reminders" accessibilityState={{ disabled: disabled || saving || (!inherit && !valid), busy: saving }} disabled={disabled || saving || (!inherit && !valid)} onPress={save} style={[styles.button, { borderColor: colors.controlBorder }]}>
      <Text style={{ color: colors.text }}>{saving ? 'Saving…' : 'Save reminders'}</Text>
    </Pressable>
  </View>;
}
const styles = StyleSheet.create({
  form: { gap: spacing.sm },
  input: { minHeight: 44, borderWidth: 1, borderRadius: radius.md, padding: spacing.sm },
  button: { minHeight: 44, borderWidth: 1, borderRadius: radius.md, padding: spacing.sm, justifyContent: 'center', alignItems: 'center' }
});
