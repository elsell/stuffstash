import { describe, expect, it } from 'vitest';
import {
  StuffStashAPIError,
  type InvitationAcceptance,
  type InventoryAccessInvitationPreview
} from '@stuff-stash/api-client';
import {
  InventoryInvitationAuthenticationRequiredError,
  InventoryInvitationEmailMismatchError,
  InventoryInvitationInvalidError,
  InventoryInvitationInvalidResponseError,
  type InventoryInvitationReference
} from '../../application/invitations/InventoryInvitationRepository';
import { ApiInventoryInvitationRepository } from './ApiInventoryInvitationRepository';

const reference: InventoryInvitationReference = {
  tenantId: 'tenant-one',
  inventoryId: 'inventory-one',
  invitationId: 'invite-one',
  acceptanceToken: 'raw-token'
};

describe('ApiInventoryInvitationRepository', () => {
  it('previews using the scoped generated-client operation and maps safe metadata', async () => {
    const calls: unknown[][] = [];
    const preview: InventoryAccessInvitationPreview = {
      inventoryId: 'inventory-one',
      inventoryName: 'Household',
      relationship: 'viewer',
      status: 'pending',
      isExpired: false,
      expiresAt: '2026-08-01T00:00:00Z'
    };
    const repository = new ApiInventoryInvitationRepository({
      previewInventoryAccessInvitation: async (...input) => {
        calls.push(input);
        return preview;
      },
      acceptInventoryAccessInvitation: async () => {
        throw new Error('Acceptance must not run during preview.');
      }
    });

    await expect(repository.preview(reference)).resolves.toEqual(preview);
    expect(calls).toEqual([
      ['tenant-one', 'inventory-one', 'invite-one', 'raw-token']
    ]);
  });

  it('accepts using the scoped generated-client operation and maps the new access', async () => {
    const calls: unknown[][] = [];
    const acceptance: InvitationAcceptance = {
      grant: {
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        principalId: 'principal-new',
        relationship: 'editor'
      },
      invitation: {
        id: 'invite-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        email: 'member@example.test',
        relationship: 'editor',
        status: 'accepted',
        isExpired: false,
        expiresAt: '2026-08-01T00:00:00Z',
        inviterPrincipalId: 'principal-owner',
        acceptedPrincipalId: 'principal-new'
      }
    };
    const repository = new ApiInventoryInvitationRepository({
      previewInventoryAccessInvitation: async () => {
        throw new Error('Preview must not run during acceptance.');
      },
      acceptInventoryAccessInvitation: async (...input) => {
        calls.push(input);
        return acceptance;
      }
    });

    await expect(repository.accept(reference)).resolves.toEqual({
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      invitationId: 'invite-one',
      principalId: 'principal-new',
      relationship: 'editor',
      status: 'accepted'
    });
    expect(calls).toEqual([
      ['tenant-one', 'inventory-one', 'invite-one', 'raw-token']
    ]);
  });

  it.each([
    [401, 'unauthenticated', InventoryInvitationAuthenticationRequiredError],
    [403, 'invitation_email_mismatch', InventoryInvitationEmailMismatchError],
    [404, 'invitation_invalid', InventoryInvitationInvalidError]
  ])('maps invitation API failures to safe mobile-owned errors', async (status, code, ErrorType) => {
    const repository = new ApiInventoryInvitationRepository({
      previewInventoryAccessInvitation: async () => {
        throw new StuffStashAPIError(status, code, `Unsafe server detail: ${reference.acceptanceToken}`);
      },
      acceptInventoryAccessInvitation: async () => {
        throw new StuffStashAPIError(status, code, `Unsafe server detail: ${reference.acceptanceToken}`);
      }
    });

    const previewError = await repository.preview(reference).catch((error: unknown) => error);
    const acceptanceError = await repository.accept(reference).catch((error: unknown) => error);

    expect(previewError).toBeInstanceOf(ErrorType);
    expect(acceptanceError).toBeInstanceOf(ErrorType);
    expect((previewError as Error).message).not.toContain(reference.acceptanceToken);
    expect((acceptanceError as Error).message).not.toContain(reference.acceptanceToken);
  });

  it('preserves unknown API failures for retry handling', async () => {
    const failure = new StuffStashAPIError(503, 'unavailable', 'Please try again.');
    const repository = new ApiInventoryInvitationRepository({
      previewInventoryAccessInvitation: async () => {
        throw failure;
      },
      acceptInventoryAccessInvitation: async () => {
        throw failure;
      }
    });

    await expect(repository.preview(reference)).rejects.toBe(failure);
    await expect(repository.accept(reference)).rejects.toBe(failure);
  });

  it('fails closed when a preview response does not match the requested inventory scope', async () => {
    const repository = new ApiInventoryInvitationRepository({
      previewInventoryAccessInvitation: async () => ({
        inventoryId: 'inventory-other',
        inventoryName: 'Other inventory',
        relationship: 'viewer',
        status: 'pending',
        isExpired: false,
        expiresAt: '2026-08-01T00:00:00Z'
      }),
      acceptInventoryAccessInvitation: async () => {
        throw new Error('not used');
      }
    });

    await expect(repository.preview(reference)).rejects.toBeInstanceOf(
      InventoryInvitationInvalidResponseError
    );
  });

  it.each([
    ['grant tenant', (value: InvitationAcceptance) => ({ ...value, grant: { ...value.grant, tenantId: 'tenant-other' } })],
    ['grant inventory', (value: InvitationAcceptance) => ({ ...value, grant: { ...value.grant, inventoryId: 'inventory-other' } })],
    ['invitation ID', (value: InvitationAcceptance) => ({ ...value, invitation: { ...value.invitation, id: 'invite-other' } })],
    ['invitation tenant', (value: InvitationAcceptance) => ({ ...value, invitation: { ...value.invitation, tenantId: 'tenant-other' } })],
    ['invitation inventory', (value: InvitationAcceptance) => ({ ...value, invitation: { ...value.invitation, inventoryId: 'inventory-other' } })],
    ['invitation relationship', (value: InvitationAcceptance) => ({ ...value, invitation: { ...value.invitation, relationship: 'viewer' as const } })],
    ['invitation status', (value: InvitationAcceptance) => ({ ...value, invitation: { ...value.invitation, status: 'pending' as const } })]
  ])('fails closed when acceptance has a mismatched %s', async (_label, change) => {
    const valid: InvitationAcceptance = {
      grant: {
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        principalId: 'principal-new',
        relationship: 'editor'
      },
      invitation: {
        id: 'invite-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        email: 'member@example.test',
        relationship: 'editor',
        status: 'accepted',
        isExpired: false,
        expiresAt: '2026-08-01T00:00:00Z',
        inviterPrincipalId: 'principal-owner',
        acceptedPrincipalId: 'principal-new'
      }
    };
    const repository = new ApiInventoryInvitationRepository({
      previewInventoryAccessInvitation: async () => {
        throw new Error('not used');
      },
      acceptInventoryAccessInvitation: async () => change(valid)
    });

    await expect(repository.accept(reference)).rejects.toBeInstanceOf(
      InventoryInvitationInvalidResponseError
    );
  });
});
