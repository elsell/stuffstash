import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import InventoryAccessManager from './InventoryAccessManager.svelte';
import type {
  CreatedInventoryAccessInvitation,
  Inventory,
  InventoryAccessGrant,
  InventoryAccessInvitation,
  InventoryAccessRelationship,
  InvitationStatusFilter,
  Tenant
} from '$lib/domain/inventory';
import type { InventoryAccessPage, InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryAccessManager', () => {
  it('loads grants and invitations when share access is available', async () => {
    const { repository, calls } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    expect(calls).toEqual(['list-grants:tenant-one:inventory-one:', 'list-invitations:tenant-one:inventory-one:all:']);
    expect(document.body.textContent).toContain('principal-two');
    expect(document.body.textContent).toContain('friend@example.test');
  });

  it('submits direct grants and invitations through the access repository port', async () => {
    const { repository, calls } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    await setInput('#grant-principal', 'principal-three');
    clickButton('Grant access');
    await flush();

    await setInput('#invite-email', 'new@example.test');
    clickButton('Create invite');
    await flush();

    expect(calls).toContain('grant:tenant-one:inventory-one:principal-three:viewer');
    expect(calls).toContain('invite:tenant-one:inventory-one:new@example.test:viewer');
    expect(document.body.textContent).toContain('principal-three');
    expect(document.body.textContent).toContain('new@example.test');
    expect(document.body.textContent).toContain('raw-token');
  });

  it('keeps acceptance token out of persistent invitation rows', async () => {
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    await setInput('#invite-email', 'new@example.test');
    clickButton('Create invite');
    await flush();

    expect(document.body.querySelector('.one-time-token')?.textContent).toContain('raw-token');
    expect(document.body.querySelector('[aria-label="Invitations"]')?.textContent).not.toContain('raw-token');
  });

  it('revokes grants through the access repository port', async () => {
    const { repository, calls } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    clickButton('Revoke');
    await flush();

    expect(calls).toContain('revoke:tenant-one:inventory-one:principal-two:viewer');
  });

  it('exposes route-backed invitation action links and opens them in-app', async () => {
    const { repository, calls } = fakeAccessRepository();
    const opened: string[] = [];

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository,
        invitationStatus: 'pending',
        onInvitationActionOpen: (action, invitationId) => {
          opened.push(`${action}:${invitationId}`);
        }
      }
    });
    await flush();

    expect(link('Expire').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access/invitations/invite-one/expire?invitationStatus=pending'
    );
    expect(link('Cancel').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access/invitations/invite-one/cancel?invitationStatus=pending'
    );
    expect(iconLink('Delete invitation for friend@example.test').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access/invitations/invite-one/delete?invitationStatus=pending'
    );

    link('Expire').click();
    await flush();

    const modifiedClick = new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true });
    link('Cancel').dispatchEvent(modifiedClick);
    await flush();

    expect(opened).toEqual(['expire:invite-one']);
    expect(modifiedClick.defaultPrevented).toBe(false);
    expect(calls).not.toContain('expire:tenant-one:inventory-one:invite-one:1970-01-01T00:00:00.000Z');
    expect(calls).not.toContain('cancel:tenant-one:inventory-one:invite-one');
  });

  it('confirms invitation actions from route state', async () => {
    const closed: string[] = [];

    let fake = fakeAccessRepository();
    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository: fake.repository,
        accessInvitationAction: 'expire',
        accessInvitationId: 'invite-one',
        onInvitationActionClose: () => {
          closed.push('expire');
        }
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Expire invitation');
    expect(document.activeElement?.textContent).toContain('Expire invitation');
    clickButton('Expire');
    await flush();
    expect(fake.calls).toContain('expire:tenant-one:inventory-one:invite-one:1970-01-01T00:00:00.000Z');

    await unmount(component);
    fake = fakeAccessRepository();
    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository: fake.repository,
        invitationStatus: 'pending',
        accessInvitationAction: 'cancel',
        accessInvitationId: 'invite-one',
        onInvitationActionClose: () => {
          closed.push('cancel');
        }
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Cancel invitation');
    expect(link('Cancel').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending'
    );
    clickButton('Cancel invitation');
    await flush();
    expect(fake.calls).toContain('cancel:tenant-one:inventory-one:invite-one');
    expect(document.body.textContent).not.toContain('friend@example.test');

    await unmount(component);
    fake = fakeAccessRepository();
    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository: fake.repository,
        accessInvitationAction: 'delete',
        accessInvitationId: 'invite-one',
        onInvitationActionClose: () => {
          closed.push('delete');
        }
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Delete invitation');
    clickButton('Delete');
    await flush();
    expect(fake.calls).toContain('delete-invitation:tenant-one:inventory-one:invite-one');
    expect(closed).toEqual(['expire', 'cancel', 'delete']);
  });

  it('shows an unavailable invitation action route when the target is not loaded', async () => {
    const { repository } = fakeAccessRepository();
    let closed = 0;

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository,
        accessInvitationAction: 'delete',
        accessInvitationId: 'missing-invite',
        onInvitationActionClose: () => {
          closed += 1;
        }
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Invitation unavailable');
    link('Back to invitations').click();
    await flush();
    expect(closed).toBe(1);
  });

  it('shows non-pending expire and cancel action routes as unavailable', async () => {
    const { repository } = fakeAccessRepository({ invitationStatus: 'accepted' });

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository,
        accessInvitationAction: 'cancel',
        accessInvitationId: 'invite-one'
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Invitation unavailable');
    expect(document.body.textContent).not.toContain('Cancel invitation');
  });

  it('supports invitation status filtering and paged load-more actions', async () => {
    const { repository, calls } = fakeAccessRepository({ hasMore: true, invitationStatus: 'revoked' });

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    clickButton('Load more grants');
    await flush();
    clickButton('Revoked');
    await flush();
    clickButton('Load more invitations');
    await flush();

    expect(calls).toEqual([
      'list-grants:tenant-one:inventory-one:',
      'list-invitations:tenant-one:inventory-one:all:',
      'list-grants:tenant-one:inventory-one:next-grants',
      'list-invitations:tenant-one:inventory-one:revoked:',
      'list-invitations:tenant-one:inventory-one:revoked:next-invitations'
    ]);
  });

  it('does not render stale access data after context changes', async () => {
    let resolveGrants: (value: InventoryAccessPage<InventoryAccessGrant>) => void = () => {};
    let resolveInvitations: (value: InventoryAccessPage<InventoryAccessInvitation>) => void = () => {};
    const repository: InventoryAccessRepository = {
      listInventoryAccessGrants: async () => new Promise((resolve) => { resolveGrants = resolve; }),
      listInventoryAccessInvitations: async () => new Promise((resolve) => { resolveInvitations = resolve; }),
      grantInventoryAccess: async () => failRepositoryCall(),
      revokeInventoryAccess: async () => failRepositoryCall(),
      createInventoryAccessInvitation: async () => failRepositoryCall(),
      updateInventoryAccessInvitationExpiration: async () => failRepositoryCall(),
      cancelInventoryAccessInvitation: async () => failRepositoryCall(),
      deleteInventoryAccessInvitation: async () => failRepositoryCall()
    };

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await tick();
    await unmount(component);
    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-two'), inventory: inventory('tenant-two', 'inventory-two', ['view']), repository }
    });
    await flush();

    resolveGrants(page([{ tenantId: 'tenant-one', inventoryId: 'inventory-one', principalId: 'stale-user', relationship: 'viewer' }]));
    resolveInvitations(page([invitation('invite-stale', 'stale@example.test')]));
    await flush();

    expect(document.body.textContent).not.toContain('stale-user');
    expect(document.body.textContent).toContain('you cannot manage sharing');
  });

  it('shows a denied state without loading access data when share access is missing', async () => {
    const { repository, calls } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view']), repository }
    });
    await flush();

    expect(calls).toEqual([]);
    expect(document.body.textContent).toContain('you cannot manage sharing');
  });

  it('uses accessible selected controls for relationship selection', async () => {
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    const grantField = segmentedGroup('Grant relationship');
    const invitationField = segmentedGroup('Invitation relationship');

    expect(grantField?.querySelectorAll('button[aria-pressed]')).toHaveLength(2);
    expect(grantField?.querySelector('button[aria-pressed="true"]')?.textContent).toBe('Viewer');
    expect(invitationField?.querySelectorAll('button[aria-pressed]')).toHaveLength(2);
  });

  it('uses the shared segmented control for invitation status filters', async () => {
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    const statusFilter = segmentedGroup('Invitation status');
    expect(statusFilter?.querySelectorAll('button[aria-pressed]')).toHaveLength(6);
    expect(statusFilter?.querySelector('button[aria-pressed="true"]')?.textContent).toBe('All');
  });

  it('exposes route-backed invitation status filter links when hrefs are provided', async () => {
    const { repository, calls } = fakeAccessRepository();
    let selectedStatus: InvitationStatusFilter | null = null;

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository,
        invitationStatus: 'pending',
        invitationStatusHref: (status) =>
          status === 'all'
            ? '/tenants/tenant-one/inventories/inventory-one/settings/access'
            : `/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=${status}`,
        onInvitationStatusChange: (status) => {
          selectedStatus = status;
        }
      }
    });
    await flush();

    const statusFilter = segmentedGroup('Invitation status');
    expect(statusFilter?.querySelectorAll('a[aria-current], a[data-selected]')).toHaveLength(6);
    expect(link('Pending').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=pending');
    expect(link('Pending').getAttribute('aria-current')).toBe('page');

    link('Revoked').click();
    await flush();

    expect(selectedStatus).toBe('revoked');
    expect(calls).toContain('list-invitations:tenant-one:inventory-one:revoked:');
  });
});

