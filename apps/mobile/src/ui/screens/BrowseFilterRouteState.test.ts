import { expect, it } from 'vitest';
import { loadBrowseFilterTags, verifyBrowseFilterScope, browseFilterApplyParams } from './BrowseFilterRouteState';
const target = { tenantId: 'tenant', inventoryId: 'inventory', sessionScope: 'session' };
it('loads choices only for the requested inventory and session', async () => {
  let loaded = 0;
  expect(await loadBrowseFilterTags(target, 'session', async () => target, async () => { loaded++; return []; })).toEqual([]);
  for (const selected of [{ ...target, tenantId: 'other' }, { ...target, inventoryId: 'other' }]) {
    await expect(loadBrowseFilterTags(target, 'session', async () => selected, async () => { loaded++; return []; })).rejects.toThrow();
  }
  await expect(loadBrowseFilterTags(target, 'new-session', async () => target, async () => { loaded++; return []; })).rejects.toThrow();
  expect(loaded).toBe(1);
});
it('rejects a scope change during choice loading or before applying', async () => {
  let current = target;
  await expect(loadBrowseFilterTags(target, 'session', async () => current, async () => { current = { ...target, inventoryId: 'new' }; return []; })).rejects.toThrow();
  expect(() => verifyBrowseFilterScope(target, 'session', current)).toThrow();
});
it('writes explicit defaults so Reset clears previous route selections', () => {
  expect(browseFilterApplyParams({ scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [], sort: 'updated_desc' }, '')).toMatchObject({
    scope: 'all', tagId: [''], query: '', lifecycleState: 'active', checkoutState: 'any', sort: 'updated_desc', surface: 'list'
  });
});
