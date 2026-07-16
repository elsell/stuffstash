import { describe, expect, it } from 'vitest';
import {
  CancelInventoryInvitationCommand,
  CreateInventoryInvitationCommand,
  InventorySharingPermissionError,
  ListInventoryInvitationsQuery,
  type InventoryInvitationManagementRepository,
  type InventorySharingScope
} from './InventorySharing';

const ownerScope: InventorySharingScope = {
  tenantId: 'tenant-home',
  inventoryId: 'inventory-home',
  inventoryName: 'Household',
  permissions: ['view', 'share']
};

class FakeInvitationRepository implements InventoryInvitationManagementRepository {
  readonly calls: string[] = [];

  async list(scope: InventorySharingScope) {
    this.calls.push(`list:${scope.inventoryId}`);
    return [{
      id: 'invite-one',
      email: 'friend@example.com',
      relationship: 'viewer' as const,
      status: 'pending' as const,
      isExpired: false,
      expiresAt: '2026-07-21T12:00:00Z'
    }];
  }

  async create(scope: InventorySharingScope, input: { email: string; relationship: 'viewer' | 'editor' }) {
    this.calls.push(`create:${scope.inventoryId}:${input.email}:${input.relationship}`);
    return {
      id: 'invite-two',
      email: input.email,
      relationship: input.relationship,
      status: 'pending' as const,
      isExpired: false,
      expiresAt: '2026-07-21T12:00:00Z',
      inviteUrl: 'https://stash.example/invitations/accept?tenant=t&inventory=i&invitation=x#token=secret'
    };
  }

  async cancel(scope: InventorySharingScope, invitationId: string) {
    this.calls.push(`cancel:${scope.inventoryId}:${invitationId}`);
  }
}

describe('mobile inventory sharing actions', () => {
  it('lists, creates, and cancels invitations through the mobile-owned port', async () => {
    const repository = new FakeInvitationRepository();
    await expect(new ListInventoryInvitationsQuery(repository).execute(ownerScope))
      .resolves.toHaveLength(1);
    await expect(new CreateInventoryInvitationCommand(repository).execute(ownerScope, {
      email: ' friend@example.com ',
      relationship: 'editor'
    })).resolves.toMatchObject({ email: 'friend@example.com', relationship: 'editor' });
    await new CancelInventoryInvitationCommand(repository).execute(ownerScope, 'invite-one');
    expect(repository.calls).toEqual([
      'list:inventory-home',
      'create:inventory-home:friend@example.com:editor',
      'cancel:inventory-home:invite-one'
    ]);
  });

  it('rejects missing share permission before calling the adapter', async () => {
    const repository = new FakeInvitationRepository();
    const viewerScope = { ...ownerScope, permissions: ['view'] };
    await expect(new ListInventoryInvitationsQuery(repository).execute(viewerScope))
      .rejects.toBeInstanceOf(InventorySharingPermissionError);
    await expect(new CreateInventoryInvitationCommand(repository).execute(viewerScope, {
      email: 'friend@example.com', relationship: 'viewer'
    })).rejects.toBeInstanceOf(InventorySharingPermissionError);
    await expect(new CancelInventoryInvitationCommand(repository).execute(viewerScope, 'invite-one'))
      .rejects.toBeInstanceOf(InventorySharingPermissionError);
    expect(repository.calls).toEqual([]);
  });

  it('validates invitation input before calling the adapter', async () => {
    const repository = new FakeInvitationRepository();
    await expect(new CreateInventoryInvitationCommand(repository).execute(ownerScope, {
      email: 'not-an-email', relationship: 'viewer'
    })).rejects.toThrow('Enter a valid email address.');
    await expect(new CancelInventoryInvitationCommand(repository).execute(ownerScope, '   '))
      .rejects.toThrow('Invitation ID must not be empty.');
    expect(repository.calls).toEqual([]);
  });
});
