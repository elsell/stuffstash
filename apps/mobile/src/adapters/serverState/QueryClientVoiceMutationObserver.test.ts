import { expect, it } from 'vitest';
import { createMobileQueryClient, mobileQueryKeys as keys } from './MobileQueryClient';
import { QueryClientVoiceMutationObserver } from './QueryClientVoiceMutationObserver';
it('reconciles executed voice changes without refreshing unrelated inventory, photos or settings', () => {
  const client = createMobileQueryClient();
  const affected = [keys.home('scope', 'tenant', 'inventory'), keys.assetCore('scope', 'tenant', 'inventory', 'promoted-parent'), keys.assetCore('scope', 'tenant', 'inventory', 'changed'), keys.assetContents('scope', 'tenant', 'inventory', 'parent', 'root')];
  const fresh = [keys.assetHistory('scope', 'tenant', 'inventory', 'other', 'all'), keys.assetPhotos('scope', 'tenant', 'inventory', 'changed'), keys.home('scope', 'other', 'inventory'), keys.voiceConfiguration('scope', 'tenant')];
  for (const key of [...affected, ...fresh]) client.setQueryData(key, {});
  new QueryClientVoiceMutationObserver(client, 'scope').onVoicePlanExecuted({ tenantId: 'tenant', inventoryId: 'inventory', assetIds: ['changed'] });
  for (const key of affected) expect(client.getQueryState(key)?.isInvalidated).toBe(true);
  for (const key of fresh) expect(client.getQueryState(key)?.isInvalidated).toBe(false);
  client.clear();
});

it('refreshes effective configuration after an approved schema creation', () => {
 const client = createMobileQueryClient();
 const fields = keys.customization('scope','tenant','inventory','inventory','field','active');
 const types = keys.customization('scope','tenant','inventory','inventory','asset-type-choices','active');
 const other = keys.customization('scope','tenant','other','inventory','field','active');
 for (const key of [fields,types,other]) client.setQueryData(key,{});
 new QueryClientVoiceMutationObserver(client,'scope').onVoicePlanExecuted({tenantId:'tenant',inventoryId:'inventory',assetIds:[],configurationChanged:true});
 expect(client.getQueryState(fields)?.isInvalidated).toBe(true);
 expect(client.getQueryState(types)?.isInvalidated).toBe(true);
 expect(client.getQueryState(other)?.isInvalidated).toBe(false);
 client.clear();
});
