import { FilterLoadingScreen } from '../ui/components/FilterLoadingScreen';
import { returnToPreviousOrHome } from '../ui/navigation/returnToPreviousOrHome';
import { NativeCommandButton } from '../ui/components/NativeCommandButton';
import { useQuery } from '@tanstack/react-query';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { ScrollView, Text } from 'react-native';
import { useAppServices } from '../ui/navigation/AppServicesContext';
import { ExpirationFiltersScreen } from '../ui/expiration/ExpirationFiltersScreen';
import { parseExpirationRoute, expirationRouteParams, type ExpirationRouteParams } from '../ui/expiration/ExpirationRouteState';
import { useMobileServerStateScope } from '../ui/navigation/MobileServerStateProvider';
import { mobileQueryKeys } from '../adapters/serverState/MobileQueryClient';
import { useSettingsListStyles } from '../ui/screens/SettingsList';
export default function ExpirationFiltersRoute() {
 const services = useAppServices(); const router = useRouter(); const scope = useMobileServerStateScope(); const { styles } = useSettingsListStyles();
 const cancel = () => returnToPreviousOrHome(router);
 const { tenantId = '', inventoryId = '', filter } = parseExpirationRoute(useLocalSearchParams<ExpirationRouteParams>());
 const state = useQuery({ queryKey: [...mobileQueryKeys.inventory(scope.scopeId, tenantId, inventoryId), 'expiration', 'choices'], queryFn: async ({ signal }) => {
  const selected = await scope.loadInventoryScope({ signal });
  if (selected.tenantId !== tenantId || selected.inventoryId !== inventoryId) throw new Error('Inventory changed. Reopen the expiration view.');
  const [types, tags, locations] = await Promise.all([services.inventoryAssetTypesQuery.execute(tenantId, inventoryId, { signal }), services.inventoryAssetTagsQuery.execute({ signal }), services.locationsQuery.execute({ signal })]);
  const current = await scope.loadInventoryScope({ signal });
  if (current.tenantId !== tenantId || current.inventoryId !== inventoryId) throw new Error('Inventory changed. Reopen the expiration view.');
  return { types: types.map(type => ({ id: type.id, label: type.displayName })), tags, locations: locations.locations.map(location => ({ id: location.id, label: location.pathLabel ?? location.title })) };
 } });
 if (state.isPending) return <FilterLoadingScreen onCancel={cancel} />;
 if (state.isError) return <ScrollView style={styles.shell} contentContainerStyle={{ flexGrow: 1 }} contentInsetAdjustmentBehavior="automatic"><Text accessibilityRole="alert" style={styles.errorMessage}>Filters could not be loaded.</Text><NativeCommandButton label="Retry" onPress={() => { void state.refetch(); }} /><NativeCommandButton label="Cancel" onPress={cancel} /></ScrollView>;
 return <ExpirationFiltersScreen key={JSON.stringify([scope.scopeId, tenantId, inventoryId])} initial={filter} choices={state.data} onCancel={cancel} onApply={draft => router.dismissTo({ pathname: '/expiration', params: expirationRouteParams(tenantId, inventoryId, draft) })} />;
}
