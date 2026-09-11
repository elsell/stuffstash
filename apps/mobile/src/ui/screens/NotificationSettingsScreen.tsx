import { TimeZonePicker } from '../components/TimeZonePicker';
import { SelectionRow } from '../components/SelectionRow';
import { Stack } from 'expo-router';
import type { PushSession } from '../../application/notifications/PushSession';
import { AppSwitchField } from '../components/AppSwitchField';
import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, Linking, Pressable, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { NotificationPreferencesSession } from '../../application/notifications/NotificationPreferencesSession';
import type { InventoryAssetTypesQuery } from '../../application/assets/InventoryAssetTypesQuery';
import type { NotificationPreferences } from '../../domain/notifications/Notification';
import type { CustomAssetTypeDefinition } from '../../domain/customization/Customization';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import { ExpirationReminderEditor } from '../components/ExpirationReminderEditor';
import { appKeyboardDismissMode } from '../components/AppTextInput';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';

export function NotificationSettingsScreen({ tenantId, inventoryId, session, assetTypesQuery, pushSession, onChanged }: {
  readonly tenantId: string; readonly inventoryId: string;
  readonly onChanged?: () => void;
  readonly session: NotificationPreferencesSession;
  readonly pushSession?: Pick<PushSession, 'enable'>;
  readonly assetTypesQuery: Pick<InventoryAssetTypesQuery, 'execute'>;
}) {
  const colors = useAppearancePalette();
  const [selectedType, setSelectedType] = useState<string | null>(null);
  const [preferences, setPreferences] = useState<NotificationPreferences | null>(null);
  const [types, setTypes] = useState<readonly CustomAssetTypeDefinition[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [pushMessage, setPushMessage] = useState('');
  const pending = useRef(false);
  const mounted = useRef(true);
  const controller = useRef<AbortController | null>(null);
  useEffect(() => { mounted.current = true; void load(); return () => { mounted.current = false; controller.current?.abort(); }; }, []);
  const buttonStyle = [styles.button, { borderColor: colors.controlBorder }];
  async function load() {
    if (pending.current) return;
    pending.current = true; setBusy(true); setError('');
    const request = new AbortController(); controller.current = request;
    try {
      const loadedTypes = await assetTypesQuery.execute(tenantId, inventoryId, { signal: request.signal });
      const first = session.snapshot === null;
      const loaded = first ? await session.initialize(Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC', { signal: request.signal }) : await session.refresh({ signal: request.signal });
      if (mounted.current && !request.signal.aborted) {
        setTypes(loadedTypes.filter((type) => type.expirationEnabled)); setPreferences(loaded);
      }
    } catch (caught) { if (mounted.current && !request.signal.aborted) setError(message(caught, 'Reminder settings could not be loaded. Try again.')); }
    finally { pending.current = false; if (mounted.current) setBusy(false); }
  }
  async function save(operation: (signal: AbortSignal) => Promise<NotificationPreferences>) {
    if (pending.current) throw new Error('Another settings request is in progress.');
    pending.current = true; setBusy(true); setError('');
    const request = new AbortController(); controller.current = request;
    try {
      const loaded = await operation(request.signal);
      if (mounted.current && !request.signal.aborted) { setPreferences(loaded); onChanged?.(); }
    } catch (caught) {
      if (mounted.current && !request.signal.aborted) setError(message(caught, 'Your changes are still here. Refresh saved settings before trying again.'));
      throw caught;
    } finally { pending.current = false; if (mounted.current) setBusy(false); }
  }
  async function changePush(enabled: boolean) {
    if (!pushSession || pending.current) return;
    setPushMessage('');
    try {
      await save(async (signal) => {
        if (!enabled) return session.savePushEnabled(false, { signal });
        const result = await pushSession.enable(tenantId, inventoryId, { signal });
        if (result === 'denied' && mounted.current && !signal.aborted) {
          setPushMessage('Allow notifications for Stuff Stash in your device settings, then try again.');
        }
        const refreshed = await session.refresh({ signal });
        if (result === 'enabled' && mounted.current && !signal.aborted) setPushMessage('Permission allowed on this device. Delivery also requires your server to support push notifications.');
        return refreshed;
      });
    } catch { /* The shared save path presents errors without changing the switch optimistically. */ }
  }
  return <><Stack.Screen options={{ title: 'Expiration reminders' }} /><ScrollView automaticallyAdjustKeyboardInsets contentInsetAdjustmentBehavior="automatic" refreshControl={<RefreshControl refreshing={busy && !!preferences} onRefresh={() => void load()} tintColor={colors.action} />} keyboardShouldPersistTaps="handled" keyboardDismissMode={appKeyboardDismissMode()} style={{ backgroundColor: colors.background }} contentContainerStyle={styles.content}>
    <Text style={{ color: colors.textMuted }}>Your reminders for this inventory. Other members have their own settings.</Text>
    {error ? <Text accessibilityRole="alert" style={{ color: colors.danger }}>{error}</Text> : null}
    {busy && !preferences ? <ActivityIndicator accessibilityLabel="Loading reminders" color={colors.action} /> : null}
    {!preferences && !busy ? <Pressable accessibilityRole="button" accessibilityLabel="Retry loading reminders" onPress={() => void load()} style={buttonStyle}><Text style={{ color: colors.action }}>Retry</Text></Pressable> : null}
    {preferences ? <>
      {pushSession ? <>
        <AppSwitchField label="Push notifications" description="For this inventory. In-app reminders are independent." value={preferences.pushEnabled} disabled={busy} onValueChange={(enabled) => { void changePush(enabled); }} />
        {preferences.pushEnabled ? <Pressable accessibilityRole="button" accessibilityLabel="Set up alerts on this device" disabled={busy} onPress={() => { void changePush(true); }} style={buttonStyle}><Text style={{ color: colors.text }}>Set up alerts on this device</Text></Pressable> : null}
        {busy ? <ActivityIndicator accessibilityLabel="Saving notification settings" color={colors.action} /> : null}
        {pushMessage ? <Text accessibilityLiveRegion="polite" style={{ color: colors.textMuted }}>{pushMessage}</Text> : null}
        {pushMessage.startsWith('Allow notifications') ? <Pressable accessibilityRole="button" accessibilityLabel="Open notification system settings" onPress={() => void Linking.openSettings()} style={buttonStyle}><Text style={{ color: colors.action }}>Open Settings</Text></Pressable> : null}
      </> : null}
      <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>Inventory defaults</Text>
      <ExpirationReminderEditor initialPolicy={preferences.defaults} disabled={busy} onSave={async (policy) => { if (policy) await save((signal) => session.saveDefaults(policy, { signal })); }} />
      <TimeZonePicker value={preferences.timezone} disabled={busy} onChange={zone => save(signal => session.saveTimezone(zone, { signal }))} />
      <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>Asset type reminders</Text>
      {!preferences.defaults.enabled && preferences.overrides.some(entry => entry.settings.enabled) ? <Text style={{ color: colors.textMuted }}>Default reminders are off. Some types use custom reminders.</Text> : null}
      {types.map(type => {
        const override = preferences.overrides.find(entry => entry.customAssetTypeId === type.id)?.settings;
        return <SelectionRow key={type.id} label={type.displayName} value={override ? override.enabled ? 'Custom' : 'Off' : 'Uses defaults'} expanded={selectedType === type.id} disabled={busy} onPress={() => setSelectedType(current => current === type.id ? null : type.id)}>
          <ExpirationReminderEditor initialPolicy={override ?? null} inheritedPolicy={preferences.defaults} disabled={busy} onSave={policy => save(signal => session.saveTypeOverride(type.id, policy, { signal }))} />
        </SelectionRow>;
      })}
      {!types.length ? <Text style={{ color: colors.textMuted }}>Enable expiration tracking on an asset type to customize its reminders here.</Text> : null}
    </> : null}
  </ScrollView></>;
}
function message(error: unknown, fallback: string) { return error instanceof NotificationFailure ? error.message : fallback; }
const styles = StyleSheet.create({
  content: { padding: spacing.lg, gap: spacing.md },
  heading: { fontSize: 18, fontWeight: '600' },
  button: { minHeight: 44, padding: spacing.sm, justifyContent: 'center' }
});
