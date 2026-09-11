import { useEffect, useRef, useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import type { ExpirationReminderPolicy } from '../../domain/notifications/Notification';
import { AppSwitchField } from './AppSwitchField';
import { AppTextInput } from './AppTextInput';
import { NativeSegmentedControl } from './NativeSegmentedControl';
import { SelectionRow } from './SelectionRow';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';

export function reminderSummary(policy: ExpirationReminderPolicy): string {
  if (!policy.enabled) return 'Off';
  return [policy.upcoming ? `${policy.advanceDays} days before` : '', policy.expired ? 'when expired' : ''].filter(Boolean).join(' and ') || 'No reminders selected';
}
type Mode = 'defaults' | 'custom' | 'off';
export function ExpirationReminderEditor({ initialPolicy, inheritedPolicy, disabled = false, onSave }: {
  readonly initialPolicy: ExpirationReminderPolicy | null;
  readonly inheritedPolicy?: ExpirationReminderPolicy;
  readonly disabled?: boolean;
  readonly onSave: (value: ExpirationReminderPolicy | null) => Promise<void>;
}) {
  const initial = initialPolicy ?? inheritedPolicy;
  if (!initial) throw new Error('A reminder policy is required.');
  const colors = useAppearancePalette();
  const [draft, setDraft] = useState({ ...initial });
  const [mode, setMode] = useState<Mode>(initialPolicy === null ? 'defaults' : initialPolicy.enabled ? 'custom' : 'off');
  const [days, setDays] = useState(String(initial.advanceDays));
  const [editingDays, setEditingDays] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);
  const pending = useRef(false);
  const mounted = useRef(true);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  const source = JSON.stringify([initialPolicy, inheritedPolicy]);
  useEffect(() => {
    if (dirty || pending.current) return;
    const policy = initialPolicy ?? inheritedPolicy;
    if (!policy) return;
    setDraft({ ...policy }); setDays(String(policy.advanceDays));
    setMode(initialPolicy === null ? 'defaults' : policy.enabled ? 'custom' : 'off');
  }, [source, dirty]);
  const displayed = mode === 'defaults' && inheritedPolicy ? inheritedPolicy : draft;
  const locked = disabled || saving;
  const valid = /^\d+$/.test(days) && Number(days) <= 3650;
  async function commit(next: ExpirationReminderPolicy, nextMode = mode) {
    if (disabled || pending.current) return;
    if (dirty && !valid) { setError('Finish entering reminder days or cancel this edit.'); return; }
    if (dirty) next = { ...next, advanceDays: Number(days) };
    pending.current = true; setSaving(true); setError(''); setDirty(true);
    setDraft(next); setMode(nextMode);
    try {
      await onSave(nextMode === 'defaults' ? null : { ...next, enabled: nextMode !== 'off' });
      if (mounted.current) { setDirty(false); setEditingDays(false); }
    } catch { if (mounted.current) setError('Could not save. Your changes are still here. Retry or reload saved settings.'); }
    finally { pending.current = false; if (mounted.current) setSaving(false); }
  }
  return <View style={{ gap: spacing.sm }}>
    {inheritedPolicy ? <NativeSegmentedControl colors={colors} value={mode} disabled={locked} segments={[{ value: 'defaults', label: 'Use defaults' }, { value: 'custom', label: 'Custom' }, { value: 'off', label: 'Off' }]} onChange={next => void commit({ ...displayed }, next)} />
      : <AppSwitchField label="Default reminders" value={mode !== 'off'} disabled={locked} onValueChange={enabled => void commit(draft, enabled ? 'custom' : 'off')} />}
    {mode === 'defaults' ? <Text style={{ color: colors.textMuted }}>{reminderSummary(displayed)}</Text> : mode === 'custom' ? <>
      <SelectionRow label="Before expiration" value={draft.upcoming ? `${draft.advanceDays} days` : 'Off'} expanded={editingDays} disabled={locked} onPress={() => setEditingDays(value => !value)}>
        <AppSwitchField label="Remind before expiration" value={draft.upcoming} disabled={locked} onValueChange={upcoming => void commit({ ...draft, upcoming })} />
        {draft.upcoming ? <>
          <Text style={{ color: colors.textMuted }}>Days before expiration</Text>
          <AppTextInput accessibilityLabel="Days before expiration" keyboardType="number-pad" editable={!locked} value={days} onChangeText={value => { setDays(value); setDirty(true); }} style={{ color: colors.text, minHeight: 44, padding: spacing.sm, backgroundColor: colors.surface }} />
          {!valid ? <Text accessibilityRole="alert" style={{ color: colors.danger }}>Enter a whole number from 0 to 3650.</Text> : null}
          <Pressable accessibilityRole="button" accessibilityLabel="Save reminder days" disabled={locked || !valid} onPress={() => void commit({ ...draft, advanceDays: Number(days) })} style={{ minHeight: 44, justifyContent: 'center' }}><Text style={{ color: colors.action }}>Done</Text></Pressable>
          <Pressable accessibilityRole="button" accessibilityLabel="Cancel reminder days" disabled={locked} onPress={() => { setDays(String(draft.advanceDays)); setEditingDays(false); setDirty(false); setError(''); }} style={{ minHeight: 44, justifyContent: 'center' }}><Text style={{ color: colors.action }}>Cancel</Text></Pressable>
        </> : null}
      </SelectionRow>
      <AppSwitchField label="When expired" value={draft.expired} disabled={locked} onValueChange={expired => void commit({ ...draft, expired })} />
    </> : null}
    {saving ? <Text accessibilityLiveRegion="polite" style={{ color: colors.textMuted }}>Saving…</Text> : null}
    {error ? <View><Text accessibilityRole="alert" style={{ color: colors.danger }}>{error}</Text>
      <Pressable accessibilityRole="button" accessibilityLabel="Retry saving reminders" disabled={locked} onPress={() => void commit({ ...draft, advanceDays: valid ? Number(days) : draft.advanceDays })} style={{ minHeight: 44 }}><Text style={{ color: colors.action }}>Retry</Text></Pressable>
      <Pressable accessibilityRole="button" accessibilityLabel="Discard reminder changes" disabled={locked} onPress={() => { setDirty(false); setError(''); setEditingDays(false); }} style={{ minHeight: 44 }}><Text style={{ color: colors.action }}>Discard changes</Text></Pressable>
    </View> : null}
  </View>;
}
