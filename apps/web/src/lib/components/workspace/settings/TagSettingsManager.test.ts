import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, unmount } from 'svelte';
import type { Inventory, ManagedAssetTag } from '$lib/domain/inventory';
import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
import type { InventoryTagRepository } from '$lib/ports/inventoryTagRepository';
import TagSettingsManager from './TagSettingsManager.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('TagSettingsManager', () => {
  it('renders compact rows with only the tag name visible and color exposed accessibly', async () => {
    component = mount(TagSettingsManager, {
      target: document.body,
      props: {
        inventory,
        repository: tagRepository([coloredTag, uncoloredTag]),
        observer,
        onNavigate: () => {},
        onTagsChange: () => {},
        onPermissionDenied: async () => {}
      }
    });

    await vi.waitFor(() => expect(document.body.querySelectorAll('.settings-tag-resource-row')).toHaveLength(2));

    const rows = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('.settings-tag-resource-row'));
    expect(rows.map((row) => row.textContent?.trim())).toEqual(['Reference', 'Workshop']);
    const referenceRow = rows.find((row) => row.textContent?.trim() === 'Reference');
    const workshopRow = rows.find((row) => row.textContent?.trim() === 'Workshop');
    expect(workshopRow?.getAttribute('aria-label')).toBe('Workshop Blue color (#2F80ED)');
    expect(referenceRow?.getAttribute('aria-label')).toBe('Reference No color');
    expect(referenceRow?.querySelector('.settings-tag-color-indicator')?.classList.contains('settings-tag-color-empty')).toBe(true);
  });
});

const inventory: Inventory = {
  id: 'inventory-household',
  tenantId: 'tenant-home',
  name: 'Household',
  access: { relationship: 'editor', permissions: ['view', 'edit_asset'] }
};

const coloredTag: ManagedAssetTag = {
  id: 'tag-workshop',
  tenantId: inventory.tenantId,
  inventoryId: inventory.id,
  key: 'workshop',
  displayName: 'Workshop',
  color: '#2F80ED',
  lifecycleState: 'active',
  createdAt: '2026-07-16T12:00:00Z',
  updatedAt: '2026-07-16T12:00:00Z'
};

const uncoloredTag: ManagedAssetTag = {
  ...coloredTag,
  id: 'tag-reference',
  key: 'reference',
  displayName: 'Reference',
  color: undefined
};

const observer: WorkspaceObserver = { record: () => {} };

function tagRepository(items: ManagedAssetTag[]): InventoryTagRepository {
  return {
    listManagedAssetTags: async () => ({ items, pagination: { limit: 50, hasMore: false, nextCursor: null } }),
    createManagedAssetTag: async () => coloredTag,
    updateManagedAssetTag: async () => coloredTag,
    archiveManagedAssetTag: async () => coloredTag
  };
}
