import { t } from '../../presentation/localization';
import { usePullRefresh } from '../serverState/usePullRefresh';
import { useCallback, useRef, useState } from 'react';
import { Stack, useFocusEffect } from 'expo-router';
import { AppState, Linking, RefreshControl, ScrollView, Text, View } from 'react-native';
import type { PushSession } from '../../application/notifications/PushSession';
import type { NotificationPreferencesSession } from '../../application/notifications/NotificationPreferencesSession';
import type { InventoryAssetTypesQuery } from '../../application/assets/InventoryAssetTypesQuery';
import type { NotificationPreferences } from '../../domain/notifications/Notification';
import type { CustomAssetTypeDefinition } from '../../domain/customization/Customization';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import { isAccessFailure } from '../serverState/isAccessFailure';
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
  const [error, setError] = useState('');
  const [pushOutcome, setPushOutcome] = useState<'enabled' | 'denied'>();
  const [openingSettings, setOpeningSettings] = useState(false);
  const [settingsError, setSettingsError] = useState('');
  const settingsLaunch = useRef<object | undefined>(undefined);
  const pushMessage = pushOutcome === 'enabled' ? t('mobile.NotificationSettingsScreen.deviceSetupCompleted') : pushOutcome === 'denied' ? t('mobile.NotificationSettingsScreen.allowNotificationsForStuffStashInYourDeviceSettings') : '';
  const pending = useRef(false); const mounted = useRef(true);
  const accessLost = useRef(false);
  const feedbackGeneration = useRef(0);
  const controller = useRef<AbortController | null>(null);
  useFocusEffect(useCallback(() => {
    mounted.current = true; feedbackGeneration.current += 1; setPushOutcome(undefined);
    settingsLaunch.current = undefined; setOpeningSettings(false); setSettingsError(''); void load();
    const appState = AppState.addEventListener('change', state => {
      if (state === 'background') { feedbackGeneration.current += 1; setPushOutcome(undefined); settingsLaunch.current = undefined; setOpeningSettings(false); setSettingsError(''); }
    });
    return () => { appState.remove(); feedbackGeneration.current += 1; mounted.current = false; settingsLaunch.current = undefined; controller.current?.abort(); controller.current = null; pending.current = false; };
  }, [session, tenantId, inventoryId]));
  async function openDeviceSettings() {
    if (!mounted.current || settingsLaunch.current) return;
    const launch = {}; settingsLaunch.current = launch;
    const generation = feedbackGeneration.current;
    setOpeningSettings(true); setSettingsError('');
    try { await Linking.openSettings(); }
    catch {
      if (mounted.current && settingsLaunch.current === launch && feedbackGeneration.current === generation) {
        setSettingsError(t('mobile.NotificationSettingsScreen.deviceSettingsCouldNotBeOpenedOpenSettingsOn'));
      }
    } finally {
      if (settingsLaunch.current === launch) { settingsLaunch.current = undefined; if (mounted.current) setOpeningSettings(false); }
    }
  }
  async function load() {
    if (pending.current) return;
    pending.current = true; setBusy(true); setError('');
    const request = new AbortController(); controller.current = request;
    try {
      const loadedTypes = await assetTypesQuery.execute(tenantId, inventoryId, { signal: request.signal });
      const first = session.snapshot === null;
      const loaded = first ? await session.initialize(Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC', { signal: request.signal }) : await session.refresh({ signal: request.signal });
      if (mounted.current && !request.signal.aborted) {
        accessLost.current = false;
        setTypes(loadedTypes.filter((type) => type.expirationEnabled)); setPreferences(loaded);
      }
    } catch (caught) { if (mounted.current && !request.signal.aborted) presentFailure(caught, t('mobile.NotificationSettingsScreen.reminderSettingsCouldNotBeLoadedTryAgain')); }
    finally { if (controller.current === request) { pending.current = false; if (mounted.current) setBusy(false); } }
  }
  async function save(operation: (signal: AbortSignal) => Promise<NotificationPreferences>) {
    if (!mounted.current || accessLost.current) throw new Error('Reload authorized reminder settings before saving.');
    if (pending.current) throw new Error('Another settings request is in progress.');
    pending.current = true; setBusy(true); setError('');
    const request = new AbortController(); controller.current = request;
    try {
      const loaded = await operation(request.signal);
      if (!mounted.current || request.signal.aborted) throw new Error('Settings navigation changed.');
      setPreferences(loaded); onChanged?.();
    } catch (caught) {
      if (mounted.current && !request.signal.aborted) presentFailure(caught, t('mobile.NotificationSettingsScreen.yourChangesAreStillHereRefreshSavedSettingsBefore'));
      throw caught;
    } finally { if (controller.current === request) { pending.current = false; if (mounted.current) setBusy(false); } }
  }
  function presentFailure(caught: unknown, fallback: string) {
    if (isAccessFailure(caught) || (caught instanceof NotificationFailure && (caught.kind === 'authentication-required' || caught.kind === 'permission-denied'))) {
      accessLost.current = true;
      setPreferences(null); setTypes([]); setPushOutcome(undefined);
    }
    setError(message(caught, fallback));
  }
  async function changePush(enabled: boolean) {
    if (!pushSession || pending.current) return;
    const feedbackOwner = feedbackGeneration.current;
    setPushOutcome(undefined);
    try {
      await save(async (signal) => {
        if (!enabled) return session.savePushEnabled(false, { signal });
        const result = await pushSession.enable(tenantId, inventoryId, { signal });
        if (result === 'denied' && mounted.current && !signal.aborted && feedbackOwner === feedbackGeneration.current) {
          setPushOutcome('denied');
        }
        const refreshed = await session.refresh({ signal });
        if (result === 'enabled' && mounted.current && !signal.aborted && feedbackOwner === feedbackGeneration.current) setPushOutcome('enabled');
        return refreshed;
      });
    } catch { /* The shared save path presents errors without changing the switch optimistically. */ }
  }
  const { refreshing, refresh } = usePullRefresh(async () => {
    if (!pending.current) await load();
  });
  const typeId = page.kind === 'type' || page.kind === 'timing' ? page.typeId : undefined;
  const type = typeId ? types.find(entry => entry.id === typeId) : undefined;
  const override = preferences?.overrides.find(entry => entry.customAssetTypeId === typeId)?.settings;
  const policy = override ?? preferences?.defaults;
  const unavailable = !!typeId && !type;
  const title = page.kind === 'timezone' ? t('mobile.NotificationSettingsScreen.timeZone') : page.kind === 'timing' ? t('mobile.NotificationSettingsScreen.beforeExpiration') : page.kind === 'type' ? type?.displayName ?? t('mobile.NotificationSettingsScreen.typeReminders') : t('mobile.NotificationSettingsScreen.expirationReminders');
  return <><Stack.Screen options={{ title }} /><ScrollView style={styles.shell} contentContainerStyle={[styles.content, {flexGrow:1}]} automaticallyAdjustKeyboardInsets alwaysBounceVertical contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled" keyboardDismissMode={appKeyboardDismissMode()}
    refreshControl={page.kind === 'overview' ? <RefreshControl refreshing={refreshing} onRefresh={() => void refresh()} tintColor={colors.action} /> : undefined}>
    {error ? <View style={styles.detailHeader}><Text accessibilityRole="alert" style={{color:colors.danger}}>{error}</Text></View> : null}
    {settingsError ? <View style={styles.detailHeader}><Text accessibilityRole="alert" style={{color:colors.danger}}>{settingsError}</Text></View> : null}
    {busy && !preferences ? <SettingsSection><SettingsLoadingRow label={t('mobile.NotificationSettingsScreen.loadingReminders')} /></SettingsSection> : null}
    {!preferences && !busy ? <SettingsSection><SettingsActionRow accessibilityLabel={t('mobile.NotificationSettingsScreen.retryLoadingReminders')} label={t('mobile.NotificationSettingsScreen.retry')} onPress={() => void load()} /></SettingsSection> : null}
    {preferences && unavailable ? <SettingsSection footer={t('mobile.NotificationSettingsScreen.thisAssetTypeIsNoLongerAvailableForExpiration')}><SettingsActionRow label={t('mobile.NotificationSettingsScreen.back')} onPress={onBack} /></SettingsSection> : null}
    {preferences && !unavailable ? page.kind === 'timezone' ? <TimeZonePicker value={preferences.timezone} disabled={busy} onChange={async zone => { await save(signal => session.saveTimezone(zone,{signal})); onBack(); }} />
      : page.kind === 'timing' && policy ? <ReminderTimingEditor policy={policy} disabled={busy} onDone={onBack} onSave={async value => { await save(signal => typeId ? session.saveTypeOverride(typeId,value,{signal}) : session.saveDefaults(value,{signal})); }} />
      : page.kind === 'type' ? <ExpirationReminderEditor initialPolicy={override ?? null} inheritedPolicy={preferences.defaults} disabled={busy} onEditDays={() => onNavigate({kind:'timing',typeId:page.typeId})} onSave={value => save(signal => session.saveTypeOverride(page.typeId,value,{signal}))} />
      : <>
        <View style={styles.detailHeader}><Text style={styles.secondaryText}>{t('mobile.NotificationSettingsScreen.yourRemindersForThisInventory')}</Text></View>
        {pushSession ? <SettingsSection title={t('mobile.NotificationSettingsScreen.notifications')} footer={pushMessage || t('mobile.NotificationSettingsScreen.pushAlertsAlsoNeedServerDeliverySupportInApp')}>
          <SettingsSwitchRow label={t('mobile.NotificationSettingsScreen.pushNotifications')} value={preferences.pushEnabled} disabled={busy} onValueChange={enabled => void changePush(enabled)} />
          {preferences.pushEnabled ? <><SettingsSeparator /><SettingsActionRow accessibilityLabel={pushOutcome === 'enabled' ? t('mobile.NotificationSettingsScreen.openDeviceSettings') : t('mobile.NotificationSettingsScreen.setUpAlertsOnThisDevice')} label={pushOutcome === 'enabled' ? t('mobile.NotificationSettingsScreen.openDeviceSettings') : t('mobile.NotificationSettingsScreen.enableOnThisDevice')} disabled={busy || openingSettings} onPress={() => { if (pushOutcome === 'enabled') void openDeviceSettings(); else void changePush(true); }} /></> : null}
          {pushOutcome === 'denied' ? <><SettingsSeparator /><SettingsActionRow accessibilityLabel={t('mobile.NotificationSettingsScreen.openNotificationSystemSettings')} label={t('mobile.NotificationSettingsScreen.openSettings')} disabled={openingSettings} onPress={() => void openDeviceSettings()} /></> : null}
        </SettingsSection> : null}
        <ExpirationReminderEditor initialPolicy={preferences.defaults} disabled={busy} onEditDays={() => onNavigate({kind:'timing'})} onSave={async value => { if (value) await save(signal => session.saveDefaults(value,{signal})); }} />
        <SettingsSection footer={t('mobile.NotificationSettingsScreen.expirationDatesUseThisTimeZoneEvenWhenYou')}><SettingsNavigationRow label={t('mobile.NotificationSettingsScreen.timeZone')} accessibilityLabel={t('mobile.NotificationSettingsScreen.timeZone')} value={readableTimeZone(preferences.timezone)} disabled={busy} onPress={() => onNavigate({kind:'timezone'})} /></SettingsSection>
        <SettingsSection title={t('mobile.NotificationSettingsScreen.assetTypeReminders')} footer={!types.length ? t('mobile.NotificationSettingsScreen.enableExpirationTrackingOnATypeToCustomizeIts') : t('mobile.NotificationSettingsScreen.eachTypeUsesYourInventoryDefaultsUnlessYouChoose')}>
          {types.map((entry,index) => {
            const rule=preferences.overrides.find(value => value.customAssetTypeId === entry.id)?.settings;
            return <View key={entry.id}>{index ? <SettingsSeparator /> : null}<SettingsNavigationRow label={entry.displayName} accessibilityLabel={t('mobile.NotificationSettingsScreen.reminders', { displayName: String(entry.displayName) })} value={rule ? rule.enabled ? t('mobile.NotificationSettingsScreen.custom') : t('mobile.NotificationSettingsScreen.off') : t('mobile.NotificationSettingsScreen.usesDefaults')} disabled={busy} onPress={() => onNavigate({kind:'type',typeId:entry.id})} /></View>;
          })}
        </SettingsSection>
      </> : null}
  </ScrollView></>;
}
function message(error: unknown, fallback: string) { return error instanceof NotificationFailure ? error.message : fallback; }
