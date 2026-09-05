import { QueryClient } from '@tanstack/react-query';
import { MobileAuthenticationRequiredError } from '../../application/auth/MobileAuthSession';

export const mobileServerStateDefaults = {
  staleTime: 30_000,
  gcTime: 300_000
} as const;

export function createMobileQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: mobileServerStateDefaults.staleTime,
        gcTime: mobileServerStateDefaults.gcTime,
        retry: shouldRetryMobileQuery,
        refetchOnReconnect: true,
        refetchOnWindowFocus: true
      },
      mutations: {
        retry: false
      }
    }
  });
}

export function shouldRetryMobileQuery(failureCount: number, error: unknown): boolean {
  if (failureCount >= 1 || error instanceof MobileAuthenticationRequiredError) {
    return false;
  }
  const status = errorStatus(error);
  if (status !== undefined) {
    return status >= 500;
  }
  return error instanceof TypeError || (
    error instanceof Error && error.message.startsWith('Network request timed out')
  );
}

function errorStatus(error: unknown): number | undefined {
  if (typeof error !== 'object' || error === null || !('status' in error)) {
    return undefined;
  }
  return typeof error.status === 'number' ? error.status : undefined;
}

export async function disposeMobileQueryClient(client: QueryClient): Promise<void> {
  await client.cancelQueries();
  client.clear();
}

export async function resetMobileInventorySelection(
  client: QueryClient,
  compositionScopeId: string
): Promise<void> {
  const queryKey = mobileQueryKeys.root(compositionScopeId);
  await client.cancelQueries({ queryKey });
  client.removeQueries({ queryKey });
}

export const mobileQueryKeys = {
  root: (compositionScopeId: string) => ['mobile', compositionScopeId] as const,
  home: (compositionScopeId: string, tenantId: string, inventoryId: string) => [
    ...mobileQueryKeys.inventory(compositionScopeId, tenantId, inventoryId),
    'home'
  ] as const,
  inventoryScope: (compositionScopeId: string) => [
    ...mobileQueryKeys.root(compositionScopeId),
    'inventory-scope'
  ] as const,
  inventory: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string
  ) => [
    ...mobileQueryKeys.root(compositionScopeId),
    'tenant',
    tenantId,
    'inventory',
    inventoryId
  ] as const,
  asset: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string,
    assetId: string
  ) => [
    ...mobileQueryKeys.inventory(compositionScopeId, tenantId, inventoryId),
    'asset',
    assetId
  ] as const,
  assetCore: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string,
    assetId: string
  ) => [...mobileQueryKeys.asset(compositionScopeId, tenantId, inventoryId, assetId), 'core'] as const,
  assetContents: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string,
    assetId: string,
    containmentIdentity: string
  ) => [
    ...mobileQueryKeys.asset(compositionScopeId, tenantId, inventoryId, assetId),
    'contents',
    containmentIdentity
  ] as const,
  tenant: (scope: string, tenant: string) => [...mobileQueryKeys.root(scope), 'tenant', tenant] as const,
  providerProfiles: (scope: string, tenant: string) => [...mobileQueryKeys.tenant(scope, tenant), 'provider-profiles'] as const,
  voiceConfiguration: (scope: string, tenant: string) => [...mobileQueryKeys.tenant(scope, tenant), 'voice-configuration'] as const,
  principal: (scope: string) => [...mobileQueryKeys.root(scope), 'principal'] as const,
  settingsScope: (scope: string, tenant: string, inventory: string) => [...mobileQueryKeys.inventory(scope, tenant, inventory), 'settings-scope'] as const,
  assetPlacement: (scope: string, tenant: string, inventory: string, asset: string, parent: string) => [
    ...mobileQueryKeys.asset(scope, tenant, inventory, asset), 'placement', parent
  ] as const,
  parentCandidates: (scope: string, tenant: string, inventory: string, query: string) => [
    ...mobileQueryKeys.inventory(scope, tenant, inventory), 'parent-candidates', query.trim()
  ] as const,
  assetHistory: (scopeId: string, tenantId: string, inventoryId: string, assetId: string, view: 'changes' | 'all') => [
    ...mobileQueryKeys.asset(scopeId, tenantId, inventoryId, assetId), 'history', view
  ] as const,
  assetActivity: (scopeId: string, tenantId: string, inventoryId: string, assetId: string, activityId: string) => [
    ...mobileQueryKeys.asset(scopeId, tenantId, inventoryId, assetId), 'activity', activityId
  ] as const,
  assetCheckouts: (scopeId: string, tenantId: string, inventoryId: string, assetId: string) => [
    ...mobileQueryKeys.asset(scopeId, tenantId, inventoryId, assetId), 'checkouts'
  ] as const,
  assetPhotos: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string,
    assetId: string
  ) => [...mobileQueryKeys.asset(compositionScopeId, tenantId, inventoryId, assetId), 'photos'] as const,
  inventoryMap: (scopeId: string, tenantId: string, inventoryId: string) => [
    ...mobileQueryKeys.inventory(scopeId, tenantId, inventoryId), 'map'
  ] as const,
  inventoryContext: (scopeId: string, tenantId: string, inventoryId: string) => [
    ...mobileQueryKeys.inventory(scopeId, tenantId, inventoryId), 'context'
  ] as const,
  browsePages: (scopeId: string, tenantId: string, inventoryId: string, identity: BrowseQueryIdentityInput) => [
    ...mobileQueryKeys.inventory(scopeId, tenantId, inventoryId), 'browse', normalizeBrowseQueryIdentity(identity), 'pages'
  ] as const,
  inventoryAssets: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string
  ) => [...mobileQueryKeys.inventory(compositionScopeId, tenantId, inventoryId), 'assets'] as const,
  assetTags: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string
  ) => [...mobileQueryKeys.inventory(compositionScopeId, tenantId, inventoryId), 'asset-tags'] as const,
  addContext: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string
  ) => [...mobileQueryKeys.inventory(compositionScopeId, tenantId, inventoryId), 'add-context'] as const,
  locations: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string
  ) => [...mobileQueryKeys.inventory(compositionScopeId, tenantId, inventoryId), 'locations'] as const,
  locationAssets: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string,
    locationId: string
  ) => [
    ...mobileQueryKeys.inventory(compositionScopeId, tenantId, inventoryId),
    'location',
    locationId,
    'assets'
  ] as const,
  browse: (
    compositionScopeId: string,
    tenantId: string,
    inventoryId: string,
    identity: BrowseQueryIdentityInput,
    cursor?: string
  ) => [
    ...mobileQueryKeys.inventory(compositionScopeId, tenantId, inventoryId),
    'browse',
    normalizeBrowseQueryIdentity(identity),
    cursor ?? 'first'
  ] as const
};

export type BrowseQueryIdentityInput = {
  readonly query?: string;
  readonly tagIds?: readonly string[];
  readonly lifecycleState?: string;
  readonly checkoutState?: string;
  readonly sort?: string;
  readonly scope?: string;
};

export type BrowseQueryIdentity = {
  readonly query: string;
  readonly tagIds: readonly string[];
  readonly lifecycleState: string;
  readonly checkoutState: string;
  readonly sort: string;
  readonly scope: string;
};

export function normalizeBrowseQueryIdentity(input: BrowseQueryIdentityInput): BrowseQueryIdentity {
  return {
    query: input.query?.trim() ?? '',
    tagIds: [...new Set(input.tagIds ?? [])].sort(),
    lifecycleState: input.lifecycleState ?? 'active',
    checkoutState: input.checkoutState ?? 'any',
    sort: input.sort ?? 'default',
    scope: input.scope ?? 'all'
  };
}
