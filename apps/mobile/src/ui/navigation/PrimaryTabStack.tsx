import { t } from '../../presentation/localization';
import { Stack } from 'expo-router';
import { Platform } from 'react-native';
import { AppNoticeScreenLayout } from '../feedback/AppNoticeScreenLayout';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { VoiceTabContent } from './VoiceTabContent';
import { VoiceAccessoryContent } from './VoiceAccessoryContent';
import { nativeTabHeaderOptions } from './NativeTabHeader';

/** Ordinary destinations stay in their originating tab; tasks live above tabs. */
export function PrimaryTabStack({ root }: { readonly root: 'index' | 'search' }) {
  const palette = useAppearancePalette();
  const home = root === 'index';
  return <VoiceTabContent platform={Platform.OS} version={Platform.Version}
    accessory={<VoiceAccessoryContent placement="regular" />}>
    <Stack screenLayout={AppNoticeScreenLayout} screenOptions={{
      contentStyle: { backgroundColor: palette.background }, headerBackTitle: 'Back',
      headerStyle: { backgroundColor: palette.surface }, headerTintColor: palette.action,
      headerTitleStyle: { color: palette.text, fontWeight: '700' }
    }}>
      <Stack.Screen name={root} options={{ title: home ? 'Home' : 'Browse',
        ...nativeTabHeaderOptions(palette, Platform.OS, Platform.Version, home ? palette.background : palette.surface) }} />
      <Stack.Screen name="expiration" options={{ title: t('mobile.PrimaryTabStack.expiration') }} />
      <Stack.Screen name="settings/index" options={{ title: t('mobile.PrimaryTabStack.settings') }} />
      <Stack.Screen name="settings/account" options={{ title: t('mobile.PrimaryTabStack.account') }} />
      <Stack.Screen name="settings/appearance" options={{ title: t('mobile.PrimaryTabStack.appearance') }} />
      <Stack.Screen name="settings/sharing" options={{ title: t('mobile.PrimaryTabStack.sharing') }} />
      <Stack.Screen name="settings/connection" options={{ title: t('mobile.PrimaryTabStack.stuffStashServer') }} />
      <Stack.Screen name="settings/about" options={{ title: t('mobile.PrimaryTabStack.about') }} />
      <Stack.Screen name="settings/diagnostics" options={{ title: t('mobile.PrimaryTabStack.diagnostics') }} />
      <Stack.Screen name="notifications" options={{ title: t('mobile.PrimaryTabStack.notifications') }} />
      <Stack.Screen name="settings/inventory/notification-editor" options={{ title: t('mobile.PrimaryTabStack.reminders') }} />
      <Stack.Screen name="settings/inventory/notifications" options={{ title: t('mobile.PrimaryTabStack.notifications') }} />
      <Stack.Screen name="settings/inventory/index" options={{ title: t('mobile.PrimaryTabStack.inventorySettings') }} />
      <Stack.Screen name="settings/household/index" options={{ title: t('mobile.PrimaryTabStack.householdSettings') }} />
      <Stack.Screen name="settings/inventory/tags/index" options={{ title: t('mobile.PrimaryTabStack.tags') }} />
      <Stack.Screen name="settings/inventory/tags/new" options={{ title: t('mobile.PrimaryTabStack.addTag') }} />
      <Stack.Screen name="settings/inventory/tags/[resourceId]" options={{ title: t('mobile.PrimaryTabStack.tag') }} />
      <Stack.Screen name="settings/inventory/fields/index" options={{ title: t('mobile.PrimaryTabStack.customFields') }} />
      <Stack.Screen name="settings/inventory/fields/new" options={{ title: t('mobile.PrimaryTabStack.addField') }} />
      <Stack.Screen name="settings/inventory/fields/[resourceId]" options={{ title: t('mobile.PrimaryTabStack.customField') }} />
      <Stack.Screen name="settings/inventory/asset-types/index" options={{ title: t('mobile.PrimaryTabStack.assetTypes') }} />
      <Stack.Screen name="settings/inventory/asset-types/new" options={{ title: t('mobile.PrimaryTabStack.addAssetType') }} />
      <Stack.Screen name="settings/inventory/asset-types/[resourceId]" options={{ title: t('mobile.PrimaryTabStack.assetType') }} />
      <Stack.Screen name="settings/household/fields/index" options={{ title: t('mobile.PrimaryTabStack.customFields') }} />
      <Stack.Screen name="settings/household/fields/new" options={{ title: t('mobile.PrimaryTabStack.addField') }} />
      <Stack.Screen name="settings/household/fields/[resourceId]" options={{ title: t('mobile.PrimaryTabStack.customField') }} />
      <Stack.Screen name="settings/household/asset-types/index" options={{ title: t('mobile.PrimaryTabStack.assetTypes') }} />
      <Stack.Screen name="settings/household/asset-types/new" options={{ title: t('mobile.PrimaryTabStack.addAssetType') }} />
      <Stack.Screen name="settings/household/asset-types/[resourceId]" options={{ title: t('mobile.PrimaryTabStack.assetType') }} />
      <Stack.Screen name="settings/voice/index" options={{ title: t('mobile.PrimaryTabStack.voiceSetup') }} />
      <Stack.Screen name="settings/voice/[capability]" options={{ title: t('mobile.PrimaryTabStack.voiceStage') }} />
      <Stack.Screen name="settings/voice/profiles/index" options={{ title: t('mobile.PrimaryTabStack.providerProfiles') }} />
      <Stack.Screen name="settings/voice/profiles/add" options={{ title: t('mobile.PrimaryTabStack.addProfile') }} />
      <Stack.Screen name="settings/voice/profiles/[providerProfileId]/index" options={{ title: t('mobile.PrimaryTabStack.providerProfile') }} />
      <Stack.Screen name="settings/voice/profiles/[providerProfileId]/credential" options={{ title: t('mobile.PrimaryTabStack.credential') }} />
      <Stack.Screen name="settings/voice/profiles/[providerProfileId]/prompt" options={{ title: t('mobile.PrimaryTabStack.promptGuidance') }} />
      <Stack.Screen name="assets/[assetId]/history/index" options={{ title: t('mobile.PrimaryTabStack.history') }} />
      <Stack.Screen name="assets/[assetId]/history/[activityId]" options={{ title: t('mobile.PrimaryTabStack.historyDetail') }} />
    </Stack>
  </VoiceTabContent>;
}