async function setInput(selector: string, value: string): Promise<void> {
  const input = document.querySelector<HTMLInputElement>(selector);
  if (!input) {
    throw new Error(`Missing input ${selector}`);
  }
  input.value = value;
  input.dispatchEvent(new Event('input', { bubbles: true }));
  await flush();
}

function clickButton(text: string): void {
  const button = Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent === text);
  if (!button) {
    throw new Error(`Missing button ${text}`);
  }
  button.click();
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent === text);
  if (!target) {
    throw new Error(`Missing link ${text}`);
  }
  return target;
}

function iconLink(label: string): HTMLAnchorElement {
  const target = document.body.querySelector<HTMLAnchorElement>(`a[aria-label="${label}"]`);
  if (!target) {
    throw new Error(`Missing icon link ${label}`);
  }
  return target;
}

function segmentedGroup(label: string): HTMLElement | null {
  return document.body.querySelector<HTMLElement>(`[role="group"][aria-label="${label}"]`);
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}

function tenant(id: string): Tenant {
  return {
    id,
    name: 'Home',
    access: { relationship: 'owner', permissions: ['view', 'configure'] }
  };
}

function inventory(tenantId: string, id: string, permissions: string[]): Inventory {
  return {
    id,
    tenantId,
    name: 'Household',
    access: { relationship: permissions.includes('share') ? 'owner' : 'viewer', permissions }
  };
}

