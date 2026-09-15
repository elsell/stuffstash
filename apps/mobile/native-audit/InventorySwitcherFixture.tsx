import { useEffect, useState } from 'react';
import { createMobileQueryClient, mobileQueryKeys } from '../src/adapters/serverState/MobileQueryClient';
import { HomeDashboardQuery, type HomeDashboardViewModel } from '../src/application/home/HomeDashboardQuery';
import { SelectInventoryCommand } from '../src/application/home/SelectInventoryCommand';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { TenantSwitcherSheetScreen } from '../src/ui/screens/TenantSwitcherSheetScreen';

const scope = { tenantId: 'audit-home', inventoryId: 'audit-main' };
const dashboard: HomeDashboardViewModel = {
  ...scope, tenantName: 'Maple Street household with a long shared name', inventoryName: 'Main',
  canAdd: true, canReturn: true, recentAssets: [], checkedOutAssets: [],
  tenants: [{ id: scope.tenantId, name: 'Maple Street household with a long shared name' }, { id: 'audit-workshop', name: 'Workshop household' }],
  inventories: [
    { id: scope.inventoryId, tenantId: scope.tenantId, tenantName: 'Maple Street', name: 'Main', roleLabel: 'Owner', updatedAtLabel: 'Today' },
    { id: 'audit-tools', tenantId: 'audit-workshop', tenantName: 'Workshop household', name: 'Workshop tools', roleLabel: 'Contributor', updatedAtLabel: 'Today' }
  ]
};

/** Warm synthetic data isolates native presentation; it does not test discovery. */
export function InventorySwitcherFixture() {
  const [state] = useState(() => {
    const client = createMobileQueryClient();
    const key = mobileQueryKeys.home('audit-switcher', scope.tenantId, scope.inventoryId);
    client.setQueryDefaults(key, { staleTime: Infinity });
    client.setQueryData(mobileQueryKeys.inventoryScope('audit-switcher'), scope);
    client.setQueryData(key, dashboard);
    let attempts = 0;
    return {
      client,
      query: new HomeDashboardQuery({ async getHomeDashboardSnapshot() { throw new Error('The presentation fixture must use its seeded dashboard'); } }),
      command: new SelectInventoryCommand({ async selectInventory() { if (++attempts === 1) throw new Error('Synthetic selection failure'); } })
    };
  });
  useEffect(() => () => state.client.clear(), [state]);
  return <MobileServerStateProvider client={state.client} scopeId="audit-switcher" loadInventoryScope={async () => scope}>
    <TenantSwitcherSheetScreen dashboardQuery={state.query} selectInventoryCommand={state.command} />
  </MobileServerStateProvider>;
}
