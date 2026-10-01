import { t } from '../../presentation/localization';
import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { SettingsPickerRow } from './SettingsPickerRow';
import { useEffect, useRef, useState } from 'react';
import type { ExpirationReminderPolicy } from '../../domain/notifications/Notification';
import { SettingsActionRow, SettingsLoadingRow, SettingsNavigationRow, SettingsSection, SettingsSeparator, SettingsSwitchRow } from '../screens/SettingsList';

export function reminderDaysLabel(days: number): string { return days === 0 ? t('reminder.sameDay') : t('reminder.days', { count: days }); }
export function reminderSummary(policy: ExpirationReminderPolicy): string {
  if (!policy.enabled) return t('reminder.off');
  if (!policy.upcoming) return policy.expired ? t('reminder.expired') : t('reminder.none');
  if (policy.advanceDays === 0) return policy.expired ? t('reminder.todayAndExpired') : t('reminder.today');
  return t(policy.expired ? 'reminder.beforeAndExpired' : 'reminder.before', { count: policy.advanceDays });
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
  const capturePresentation = useTaskPresentation();
  const initial = initialPolicy ?? inheritedPolicy;
  if (!initial) throw new Error('A reminder policy is required.');
  const [draft, setDraft] = useState({ ...initial });
  const [mode, setMode] = useState<Mode>(policyMode(initialPolicy));
  const [error, setError] = useState(false);
  const [saving, setSaving] = useState(false);
  const [reconciliation, setReconciliation] = useState(0);
  const pending = useRef(false);
  const mounted = useRef(true);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  const source = JSON.stringify([initialPolicy, inheritedPolicy]);
  useEffect(() => {
    if (error || pending.current) return;
    const policy = initialPolicy ?? inheritedPolicy;
    if (!policy) return;
    setDraft({ ...policy }); setMode(policyMode(initialPolicy));
  }, [source, error, reconciliation]);
  const displayed = mode === 'defaults' && inheritedPolicy ? inheritedPolicy : draft;
  const locked = disabled || saving;
  async function commit(next: ExpirationReminderPolicy, nextMode = mode) {
    const canPresent = capturePresentation();
    if (!canPresent() || disabled || pending.current) return;
    pending.current = true; setSaving(true); setError(false);
    const value = { ...next, enabled: nextMode !== 'off' };
    setDraft(value); setMode(nextMode);
    try { await onSave(nextMode === 'defaults' ? null : value); }
    catch { if (canPresent()) setError(true); }
    finally {
      pending.current = false;
      if (mounted.current) {
        if (!canPresent()) setReconciliation(current => current + 1);
        setSaving(false);
      }
    }
  }
  return <>
    {inheritedPolicy ? <SettingsSection footer={mode === 'defaults' ? t('mobile.ExpirationReminderEditor.inventoryDefaults2', { value: String(reminderSummary(displayed)) }) : undefined}>
      <SettingsPickerRow label={t('mobile.ExpirationReminderEditor.reminders')} accessibilityLabel={t('mobile.ExpirationReminderEditor.chooseReminderMode')} value={mode}
        options={[{value:'defaults',label:t('mobile.ExpirationReminderEditor.useDefaults')},{value:'custom',label:t('mobile.ExpirationReminderEditor.custom')},{value:'off',label:t('mobile.ExpirationReminderEditor.off')}] as const}
        disabled={locked || error} onChange={nextMode => { if (nextMode !== mode) void commit(displayed, nextMode); }} />
    </SettingsSection> : <SettingsSection title={t('mobile.ExpirationReminderEditor.inventoryDefaults')} footer={t('mobile.ExpirationReminderEditor.typesUseTheseRulesUnlessYouCustomizeThemBelow')}>
      <SettingsSwitchRow label={t('mobile.ExpirationReminderEditor.defaultReminders')} value={mode !== 'off'} disabled={locked || error} onValueChange={enabled => void commit(draft, enabled ? 'custom' : 'off')} />
    </SettingsSection>}
    {mode === 'custom' ? <SettingsSection>
      <SettingsNavigationRow label={t('mobile.ExpirationReminderEditor.beforeExpiration')} accessibilityLabel={t('mobile.ExpirationReminderEditor.beforeExpiration')} value={draft.upcoming ? reminderDaysLabel(draft.advanceDays) : 'Off'} disabled={locked || error} onPress={onEditDays} />
      <SettingsSeparator /><SettingsSwitchRow label={t('mobile.ExpirationReminderEditor.whenExpired')} value={draft.expired} disabled={locked || error} onValueChange={expired => void commit({ ...draft, expired })} />
    </SettingsSection> : null}
    {saving ? <SettingsSection><SettingsLoadingRow label={t('mobile.ExpirationReminderEditor.savingReminders')} /></SettingsSection> : null}
    {error ? <SettingsSection footer={t('mobile.ExpirationReminderEditor.couldNotSaveYourChangeIsKeptHereUntil')}>
      <SettingsActionRow accessibilityLabel={t('mobile.ExpirationReminderEditor.retrySavingReminders')} label={t('mobile.ExpirationReminderEditor.retry')} disabled={locked} onPress={() => void commit(draft)} />
      <SettingsSeparator /><SettingsActionRow accessibilityLabel={t('mobile.ExpirationReminderEditor.discardReminderChanges')} label={t('mobile.ExpirationReminderEditor.discardChange')} disabled={locked} onPress={() => setError(false)} />
    </SettingsSection> : null}
  </>;
}
