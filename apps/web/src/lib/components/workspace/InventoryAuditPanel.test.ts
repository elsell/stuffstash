import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import InventoryAuditPanel from './InventoryAuditPanel.svelte';
import type { AuditRecord, AuditScope, Inventory, Tenant } from '$lib/domain/inventory';
import type { AuditRecordPage, InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryAuditPanel', () => {
  it('loads inventory audit records by default and supports pagination', async () => {
    const { repository, calls } = fakeAuditRepository({ hasMore: true });

    component = mount(InventoryAuditPanel, {
      target: document.body,
      props: { tenant: tenant(['view', 'configure']), inventory: inventory(['view']), repository }
    });
    await flush();
    clickButton('Load more history');
    await flush();

    expect(calls).toEqual([
      'inventory:tenant-one:inventory-one:',
      'inventory:tenant-one:inventory-one:next-page'
    ]);
    expect(document.body.textContent).toContain('asset.created');
  });

  it('switches to tenant audit only when tenant configure access exists', async () => {
    const { repository, calls } = fakeAuditRepository();

    component = mount(InventoryAuditPanel, {
      target: document.body,
      props: { tenant: tenant(['view', 'configure']), inventory: inventory(['view']), repository }
    });
    await flush();
    clickButton('Tenant');
    await flush();

    expect(calls).toEqual(['inventory:tenant-one:inventory-one:', 'tenant:tenant-one:']);
  });

  it('uses the shared segmented control for audit scope', async () => {
    const { repository } = fakeAuditRepository();

    component = mount(InventoryAuditPanel, {
      target: document.body,
      props: { tenant: tenant(['view', 'configure']), inventory: inventory(['view']), repository }
    });
    await flush();

    const scopeFilter = document.body.querySelector<HTMLElement>('[role="group"][aria-label="Audit scope"]');
    expect(scopeFilter?.querySelectorAll('button[aria-pressed]')).toHaveLength(2);
    expect(scopeFilter?.querySelector('button[aria-pressed="true"]')?.textContent).toBe('Inventory');
  });

  it('exposes route-backed audit scope filter links when hrefs are provided', async () => {
    const { repository, calls } = fakeAuditRepository();
    let selectedScope: AuditScope | null = null;

    component = mount(InventoryAuditPanel, {
      target: document.body,
      props: {
        tenant: tenant(['view', 'configure']),
        inventory: inventory(['view']),
        repository,
        scope: 'tenant',
        scopeHref: (scope) =>
          scope === 'inventory'
            ? '/tenants/tenant-one/inventories/inventory-one/settings/activity'
            : '/tenants/tenant-one/inventories/inventory-one/settings/activity?auditScope=tenant',
        onScopeChange: (scope) => {
          selectedScope = scope;
        }
      }
    });
    await flush();

    const scopeFilter = document.body.querySelector<HTMLElement>('[role="group"][aria-label="Audit scope"]');
    expect(scopeFilter?.querySelectorAll('a[aria-current], a[data-selected]')).toHaveLength(2);
    expect(link('Tenant').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/activity?auditScope=tenant');
    expect(link('Tenant').getAttribute('aria-current')).toBe('page');

    link('Inventory').click();
    await flush();

    expect(selectedScope).toBe('inventory');
    expect(calls).toContain('inventory:tenant-one:inventory-one:');
  });

  it('disables unavailable audit scopes through the shared segmented control', async () => {
    const { repository } = fakeAuditRepository();

    component = mount(InventoryAuditPanel, {
      target: document.body,
      props: { tenant: tenant(['view', 'configure']), inventory: null, repository }
    });
    await flush();

    const scopeFilter = document.body.querySelector<HTMLElement>('[role="group"][aria-label="Audit scope"]');
    expect(buttonIn(scopeFilter, 'Inventory')?.disabled).toBe(true);
    expect(buttonIn(scopeFilter, 'Tenant')?.disabled).toBe(false);
  });

  it('shows an authorization-aware denied state for tenant audit', async () => {
    const { repository } = fakeAuditRepository();

    component = mount(InventoryAuditPanel, {
      target: document.body,
      props: { tenant: tenant(['view']), inventory: inventory(['view']), repository }
    });
    await flush();
    clickButton('Tenant');
    await flush();

    expect(document.body.textContent).toContain('Tenant audit history requires tenant configuration access.');
  });

  it('aborts stale audit reads when the selected scope changes', async () => {
    let inventoryAborted = false;
    let tenantLoads = 0;
    const repository: InventoryAuditRepository = {
      listTenantAuditRecords: async () => {
        tenantLoads += 1;
        return page(null);
      },
      listInventoryAuditRecords: async (_tenantId, _inventoryId, _cursor, signal) =>
        new Promise<AuditRecordPage>((_resolve, reject) => {
          signal?.addEventListener('abort', () => {
            inventoryAborted = true;
            const error = new Error('Aborted');
            error.name = 'AbortError';
            reject(error);
          });
        })
    };

    component = mount(InventoryAuditPanel, {
      target: document.body,
      props: { tenant: tenant(['view', 'configure']), inventory: inventory(['view']), repository }
    });
    await tick();

    clickButton('Tenant');
    await flush();

    expect(inventoryAborted).toBe(true);
    expect(tenantLoads).toBe(1);
    expect(document.body.textContent).toContain('asset.created');
  });

  it('aborts pending pagination reads when the panel unmounts', async () => {
    let paginationAborted = false;
    const repository: InventoryAuditRepository = {
      listTenantAuditRecords: async () => page(null),
      listInventoryAuditRecords: async (_tenantId, _inventoryId, cursor, signal) => {
        if (!cursor) {
          return page('next-page');
        }
        return new Promise<AuditRecordPage>((_resolve, reject) => {
          signal?.addEventListener('abort', () => {
            paginationAborted = true;
            const error = new Error('Aborted');
            error.name = 'AbortError';
            reject(error);
          });
        });
      }
    };

    component = mount(InventoryAuditPanel, {
      target: document.body,
      props: { tenant: tenant(['view', 'configure']), inventory: inventory(['view']), repository }
    });
    await flush();

    clickButton('Load more history');
    await tick();
    unmount(component);
    component = null;
    await flush();

    expect(paginationAborted).toBe(true);
  });
});

