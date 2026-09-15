import { useLayoutEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { ExpirationWorkspaceQuery } from '../../application/expiration/ExpirationWorkspaceQuery';
import type { ExpirationMode } from '../../application/expiration/ExpirationRepository';
import { useMobileServerStateScope } from '../navigation/MobileServerStateProvider';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { expirationRefreshDelay } from '../serverState/expirationRefreshDelay';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { ExpirationHomeSection } from './ExpirationHomeSection';
export function ExpirationHomeContent({ query, onOpen, onOpenAsset }: {
 readonly query: Pick<ExpirationWorkspaceQuery, 'home'>;
 readonly onOpen: (tenantId: string, inventoryId: string, mode: ExpirationMode) => void;
 readonly onOpenAsset: (id: string) => void;
}) {
 const scope = useMobileServerStateScope();
 const inventory = useQuery({ queryKey: mobileQueryKeys.inventoryScope(scope.scopeId), queryFn: ({ signal }) => scope.loadInventoryScope({ signal }), staleTime: Infinity });
 const tenantId = inventory.data?.tenantId ?? '';
 const inventoryId = inventory.data?.inventoryId ?? '';
 const usableScope = inventory.isSuccess && !!tenantId && !!inventoryId;
 const state = useQuery({ queryKey: [...mobileQueryKeys.inventory(scope.scopeId, tenantId, inventoryId), 'expiration', 'home'], enabled: usableScope, queryFn: ({ signal }) => query.home(tenantId, inventoryId, signal), refetchInterval: query => expirationRefreshDelay({ data: query.state.data, expirationContext: { timezone: query.state.data?.timezone } }, new Date(), new Date(query.state.dataUpdatedAt)), refetchIntervalInBackground: false });
 const owner = JSON.stringify([scope.scopeId, tenantId, inventoryId]);
 const current = useRef<{ owner: string; canOpenAsset: boolean } | undefined>(undefined);
 useLayoutEffect(() => {
  current.current = usableScope ? { owner, canOpenAsset: !isAccessFailure(state.error) } : undefined;
  return () => { current.current = undefined; };
 }, [owner, usableScope, state.error]);
 return <ExpirationHomeSection canOpen={usableScope} data={!usableScope || isAccessFailure(state.error) ? undefined : state.data} error={inventory.error || state.error ? 'Expiration could not be refreshed. Try again.' : undefined} onRetry={() => { if (inventory.isError) void inventory.refetch(); else void state.refetch(); }} onOpen={mode => { if (current.current?.owner === owner) onOpen(tenantId, inventoryId, mode); }} onOpenAsset={id => { if (current.current?.owner === owner && current.current.canOpenAsset) onOpenAsset(id); }} />;
}
