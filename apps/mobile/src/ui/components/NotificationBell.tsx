import type { ReactNode } from 'react';
import type { NativeHeaderAction } from './NativeHeaderActions.types';
import { useQuery } from '@tanstack/react-query';
import { Bell } from 'lucide-react-native';
import { Pressable, StyleSheet, Text } from 'react-native';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScopeId } from '../navigation/MobileServerStateProvider';
import { useAppearancePalette } from '../theme/AppearanceContext';

export function NotificationBell({ tenantId, inventoryId, initialize, count, onOpen, renderAction }: {
  readonly tenantId: string; readonly inventoryId: string;
  readonly initialize: (signal: AbortSignal) => Promise<void>;
  readonly count: (signal: AbortSignal) => Promise<number>;
  readonly onOpen: () => void;
  readonly renderAction?: (action: NativeHeaderAction) => ReactNode;
}) {
  const scopeId = useMobileServerStateScopeId();
  const colors = useAppearancePalette();
  const registration = useQuery({ queryKey: mobileQueryKeys.notificationRegistration(scopeId, tenantId, inventoryId), queryFn: async ({ signal }) => { await initialize(signal); return true; }, staleTime: Infinity });
  const unread = useQuery({ queryKey: mobileQueryKeys.notificationCount(scopeId, tenantId, inventoryId), queryFn: ({ signal }) => count(signal), enabled: registration.isSuccess, refetchInterval: 30_000, refetchIntervalInBackground: false });
  const failed = registration.isError || unread.isError;
  const total = unread.data;
  const label = failed ? 'Notifications, unread count unavailable' : total === undefined ? 'Notifications, loading unread count' : `Notifications, ${total} unread`;
  const onPress = () => { if (registration.isError) void registration.refetch(); if (registration.isSuccess) void unread.refetch(); onOpen(); };
  if (renderAction) return renderAction({ kind: 'notifications', label, badgeCount: total, onPress });
  return <Pressable accessibilityRole="button" accessibilityLabel={label} onPress={onPress} style={styles.button}>
    <Bell color={colors.text} size={24} strokeWidth={2} />
    {total !== undefined && total > 0 ? <Text style={[styles.badge, { backgroundColor: colors.action, color: colors.background }]}>{total > 99 ? '99+' : total}</Text> : null}
  </Pressable>;
}
const styles = StyleSheet.create({ button: { minWidth: 44, minHeight: 44, alignItems: 'center', justifyContent: 'center' }, badge: { position: 'absolute', top: 0, right: 0, minWidth: 18, borderRadius: 9, paddingHorizontal: 3, textAlign: 'center', fontSize: 11, fontWeight: '700' } });
