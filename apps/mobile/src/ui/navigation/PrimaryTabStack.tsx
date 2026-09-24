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
      <Stack.Screen name="expiration" options={{ title: 'Expiration' }} />
      <Stack.Screen name="settings/index" options={{ title: 'Settings' }} />
      <Stack.Screen name="settings/account" options={{ title: 'Account' }} />
      <Stack.Screen name="settings/appearance" options={{ title: 'Appearance' }} />
      <Stack.Screen name="settings/sharing" options={{ title: 'Sharing' }} />
      <Stack.Screen name="settings/connection" options={{ title: 'Stuff Stash Server' }} />
      <Stack.Screen name="settings/about" options={{ title: 'About' }} />
      <Stack.Screen name="settings/diagnostics" options={{ title: 'Diagnostics' }} />
      <Stack.Screen name="notifications" options={{ title: 'Notifications' }} />
      <Stack.Screen name="settings/inventory/notification-editor" options={{ title: 'Reminders' }} />
      <Stack.Screen name="settings/inventory/notifications" options={{ title: 'Notifications' }} />
      <Stack.Screen name="settings/inventory/index" options={{ title: 'Inventory Settings' }} />
      <Stack.Screen name="settings/household/index" options={{ title: 'Household Settings' }} />
      <Stack.Screen name="settings/inventory/tags/index" options={{ title: 'Tags' }} />
      <Stack.Screen name="settings/inventory/tags/new" options={{ title: 'Add Tag' }} />
      <Stack.Screen name="settings/inventory/tags/[resourceId]" options={{ title: 'Tag' }} />
      <Stack.Screen name="settings/inventory/fields/index" options={{ title: 'Custom Fields' }} />
      <Stack.Screen name="settings/inventory/fields/new" options={{ title: 'Add Field' }} />
      <Stack.Screen name="settings/inventory/fields/[resourceId]" options={{ title: 'Custom Field' }} />
      <Stack.Screen name="settings/inventory/asset-types/index" options={{ title: 'Asset Types' }} />
      <Stack.Screen name="settings/inventory/asset-types/new" options={{ title: 'Add Asset Type' }} />
      <Stack.Screen name="settings/inventory/asset-types/[resourceId]" options={{ title: 'Asset Type' }} />
      <Stack.Screen name="settings/household/fields/index" options={{ title: 'Custom Fields' }} />
      <Stack.Screen name="settings/household/fields/new" options={{ title: 'Add Field' }} />
      <Stack.Screen name="settings/household/fields/[resourceId]" options={{ title: 'Custom Field' }} />
      <Stack.Screen name="settings/household/asset-types/index" options={{ title: 'Asset Types' }} />
      <Stack.Screen name="settings/household/asset-types/new" options={{ title: 'Add Asset Type' }} />
      <Stack.Screen name="settings/household/asset-types/[resourceId]" options={{ title: 'Asset Type' }} />
      <Stack.Screen name="settings/voice/index" options={{ title: 'Voice Setup' }} />
      <Stack.Screen name="settings/voice/[capability]" options={{ title: 'Voice Stage' }} />
      <Stack.Screen name="settings/voice/profiles/index" options={{ title: 'Provider Profiles' }} />
      <Stack.Screen name="settings/voice/profiles/add" options={{ title: 'Add Profile' }} />
      <Stack.Screen name="settings/voice/profiles/[providerProfileId]/index" options={{ title: 'Provider Profile' }} />
      <Stack.Screen name="settings/voice/profiles/[providerProfileId]/credential" options={{ title: 'Credential' }} />
      <Stack.Screen name="settings/voice/profiles/[providerProfileId]/prompt" options={{ title: 'Prompt Guidance' }} />
      <Stack.Screen name="assets/[assetId]/history/index" options={{ title: 'History' }} />
      <Stack.Screen name="assets/[assetId]/history/[activityId]" options={{ title: 'History detail' }} />
    </Stack>
  </VoiceTabContent>;
}
