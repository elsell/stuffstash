import { useEffect, useState } from 'react';
import { AssetCheckoutCommand } from '../src/application/assets/AssetCheckoutCommand';
import { HomeDashboardQuery } from '../src/application/home/HomeDashboardQuery';
import type { HomeDashboardSnapshot } from '../src/application/home/InventorySummaryRepository';
import { assetId, type AssetSummary } from '../src/domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../src/domain/inventories/InventorySummary';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { HomeScreen } from '../src/ui/screens/HomeScreen';
import { QueryReadinessDiagnostics } from './QueryReadinessDiagnostics';

const drill: AssetSummary = {
  id: assetId('audit-drill'), title: 'Audit drill', kind: 'item', lifecycleState: 'active', description: '',
  locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false,
  currentCheckout: { id: 'audit-checkout', state: 'open', checkedOutAt: '2026-09-10T12:00:00Z', checkedOutByPrincipalId: 'audit-user' }
};
export function HomeReturnFixture() {
  const [fixture] = useState(() => {
    let returned = false;
    let rejectDetails = true;
    const client = createMobileQueryClient();
    const query = new HomeDashboardQuery({ getHomeDashboardSnapshot: async (): Promise<HomeDashboardSnapshot> => ({
      checkedOutAssets: returned ? [] : [drill], workspace: {
        tenants: [{ id: tenantId('audit-tenant'), name: 'Audit home' }], defaultInventoryId: inventoryId('audit-inventory'),
        inventories: [{ id: inventoryId('audit-inventory'), tenantId: tenantId('audit-tenant'), name: 'Audit inventory',
          role: 'owner', permissions: ['view', 'create_asset', 'edit_asset'], description: '', updatedAtLabel: '',
          locationCount: 0, locations: [], assets: [drill] }]
      }
    }) });
    const command = new AssetCheckoutCommand({
      returnAsset: async () => { returned = true; return { id: 'audit-checkout', assetId: drill.id, undoableOperationId: 'audit-return' }; },
      updateReturnedCheckoutDetails: async () => {
        if (rejectDetails) { rejectDetails = false; throw new Error('Audit details save failed. Retry to finish.'); }
        return { id: 'audit-checkout', assetId: drill.id };
      },
      undoInventoryOperation: async () => { returned = false; }
    });
    return { client, query, command };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <HomeScreen dashboardQuery={fixture.query} assetCheckoutCommand={fixture.command} />
    <QueryReadinessDiagnostics client={fixture.client} />
  </MobileServerStateProvider>;
}
