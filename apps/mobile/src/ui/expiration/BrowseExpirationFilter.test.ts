import { expect, it } from 'vitest';
import { browseExpirationFilter } from './BrowseExpirationFilter';
it('carries Browse text, all tags, kind and checkout into chronological expiration review', () => {
 expect(browseExpirationFilter('soon',{query:'drops',tagIds:['medicine','kids'],scope:'items',checkoutState:'available'})).toEqual({mode:'soon',query:'drops',tagIds:['medicine','kids'],kind:'item',checkoutState:'available'});
});
