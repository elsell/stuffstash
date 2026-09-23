import { expect, it } from 'vitest';
import { AddAssetScreen } from './AddAssetScreen';
import { AddAssetContextQuery } from '../../application/add/AddAssetContextQuery';
import { AddDraftScopeQuery } from '../../application/add/AddDraftScopeQuery';
import { InMemoryAddAssetDraftStore } from '../../application/add/AddAssetDraftStore';
import { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';

it('keeps the chosen parent when searching and leaving without a new selection', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const store = new InMemoryAddAssetDraftStore('scope');
  const scope = { tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' };
  const context = { ...scope, tenantName: 'Home', inventoryName: 'Inventory', canAdd: true, assetTags: [] };
  const garage = { id: 'garage', title: 'Garage', kind: 'location' as const, subtitle: 'Garage', pathLabel: 'Garage', selectionHint: 'Location', willPromoteToContainer: false };
  store.save(scope, { title: 'Tent', description: 'Packed', parentAssetId: garage.id, parentQuery: garage.title, lastParent: garage, selectedPhotos: [], showDetails: false });
  const submissions: unknown[] = [];
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
      inventoryAssetTypesQuery={{ execute: async () => [] }} addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
      addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={store}
      createAssetCommand={{ execute: async input => { submissions.push(input); throw new Error('Retain for retry'); } }}
      parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })}
      photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} /></AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    const chooser = h.byText('Garage')?.parent?.parent ?? undefined;
    await h.press(chooser);
    await h.changeText(h.byLabel('Search parent'), 'Another place');
    expect(store.load(scope)).toMatchObject({ parentAssetId: 'garage', parentQuery: 'Garage', title: 'Tent', description: 'Packed' });
    await h.press(chooser);
    await h.press(h.byLabel('Save item'));
    expect(submissions).toEqual([expect.objectContaining({ parentAssetId: 'garage', title: 'Tent', description: 'Packed' })]);
  } finally { await h.unmount(); client.clear(); }
});
