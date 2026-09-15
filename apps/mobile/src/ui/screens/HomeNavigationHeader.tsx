import React, { useMemo } from 'react';
import { Stack, router } from 'expo-router';
import { Pressable, Text, View, useWindowDimensions } from 'react-native';
import { ChevronDown } from 'lucide-react-native';
import type { HomeDashboardViewModel } from '../../application/home/HomeDashboardQuery';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import type { NativeHeaderAction } from '../components/NativeHeaderActions.types';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { homeInventoryControlWidth } from './HomeHeaderLayout';
import { createHomeScreenStyles } from './HomeScreen.styles';

export function HomeNavigationHeader({ dashboard, notificationAction }: {
  readonly dashboard?: Pick<HomeDashboardViewModel, 'inventoryName' | 'tenantName' | 'canAdd'>;
  readonly notificationAction?: NativeHeaderAction;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = useMemo(() => createHomeScreenStyles(colors), [colors]);
  const { width, fontScale } = useWindowDimensions();
  const actions: NativeHeaderAction[] = [
    ...(dashboard?.canAdd ? [{ kind: 'add' as const, label: 'Add an asset', onPress: () => router.push('/add') }] : []),
    ...(notificationAction ? [notificationAction] : []),
    { kind: 'account', label: 'Open account and settings', onPress: () => router.push('/settings') }
  ];
  const actionOptions = useNativeHeaderActionOptions(actions);
  const inventoryName = dashboard?.inventoryName;
  const tenantName = dashboard?.tenantName;
  const hasDashboard = dashboard !== undefined;
  const headerLeft = useMemo(() => hasDashboard ? () => <Pressable
      accessibilityLabel={`Current inventory ${inventoryName}, tenant ${tenantName}. Switch inventory`}
      accessibilityRole="button" onPress={() => router.push('/tenant-switcher')}
      style={[styles.contextControl, { flex: 0, width: homeInventoryControlWidth(width, actions.length) }]}
    >
      <View style={styles.contextText}>
        <Text numberOfLines={1} style={styles.contextInventory}>{inventoryName}</Text>
        {fontScale <= 1.05 ? <Text numberOfLines={1} style={styles.contextTenantPrefix}>{tenantName}</Text> : null}
      </View>
      <ChevronDown color={colors.textMuted} size={18} strokeWidth={2} />
    </Pressable> : undefined, [hasDashboard, inventoryName, tenantName, styles, colors.textMuted, width, fontScale, actions.length]);
  const options = useMemo(() => ({ title: hasDashboard ? '' : 'Home', headerLeft, ...actionOptions }), [hasDashboard, headerLeft, actionOptions]);
  return <Stack.Screen options={options} />;
}
