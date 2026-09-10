import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { NotificationPreferencesSession } from '../../application/notifications/NotificationPreferencesSession';
import type { InventoryAssetTypesQuery } from '../../application/assets/InventoryAssetTypesQuery';
import type { NotificationPreferences } from '../../domain/notifications/Notification';
import type { CustomAssetTypeDefinition } from '../../domain/customization/Customization';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import { ExpirationReminderEditor } from '../components/ExpirationReminderEditor';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing } from '../theme/tokens';

export function NotificationSettingsScreen({ tenantId, inventoryId, session, assetTypesQuery }: {
  readonly tenantId: string; readonly inventoryId: string;
  readonly session: NotificationPreferencesSession;
  readonly assetTypesQuery: Pick<InventoryAssetTypesQuery, 'execute'>;
}) {
  const colors = useAppearancePalette();
  const [preferences, setPreferences] = useState<NotificationPreferences | null>(null);
  const [types, setTypes] = useState<readonly CustomAssetTypeDefinition[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [timezone, setTimezone] = useState('');
  const [timezoneSaved, setTimezoneSaved] = useState(false);
  const timezoneInitialized = useRef(false);
  const pending = useRef(false);
  const mounted = useRef(true);
  const controller = useRef<AbortController | null>(null);
  useEffect(() => { mounted.current = true; void load(); return () => { mounted.current = false; controller.current?.abort(); }; }, []);
  const buttonStyle = [styles.button, { borderColor: colors.controlBorder }];
  const validTimezone = validZone(timezone);
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
        if (!timezoneInitialized.current) { setTimezone(loaded.timezone); timezoneInitialized.current = true; }
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
      if (mounted.current && !request.signal.aborted) setPreferences(loaded);
    } catch (caught) {
      if (mounted.current && !request.signal.aborted) setError(message(caught, 'Your changes are still here. Refresh saved settings before trying again.'));
      throw caught;
    } finally { pending.current = false; if (mounted.current) setBusy(false); }
  }
  async function saveTimezone() {
    if (!validTimezone || pending.current) return;
    setTimezoneSaved(false);
    try { await save((signal) => session.saveTimezone(timezone, { signal })); if (mounted.current) setTimezoneSaved(true); }
    catch { /* The shared save path presents the failure and preserves the draft. */ }
  }
  return <ScrollView keyboardShouldPersistTaps="handled" keyboardDismissMode={appKeyboardDismissMode()} style={{ backgroundColor: colors.background }} contentContainerStyle={styles.content}>
    <Text accessibilityRole="header" style={[styles.title, { color: colors.text }]}>Notifications</Text>
    <Text style={{ color: colors.textMuted }}>Your reminders for this inventory. Other members have their own settings.</Text>
    {error ? <Text accessibilityRole="alert" style={{ color: colors.danger }}>{error}</Text> : null}
    {busy && !preferences ? <ActivityIndicator accessibilityLabel="Loading reminders" color={colors.action} /> : null}
    <Pressable accessibilityRole="button" accessibilityLabel={preferences ? 'Refresh saved settings' : 'Retry loading reminders'} disabled={busy} onPress={load} style={buttonStyle}><Text style={{ color: colors.text }}>{preferences ? 'Refresh saved settings' : 'Retry loading reminders'}</Text></Pressable>
    {preferences ? <>
      <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>Inventory defaults</Text>
      <ExpirationReminderEditor initialPolicy={preferences.defaults} disabled={busy} onSave={async (policy) => { if (policy) await save((signal) => session.saveDefaults(policy, { signal })); }} />
      <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>Calendar timezone</Text>
      <AppTextInput accessibilityLabel="Timezone" value={timezone} editable={!busy} autoCapitalize="none" autoCorrect={false} onChangeText={(value) => { setTimezone(value); setTimezoneSaved(false); }} style={[styles.input, { color: colors.text, borderColor: colors.controlBorder }]} />
      <Text style={{ color: colors.textMuted }}>{validTimezone ? `Saved timezone: ${preferences.timezone}` : 'Enter a timezone such as America/New_York.'}</Text>
      <Pressable accessibilityRole="button" accessibilityLabel="Save timezone" disabled={busy || !validTimezone} onPress={saveTimezone} style={buttonStyle}><Text style={{ color: colors.text }}>Save timezone</Text></Pressable>
      {timezoneSaved ? <Text accessibilityLiveRegion="polite" style={{ color: colors.textMuted }}>Timezone saved.</Text> : null}
      <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>Asset type reminders</Text>
      {types.map((type) => <View key={type.id} style={styles.type}>
        <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>{type.displayName}</Text>
        <ExpirationReminderEditor initialPolicy={preferences.overrides.find((entry) => entry.customAssetTypeId === type.id)?.settings ?? null} inheritedPolicy={preferences.defaults} disabled={busy} onSave={(policy) => save((signal) => session.saveTypeOverride(type.id, policy, { signal }))} />
      </View>)}
      {!types.length ? <Text style={{ color: colors.textMuted }}>Enable expiration tracking on an asset type to customize its reminders here.</Text> : null}
    </> : null}
  </ScrollView>;
}
function validZone(value: string) { try { new Intl.DateTimeFormat('en', { timeZone: value }); return value.trim().length > 0; } catch { return false; } }
function message(error: unknown, fallback: string) { return error instanceof NotificationFailure ? error.message : fallback; }
const styles = StyleSheet.create({
  content: { padding: spacing.lg, gap: spacing.md }, type: { gap: spacing.sm },
  title: { fontSize: 24, fontWeight: '700' }, heading: { fontSize: 18, fontWeight: '600' },
  input: { minHeight: 44, borderWidth: 1, borderRadius: radius.md, padding: spacing.sm },
  button: { minHeight: 44, borderWidth: 1, borderRadius: radius.md, padding: spacing.sm, justifyContent: 'center' }
});
