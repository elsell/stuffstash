import { describe, expect, it } from 'vitest';
import type { InventoryAccessInvitation } from '$lib/domain/inventory';
import { invitationActionIsAvailable } from './invitationActionPolicy';

describe('invitationActionIsAvailable', () => {
  it('allows expire and cancel only for unexpired pending invitations', () => {
    expect(invitationActionIsAvailable('expire', invitation('pending', false))).toBe(true);
    expect(invitationActionIsAvailable('cancel', invitation('pending', false))).toBe(true);

    expect(invitationActionIsAvailable('expire', invitation('pending', true))).toBe(false);
    expect(invitationActionIsAvailable('cancel', invitation('accepted', false))).toBe(false);
  });

  it('allows delete for retained invitation records regardless of status', () => {
    expect(invitationActionIsAvailable('delete', invitation('accepted', true))).toBe(true);
  });
});

function invitation(status: InventoryAccessInvitation['status'], isExpired: boolean): InventoryAccessInvitation {
  return {
    id: 'invite-one',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    email: 'friend@example.test',
    relationship: 'viewer',
    status,
    isExpired,
    expiresAt: '2026-06-30T00:00:00Z',
    inviterPrincipalId: 'principal-one'
  };
}
