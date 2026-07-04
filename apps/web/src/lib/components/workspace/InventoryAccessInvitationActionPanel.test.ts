import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { InventoryAccessInvitation } from '$lib/domain/inventory';
import InventoryAccessInvitationActionPanel, { type InventoryAccessInvitationActionPanelProps } from './InventoryAccessInvitationActionPanel.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryAccessInvitationActionPanel', () => {
  it('renders an available route-backed action with durable cancel href', async () => {
    const confirmed: string[] = [];
    component = mount(InventoryAccessInvitationActionPanel, {
      target: document.body,
      props: panelProps({
        action: 'expire',
        invitation: invitation(),
        onConfirm: async (action, target) => {
          confirmed.push(`${action}:${target.id}`);
        }
      })
    });

    expect(document.body.textContent).toContain('Expire invitation');
    expect(document.body.textContent).toContain('Set the invitation for friend@example.test to expire immediately.');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending');

    requiredButton('Expire').click();
    await tick();

    expect(confirmed).toEqual(['expire:invite-one']);
  });

  it('renders delete as the destructive action', () => {
    component = mount(InventoryAccessInvitationActionPanel, {
      target: document.body,
      props: panelProps({
        action: 'delete',
        invitation: invitation('accepted')
      })
    });

    expect(document.body.textContent).toContain('Delete invitation');
    expect(requiredButton('Delete')).not.toBeNull();
  });

  it('renders unavailable state for missing or invalid routed invitations', () => {
    component = mount(InventoryAccessInvitationActionPanel, {
      target: document.body,
      props: panelProps({
        action: 'cancel',
        invitation: invitation('accepted')
      })
    });

    expect(document.body.textContent).toContain('Invitation unavailable');
    expect(link('Back to invitations').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending');
    expect(optionalButton('Cancel invitation')).toBeNull();
  });
});

function panelProps(overrides: Partial<InventoryAccessInvitationActionPanelProps> = {}): InventoryAccessInvitationActionPanelProps {
  return {
    action: 'expire',
    invitation: invitation(),
    busy: false,
    accessHref: '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending',
    panelElement: null,
    onClose: () => {},
    onConfirm: async () => {},
    ...overrides
  };
}

function invitation(status: InventoryAccessInvitation['status'] = 'pending'): InventoryAccessInvitation {
  return {
    id: 'invite-one',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    email: 'friend@example.test',
    relationship: 'viewer',
    status,
    isExpired: false,
    expiresAt: '2026-06-30T00:00:00Z',
    inviterPrincipalId: 'principal-one'
  };
}

function requiredButton(text: string): HTMLButtonElement {
  const target = optionalButton(text);
  if (!target) {
    throw new Error(`Missing button ${text}`);
  }
  return target;
}

function optionalButton(text: string): HTMLButtonElement | null {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) => candidate.textContent === text) ?? null;
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent === text);
  if (!target) {
    throw new Error(`Missing link ${text}`);
  }
  return target;
}
