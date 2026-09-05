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
