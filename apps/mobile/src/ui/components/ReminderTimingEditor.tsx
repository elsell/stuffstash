import { useEffect, useRef, useState } from 'react';
import { Stack } from 'expo-router';
import { Pressable, Text, View } from 'react-native';
import type { ExpirationReminderPolicy } from '../../domain/notifications/Notification';
import { AppTextInput } from './AppTextInput';
import { reminderDaysLabel } from './ExpirationReminderEditor';
import { SettingsChoiceRow, SettingsLoadingRow, SettingsSection, SettingsSeparator, useSettingsListStyles } from '../screens/SettingsList';

const presets = [0, 1, 7, 14, 30, 60, 90] as const;
export function ReminderTimingEditor({ policy, disabled = false, onSave, onDone }: {
  readonly policy: ExpirationReminderPolicy; readonly disabled?: boolean;
  readonly onSave: (policy: ExpirationReminderPolicy) => Promise<void>; readonly onDone: () => void;
}) {
  const { styles, palette } = useSettingsListStyles();
  const [selection, setSelection] = useState({upcoming:policy.upcoming,advanceDays:policy.advanceDays});
  const [custom, setCustom] = useState(policy.upcoming && !presets.some(days => days === policy.advanceDays));
  const [days, setDays] = useState(String(policy.advanceDays));
  const [dirty, setDirty] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const pending = useRef(false); const mounted = useRef(true);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  useEffect(() => { if (!dirty && !pending.current) { setDays(String(policy.advanceDays)); setCustom(policy.upcoming && !presets.some(days => days === policy.advanceDays)); setSelection({upcoming:policy.upcoming,advanceDays:policy.advanceDays}); } }, [policy.advanceDays, policy.upcoming, dirty]);
  const valid = /^\d+$/.test(days) && Number(days) <= 3650;
  const locked = disabled || saving;
  async function save(upcoming: boolean, advanceDays: number) {
    if (locked || pending.current) return;
    pending.current = true; setSaving(true); setError(''); setDirty(true); setSelection({upcoming,advanceDays});
    try { await onSave({ ...policy, upcoming, advanceDays }); if (mounted.current) onDone(); }
    catch { if (mounted.current) setError('Could not save. Your selection is still here. Try again.'); }
    finally { pending.current = false; if (mounted.current) setSaving(false); }
  }
  return <>
    <Stack.Screen options={{ title: 'Before expiration', gestureEnabled: !saving, headerBackVisible: !saving, headerRight: () => custom ?
      <Pressable accessibilityRole="button" accessibilityLabel="Save reminder days" accessibilityState={{ disabled: locked || !valid }} disabled={locked || !valid} onPress={() => void save(true, Number(days))} style={[styles.iconButton, { opacity: locked || !valid ? 0.5 : 1 }]}><Text style={styles.actionText}>Done</Text></Pressable> : null }} />
    <SettingsSection footer="Choose when to remind you before the expiration date. Expired reminders are set separately.">
      <SettingsChoiceRow label="Off" selected={!selection.upcoming && !custom} disabled={locked} onPress={() => { setCustom(false); void save(false, policy.advanceDays); }} />
      {presets.map(value => <View key={value}><SettingsSeparator /><SettingsChoiceRow label={value === 0 ? 'On the expiration date' : `${reminderDaysLabel(value)} before`} selected={!custom && selection.upcoming && selection.advanceDays === value} disabled={locked} onPress={() => { setCustom(false); void save(true, value); }} /></View>)}
      <SettingsSeparator /><SettingsChoiceRow label="Custom…" accessibilityLabel="Custom days" selected={custom} disabled={locked} onPress={() => { setCustom(true); setDirty(true); }} />
    </SettingsSection>
    {custom ? <SettingsSection footer="Enter 0–3650 days. Tap Done to save; go back to cancel.">
      <View style={styles.navigationRow}><View style={styles.navigationRowContent}>
        <Text style={[styles.rowLabel, styles.rowText]}>Days before</Text>
        <AppTextInput accessibilityLabel="Days before expiration" keyboardType="number-pad" editable={!locked} value={days} selectTextOnFocus onChangeText={value => { setDays(value); setDirty(true); }} style={{ color: palette.text, fontSize: 17, minHeight: 44, minWidth: 88, textAlign: 'right' }} />
      </View></View>
      {!valid ? <Text accessibilityRole="alert" style={[styles.sectionFooter, { color: palette.danger }]}>Enter a whole number from 0 to 3650.</Text> : null}
    </SettingsSection> : null}
    {saving ? <SettingsSection><SettingsLoadingRow label="Saving timing…" /></SettingsSection> : null}
    {error ? <View style={styles.detailHeader}><Text accessibilityRole="alert" style={{ color: palette.danger }}>{error}</Text></View> : null}
  </>;
}
