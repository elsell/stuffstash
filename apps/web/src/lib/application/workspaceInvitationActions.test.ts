import { describe, expect, it } from 'vitest';
import type { InventoryAccessInvitation } from '$lib/domain/inventory';
import {
  accessInvitationsHref,
  invitationActionConfirmation,
  invitationActionHref,
  invitationActionIsAvailable,
  invitationActionOptions
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

  it('builds row action options with hrefs, labels, tone, and availability', () => {
    expect(
      invitationActionOptions({
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        invitationStatus: 'pending',
        invitation: invitation('pending', false),
        busy: false
      })
    ).toEqual([
      {
        action: 'expire',
        label: 'Expire',
        ariaLabel: undefined,
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access/invitations/invite-one/expire?invitationStatus=pending',
        disabled: false,
        destructive: false,
        iconOnly: false
      },
      {
        action: 'cancel',
        label: 'Cancel',
        ariaLabel: undefined,
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access/invitations/invite-one/cancel?invitationStatus=pending',
        disabled: false,
        destructive: false,
        iconOnly: false
      },
      {
        action: 'delete',
        label: 'Delete',
        ariaLabel: 'Delete invitation for friend@example.test',
        href: '/tenants/tenant-one/inventories/inventory-one/settings/access/invitations/invite-one/delete?invitationStatus=pending',
        disabled: false,
        destructive: true,
        iconOnly: true
      }
    ]);
  });

  it('disables unavailable or busy row action options', () => {
    expect(
      invitationActionOptions({
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        invitationStatus: 'expired',
        invitation: invitation('pending', true),
        busy: false
      }).map((option) => [option.action, option.disabled])
    ).toEqual([
      ['expire', true],
      ['cancel', true],
      ['delete', false]
    ]);

    expect(
      invitationActionOptions({
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        invitationStatus: 'pending',
        invitation: invitation('pending', false),
        busy: true
      }).map((option) => option.disabled)
    ).toEqual([true, true, true]);
  });

  it('builds confirmation copy and disabled state for routed actions', () => {
    expect(invitationActionConfirmation('expire', invitation('pending', false), false)).toEqual({
      title: 'Expire invitation',
      description: 'Set the invitation for friend@example.test to expire immediately.',
      buttonLabel: 'Expire',
      destructive: false,
      disabled: false
    });
    expect(invitationActionConfirmation('delete', invitation('accepted', true), true)).toEqual({
      title: 'Delete invitation',
      description: 'Permanently remove the invitation record for friend@example.test.',
      buttonLabel: 'Delete',
      destructive: true,
      disabled: true
    });
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
