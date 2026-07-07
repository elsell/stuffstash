import { StatusBar } from 'expo-status-bar';
import { Stack } from 'expo-router';
import { AppServicesProvider } from '../ui/navigation/AppServicesContext';
import { colors } from '../ui/theme/tokens';
import {
  assetAuditNativeSheetOptions,
  assetCheckoutHistoryNativeSheetOptions,
  assetEditNativeSheetOptions,
  assetMoveHereNativeSheetOptions,
  assetMoveNativeSheetOptions
} from '../ui/screens/AssetNativeSheetOptions';

export default function RootLayout() {
  return (
    <AppServicesProvider>
      <StatusBar style="dark" />
      <Stack
        screenOptions={{
          contentStyle: { backgroundColor: colors.background },
          headerBackTitle: 'Back',
          headerTintColor: colors.action,
          headerTitleStyle: {
            color: colors.text,
            fontWeight: '700'
          }
        }}
      >
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen
          name="voice"
          options={{
            contentStyle: { backgroundColor: colors.surface },
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
          options={assetEditNativeSheetOptions}
        />
        <Stack.Screen
          name="assets/[assetId]/move"
          options={assetMoveNativeSheetOptions}
        />
        <Stack.Screen
          name="assets/[assetId]/move-here"
          options={assetMoveHereNativeSheetOptions}
        />
        <Stack.Screen
          name="assets/[assetId]/audit"
          options={assetAuditNativeSheetOptions}
        />
        <Stack.Screen
          name="assets/[assetId]/checkouts"
          options={assetCheckoutHistoryNativeSheetOptions}
        />
        <Stack.Screen
          name="tenant-switcher"
          options={{
            contentStyle: { backgroundColor: colors.surface },
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
