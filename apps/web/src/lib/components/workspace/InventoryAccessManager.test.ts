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

  it('revokes grants and expires, cancels, and deletes invitations', async () => {
    const { repository, calls } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    clickButton('Revoke');
    await flush();
    clickButton('Expire');
    await flush();
    clickButton('Cancel');
    await flush();
    clickIconButton('Delete invitation for friend@example.test');
    await flush();

    expect(calls).toContain('revoke:tenant-one:inventory-one:principal-two:viewer');
    expect(calls).toContain('expire:tenant-one:inventory-one:invite-one:1970-01-01T00:00:00.000Z');
    expect(calls).toContain('cancel:tenant-one:inventory-one:invite-one');
    expect(calls).toContain('delete-invitation:tenant-one:inventory-one:invite-one');
  });

  it('supports invitation status filtering and paged load-more actions', async () => {
    const { repository, calls } = fakeAccessRepository({ hasMore: true });

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

function clickIconButton(label: string): void {
  const button = document.body.querySelector<HTMLButtonElement>(`button[aria-label="${label}"]`);
  if (!button) {
    throw new Error(`Missing icon button ${label}`);
  }
  button.click();
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

function fakeAccessRepository(options: { hasMore?: boolean } = {}): { repository: InventoryAccessRepository; calls: string[] } {
  const calls: string[] = [];
  let grants: InventoryAccessGrant[] = [
    { tenantId: 'tenant-one', inventoryId: 'inventory-one', principalId: 'principal-two', relationship: 'viewer' }
  ];
  let invitations: InventoryAccessInvitation[] = [invitation('invite-one', 'friend@example.test')];
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
        return page(invitations, options.hasMore && !cursor ? 'next-invitations' : null);
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
  relationship: InventoryAccessRelationship = 'viewer'
): InventoryAccessInvitation {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    email,
    relationship,
    status: 'pending',
    isExpired: false,
    expiresAt: '2026-06-30T00:00:00Z',
    inviterPrincipalId: 'principal-one'
  };
}

function page<T>(items: T[], nextCursor: string | null = null): InventoryAccessPage<T> {
  return { items, pagination: { limit: 50, nextCursor, hasMore: nextCursor !== null } };
}

function failRepositoryCall(): never {
  throw new Error('Unexpected repository call.');
}
