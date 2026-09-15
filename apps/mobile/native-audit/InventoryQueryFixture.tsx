import { useState } from 'react';
import { ScrollView, Text } from 'react-native';
import { createMobileQueryClient, mobileQueryKeys } from '../src/adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { useMobileInventoryServerQuery } from '../src/ui/serverState/useMobileInventoryServerQuery';
import { QueryReadinessDiagnostics } from './QueryReadinessDiagnostics';

/** Cold-cache comparison; no production session or remote service is used. */
export function InventoryQueryFixture() {
  const [client] = useState(createMobileQueryClient);
  return <MobileServerStateProvider client={client} scopeId="audit-query"
    loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <InventoryQueries />
    <QueryReadinessDiagnostics client={client} />
  </MobileServerStateProvider>;
}

function InventoryQueries() {
  const first = useMobileInventoryServerQuery({ key: mobileQueryKeys.addContext, query: async () => 'First query ready' });
  const second = useMobileInventoryServerQuery({ key: mobileQueryKeys.locations, enabled: Boolean(first.data), query: async () => 'Dependent query ready' });
  return <ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={{ padding: 20, gap: 16 }}>
    <Text>{first.data ?? 'First query pending'}</Text>
    <Text>{second.data ?? 'Dependent query pending'}</Text>
  </ScrollView>;
}
