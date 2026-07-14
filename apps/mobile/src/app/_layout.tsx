import { StatusBar } from 'expo-status-bar';
import { Stack } from 'expo-router';
import { View } from 'react-native';
import { getAppearancePreferenceController } from '../bootstrap/mobileComposition';
import { AppServicesProvider } from '../ui/navigation/AppServicesContext';
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
        <Stack.Screen name="settings" options={{ title: 'Settings' }} />
        <Stack.Screen name="provider-profiles" options={{ title: 'Voice providers' }} />
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
        <Stack.Screen
          name="assets/[assetId]/audit"
          options={sheetOptions.audit}
        />
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
  );
}
