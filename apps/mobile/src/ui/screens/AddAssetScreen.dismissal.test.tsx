import React from 'react';
import { expect, it } from 'vitest';
import { AddAssetScreen } from './AddAssetScreen';
import { AddAssetContextQuery } from '../../application/add/AddAssetContextQuery';
import { AddDraftScopeQuery } from '../../application/add/AddDraftScopeQuery';
import { InMemoryAddAssetDraftStore } from '../../application/add/AddAssetDraftStore';
import { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';

it('keeps dirty Add parent/title across metadata refresh and exposes dismissal', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let contextReads = 0; let principals = 0; let dismissed = 0; let tags: readonly [] = [];
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Garage', canAdd: true, assetTags: tags };
  const store = new InMemoryAddAssetDraftStore('scope');
  const query = new AddAssetContextQuery({ getAddAssetContext: async () => ({ ...context, inventoryName: `Garage ${++contextReads}`, assetTags: tags }) });
  const scope = new AddDraftScopeQuery({ getCurrentPrincipal: async () => { principals++; return { id: 'principal' }; } });
  const settle = () => h.run(() => new Promise(r => setTimeout(r, 10)));
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen addAssetContextQuery={query} addDraftScopeQuery={scope} addAssetDraftStore={store} createAssetCommand={{ execute: async () => ({ id: 'new', title: 'New', message: 'Saved' }) }} parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })} photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} initialParent={{ id: 'initial', title: 'Initial', kind: 'container', pathLabel: 'Initial', selectionHint: '', subtitle: '', willPromoteToContainer: false }} onDismiss={() => { dismissed++; }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(); await settle();
    await h.changeText(h.byLabel('Asset name'), 'My dirty draft');
    // The draft store is a persistence port, not a live source of form state.
    store.save({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' }, { title: 'Stale stored draft', description: '', parentQuery: '', selectedPhotos: [], showDetails: false });
    await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.addContext('scope', 'tenant', 'inventory') })); await settle();
    expect(h.byLabel('Asset name')?.props.value).toBe('My dirty draft'); expect(principals).toBe(1);
    await h.press(h.byLabel('Close Add')); expect(dismissed).toBe(1);
  } finally { await h.unmount(); }
});
