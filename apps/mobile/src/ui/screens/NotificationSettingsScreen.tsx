import { useCallback, useRef, useState } from 'react';
import { Stack, useFocusEffect } from 'expo-router';
import { Linking, RefreshControl, ScrollView, Text, View } from 'react-native';
import type { PushSession } from '../../application/notifications/PushSession';
import type { NotificationPreferencesSession } from '../../application/notifications/NotificationPreferencesSession';
import type { InventoryAssetTypesQuery } from '../../application/assets/InventoryAssetTypesQuery';
import type { NotificationPreferences } from '../../domain/notifications/Notification';
import type { CustomAssetTypeDefinition } from '../../domain/customization/Customization';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import { ExpirationReminderEditor } from '../components/ExpirationReminderEditor';
import { ReminderTimingEditor } from '../components/ReminderTimingEditor';
import { TimeZonePicker, readableTimeZone } from '../components/TimeZonePicker';
import { appKeyboardDismissMode } from '../components/AppTextInput';
import { SettingsActionRow, SettingsLoadingRow, SettingsNavigationRow, SettingsSection, SettingsSeparator, SettingsSwitchRow, useSettingsListStyles } from './SettingsList';

import type { NotificationSettingsPage, NotificationSettingsDestination } from '../presentation/NotificationSettingsDestination';
const overview: NotificationSettingsPage = {kind:'overview'};
export function NotificationSettingsScreen({ tenantId, inventoryId, session, assetTypesQuery, pushSession, onChanged, page = overview, onNavigate, onBack }: {
  readonly tenantId: string; readonly inventoryId: string; readonly page?: NotificationSettingsPage;
  readonly onNavigate: (page: NotificationSettingsDestination) => void; readonly onBack: () => void;
  readonly onChanged?: () => void; readonly session: NotificationPreferencesSession;
  readonly pushSession?: Pick<PushSession, 'enable'>; readonly assetTypesQuery: Pick<InventoryAssetTypesQuery, 'execute'>;
}) {
  const { palette: colors, styles } = useSettingsListStyles();
  const [preferences, setPreferences] = useState<NotificationPreferences | null>(null);
  const [types, setTypes] = useState<readonly CustomAssetTypeDefinition[]>([]);
  const [busy, setBusy] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState('');
  const [pushMessage, setPushMessage] = useState('');
  const pending = useRef(false); const mounted = useRef(true);
  const controller = useRef<AbortController | null>(null);
  useFocusEffect(useCallback(() => {
    mounted.current = true; void load();
    return () => { mounted.current = false; controller.current?.abort(); controller.current = null; pending.current = false; };
  }, [session, tenantId, inventoryId]));
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
    finally { if (controller.current === request) { pending.current = false; if (mounted.current) setBusy(false); } }
  }
  async function save(operation: (signal: AbortSignal) => Promise<NotificationPreferences>) {
    if (pending.current) throw new Error('Another settings request is in progress.');
    pending.current = true; setBusy(true); setError('');
    const request = new AbortController(); controller.current = request;
    try {
      const loaded = await operation(request.signal);
      if (!mounted.current || request.signal.aborted) throw new Error('Settings navigation changed.');
      setPreferences(loaded); onChanged?.();
    } catch (caught) {
      if (mounted.current && !request.signal.aborted) setError(message(caught, 'Your changes are still here. Refresh saved settings before trying again.'));
      throw caught;
    } finally { if (controller.current === request) { pending.current = false; if (mounted.current) setBusy(false); } }
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
        if (result === 'enabled' && mounted.current && !signal.aborted) setPushMessage('Notifications allowed on this device.');
        return refreshed;
      });
    } catch { /* The shared save path presents errors without changing the switch optimistically. */ }
  }
  async function refresh() {
    if (pending.current) return;
    setRefreshing(true);
    try { await load(); } finally { if (mounted.current) setRefreshing(false); }
  }
  const typeId = page.kind === 'type' || page.kind === 'timing' ? page.typeId : undefined;
  const type = typeId ? types.find(entry => entry.id === typeId) : undefined;
  const override = preferences?.overrides.find(entry => entry.customAssetTypeId === typeId)?.settings;
  const policy = override ?? preferences?.defaults;
  const unavailable = !!typeId && !type;
  const title = page.kind === 'timezone' ? 'Time zone' : page.kind === 'timing' ? 'Before expiration' : page.kind === 'type' ? type?.displayName ?? 'Type reminders' : 'Expiration reminders';
  return <><Stack.Screen options={{ title }} /><ScrollView style={styles.shell} contentContainerStyle={[styles.content, {flexGrow:1}]} automaticallyAdjustKeyboardInsets alwaysBounceVertical contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled" keyboardDismissMode={appKeyboardDismissMode()}
    refreshControl={page.kind === 'overview' ? <RefreshControl refreshing={refreshing} onRefresh={() => void refresh()} tintColor={colors.action} /> : undefined}>
    {error ? <View style={styles.detailHeader}><Text accessibilityRole="alert" style={{color:colors.danger}}>{error}</Text></View> : null}
    {busy && !preferences ? <SettingsSection><SettingsLoadingRow label="Loading reminders" /></SettingsSection> : null}
    {!preferences && !busy ? <SettingsSection><SettingsActionRow accessibilityLabel="Retry loading reminders" label="Retry" onPress={() => void load()} /></SettingsSection> : null}
    {preferences && unavailable ? <SettingsSection footer="This asset type is no longer available for expiration reminders."><SettingsActionRow label="Back" onPress={onBack} /></SettingsSection> : null}
    {preferences && !unavailable ? page.kind === 'timezone' ? <TimeZonePicker value={preferences.timezone} disabled={busy} onChange={async zone => { await save(signal => session.saveTimezone(zone,{signal})); onBack(); }} />
      : page.kind === 'timing' && policy ? <ReminderTimingEditor policy={policy} disabled={busy} onDone={onBack} onSave={async value => { await save(signal => typeId ? session.saveTypeOverride(typeId,value,{signal}) : session.saveDefaults(value,{signal})); }} />
      : page.kind === 'type' ? <ExpirationReminderEditor initialPolicy={override ?? null} inheritedPolicy={preferences.defaults} disabled={busy} onEditDays={() => onNavigate({kind:'timing',typeId:page.typeId})} onSave={value => save(signal => session.saveTypeOverride(page.typeId,value,{signal}))} />
      : <>
        <View style={styles.detailHeader}><Text style={styles.secondaryText}>Your reminders for this inventory.</Text></View>
        {pushSession ? <SettingsSection title="Notifications" footer={pushMessage || 'Push alerts also need server delivery support. In-app reminders are independent.'}>
          <SettingsSwitchRow label="Push notifications" value={preferences.pushEnabled} disabled={busy} onValueChange={enabled => void changePush(enabled)} />
          {preferences.pushEnabled ? <><SettingsSeparator /><SettingsActionRow accessibilityLabel="Set up alerts on this device" label={pushMessage === 'Notifications allowed on this device.' ? 'Open device settings' : 'Enable on this device'} disabled={busy} onPress={() => { if (pushMessage === 'Notifications allowed on this device.') void Linking.openSettings(); else void changePush(true); }} /></> : null}
          {pushMessage.startsWith('Allow notifications') ? <><SettingsSeparator /><SettingsActionRow accessibilityLabel="Open notification system settings" label="Open Settings" onPress={() => void Linking.openSettings()} /></> : null}
        </SettingsSection> : null}
        <ExpirationReminderEditor initialPolicy={preferences.defaults} disabled={busy} onEditDays={() => onNavigate({kind:'timing'})} onSave={async value => { if (value) await save(signal => session.saveDefaults(value,{signal})); }} />
        <SettingsSection footer="Expiration dates use this time zone, even when you travel."><SettingsNavigationRow label="Time zone" accessibilityLabel="Time zone" value={readableTimeZone(preferences.timezone)} disabled={busy} onPress={() => onNavigate({kind:'timezone'})} /></SettingsSection>
        <SettingsSection title="Asset type reminders" footer={!types.length ? 'Enable expiration tracking on a type to customize its reminders.' : 'Each type uses your inventory defaults unless you choose Custom or Off.'}>
          {types.map((entry,index) => {
            const rule=preferences.overrides.find(value => value.customAssetTypeId === entry.id)?.settings;
            return <View key={entry.id}>{index ? <SettingsSeparator /> : null}<SettingsNavigationRow label={entry.displayName} accessibilityLabel={`${entry.displayName} reminders`} value={rule ? rule.enabled ? 'Custom' : 'Off' : 'Uses defaults'} disabled={busy} onPress={() => onNavigate({kind:'type',typeId:entry.id})} /></View>;
          })}
        </SettingsSection>
      </> : null}
  </ScrollView></>;
}
function message(error: unknown, fallback: string) { return error instanceof NotificationFailure ? error.message : fallback; }
