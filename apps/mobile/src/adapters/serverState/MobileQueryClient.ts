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
  return status === undefined || status >= 500;
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

export const mobileQueryKeys = {
  root: (compositionScopeId: string) => ['mobile', compositionScopeId] as const,
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
    checkoutState: input.checkoutState ?? 'all',
    sort: input.sort ?? 'default',
    scope: input.scope ?? 'all'
  };
}
