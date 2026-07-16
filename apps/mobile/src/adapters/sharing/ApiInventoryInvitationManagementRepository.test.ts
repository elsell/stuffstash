import { describe, expect, it } from 'vitest';
import type { CreatedInventoryAccessInvitation, InventoryAccessInvitation, Page } from '@stuff-stash/api-client';
import type { InventorySharingScope } from '../../application/sharing/InventorySharing';
import { ApiInventoryInvitationManagementRepository } from './ApiInventoryInvitationManagementRepository';

const scope: InventorySharingScope = {
  tenantId: 'tenant-home', inventoryId: 'inventory-home', inventoryName: 'Household', permissions: ['share']
};
const token = 'a'.repeat(43);
const inviteUrl = `https://stash.example/invitations/accept?tenant=${scope.tenantId}&inventory=${scope.inventoryId}&invitation=invite-one#token=${token}`;
const trustedInvitationOrigin = 'https://stash.example';

const invitation: InventoryAccessInvitation = {
  id: 'invite-one', tenantId: scope.tenantId, inventoryId: scope.inventoryId,
  email: 'friend@example.com', relationship: 'viewer', status: 'pending', isExpired: false,
  expiresAt: '2026-07-21T12:00:00Z', inviterPrincipalId: 'owner-one'
};

describe('ApiInventoryInvitationManagementRepository', () => {
  it('paginates safe invitation metadata and never invents listed invite links', async () => {
    const calls: Array<string | undefined> = [];
    const repository = new ApiInventoryInvitationManagementRepository({
      listInventoryAccessInvitations: async (_tenant, _inventory, options): Promise<Page<InventoryAccessInvitation>> => {
        calls.push(options?.cursor);
        return options?.cursor
          ? { items: [{ ...invitation, id: 'invite-two', status: 'accepted' }], pagination: { limit: 50, nextCursor: null, hasMore: false } }
          : { items: [invitation], pagination: { limit: 50, nextCursor: 'next', hasMore: true } };
      },
      createInventoryAccessInvitation: async () => ({ ...invitation, inviteUrl }),
      cancelInventoryAccessInvitation: async () => undefined
    }, trustedInvitationOrigin);

    await expect(repository.list(scope)).resolves.toEqual([
      expect.objectContaining({ id: 'invite-one' }),
      expect.objectContaining({ id: 'invite-two' })
    ]);
    expect(JSON.stringify(await repository.list(scope))).not.toContain('inviteUrl');
    expect(calls).toEqual([undefined, 'next', undefined, 'next']);
  });

  it('requires the one-time complete URL in creation responses', async () => {
    const repository = new ApiInventoryInvitationManagementRepository({
      listInventoryAccessInvitations: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
      createInventoryAccessInvitation: async () => invitation as unknown as CreatedInventoryAccessInvitation,
      cancelInventoryAccessInvitation: async () => undefined
    }, trustedInvitationOrigin);
    await expect(repository.create(scope, { email: 'friend@example.com', relationship: 'viewer' }))
      .rejects.toThrow('Stuff Stash did not return the one-time invitation link.');
  });

  it('scopes create and cancel calls to the selected inventory', async () => {
    const calls: string[] = [];
    const repository = new ApiInventoryInvitationManagementRepository({
      listInventoryAccessInvitations: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
      createInventoryAccessInvitation: async (tenant, inventory, input) => {
        calls.push(`create:${tenant}:${inventory}:${input.relationship}`);
        return { ...invitation, inviteUrl };
      },
      cancelInventoryAccessInvitation: async (tenant, inventory, id) => {
        calls.push(`cancel:${tenant}:${inventory}:${id}`);
      }
    }, trustedInvitationOrigin);
    await repository.create(scope, { email: 'friend@example.com', relationship: 'editor' });
    await repository.cancel(scope, 'invite-one');
    expect(calls).toEqual([
      'create:tenant-home:inventory-home:editor',
      'cancel:tenant-home:inventory-home:invite-one'
    ]);
  });

  it('rejects a one-time URL with a mismatched scope, invitation ID, duplicate, or unknown fields', async () => {
    for (const invalidUrl of [
      inviteUrl.replace('inventory=inventory-home', 'inventory=other'),
      inviteUrl.replace('invitation=invite-one', 'invitation=other'),
      inviteUrl.replace('tenant=tenant-home', 'tenant=tenant-home&tenant=tenant-home'),
      inviteUrl.replace('&inventory=', '&campaign=x&inventory=')
    ]) {
      const repository = new ApiInventoryInvitationManagementRepository({
        listInventoryAccessInvitations: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
        createInventoryAccessInvitation: async () => ({ ...invitation, inviteUrl: invalidUrl }),
        cancelInventoryAccessInvitation: async () => undefined
      }, trustedInvitationOrigin);
      await expect(repository.create(scope, { email: 'friend@example.com', relationship: 'viewer' }))
        .rejects.toThrow('Stuff Stash did not return the one-time invitation link.');
    }
  });

  it('rejects a creation response whose invitation metadata is outside the selected scope', async () => {
    const repository = new ApiInventoryInvitationManagementRepository({
      listInventoryAccessInvitations: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
      createInventoryAccessInvitation: async () => ({ ...invitation, inventoryId: 'inventory-other', inviteUrl }),
      cancelInventoryAccessInvitation: async () => undefined
    }, trustedInvitationOrigin);
    await expect(repository.create(scope, { email: 'friend@example.com', relationship: 'viewer' }))
      .rejects.toThrow('Stuff Stash did not return the one-time invitation link.');
  });

  it('rejects repeated pagination cursors instead of looping forever', async () => {
    const repository = new ApiInventoryInvitationManagementRepository({
      listInventoryAccessInvitations: async () => ({
        items: [], pagination: { limit: 50, nextCursor: 'repeat', hasMore: true }
      }),
      createInventoryAccessInvitation: async () => ({ ...invitation, inviteUrl }),
      cancelInventoryAccessInvitation: async () => undefined
    }, trustedInvitationOrigin);
    await expect(repository.list(scope)).rejects.toThrow('invalid invitation page');
  });

  it('rejects listed invitation metadata outside the selected inventory scope', async () => {
    const repository = new ApiInventoryInvitationManagementRepository({
      listInventoryAccessInvitations: async () => ({
        items: [{ ...invitation, tenantId: 'tenant-other' }],
        pagination: { limit: 50, nextCursor: null, hasMore: false }
      }),
      createInventoryAccessInvitation: async () => ({ ...invitation, inviteUrl }),
      cancelInventoryAccessInvitation: async () => undefined
    }, trustedInvitationOrigin);
    await expect(repository.list(scope)).rejects.toThrow('invalid invitation response');
  });
});
