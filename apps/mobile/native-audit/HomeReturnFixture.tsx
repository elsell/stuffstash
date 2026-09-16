import { useEffect, useState } from 'react';
import { Text } from 'react-native';
import { AssetCheckoutCommand } from '../src/application/assets/AssetCheckoutCommand';
import { HomeDashboardQuery } from '../src/application/home/HomeDashboardQuery';
import type { HomeDashboardSnapshot } from '../src/application/home/InventorySummaryRepository';
import { assetId, type AssetSummary } from '../src/domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../src/domain/inventories/InventorySummary';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { HomeScreen } from '../src/ui/screens/HomeScreen';
import { QueryReadinessDiagnostics } from './QueryReadinessDiagnostics';
import { toAssetCardViewModel } from '../src/application/assets/AssetViewModels';
import { ExpirationHomeSection } from '../src/ui/expiration/ExpirationHomeSection';

const drill: AssetSummary = {
  id: assetId('audit-drill'), title: 'Audit drill', kind: 'item', lifecycleState: 'active', description: '',
  locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false,
  currentCheckout: { id: 'audit-checkout', state: 'open', checkedOutAt: '2026-09-10T12:00:00Z', checkedOutByPrincipalId: 'audit-user' }
};
const headerAssets: readonly AssetSummary[] = [drill, ...['Audit camping equipment', 'Audit garden tools'].map((title, index) => ({
  ...drill, id: assetId(`audit-header-${index}`), title,
  currentCheckout: { ...drill.currentCheckout!, id: `audit-header-checkout-${index}` }
}))];

export function HomeHeaderFixture() { return <HomeReturnFixture headerAudit />; }

// Observe the production header's actual router destinations without replacing
// its handlers or claiming acceptance of the destination workflows.
export function HomeAddProbeDestination() { return <Text>Header Add destination</Text>; }
export function HomeProfileProbeDestination() { return <Text>Header Profile destination</Text>; }

export function HomeReturnFixture({ headerAudit = false }: { readonly headerAudit?: boolean }) {
  const [notificationActivations, setNotificationActivations] = useState(0);
  const [fixture] = useState(() => {
    let returned = false;
    let rejectDetails = true;
    const client = createMobileQueryClient();
    const query = new HomeDashboardQuery({ getHomeDashboardSnapshot: async (): Promise<HomeDashboardSnapshot> => ({
      checkedOutAssets: returned ? [] : headerAudit ? headerAssets : [drill], workspace: {
        tenants: [{ id: tenantId('audit-tenant'), name: 'Audit home' }], defaultInventoryId: inventoryId('audit-inventory'),
        inventories: [{ id: inventoryId('audit-inventory'), tenantId: tenantId('audit-tenant'), name: headerAudit ? 'Main inventory with a long household name' : 'Audit inventory',
          role: 'owner', permissions: ['view', 'create_asset', 'edit_asset'], description: '', updatedAtLabel: '',
          locationCount: 0, locations: [], assets: headerAudit ? headerAssets : [drill] }]
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
    <HomeScreen dashboardQuery={fixture.query} assetCheckoutCommand={fixture.command}
      notificationAction={headerAudit ? { kind: 'notifications', label: 'Notifications, 2 unread', badgeCount: 2, onPress: () => setNotificationActivations(count => count + 1) } : undefined}
      expirationSection={headerAudit ? <ExpirationHomeSection data={{ items: headerAssets.map(toAssetCardViewModel), counts: { expired: 3, soon: 0, all: 3 }, timezone: 'UTC' }} onOpen={() => undefined} onOpenAsset={() => undefined} onRetry={() => undefined} /> : undefined}
    />
    <QueryReadinessDiagnostics client={fixture.client} />
    {headerAudit ? <Text pointerEvents="none" style={{ position: 'absolute', left: 8, bottom: 24, fontSize: 10 }}>
      {`Header notification activations: ${notificationActivations}`}
    </Text> : null}
  </MobileServerStateProvider>;
}
