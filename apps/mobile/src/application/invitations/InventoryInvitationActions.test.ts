import { describe, expect, it } from 'vitest';
import type {
  InventoryInvitationAcceptance,
  InventoryInvitationPreview,
  InventoryInvitationReference,
  InventoryInvitationRepository
} from './InventoryInvitationRepository';
import { AcceptInventoryInvitationCommand } from './AcceptInventoryInvitationCommand';
import { PreviewInventoryInvitationQuery } from './PreviewInventoryInvitationQuery';

const reference: InventoryInvitationReference = {
  tenantId: 'tenant-one',
  inventoryId: 'inventory-one',
  invitationId: 'invite-one',
  acceptanceToken: 'raw-token'
};

class FakeInventoryInvitationRepository implements InventoryInvitationRepository {
  readonly previewCalls: InventoryInvitationReference[] = [];
  readonly acceptanceCalls: InventoryInvitationReference[] = [];

  async preview(input: InventoryInvitationReference): Promise<InventoryInvitationPreview> {
    this.previewCalls.push(input);
    return {
      inventoryId: input.inventoryId,
      inventoryName: 'Household',
      relationship: 'editor',
      status: 'pending',
      isExpired: false,
      expiresAt: '2026-08-01T00:00:00Z'
    };
  }

  async accept(input: InventoryInvitationReference): Promise<InventoryInvitationAcceptance> {
    this.acceptanceCalls.push(input);
    return {
      tenantId: input.tenantId,
      inventoryId: input.inventoryId,
      invitationId: input.invitationId,
      principalId: 'principal-new',
      relationship: 'editor',
      status: 'accepted'
    };
  }
}

describe('inventory invitation application actions', () => {
  it('previews through the repository port without accepting', async () => {
    const repository = new FakeInventoryInvitationRepository();

    await expect(new PreviewInventoryInvitationQuery(repository).execute(reference)).resolves.toEqual({
      inventoryId: 'inventory-one',
      inventoryName: 'Household',
      relationship: 'editor',
      status: 'pending',
      isExpired: false,
      expiresAt: '2026-08-01T00:00:00Z'
    });
    expect(repository.previewCalls).toEqual([reference]);
    expect(repository.acceptanceCalls).toEqual([]);
  });

  it('accepts only when the explicit command is executed', async () => {
    const repository = new FakeInventoryInvitationRepository();
    const command = new AcceptInventoryInvitationCommand(repository);

    expect(repository.acceptanceCalls).toEqual([]);
    await expect(command.execute(reference)).resolves.toEqual({
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      invitationId: 'invite-one',
      principalId: 'principal-new',
      relationship: 'editor',
      status: 'accepted'
    });
    expect(repository.acceptanceCalls).toEqual([reference]);
  });
});
