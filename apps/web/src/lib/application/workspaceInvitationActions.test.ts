import { describe, expect, it } from 'vitest';
import type { InventoryAccessInvitation } from '$lib/domain/inventory';
import {
  accessInvitationsHref,
  invitationActionHref,
  invitationActionIsAvailable
} from './workspaceInvitationActions';

describe('workspace invitation actions', () => {
  it('allows expire and cancel only for unexpired pending invitations', () => {
    expect(invitationActionIsAvailable('expire', invitation('pending', false))).toBe(true);
    expect(invitationActionIsAvailable('cancel', invitation('pending', false))).toBe(true);

    expect(invitationActionIsAvailable('expire', invitation('pending', true))).toBe(false);
    expect(invitationActionIsAvailable('cancel', invitation('accepted', false))).toBe(false);
  });

  it('allows delete for retained invitation records regardless of status', () => {
    expect(invitationActionIsAvailable('delete', invitation('accepted', true))).toBe(true);
  });

  it('builds canonical access invitation cancel and action hrefs', () => {
    expect(accessInvitationsHref('tenant-one', 'inventory-one', 'pending')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending'
    );
    expect(accessInvitationsHref('tenant-one', 'inventory-one', 'all')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access'
    );
    expect(invitationActionHref('tenant-one', 'inventory-one', 'pending', invitation('pending', false), 'delete')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access/invitations/invite-one/delete?invitationStatus=pending'
    );
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
