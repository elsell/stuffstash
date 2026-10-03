import { expect, it } from 'vitest';
import { AddAssetScreen } from './AddAssetScreen';
import { AddDestinationTaskProvider } from '../navigation/AddDestinationTask';
import { AddAssetContextQuery } from '../../application/add/AddAssetContextQuery';
import { AddDraftScopeQuery } from '../../application/add/AddDraftScopeQuery';
import { InMemoryAddAssetDraftStore } from '../../application/add/AddAssetDraftStore';
import { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { MobileRenderHarness } from '../../test-support/render';
import { PrintingFake } from '../../test-support/PrintingFake';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { resetNavigation } from '../../test-support/navigation';
import type { CreateAssetCommandInput } from '../../application/add/CreateAssetCommand';

it('initializes printing once, preserves a user choice, and retries an immutable create request after navigation', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const fake = new PrintingFake(); fake.settings = { ...fake.settings, printOnCreateDefault: true };
  const printing = fake.workspace(); const store = new InMemoryAddAssetDraftStore('scope');
  const scope = { tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' };
  const context = { ...scope, tenantName: 'Home', inventoryName: 'Home', canAdd: true, assetTags: [] };
  store.save(scope, { title: 'Lamp', description: '', parentQuery: '', selectedPhotos: [], showDetails: false });
  const requests: CreateAssetCommandInput[] = []; let lost = true;
  const command = { execute: async (input: CreateAssetCommandInput) => { requests.push(input); if (lost) { lost = false; throw new Error('Response lost after atomic commit'); } return { id: 'asset', title: input.title, message: 'Saved', printJobId: 'job' }; } };
  const screen = () => <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddDestinationTaskProvider><AddAssetScreen printing={printing}
    inventoryAssetTypesQuery={{ execute: async () => [] }} addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
    addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={store} createAssetCommand={command}
    parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })} photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} /></AddDestinationTaskProvider></AppFeedbackProvider></MobileServerStateProvider>;
  const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
  try {
    await h.render(screen()); await settle();
    expect(h.byLabel('Print label when adding an item')?.props.value).toBe(true);
    await h.run(() => h.byLabel('Print label when adding an item')?.props.onValueChange(false));
    await h.run(() => client.invalidateQueries()); await settle();
    expect(h.byLabel('Print label when adding an item')?.props.value).toBe(false);
    await h.run(() => h.byLabel('Print label when adding an item')?.props.onValueChange(true));
    await h.press(h.byLabel('Save item')); await settle();
    expect(requests[0].printRequest?.key).toBe('request'); expect(store.load(scope)?.pendingPrintCreate).toBeDefined();
    await h.render(<></>); await h.render(screen()); await settle();
    expect(h.byLabel('Asset name')?.props.editable).toBe(false);
    await h.press(h.byLabel('Save item')); await settle();
    expect(requests).toHaveLength(2); expect(requests[1]).toEqual(requests[0]);
    expect(store.load(scope)?.pendingPrintCreate).toBeUndefined();
  } finally { await h.unmount(); client.clear(); resetNavigation(); }
});
