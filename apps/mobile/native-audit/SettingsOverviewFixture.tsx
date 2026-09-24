import { useEffect, useState } from 'react';
import { router, type Href } from 'expo-router';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { SettingsQuery } from '../src/application/settings/SettingsQuery';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { SettingsScreen } from '../src/ui/screens/SettingsScreen';
import { InventorySettingsScreen } from '../src/ui/screens/ScopedSettingsScreens';
import { AccountSettingsScreen, DiagnosticsSettingsScreen } from '../src/ui/screens/SettingsDetailScreens';

/** Overview layout only; session commands and inventory leaf workflows have separate fixtures. */
export function SettingsRootOverviewFixture() { return <SettingsOverviewFixture page="root" />; }
export function SettingsAccountOverviewFixture() { return <SettingsOverviewFixture page="account" />; }
export function SettingsInventoryOverviewFixture() { return <SettingsOverviewFixture page="inventory" />; }
export function SettingsDiagnosticsOverviewFixture() { return <SettingsOverviewFixture page="diagnostics" />; }

function SettingsOverviewFixture({ page }: { readonly page: 'root' | 'account' | 'inventory' | 'diagnostics' }) {
  const [fixture] = useState(() => ({
    client: createMobileQueryClient(),
    query: new SettingsQuery(
      { getCurrentPrincipal: async () => ({ id: 'overview-principal', email: 'household.member@example.invalid' }) },
      { getDiagnostics: () => ({ apiBaseUrl: 'https://inventory.example.invalid/api', appVersion: 'overview-audit-1', authenticationMode: 'oidc-sso' }) },
      { getSelectedScope: async () => ({
        tenant: { id: 'overview-tenant', name: 'Family household', permissions: ['configure'] },
        inventory: { id: 'overview-inventory', name: 'Main Inventory', permissions: ['view', 'share', 'configure'] }
      }) }
    )
  }));
  useEffect(() => () => fixture.client.clear(), [fixture]);
  const screen = page === 'account'
    ? <AccountSettingsScreen settingsQuery={fixture.query} onSignOut={async () => { throw new Error('Use the session-action fixture to review sign out.'); }} />
    : page === 'inventory'
      ? <InventorySettingsScreen settingsQuery={fixture.query} onNavigate={() => { throw new Error('Use the inventory leaf fixtures for this workflow.'); }} />
      : page === 'diagnostics'
        ? <DiagnosticsSettingsScreen settingsQuery={fixture.query} />
        : <SettingsScreen settingsQuery={fixture.query} onNavigate={destination => {
          const pages = { account: 'account', 'inventory-settings': 'inventory', diagnostics: 'diagnostics' } as const;
          if (!(destination in pages)) throw new Error('This overview fixture covers Account, Inventory and Diagnostics.');
          router.push(`/audit-tabs/(home)/settings/${pages[destination as keyof typeof pages]}` as Href);
        }} />;
  return <MobileServerStateProvider client={fixture.client} scopeId="overview"
    loadInventoryScope={async () => ({ tenantId: 'overview-tenant', inventoryId: 'overview-inventory' })}>
    {screen}
  </MobileServerStateProvider>;
}
