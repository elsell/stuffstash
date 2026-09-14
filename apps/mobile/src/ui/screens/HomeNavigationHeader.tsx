import React from 'react';
import { Stack, router } from 'expo-router';
import { Pressable, Text, View, useWindowDimensions } from 'react-native';
import { ChevronDown } from 'lucide-react-native';
import type { HomeDashboardViewModel } from '../../application/home/HomeDashboardQuery';
import { nativeHeaderActionOptions } from '../components/NativeHeaderActions';
import type { NativeHeaderAction } from '../components/NativeHeaderActions.types';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { createHomeScreenStyles } from './HomeScreen.styles';

export function HomeNavigationHeader({ dashboard, notificationAction }: {
  readonly dashboard?: HomeDashboardViewModel;
  readonly notificationAction?: NativeHeaderAction;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createHomeScreenStyles(colors);
  const { width } = useWindowDimensions();
  const actions: NativeHeaderAction[] = [
    ...(notificationAction ? [notificationAction] : []),
    ...(dashboard?.canAdd ? [{ kind: 'add' as const, label: 'Add an asset', onPress: () => router.push('/add') }] : []),
    { kind: 'account', label: 'Open account and settings', onPress: () => router.push('/settings') }
  ];
  return <Stack.Screen options={{
    title: dashboard ? '' : 'Home',
    headerLeft: dashboard ? () => <Pressable
      accessibilityLabel={`Current inventory ${dashboard.inventoryName}, tenant ${dashboard.tenantName}. Switch inventory`}
      accessibilityRole="button" onPress={() => router.push('/tenant-switcher')}
      style={[styles.contextControl, { flex: 0, width: Math.max(100, Math.min(260, width - actions.length * 48 - 48)) }]}
    >
      <View style={styles.contextText}>
        <Text numberOfLines={1} style={styles.contextInventory}>{dashboard.inventoryName}</Text>
        <Text numberOfLines={1} style={styles.contextTenantPrefix}>{dashboard.tenantName}</Text>
      </View>
      <ChevronDown color={colors.textMuted} size={18} strokeWidth={2} />
    </Pressable> : undefined,
    ...nativeHeaderActionOptions(actions)
  }} />;
}
