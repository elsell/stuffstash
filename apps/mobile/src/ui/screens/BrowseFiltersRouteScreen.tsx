import { FilterLoadingScreen } from '../components/FilterLoadingScreen';
import { returnToPreviousOrHome } from '../navigation/returnToPreviousOrHome';
import { useQuery } from '@tanstack/react-query';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { ScrollView, Text } from 'react-native';
import type { InventoryAssetTagsQuery } from '../../application/assets/InventoryAssetTagsQuery';
import { useMobileServerStateScope } from '../navigation/MobileServerStateProvider';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useBrowseFilterNavigation } from '../screens/useBrowseFilterNavigation';
import { BrowseFiltersScreen } from '../screens/BrowseFiltersScreen';
import { parseBrowseRouteParams } from '../screens/BrowseRouteParams';
import { browseFilterApplyParams, loadBrowseFilterTags, verifyBrowseFilterScope } from '../screens/BrowseFilterRouteState';
import { useSettingsListStyles, SettingsActionRow } from '../screens/SettingsList';
import { expirationRouteParams } from '../expiration/ExpirationRouteState';
import { browseExpirationFilter } from '../expiration/BrowseExpirationFilter';

export function BrowseFiltersRouteScreen({ inventoryAssetTagsQuery }: { readonly inventoryAssetTagsQuery: Pick<InventoryAssetTagsQuery, 'execute'> }) {
  const params = useLocalSearchParams();
  const scalar = (value: string | string[] | undefined) => typeof value === 'string' ? value : value?.[0] ?? '';
  const target = { tenantId: scalar(params.tenantId), inventoryId: scalar(params.inventoryId), sessionScope: scalar(params.sessionScope) };
  const initial = parseBrowseRouteParams(params);
  const scope = useMobileServerStateScope();
  const router = useRouter();
  const { styles } = useSettingsListStyles();
  const { busy, error, navigate, cancel } = useBrowseFilterNavigation(
    JSON.stringify([scope.scopeId, target.tenantId, target.inventoryId, target.sessionScope]),
    async signal => { verifyBrowseFilterScope(target, scope.scopeId, await scope.loadInventoryScope({ signal })); }
  );
  const dismiss = () => { cancel(); returnToPreviousOrHome(router); };
  const identity = useQuery({ queryKey: mobileQueryKeys.inventoryScope(scope.scopeId), queryFn: ({ signal }) => scope.loadInventoryScope({ signal }), staleTime: Infinity });
  const choices = useQuery({
    queryKey: [...mobileQueryKeys.inventory(scope.scopeId, target.tenantId, target.inventoryId), 'browse-filter-tags', target.sessionScope],
    queryFn: ({ signal }) => loadBrowseFilterTags(target, scope.scopeId, () => scope.loadInventoryScope({ signal }), () => inventoryAssetTagsQuery.execute({ signal }))
  });
  const matches = identity.data && target.sessionScope === scope.scopeId && target.tenantId === identity.data.tenantId && target.inventoryId === identity.data.inventoryId;
  if (choices.isError || identity.isError || (identity.data && !matches)) return <ScrollView style={styles.shell} contentContainerStyle={{ flexGrow: 1 }} contentInsetAdjustmentBehavior="automatic">
    <Text accessibilityRole="alert" style={styles.errorMessage}>Filters are unavailable for this inventory.</Text>
    <SettingsActionRow label="Retry" onPress={() => { void identity.refetch(); void choices.refetch(); }} />
    <SettingsActionRow label="Cancel" onPress={dismiss} />
  </ScrollView>;
  if (!choices.data || !matches) return <FilterLoadingScreen onCancel={dismiss} />;
  return <BrowseFiltersScreen key={JSON.stringify([scope.scopeId, target.tenantId, target.inventoryId])}
      initial={{ scope: initial.initialScope, lifecycleState: initial.initialLifecycleState, checkoutState: initial.initialCheckoutState, tagIds: initial.initialTagIds, sort: initial.initialSort }}
      query={initial.initialQuery} tags={choices.data} busy={busy} error={error} onCancel={dismiss} onCancelPending={cancel}
      onApply={draft => { void navigate(() => router.dismissTo({ pathname: '/search', params: browseFilterApplyParams(draft, initial.initialQuery) })); }}
      onExpiration={(mode, draft) => { void navigate(() => router.replace({ pathname: '/expiration', params: expirationRouteParams(target.tenantId, target.inventoryId, browseExpirationFilter(mode, { ...draft, query: initial.initialQuery })) })); }}
    />;
}