function fakeAccessRepository(options: { hasMore?: boolean; invitationStatus?: InventoryAccessInvitation['status'] } = {}): {
  repository: InventoryAccessRepository;
  calls: string[];
} {
  const calls: string[] = [];
  let grants: InventoryAccessGrant[] = [
    { tenantId: 'tenant-one', inventoryId: 'inventory-one', principalId: 'principal-two', relationship: 'viewer' }
  ];
  let invitations: InventoryAccessInvitation[] = [invitation('invite-one', 'friend@example.test', 'viewer', options.invitationStatus)];
  return {
    calls,
    repository: {
      listInventoryAccessGrants: async (tenantId, inventoryId, cursor) => {
        calls.push(`list-grants:${tenantId}:${inventoryId}:${cursor ?? ''}`);
        return page(grants, options.hasMore && !cursor ? 'next-grants' : null);
      },
      grantInventoryAccess: async (tenantId, inventoryId, principalId, relationship) => {
        calls.push(`grant:${tenantId}:${inventoryId}:${principalId}:${relationship}`);
        const grant = { tenantId, inventoryId, principalId, relationship };
        grants = [grant, ...grants];
        return grant;
      },
      revokeInventoryAccess: async (tenantId, inventoryId, principalId, relationship) => {
        calls.push(`revoke:${tenantId}:${inventoryId}:${principalId}:${relationship}`);
      },
      listInventoryAccessInvitations: async (tenantId, inventoryId, status, cursor) => {
        calls.push(`list-invitations:${tenantId}:${inventoryId}:${status}:${cursor ?? ''}`);
        return page(
          invitations.filter(
            (candidate) =>
              candidate.tenantId === tenantId &&
              candidate.inventoryId === inventoryId &&
              invitationMatchesStatus(candidate, status)
          ),
          options.hasMore && !cursor ? 'next-invitations' : null
        );
      },
      createInventoryAccessInvitation: async (tenantId, inventoryId, email, relationship) => {
        calls.push(`invite:${tenantId}:${inventoryId}:${email}:${relationship}`);
        const created = invitation('invite-two', email, relationship);
        invitations = [created, ...invitations];
        return { invitation: created, acceptanceToken: 'raw-token' };
      },
      updateInventoryAccessInvitationExpiration: async (tenantId, inventoryId, invitationId, expiresAt) => {
        calls.push(`expire:${tenantId}:${inventoryId}:${invitationId}:${expiresAt}`);
        return { ...invitation(invitationId, 'friend@example.test'), expiresAt, isExpired: true };
      },
      cancelInventoryAccessInvitation: async (tenantId, inventoryId, invitationId) => {
        calls.push(`cancel:${tenantId}:${inventoryId}:${invitationId}`);
      },
      deleteInventoryAccessInvitation: async (tenantId, inventoryId, invitationId) => {
        calls.push(`delete-invitation:${tenantId}:${inventoryId}:${invitationId}`);
      }
    }
  };
}

function invitation(
  id: string,
  email: string,
  relationship: InventoryAccessRelationship = 'viewer',
  status: InventoryAccessInvitation['status'] = 'pending'
): InventoryAccessInvitation {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    email,
    relationship,
    status,
    isExpired: false,
    expiresAt: '2026-06-30T00:00:00Z',
    inviterPrincipalId: 'principal-one'
  };
}

function invitationMatchesStatus(invitation: InventoryAccessInvitation, status: InvitationStatusFilter): boolean {
  if (status === 'all') {
    return true;
  }
  if (status === 'expired') {
    return invitation.status === 'expired' || invitation.isExpired;
  }
  if (status === 'pending') {
    return invitation.status === 'pending' && !invitation.isExpired;
  }
  return invitation.status === status;
}

function page<T>(items: T[], nextCursor: string | null = null): InventoryAccessPage<T> {
  return { items, pagination: { limit: 50, nextCursor, hasMore: nextCursor !== null } };
}

function failRepositoryCall(): never {
  throw new Error('Unexpected repository call.');
}
