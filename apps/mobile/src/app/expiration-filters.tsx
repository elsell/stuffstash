import { useQuery } from '@tanstack/react-query';
import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import { useAppServices } from '../ui/navigation/AppServicesContext';
import { ExpirationFiltersScreen } from '../ui/expiration/ExpirationFiltersScreen';
import { parseExpirationRoute, expirationRouteParams, type ExpirationRouteParams } from '../ui/expiration/ExpirationRouteState';
import { useMobileServerStateScope } from '../ui/navigation/MobileServerStateProvider';
import { mobileQueryKeys } from '../adapters/serverState/MobileQueryClient';
import { useSettingsListStyles } from '../ui/screens/SettingsList';
export default function ExpirationFiltersRoute() {
 const services = useAppServices(); const router = useRouter(); const scope = useMobileServerStateScope(); const { palette, styles } = useSettingsListStyles();
 const { tenantId = '', inventoryId = '', filter } = parseExpirationRoute(useLocalSearchParams<ExpirationRouteParams>());
 const state = useQuery({ queryKey: [...mobileQueryKeys.inventory(scope.scopeId, tenantId, inventoryId), 'expiration', 'choices'], queryFn: async ({ signal }) => {
  const selected = await scope.loadInventoryScope({ signal });
  if (selected.tenantId !== tenantId || selected.inventoryId !== inventoryId) throw new Error('Inventory changed. Reopen the expiration view.');
  const [types, tags, locations] = await Promise.all([services.inventoryAssetTypesQuery.execute(tenantId, inventoryId, { signal }), services.inventoryAssetTagsQuery.execute({ signal }), services.locationsQuery.execute({ signal })]);
  const current = await scope.loadInventoryScope({ signal });
  if (current.tenantId !== tenantId || current.inventoryId !== inventoryId) throw new Error('Inventory changed. Reopen the expiration view.');
  return { types: types.map(type => ({ id: type.id, label: type.displayName })), tags, locations: locations.locations.map(location => ({ id: location.id, label: location.title })) };
 } });
 if (state.isPending) return <View style={styles.shell}><ActivityIndicator accessibilityLabel="Loading filters" color={palette.action} /></View>;
 if (state.isError) return <View style={styles.shell}><Text accessibilityRole="alert" style={styles.errorMessage}>Filters could not be loaded.</Text><Pressable accessibilityRole="button" style={styles.retryButton} onPress={() => { void state.refetch(); }}><Text style={styles.retryText}>Retry</Text></Pressable><Pressable accessibilityRole="button" style={styles.retryButton} onPress={() => router.back()}><Text style={styles.retryText}>Cancel</Text></Pressable></View>;
 return <><Stack.Screen options={{ headerShown: true }} /><ExpirationFiltersScreen key={JSON.stringify([scope.scopeId, tenantId, inventoryId])} initial={filter} choices={state.data} onCancel={() => router.back()} onApply={draft => router.dismissTo({ pathname: '/expiration', params: expirationRouteParams(tenantId, inventoryId, draft) })} /></>;
}
