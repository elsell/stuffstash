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

export default function RootLayout() {
  return (
    <AppearanceProvider controller={getAppearancePreferenceController()}>
      <ThemedApp />
    </AppearanceProvider>
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
    <AppServicesProvider>
      <StatusBar style={resolvedColorScheme === 'dark' ? 'light' : 'dark'} />
      <Stack
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
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen
          name="voice"
          options={{
            contentStyle: { backgroundColor: palette.surface },
            headerShown: false,
            presentation: 'formSheet',
            sheetAllowedDetents: [0.42, 0.88],
            sheetCornerRadius: 24,
            sheetExpandsWhenScrolledToEdge: true,
            sheetGrabberVisible: true,
            sheetInitialDetentIndex: 0,
            sheetLargestUndimmedDetentIndex: 'none'
          }}
        />
        <Stack.Screen name="settings/index" options={{ title: 'Settings' }} />
        <Stack.Screen name="settings/account" options={{ title: 'Account' }} />
        <Stack.Screen name="settings/appearance" options={{ title: 'Appearance' }} />
        <Stack.Screen name="settings/sharing" options={{ title: 'Sharing' }} />
        <Stack.Screen name="settings/connection" options={{ title: 'Stuff Stash Server' }} />
        <Stack.Screen name="settings/about" options={{ title: 'About' }} />
        <Stack.Screen name="settings/diagnostics" options={{ title: 'Diagnostics' }} />
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
        <Stack.Screen name="invitations/accept" options={{ title: 'Invitation' }} />
        <Stack.Screen name="settings/voice/index" options={{ title: 'Voice Setup' }} />
        <Stack.Screen name="settings/voice/[capability]" options={{ title: 'Voice Stage' }} />
        <Stack.Screen name="settings/voice/profiles/index" options={{ title: 'Provider Profiles' }} />
        <Stack.Screen name="settings/voice/profiles/add" options={{ title: 'Add Profile' }} />
        <Stack.Screen name="settings/voice/profiles/[providerProfileId]/index" options={{ title: 'Provider Profile' }} />
        <Stack.Screen name="settings/voice/profiles/[providerProfileId]/credential" options={{ title: 'Credential' }} />
        <Stack.Screen name="settings/voice/profiles/[providerProfileId]/prompt" options={{ title: 'Prompt Guidance' }} />
        <Stack.Screen
          name="add"
          options={{
            contentStyle: { backgroundColor: palette.background },
            headerShown: false,
            presentation: 'formSheet',
            sheetAllowedDetents: [0.92],
            sheetCornerRadius: 24,
            sheetGrabberVisible: true
          }}
        />
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
        <Stack.Screen name="assets/[assetId]/history/index" options={{ title: 'History' }} />
        <Stack.Screen name="assets/[assetId]/history/[activityId]" options={{ title: 'History detail' }} />
        <Stack.Screen
          name="assets/[assetId]/checkouts"
          options={sheetOptions.checkoutHistory}
        />
        <Stack.Screen
          name="tenant-switcher"
          options={{
            contentStyle: { backgroundColor: palette.surface },
            headerShown: false,
            presentation: 'formSheet',
            sheetAllowedDetents: 'fitToContents',
            sheetCornerRadius: 24,
            sheetGrabberVisible: true
          }}
        />
      </Stack>
    </AppServicesProvider>
    </InventoryInvitationLinkProvider>
  );
}