function clickButton(text: string): void {
  const button = Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent === text);
  if (!button) {
    throw new Error(`Missing button ${text}`);
  }
  button.click();
}

function buttonIn(root: HTMLElement | null, text: string): HTMLButtonElement | null {
  return Array.from(root?.querySelectorAll<HTMLButtonElement>('button') ?? []).find((button) => button.textContent === text) ?? null;
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent === text);
  if (!target) {
    throw new Error(`Missing link ${text}`);
  }
  return target;
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}

function tenant(permissions: string[]): Tenant {
  return {
    id: 'tenant-one',
    name: 'Home',
    access: { relationship: permissions.includes('configure') ? 'owner' : 'viewer', permissions }
  };
}

function inventory(permissions: string[]): Inventory {
  return {
    id: 'inventory-one',
    tenantId: 'tenant-one',
    name: 'Household',
    access: { relationship: permissions.includes('view') ? 'viewer' : 'none', permissions }
  };
}

function fakeAuditRepository(options: { hasMore?: boolean } = {}): { repository: InventoryAuditRepository; calls: string[] } {
  const calls: string[] = [];
  return {
    calls,
    repository: {
      listTenantAuditRecords: async (tenantId, cursor) => {
        calls.push(`tenant:${tenantId}:${cursor ?? ''}`);
        return page(options.hasMore && !cursor ? 'next-page' : null);
      },
      listInventoryAuditRecords: async (tenantId, inventoryId, cursor) => {
        calls.push(`inventory:${tenantId}:${inventoryId}:${cursor ?? ''}`);
        return page(options.hasMore && !cursor ? 'next-page' : null);
      }
    }
  };
}

function page(nextCursor: string | null): AuditRecordPage {
  return {
    items: [
      {
        id: 'audit-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        principalId: 'principal-one',
        action: 'asset.created',
        source: 'api',
        targetType: 'asset',
        targetId: 'asset-one',
        occurredAt: '2026-06-24T12:00:00Z',
        requestId: 'request-one',
        metadata: { operation_id: 'operation-one' }
      } satisfies AuditRecord
    ],
    pagination: { limit: 50, nextCursor, hasMore: nextCursor !== null }
  };
}
