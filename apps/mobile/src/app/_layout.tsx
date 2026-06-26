import { StatusBar } from 'expo-status-bar';
import { Stack } from 'expo-router';
import { AppServicesProvider } from '../ui/navigation/AppServicesContext';
import { colors } from '../ui/theme/tokens';

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
        <Stack.Screen name="settings" options={{ title: 'Settings' }} />
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
