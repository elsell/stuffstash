import { expect, it } from 'vitest';
import { VoiceInteractionPreviewQuery } from './VoiceInteractionPreviewQuery';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
it('reads only selected identity and names for the voice preview', async () => {
  let reads = 0;
  const query = new VoiceInteractionPreviewQuery({ getVoiceInventoryContext: async () => { reads++; return { tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), tenantName: 'Home', inventoryName: 'Garage' }; } });
  expect(await query.execute()).toMatchObject({ tenantName: 'Home', inventoryName: 'Garage' });
  expect(reads).toBe(1);
});
