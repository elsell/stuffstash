import { AddDestinationTaskProvider } from '../ui/navigation/AddDestinationTask';
import { AssetTagSelectionTaskProvider } from '../ui/navigation/AssetTagSelectionTask';
import { inventorySwitcherNativeOptions } from '../ui/screens/InventorySwitcherNativeOptions';
import { AppNoticeScreenLayout } from '../ui/feedback/AppNoticeScreenLayout';
import { voiceNativeSheetOptions } from '../ui/screens/VoiceNativeSheetOptions';
import { HomeReturnTaskProvider } from '../ui/navigation/HomeReturnTaskPresentation';
import { PushNotificationNavigation } from '../ui/navigation/PushNotificationNavigation';
import { VoiceConversationReturn } from '../ui/navigation/VoiceConversationReturn';
import { StatusBar } from 'expo-status-bar';
import { Stack } from 'expo-router';
import { View } from 'react-native';
import { getAppearancePreferenceController } from '../bootstrap/mobileComposition';
import { AppServicesProvider } from '../ui/navigation/AppServicesContext';
import { InventoryInvitationLinkProvider } from '../ui/navigation/InventoryInvitationLinkContext';
import { AppearanceProvider, useAppearance } from '../ui/theme/AppearanceContext';
import {
  createAssetNativeSheetOptions
} from '../ui/screens/AssetNativeSheetOptions';
import { AppKeyboardAccessory } from '../ui/components/AppKeyboardAccessory';
import { AppKeyboardProvider } from '../ui/components/AppKeyboardProvider';

export const unstable_settings = { anchor: '(tabs)' };

export default function RootLayout() {
  return (
    <AppKeyboardProvider>
      <AppearanceProvider controller={getAppearancePreferenceController()}>
        <ThemedApp />
      </AppearanceProvider>
    </AppKeyboardProvider>
  );
}

function ThemedApp() {
  const { isHydrated, palette, resolvedColorScheme } = useAppearance();
  const sheetOptions = createAssetNativeSheetOptions(palette);

  if (!isHydrated) {
    return (
      <View style={{ backgroundColor: palette.background, flex: 1 }}>
        <StatusBar style={resolvedColorScheme === 'dark' ? 'light' : 'dark'} />
      </View>
    );
  }

  return (
    <InventoryInvitationLinkProvider>
    <AppServicesProvider><HomeReturnTaskProvider><AssetTagSelectionTaskProvider><AddDestinationTaskProvider>
      <StatusBar style={resolvedColorScheme === 'dark' ? 'light' : 'dark'} />
      <Stack
        screenLayout={AppNoticeScreenLayout}
        screenOptions={{
          contentStyle: { backgroundColor: palette.background },
          headerBackTitle: 'Back',
          headerStyle: { backgroundColor: palette.surface },
          headerTintColor: palette.action,
          headerTitleStyle: {
            color: palette.text,
            fontWeight: '700'
          }
        }}
      >
        <Stack.Screen name="voice-plan-location" options={{ title: 'Containing location' }} />
        <Stack.Screen name="browse-filters" options={sheetOptions.filters} />
        <Stack.Screen name="expiration-filters" options={sheetOptions.filters} />
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen
          name="voice"
          options={voiceNativeSheetOptions(palette)}
        />
        <Stack.Screen name="invitations/accept" options={{ title: 'Invitation' }} />
        <Stack.Screen name="add" options={sheetOptions.add} />
        <Stack.Screen name="provider-profiles" options={{ headerShown: false }} />
        <Stack.Screen
          name="assets/[assetId]/edit"
          options={sheetOptions.edit}
        />
        <Stack.Screen
          name="assets/[assetId]/move"
          options={sheetOptions.move}
        />
        <Stack.Screen
          name="assets/[assetId]/move-here"
          options={sheetOptions.moveHere}
        />
        <Stack.Screen name="add-destination" options={{ ...sheetOptions.selection, title: 'Put in' }} />
        <Stack.Screen name="asset-tag-selection" options={{ ...sheetOptions.selection, title: 'Tags' }} />
        <Stack.Screen name="home-return-details" options={{ ...sheetOptions.checkoutHistory, title: 'Return details', gestureEnabled: false }} />
        <Stack.Screen
          name="assets/[assetId]/checkouts"
          options={sheetOptions.checkoutHistory}
        />
        <Stack.Screen name="tenant-switcher" options={inventorySwitcherNativeOptions(palette)} />
      </Stack>
      <PushNotificationNavigation />
      <VoiceConversationReturn />
      <AppKeyboardAccessory />
    </AddDestinationTaskProvider></AssetTagSelectionTaskProvider></HomeReturnTaskProvider></AppServicesProvider>
    </InventoryInvitationLinkProvider>
  );
}
