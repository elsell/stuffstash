import { QueryReadinessDiagnostics } from './QueryReadinessDiagnostics';
import { useEffect, useState } from 'react';
import { AssetCheckoutHistorySheetRouteScreen } from '../src/ui/screens/AssetCheckoutHistoryScreen';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AssetCoreQuery } from '../src/application/assets/AssetCoreQuery';
import { AssetCheckoutHistoryQuery } from '../src/application/assets/AssetCheckoutHistoryQuery';
import { assetId } from '../src/domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../src/domain/inventories/InventorySummary';

export function CheckoutHistoryFixture() {
  const [fixture] = useState(() => ({
    client: createMobileQueryClient(),
    core: new AssetCoreQuery({ getAssetCore: async () => ({
      tenantId: tenantId('audit-tenant'), inventoryId: inventoryId('audit-inventory'), permissions: ['view'], revision: 'audit',
      asset: { id: assetId('audit-ladder'), title: 'Audit ladder', kind: 'item', lifecycleState: 'active', description: '',
        locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
    }) }),
    history: new AssetCheckoutHistoryQuery({ listAssetCheckoutHistory: async ({ cursor }) => ({
      records: Array.from({ length: cursor ? 1 : 3 }, (_, index) => ({
        id: cursor ?? `checkout-${index}`, state: 'returned', checkedOutAt: '2026-09-10T12:00:00Z', checkedOutByPrincipalId: 'audit-user',
        returnedAt: '2026-09-11T12:00:00Z', returnedByPrincipalId: 'audit-user',
        checkoutDetails: cursor ? 'Older audit checkout' : `Audit checkout ${index + 1}: borrowed for cleaning the gutters.`,
        returnDetails: 'Returned to the garage shelf.'
      })), hasMore: !cursor, nextCursor: cursor ? undefined : 'older'
    }) })
  }));
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <AssetCheckoutHistorySheetRouteScreen assetId="audit-ladder" assetCoreQuery={fixture.core} assetCheckoutHistoryQuery={fixture.history} />
    <QueryReadinessDiagnostics client={fixture.client} />
  </MobileServerStateProvider>;
}
