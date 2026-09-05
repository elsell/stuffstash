import { expect, it } from 'vitest';
import { createMobileQueryClient, mobileQueryKeys } from './MobileQueryClient';
import { QueryClientCustomizationMutationObserver } from './QueryClientCustomizationMutationObserver';

it('reconciles inherited household definitions across its inventories and preserves unrelated tenants', () => {
  const client = createMobileQueryClient();
  const keys = [
    mobileQueryKeys.customization('scope', 'home', 'one', 'tenant', 'asset-type', 'active'),
    mobileQueryKeys.customization('scope', 'home', 'one', 'inventory', 'asset-type', 'active'),
    mobileQueryKeys.customization('scope', 'home', 'two', 'inventory', 'field', 'active'),
    mobileQueryKeys.assetCore('scope', 'home', 'two', 'item'),
    mobileQueryKeys.customization('scope', 'other', 'one', 'inventory', 'asset-type', 'active'),
    mobileQueryKeys.assetPhotos('scope', 'home', 'two', 'item')
  ];
  for (const key of keys) client.setQueryData(key, 'cached');
  new QueryClientCustomizationMutationObserver(client, 'scope').onCustomizationChanged({ tenantId: 'home', kind: 'asset-type' });
  expect(keys.map((key) => client.getQueryState(key)?.isInvalidated)).toEqual([true, true, true, true, false, false]);
  client.clear();
});
