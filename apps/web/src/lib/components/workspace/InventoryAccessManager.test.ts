import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import InventoryAccessManager from './InventoryAccessManager.svelte';
import InventoryAccessManagerTestHarness from './InventoryAccessManager.test-harness.svelte';
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
const originalClipboardDescriptor = Object.getOwnPropertyDescriptor(Navigator.prototype, 'clipboard');
const originalShareDescriptor = Object.getOwnPropertyDescriptor(Navigator.prototype, 'share');
const inviteUrl = 'https://stash.example.test/invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-two#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

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
  restoreClipboard();
  restoreShare();
  vi.restoreAllMocks();
});

describe('InventoryAccessManager', () => {
  it('renders missing inventory access status without an alert role', async () => {
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: null, repository }
    });
    await flush();

    expect(document.body.textContent).toContain('Select an inventory before managing sharing.');
    expect(document.body.querySelector('[role="alert"]')).toBeNull();
  });

  it('renders denied access status as an alert', async () => {
    const { repository, calls } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view']), repository }
    });
    await flush();

    expect(calls).toEqual([]);
    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain(
      'You can view this inventory, but you cannot manage sharing.'
    );
  });

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

  it('leads with email invitations and keeps account-id grants advanced', async () => {
    const { repository } = fakeAccessRepository();
    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    const inviteForm = requiredElement('.invitation-form');
    const advanced = requiredElement('.advanced-access');
    expect(inviteForm.textContent).toContain('Email address');
    expect(inviteForm.textContent).toContain('Access level');
    expect(inviteForm.querySelector('[role="group"][aria-label="Invitation access level"]')).not.toBeNull();
    expect(advanced.querySelector('summary')?.textContent).toContain('Advanced account grants');
    expect(advanced.textContent).toContain('Account ID');
    expect(advanced.textContent).toContain('identity provider');
    expect(advanced.querySelector('[role="group"][aria-label="Direct grant access level"]')).not.toBeNull();
    expect((advanced as HTMLDetailsElement).open).toBe(false);
    expect(requiredElement('[aria-label="Direct grants"]').closest('details')).toBe(advanced);
    expect(advanced.contains(requiredElement('[aria-label="Invitations"]'))).toBe(false);
    expect(
      Boolean(inviteForm.compareDocumentPosition(advanced) & Node.DOCUMENT_POSITION_FOLLOWING)
    ).toBe(true);
    expect(document.body.textContent).not.toContain('Principal ID');
  });

  it('renders invitation rows with separated metadata, status, and actions', async () => {
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    const row = requiredElement('.invitation-row');
    expect(row.querySelector('.access-row-main')?.textContent).toContain('friend@example.test');
    expect(row.querySelector('.access-row-meta')?.textContent).toContain('viewer');
    expect(row.querySelector('.access-row-meta')?.textContent).toContain('pending');
    expect(row.querySelector('.access-row-status')?.textContent).toContain('pending');
    expect(row.querySelector('.access-actions')?.textContent).toContain('Expire');
    expect(row.querySelector('.access-actions')?.textContent).toContain('Cancel');
  });

  it('announces initial access list loading states', async () => {
    const repository: InventoryAccessRepository = {
      listInventoryAccessGrants: async () => new Promise(() => {}),
      listInventoryAccessInvitations: async () => new Promise(() => {}),
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

    expect(document.body.querySelector('[aria-label="Direct grants"] [role="status"]')?.textContent).toBe('Loading grants...');
    expect(document.body.querySelector('[aria-label="Invitations"] [role="status"]')?.textContent).toBe('Loading invitations...');
  });

  it('does not present failed initial access loads as empty lists', async () => {
    const repository: InventoryAccessRepository = {
      listInventoryAccessGrants: async () => {
        throw new Error('Access service unavailable.');
      },
      listInventoryAccessInvitations: async () => page([]),
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
    await flush();

    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain('Access service unavailable.');
    expect(document.body.textContent).not.toContain('No direct grants.');
    expect(document.body.textContent).not.toContain('No invitations.');
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
    expect(document.body.textContent).toContain('token=');
  });

  it('keeps the one-time invitation link out of persistent invitation rows', async () => {
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    await setInput('#invite-email', 'new@example.test');
    clickButton('Create invite');
    await flush();

    expect(document.body.querySelector('.one-time-token')?.textContent).toContain('Copy link');
    expect(document.body.querySelector('.one-time-token')?.getAttribute('aria-label')).toBe('One-time invitation link');
    expect(document.body.querySelector('.one-time-token')?.getAttribute('role')).toBeNull();
    expect(document.body.querySelector('[aria-label="Invitations"]')?.textContent).not.toContain('token=');
  });

  it('copies the complete one-time invitation link when clipboard access is available', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    stubClipboard(writeText);
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    await setInput('#invite-email', 'new@example.test');
    clickButton('Create invite');
    await flush();
    clickButton('Copy link');
    await flush();

    expect(writeText).toHaveBeenCalledWith(inviteUrl);
    expect(document.body.textContent).toContain('Invitation link copied.');
  });

  it('shows a manual-copy error when clipboard copy fails', async () => {
    stubClipboard(vi.fn().mockRejectedValue(new Error('denied')));
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    await setInput('#invite-email', 'new@example.test');
    clickButton('Create invite');
    await flush();
    clickButton('Copy link');
    await flush();

    expect(document.body.textContent).toContain('Invitation link not copied. Select the link and copy it manually.');
  });

  it('clears stale copied status when a later copy attempt fails', async () => {
    const writeText = vi.fn().mockResolvedValueOnce(undefined).mockRejectedValueOnce(new Error('denied'));
    stubClipboard(writeText);
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    await setInput('#invite-email', 'new@example.test');
    clickButton('Create invite');
    await flush();
    clickButton('Copy link');
    await flush();
    expect(document.body.textContent).toContain('Invitation link copied.');

    clickButton('Copy link');
    await flush();

    expect(document.body.textContent).not.toContain('Invitation link copied.');
    expect(document.body.textContent).toContain('Invitation link not copied. Select the link and copy it manually.');
  });

  it('shows a manual-copy error when clipboard access is unavailable', async () => {
    stubClipboard(undefined);
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    await setInput('#invite-email', 'new@example.test');
    clickButton('Create invite');
    await flush();
    clickButton('Copy link');
    await flush();

    expect(document.body.textContent).toContain('Invitation link not copied. Select the link and copy it manually.');
  });

  it('shares the complete one-time invitation link through the Web Share API', async () => {
    const share = vi.fn().mockResolvedValue(undefined);
    stubShare(share);
    const { repository } = fakeAccessRepository();
    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();
    await setInput('#invite-email', 'new@example.test');
    clickButton('Create invite');
    await flush();
    clickButton('Share invitation');
    await flush();
    expect(share).toHaveBeenCalledWith(expect.objectContaining({ url: inviteUrl }));
    expect(document.body.textContent).toContain('Invitation shared.');
  });

  it('clears a prior one-time link before a later creation attempt fails', async () => {
    const { repository } = fakeAccessRepository({ failInvitationCreationAfter: 1 });
    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();
    await setInput('#invite-email', 'first@example.test');
    clickButton('Create invite');
    await flush();
    expect(document.body.textContent).toContain('token=');
    await setInput('#invite-email', 'second@example.test');
    clickButton('Create invite');
    await flush();
    expect(document.body.textContent).not.toContain('token=');
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

    expect(calls).not.toContain('revoke:tenant-one:inventory-one:principal-two:viewer');
    expect(document.body.querySelector('[role="alertdialog"]')?.textContent).toContain('Revoke access');
    clickButton('Revoke access');
    await flush();

    expect(calls).toContain('revoke:tenant-one:inventory-one:principal-two:viewer');
  });

  it('clears a pending revoke when the inventory context changes', async () => {
    const { repository, calls } = fakeAccessRepository();

    component = mount(InventoryAccessManagerTestHarness, {
      target: document.body,
      props: {
        initialTenant: tenant('tenant-one'),
        initialInventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository
      }
    });
    await flush();

    clickButton('Revoke');
    await flush();
    expect(document.body.querySelector('[role="alertdialog"]')).not.toBeNull();

    (component as unknown as { setContext: (tenant: Tenant, inventory: Inventory) => void }).setContext(
      tenant('tenant-two'),
      inventory('tenant-two', 'inventory-two', ['view', 'share'])
    );
    await flush();

    expect(document.body.querySelector('[role="alertdialog"]')).toBeNull();
    expect(calls.some((call) => call.startsWith('revoke:'))).toBe(false);
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
    expect(document.activeElement?.textContent).toBe('Cancel');
    clickButton('Expire');
    await flush();
    await waitForDialogClose(() => closed.includes('expire'));
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
    await waitForDialogClose(() => closed.includes('cancel'));
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
    await waitForDialogClose(() => closed.includes('delete'));
    expect(fake.calls).toContain('delete-invitation:tenant-one:inventory-one:invite-one');
    expect(closed).toEqual(['expire', 'cancel', 'delete']);
  });

  it('moves focus to the Sharing heading when a deep-linked invitation confirmation closes', async () => {
    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository: fakeAccessRepository().repository,
        accessInvitationAction: 'expire',
        accessInvitationId: 'invite-one'
      }
    });
    await flush();

    requiredElement('[role="alertdialog"]').dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await waitForDialogClose(() => document.activeElement?.id === 'settings-access');

    expect(document.activeElement).toBe(requiredElement('#settings-access'));
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
    await waitForDialogClose(() => closed === 1);
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

  it('disables expire and cancel actions for pending invitations that are already expired', async () => {
    const { repository } = fakeAccessRepository({ expired: true });

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository,
        invitationStatus: 'expired'
      }
    });
    await flush();

    expect(link('Expire').getAttribute('aria-disabled')).toBe('true');
    expect(link('Cancel').getAttribute('aria-disabled')).toBe('true');
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

  it('uses accessible selected controls for access-level selection', async () => {
    const { repository } = fakeAccessRepository();

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: { tenant: tenant('tenant-one'), inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']), repository }
    });
    await flush();

    const grantField = segmentedGroup('Direct grant access level');
    const invitationField = segmentedGroup('Invitation access level');

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
    expect(statusFilter?.querySelectorAll('a[aria-current], a[data-selected]')).toHaveLength(6);
    expect(link('All').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/access');
    expect(link('All').getAttribute('aria-current')).toBe('page');
  });

  it('exposes route-backed invitation status filter links', async () => {
    const { repository, calls } = fakeAccessRepository();
    let selectedStatus: InvitationStatusFilter | null = null;

    component = mount(InventoryAccessManager, {
      target: document.body,
      props: {
        tenant: tenant('tenant-one'),
        inventory: inventory('tenant-one', 'inventory-one', ['view', 'share']),
        repository,
        invitationStatus: 'pending',
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
  const dialog = document.body.querySelector<HTMLElement>('[role="alertdialog"]');
  const control = Array.from((dialog ?? document.body).querySelectorAll<HTMLElement>('button, a')).find(
    (candidate) => candidate.textContent === text
  );
  if (!control) {
    throw new Error(`Missing control ${text}`);
  }
  control.click();
}

function link(text: string): HTMLAnchorElement {
  const dialog = document.body.querySelector<HTMLElement>('[role="alertdialog"]');
  const target = Array.from((dialog ?? document.body).querySelectorAll<HTMLAnchorElement>('a')).find(
    (candidate) => candidate.textContent === text
  );
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

function stubClipboard(writeText: ((text: string) => Promise<void>) | undefined): void {
  Object.defineProperty(Navigator.prototype, 'clipboard', {
    configurable: true,
    get: () => (writeText ? { writeText } : undefined)
  });
}

function restoreClipboard(): void {
  if (originalClipboardDescriptor) {
    Object.defineProperty(Navigator.prototype, 'clipboard', originalClipboardDescriptor);
    return;
  }
  Reflect.deleteProperty(Navigator.prototype, 'clipboard');
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}

async function waitForDialogClose(condition: () => boolean): Promise<void> {
  for (let attempt = 0; attempt < 100; attempt += 1) {
    await flush();
    if (condition() && !document.body.querySelector('[role="alertdialog"]')) {
      return;
    }
    await new Promise((resolve) => window.setTimeout(resolve, 5));
  }
  throw new Error('Dialog did not finish closing.');
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

function fakeAccessRepository(options: { hasMore?: boolean; invitationStatus?: InventoryAccessInvitation['status']; expired?: boolean; failInvitationCreationAfter?: number } = {}): {
  repository: InventoryAccessRepository;
  calls: string[];
} {
  const calls: string[] = [];
  let grants: InventoryAccessGrant[] = [
    { tenantId: 'tenant-one', inventoryId: 'inventory-one', principalId: 'principal-two', relationship: 'viewer' }
  ];
  let invitations: InventoryAccessInvitation[] = [
    {
      ...invitation('invite-one', 'friend@example.test', 'viewer', options.invitationStatus),
      isExpired: options.expired ?? false
    }
  ];
  let invitationCreations = 0;
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
        if (options.failInvitationCreationAfter !== undefined && invitationCreations >= options.failInvitationCreationAfter) {
          throw new Error('creation failed');
        }
        invitationCreations += 1;
        calls.push(`invite:${tenantId}:${inventoryId}:${email}:${relationship}`);
        const created = invitation('invite-two', email, relationship);
        invitations = [created, ...invitations];
        return { invitation: created, inviteUrl };
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

function stubShare(share: typeof navigator.share | undefined): void {
  Object.defineProperty(Navigator.prototype, 'share', { configurable: true, value: share });
}

function restoreShare(): void {
  if (originalShareDescriptor) Object.defineProperty(Navigator.prototype, 'share', originalShareDescriptor);
  else delete (Navigator.prototype as unknown as { share?: typeof navigator.share }).share;
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

function requiredElement(selector: string): HTMLElement {
  const element = document.body.querySelector<HTMLElement>(selector);
  if (!element) {
    throw new Error(`Missing element ${selector}`);
  }
  return element;
}

function failRepositoryCall(): never {
  throw new Error('Unexpected repository call.');
}
