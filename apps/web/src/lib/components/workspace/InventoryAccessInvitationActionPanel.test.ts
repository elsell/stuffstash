import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { InventoryAccessInvitation } from '$lib/domain/inventory';
import InventoryAccessInvitationActionPanel, { invitationActionFocusTarget, type InventoryAccessInvitationActionPanelProps } from './InventoryAccessInvitationActionPanel.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(async () => {
  document.body.querySelector<HTMLElement>('[role="alertdialog"]')?.dispatchEvent(
    new KeyboardEvent('keydown', { key: 'Escape', bubbles: true })
  );
  await new Promise((resolve) => window.setTimeout(resolve, 20));
  if (component) {
    await unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryAccessInvitationActionPanel', () => {
  it('restores the exact surviving trigger and falls back when that trigger disappears', () => {
    const trigger = document.createElement('button');
    const heading = document.createElement('h2');
    document.body.append(trigger, heading);

    expect(invitationActionFocusTarget(trigger, heading)).toBe(trigger);
    trigger.remove();
    expect(invitationActionFocusTarget(trigger, heading)).toBe(heading);
  });

  it('does not dismiss the route after external teardown', async () => {
    vi.useFakeTimers();
    let dismissed = 0;
    let closeAutoFocused = 0;
    try {
      component = mount(InventoryAccessInvitationActionPanel, {
        target: document.body,
        props: panelProps({
          onClose: (event) => event.preventDefault(),
          onCloseAutoFocus: () => { closeAutoFocused += 1; },
          onDismiss: () => { dismissed += 1; }
        })
      });
      await tick();

      link('Cancel').click();
      await tick();
      expect(closeAutoFocused).toBe(1);
      expect(vi.getTimerCount()).toBeGreaterThan(0);
      await unmount(component);
      component = null;
      vi.runAllTimers();

      expect(dismissed).toBe(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it('renders an available route-backed action with durable cancel href', async () => {
    const confirmed: string[] = [];
    component = mount(InventoryAccessInvitationActionPanel, {
      target: document.body,
      props: panelProps({
        action: 'expire',
        invitation: invitation(),
        onConfirm: async (action, target) => {
          confirmed.push(`${action}:${target.id}`);
          return true;
        }
      })
    });
    await tick();

    expect(document.body.querySelector('[role="alertdialog"]')).not.toBeNull();
    expect(document.activeElement?.textContent).toBe('Cancel');
    expect(link('Cancel').classList.contains('min-h-11')).toBe(true);
    expect(requiredButton('Expire').classList.contains('min-h-11')).toBe(true);
    expect(document.body.textContent).toContain('Expire invitation');
    expect(document.body.textContent).toContain('Set the invitation for friend@example.test to expire immediately.');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending');

    requiredButton('Expire').click();
    await tick();

    expect(confirmed).toEqual(['expire:invite-one']);
  });

  it('renders delete as the destructive action', async () => {
    component = mount(InventoryAccessInvitationActionPanel, {
      target: document.body,
      props: panelProps({
        action: 'delete',
        invitation: invitation('accepted')
      })
    });
    await tick();

    expect(document.body.textContent).toContain('Delete invitation');
    expect(requiredButton('Delete')).not.toBeNull();
  });

  it('renders unavailable state for missing or invalid routed invitations', async () => {
    component = mount(InventoryAccessInvitationActionPanel, {
      target: document.body,
      props: panelProps({
        action: 'cancel',
        invitation: invitation('accepted')
      })
    });
    await tick();

    expect(document.body.textContent).toContain('Invitation unavailable');
    expect(link('Back to invitations').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending');
    expect(optionalButton('Cancel invitation')).toBeNull();
  });

  it('keeps an operation failure inside the open confirmation', async () => {
    component = mount(InventoryAccessInvitationActionPanel, {
      target: document.body,
      props: panelProps({
        action: 'delete',
        invitation: invitation('accepted'),
        error: 'Delete not saved. The invitation still exists.'
      })
    });
    await tick();

    const dialog = document.body.querySelector('[role="alertdialog"]');
    expect(dialog?.textContent).toContain('Delete not saved. The invitation still exists.');
    expect(dialog?.querySelector('[role="alert"]')).not.toBeNull();
  });
});

function panelProps(overrides: Partial<InventoryAccessInvitationActionPanelProps> = {}): InventoryAccessInvitationActionPanelProps {
  return {
    action: 'expire',
    invitation: invitation(),
    busy: false,
    accessHref: '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending',
    error: '',
    onClose: () => {},
    onDismiss: () => {},
    onCloseAutoFocus: () => {},
    onConfirm: async () => true,
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
