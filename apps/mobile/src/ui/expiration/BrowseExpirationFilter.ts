import type { ExpirationFilter, ExpirationMode } from '../../application/expiration/ExpirationRepository';
import type { BrowseScope } from '../screens/SearchScreenPresentation';
import type { AssetBrowseCheckoutFilter } from '../../application/home/InventorySummaryRepository';
export function browseExpirationFilter(mode: ExpirationMode, browse: { query: string; tagIds: readonly string[]; scope: BrowseScope; checkoutState: AssetBrowseCheckoutFilter }): ExpirationFilter {
 return {mode,query:browse.query,tagIds:browse.tagIds,kind:browse.scope==='items'?'item':browse.scope==='containers'?'container':browse.scope==='places'?'location':undefined,checkoutState:browse.checkoutState};
}
