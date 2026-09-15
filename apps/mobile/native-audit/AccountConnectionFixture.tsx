import { useEffect, useState } from 'react';
import { usePathname, useRouter } from 'expo-router';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { SettingsQuery } from '../src/application/settings/SettingsQuery';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { AccountSettingsScreen, ConnectionSettingsScreen } from '../src/ui/screens/SettingsDetailScreens';

/** Controlled session actions: never contact an auth provider or change real settings. */
export function AccountConnectionFixture() {
  const pathname = usePathname();
  const router = useRouter();
  const [fixture] = useState(() => {
    let attempts = 0;
    return {
      client: createMobileQueryClient(),
      query: new SettingsQuery(
        { getCurrentPrincipal: async () => ({ id: 'audit-principal', email: 'audit@example.invalid' }) },
        { getDiagnostics: () => ({ apiBaseUrl: 'https://audit.example.invalid/api', appVersion: 'audit', authenticationMode: 'oidc-sso' }) },
        { getSelectedScope: async () => { throw new Error('Account recovery must not need inventory access'); } }
      ),
      act: async () => {
        if (++attempts === 1) throw new Error('Audit session action unavailable. Try again.');
        router.back();
      }
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit-account" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    {pathname.endsWith('audit-account')
      ? <AccountSettingsScreen settingsQuery={fixture.query} onSignOut={fixture.act} />
      : <ConnectionSettingsScreen settingsQuery={fixture.query} onChangeServer={fixture.act} />}
  </MobileServerStateProvider>;
}
