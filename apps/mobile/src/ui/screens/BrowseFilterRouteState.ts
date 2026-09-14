import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type { BrowseFilterDraft } from './BrowseFiltersScreen';
export type BrowseFilterTarget = { readonly tenantId: string; readonly inventoryId: string; readonly sessionScope: string };
type InventoryScope = { readonly tenantId: string; readonly inventoryId: string };
export function verifyBrowseFilterScope(target: BrowseFilterTarget, sessionScope: string, current: InventoryScope) {
  if (!target.tenantId || !target.inventoryId || !target.sessionScope || target.sessionScope !== sessionScope ||
      target.tenantId !== current.tenantId || target.inventoryId !== current.inventoryId) {
    throw new Error('Inventory changed. Reopen Browse filters.');
  }
}
export async function loadBrowseFilterTags(target: BrowseFilterTarget, sessionScope: string, loadScope: () => Promise<InventoryScope>, loadTags: () => Promise<readonly AssetTagOptionViewModel[]>) {
  verifyBrowseFilterScope(target, sessionScope, await loadScope());
  const tags = await loadTags();
  verifyBrowseFilterScope(target, sessionScope, await loadScope());
  return tags;
}
export function browseFilterApplyParams(draft: BrowseFilterDraft, query: string) {
  return { surface: 'list', scope: draft.scope, lifecycleState: draft.lifecycleState, checkoutState: draft.checkoutState,
    sort: draft.sort, query: query.trim(), tagId: draft.tagIds.length ? [...draft.tagIds] : [''] };
}
