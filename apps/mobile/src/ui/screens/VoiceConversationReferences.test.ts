import { expect, it } from 'vitest';
import { voiceConversationReferences } from './VoiceConversationReferences';

it('exposes grounded item, container and location references without inventing draft IDs', () => {
  const result = voiceConversationReferences({ status: 'review', tenantName: 'Home', inventoryName: 'Home', debugEvents: [],
    responseArtifacts: [{ type: 'asset_reference', assetId: 'drill', title: 'Drill', assetKind: 'item' }],
    actionPlan: { planId: 'plan', status: 'proposed', confirmationSummary: 'Add drill', risks: [], commands: [
      { id: 'new-drill', kind: 'create_asset', summary: 'New drill', title: 'Drill', parentAssetId: 'box', parentTitle: 'Toolbox', parentKind: 'container' },
      { id: 'new-box', kind: 'create_asset', summary: 'New box', parentAssetId: 'office', parentTitle: 'Office', parentKind: 'location' },
      { id: 'unresolved', kind: 'create_asset', summary: 'New thing', parentCommandId: 'new-box' }
    ] } });
  expect(result.map(reference => [reference.assetId, reference.assetKind])).toEqual([['drill', 'item'], ['box', 'container'], ['office', 'location']]);
});
