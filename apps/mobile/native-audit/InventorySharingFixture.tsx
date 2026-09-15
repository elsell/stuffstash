import { useEffect, useState } from 'react';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { QueryClientInvitationMutationObserver } from '../src/adapters/serverState/QueryClientInvitationMutationObserver';
import {
  CancelInventoryInvitationCommand, CreateInventoryInvitationCommand, InventoryInvitationLinkUnavailableError,
  ListInventoryInvitationsQuery, type InventoryInvitationManagementRepository, type InventoryInvitationSummary,
  type InventorySharingScope
} from '../src/application/sharing/InventorySharing';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { InventorySharingScreen } from '../src/ui/screens/InventorySharingScreen';

const scope: InventorySharingScope = {
  tenantId: 'audit-tenant', inventoryId: 'audit-inventory', inventoryName: 'Audit inventory', permissions: ['share']
};

/** Runner-only controlled ports: no remote invitations, system clipboard, or real share action. */
export function InventorySharingFixture() {
  const [fixture] = useState(() => {
    const client = createMobileQueryClient();
    let creations = 0; let copies = 0; let cancellations = 0;
    let items: InventoryInvitationSummary[] = [];
    const repository: InventoryInvitationManagementRepository = {
      list: async () => ({ items }),
      create: async (_scope, input) => {
        const invitation: InventoryInvitationSummary = {
          id: `audit-invitation-${++creations}`, email: input.email, relationship: input.relationship,
          status: 'pending', isExpired: false, expiresAt: '2027-01-01T00:00:00Z'
        };
        items = [invitation, ...items];
        if (creations === 1) throw new InventoryInvitationLinkUnavailableError();
        return { ...invitation, inviteUrl: 'https://example.invalid/invite#token=synthetic-audit-value' };
      },
      cancel: async (_scope, id) => {
        if (++cancellations === 1) throw new Error('Audit cancellation unavailable. Try again.');
        items = items.map(item => item.id === id ? { ...item, status: 'cancelled' } : item);
      }
    };
    const observer = new QueryClientInvitationMutationObserver(client, 'audit-sharing');
    return { client, listQuery: new ListInventoryInvitationsQuery(repository),
      createCommand: new CreateInventoryInvitationCommand(repository, observer),
      cancelCommand: new CancelInventoryInvitationCommand(repository, observer),
      linkActions: {
        copy: async () => { if (++copies === 1) throw new Error('Audit copy unavailable. Try again.'); },
        share: async () => { throw new Error('Audit sharing does not open external destinations.'); }
      }
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit-sharing" loadInventoryScope={async () => scope}>
    <InventorySharingScreen scope={scope} listQuery={fixture.listQuery} createCommand={fixture.createCommand}
      cancelCommand={fixture.cancelCommand} linkActions={fixture.linkActions} />
  </MobileServerStateProvider>;
}
