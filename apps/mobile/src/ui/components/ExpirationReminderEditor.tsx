import { useEffect, useRef, useState } from 'react';
import type { ExpirationReminderPolicy } from '../../domain/notifications/Notification';
import { SettingsActionRow, SettingsChoiceRow, SettingsLoadingRow, SettingsNavigationRow, SettingsSection, SettingsSeparator, SettingsSwitchRow } from '../screens/SettingsList';

export function reminderDaysLabel(days: number): string { return days === 0 ? 'Same day' : `${days} ${days === 1 ? 'day' : 'days'}`; }
export function reminderSummary(policy: ExpirationReminderPolicy): string {
  if (!policy.enabled) return 'Off';
  return [policy.upcoming ? policy.advanceDays === 0 ? 'On the expiration date' : `${reminderDaysLabel(policy.advanceDays)} before` : '', policy.expired ? 'when expired' : ''].filter(Boolean).join(' and ') || 'No reminders selected';
}
type Mode = 'defaults' | 'custom' | 'off';
const policyMode = (policy: ExpirationReminderPolicy | null): Mode => policy === null ? 'defaults' : policy.enabled ? 'custom' : 'off';
export function ExpirationReminderEditor({ initialPolicy, inheritedPolicy, disabled = false, onSave, onEditDays }: {
  readonly initialPolicy: ExpirationReminderPolicy | null;
  readonly inheritedPolicy?: ExpirationReminderPolicy;
  readonly disabled?: boolean;
  readonly onSave: (value: ExpirationReminderPolicy | null) => Promise<void>;
  readonly onEditDays: () => void;
}) {
  const initial = initialPolicy ?? inheritedPolicy;
  if (!initial) throw new Error('A reminder policy is required.');
  const [draft, setDraft] = useState({ ...initial });
  const [mode, setMode] = useState<Mode>(policyMode(initialPolicy));
  const [error, setError] = useState(false);
  const [saving, setSaving] = useState(false);
  const pending = useRef(false);
  const mounted = useRef(true);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  const source = JSON.stringify([initialPolicy, inheritedPolicy]);
  useEffect(() => {
    if (error || pending.current) return;
    const policy = initialPolicy ?? inheritedPolicy;
    if (!policy) return;
    setDraft({ ...policy }); setMode(policyMode(initialPolicy));
  }, [source, error]);
  const displayed = mode === 'defaults' && inheritedPolicy ? inheritedPolicy : draft;
  const locked = disabled || saving;
  async function commit(next: ExpirationReminderPolicy, nextMode = mode) {
    if (disabled || pending.current) return;
    pending.current = true; setSaving(true); setError(false);
    const value = { ...next, enabled: nextMode !== 'off' };
    setDraft(value); setMode(nextMode);
    try { await onSave(nextMode === 'defaults' ? null : value); }
    catch { if (mounted.current) setError(true); }
    finally { pending.current = false; if (mounted.current) setSaving(false); }
  }
  return <>
    {inheritedPolicy ? <SettingsSection footer={mode === 'defaults' ? `Inventory defaults: ${reminderSummary(displayed)}.` : undefined}>
      <SettingsChoiceRow label="Use defaults" accessibilityLabel="Use inventory defaults" selected={mode === 'defaults'} disabled={locked || error} onPress={() => void commit(displayed, 'defaults')} />
      <SettingsSeparator /><SettingsChoiceRow label="Custom" accessibilityLabel="Custom reminders" selected={mode === 'custom'} disabled={locked || error} onPress={() => void commit(displayed, 'custom')} />
      <SettingsSeparator /><SettingsChoiceRow label="Off" accessibilityLabel="Turn off type reminders" selected={mode === 'off'} disabled={locked || error} onPress={() => void commit(displayed, 'off')} />
    </SettingsSection> : <SettingsSection title="Inventory defaults" footer="Types use these rules unless you customize them below.">
      <SettingsSwitchRow label="Default reminders" value={mode !== 'off'} disabled={locked || error} onValueChange={enabled => void commit(draft, enabled ? 'custom' : 'off')} />
    </SettingsSection>}
    {mode === 'custom' ? <SettingsSection>
      <SettingsNavigationRow label="Before expiration" accessibilityLabel="Before expiration" value={draft.upcoming ? reminderDaysLabel(draft.advanceDays) : 'Off'} disabled={locked || error} onPress={onEditDays} />
      <SettingsSeparator /><SettingsSwitchRow label="When expired" value={draft.expired} disabled={locked || error} onValueChange={expired => void commit({ ...draft, expired })} />
    </SettingsSection> : null}
    {saving ? <SettingsSection><SettingsLoadingRow label="Saving reminders…" /></SettingsSection> : null}
    {error ? <SettingsSection footer="Could not save. Your change is kept here until you retry or discard it.">
      <SettingsActionRow accessibilityLabel="Retry saving reminders" label="Retry" disabled={locked} onPress={() => void commit(draft)} />
      <SettingsSeparator /><SettingsActionRow accessibilityLabel="Discard reminder changes" label="Discard change" disabled={locked} onPress={() => setError(false)} />
    </SettingsSection> : null}
  </>;
}
