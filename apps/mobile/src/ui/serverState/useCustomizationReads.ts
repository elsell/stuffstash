import { isAccessFailure } from './isAccessFailure';
import type { CustomizationAccessPolicy } from '../../application/customization/CustomizationAccess';
import { useQuery, useQueryClient, queryOptions } from '@tanstack/react-query';
import type { CustomizationContextQuery } from '../../application/customization/CustomizationContextQuery';
import type { CustomizationCollection, CustomizationCollectionQuery } from '../../application/customization/CustomizationQueries';
import type { CustomizationContext } from '../../application/customization/CustomizationRepository';
import type { AssetTagDefinition, CustomDefinition, CustomizationKind, CustomizationScope, CustomizationLifecycle } from '../../domain/customization/Customization';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScopeId } from '../navigation/MobileServerStateProvider';
import { useMobileInventoryServerQuery } from './useMobileInventoryServerQuery';

export function useCustomizationReads(contextQuery: Pick<CustomizationContextQuery, 'execute'>, query: Pick<CustomizationCollectionQuery, 'tags' | 'fields' | 'assetTypes'>, kind: CustomizationKind, scope: CustomizationScope, lifecycle: CustomizationLifecycle, accessPolicy: CustomizationAccessPolicy, readRecord = true, includeEligibleTypes = false) {
  const scopeId = useMobileServerStateScopeId();
  const client = useQueryClient();
  const selected = useMobileInventoryServerQuery({ key: mobileQueryKeys.settingsScope, query: async (signal) => {
    const context = await contextQuery.execute({ signal });
    return { tenant: { id: context.tenantId, name: context.tenantName, permissions: context.tenantPermissions }, inventory: { id: context.inventoryId, name: context.inventoryName, permissions: context.inventoryPermissions } };
  } });
  const context = selected.data && !isAccessFailure(selected.error) ? fromSettingsScope(selected.data) : undefined;
  const options = (context: CustomizationContext, kind: CustomizationKind, scope: CustomizationScope, lifecycle: CustomizationLifecycle) => queryOptions({
    queryKey: mobileQueryKeys.customization(scopeId, context.tenantId, context.inventoryId, scope, kind, lifecycle),
    queryFn: async ({ signal }): Promise<CustomizationCollection<AssetTagDefinition | CustomDefinition>> => kind === 'tag' ? query.tags(context, { signal }) : kind === 'field' ? query.fields(context, scope, lifecycle, { signal }) : query.assetTypes(context, scope, lifecycle, { signal })
  });
  const pendingContext: CustomizationContext = { tenantId: 'pending', inventoryId: 'pending', tenantName: '', inventoryName: '', tenantPermissions: [], inventoryPermissions: [] };
  const canRead = Boolean(context && accessPolicy.canRead(context, scope));
  const resource = useQuery({ ...options(context ?? pendingContext, kind, scope, lifecycle), enabled: canRead && readRecord, subscribed: canRead && readRecord });
  const types = useQuery({ ...options(context ?? pendingContext, 'asset-type', scope, 'active'), enabled: canRead && includeEligibleTypes && kind === 'field' && Boolean(readRecord || context && accessPolicy.canMutate(context, kind, scope)), subscribed: canRead && includeEligibleTypes && kind === 'field' && Boolean(readRecord || context && accessPolicy.canMutate(context, kind, scope)) });
  return {
    context, resource, types, ownerKey: JSON.stringify(selected.resourceKey), contextError: selected.error,
    contextQuery: { execute: async () => {
      const context = fromSettingsScope(await selected.reconcile());
      const failure = client.getQueryState(mobileQueryKeys.settingsScope(scopeId, context.tenantId, context.inventoryId))?.error;
      if (isAccessFailure(failure)) throw failure;
      return context;
    } },
    refreshContext: async () => {
      const result = await selected.refetch({ throwOnError: true });
      return fromSettingsScope(result.data!);
    },
    query: {
      tags: (context: CustomizationContext) => client.fetchQuery(options(context, 'tag', 'inventory', 'active')) as ReturnType<CustomizationCollectionQuery['tags']>,
      fields: (context: CustomizationContext, scope: CustomizationScope, lifecycle: CustomizationLifecycle) => client.fetchQuery(options(context, 'field', scope, lifecycle)) as ReturnType<CustomizationCollectionQuery['fields']>,
      assetTypes: (context: CustomizationContext, scope: CustomizationScope, lifecycle: CustomizationLifecycle) => client.fetchQuery(options(context, 'asset-type', scope, lifecycle)) as ReturnType<CustomizationCollectionQuery['assetTypes']>
    },
    refreshDefinitions: async () => { await Promise.all([...(readRecord ? [resource.refetch()] : []), ...(includeEligibleTypes && kind === 'field' ? [types.refetch()] : [])]); },
    invalidateResource: async () => { await client.invalidateQueries({ queryKey: options(context ?? pendingContext, kind, scope, lifecycle).queryKey, refetchType: 'none' }); }
  };
}

function fromSettingsScope(scope: { tenant: { id: string; name: string; permissions: readonly string[] }; inventory: { id: string; name: string; permissions: readonly string[] } }): CustomizationContext {
  return { tenantId: scope.tenant.id, tenantName: scope.tenant.name, tenantPermissions: scope.tenant.permissions, inventoryId: scope.inventory.id, inventoryName: scope.inventory.name, inventoryPermissions: scope.inventory.permissions };
}
