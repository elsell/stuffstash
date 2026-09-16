import { useState } from 'react';
import { Text } from 'react-native';
import { useRouter } from 'expo-router';
import type { InventoryInvitationAcceptance, InventoryInvitationPreview } from '../src/application/invitations/InventoryInvitationRepository';
import { InventoryInvitationScreen } from '../src/ui/screens/InventoryInvitationScreen';

/** Isolated UI ports: no server, live invitation, credentials or external links. */
export function InvitationAcceptanceFixture() {
  const router = useRouter();
  const [result, setResult] = useState('');
  const [fixture] = useState(() => {
    let accepted = 0;
    let opened = 0;
    const reference = { tenantId: 'audit-tenant', inventoryId: 'audit-inventory', invitationId: 'audit-invitation', acceptanceToken: 'synthetic-fixture-token' };
    return {
      reference,
      previewQuery: { execute: async (): Promise<InventoryInvitationPreview> => ({
        inventoryId: reference.inventoryId,
        inventoryName: 'Family camping equipment and seasonal supplies shared with the household',
        relationship: 'editor', status: 'pending', isExpired: false, expiresAt: '2030-12-01T12:00:00Z'
      }) },
      acceptCommand: { execute: async (): Promise<InventoryInvitationAcceptance> => {
        accepted++;
        return { tenantId: reference.tenantId, inventoryId: reference.inventoryId, invitationId: reference.invitationId, principalId: 'audit-principal', relationship: 'editor', status: 'accepted' };
      } },
      open: async (inventoryId: string) => {
        if (++opened === 1) throw new Error('Synthetic opening failure');
        setResult(inventoryId === reference.inventoryId && accepted === 1
          ? 'Opened invitation inventory; accepted once' : 'Unexpected invitation acceptance');
      },
      dismiss: () => {
        if (accepted === 0) router.back();
        else setResult('Unexpected acceptance before dismissal');
      }
    };
  });
  if (result) return <Text>{result}</Text>;
  return <InventoryInvitationScreen initialized invalidLink={false} reference={fixture.reference}
    previewQuery={fixture.previewQuery} acceptCommand={fixture.acceptCommand}
    onAccepted={fixture.open} onDismiss={fixture.dismiss}
    onStartOver={async () => setResult('Unexpected start over')}
    onSwitchAccount={() => setResult('Unexpected account switch')} />;
}
