import { assetId } from '../../domain/assets/AssetSummary';
import { AddDestinationTaskProvider, useAddDestinationTask } from '../navigation/AddDestinationTask';
import { navigationOptions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { expect, it } from 'vitest';
import { AddAssetScreen as AddScreen } from './AddAssetScreen';
import { AddAssetContextQuery } from '../../application/add/AddAssetContextQuery';
import { AddDraftScopeQuery } from '../../application/add/AddDraftScopeQuery';
import { InMemoryAddAssetDraftStore } from '../../application/add/AddAssetDraftStore';
import { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';

function Destination() { return useAddDestinationTask()?.content ?? null; }
function AddAssetScreen(props: React.ComponentProps<typeof AddScreen>) { return <AddDestinationTaskProvider><AddScreen {...props} /><Destination /></AddDestinationTaskProvider>; }

it.each(['cancel', 'existing', 'top-level', 'create', 'permission'] as const)('preserves the item draft through destination %s and rejected save', async choice => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const store = new InMemoryAddAssetDraftStore('scope');
  const scope = { tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' };
  const context = { ...scope, tenantName: 'Home', inventoryName: 'Inventory', canAdd: true, assetTags: [] };
  const garage = { id: 'garage', title: 'Garage', kind: 'location' as const, subtitle: 'Garage', pathLabel: 'Garage', selectionHint: 'Location', willPromoteToContainer: false };
  store.save(scope, { title: 'Tent', description: 'Packed', parentAssetId: garage.id, parentQuery: garage.title, lastParent: garage, selectedPhotos: [], showDetails: false });
  const shed = { id: assetId('shed'), title: 'Shed', kind: 'location' as const, lifecycleState: 'active' as const, locationLabel: 'Yard', locationTrail: ['Inventory', 'Yard', 'Shed'], parentLocationTrail: [], description: '', updatedAtLabel: '', hasPhoto: false };
  const submissions: Array<{ kind?: string; title: string; parentAssetId?: string }> = [];
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
      inventoryAssetTypesQuery={{ execute: async () => [] }} addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
      addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={store}
      createAssetCommand={{ execute: async input => { submissions.push(input); if (input.kind === 'location') return { id: 'new-place', title: input.title, message: 'Created' }; throw new Error('Retain for retry'); } }}
      parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [shed] })}
      photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} /></AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    const chooser = h.byLabel('Choose destination');
    if (choice === 'cancel') {
      await h.run(() => setScreenFocused(false));
      await h.press(chooser); expect(h.byLabel('Choose inventory top level')).toBeUndefined();
      await h.run(() => setScreenFocused(true));
    }
    await h.press(chooser);
    await h.run(() => Object.assign({}, ...navigationOptions()).headerSearchBarOptions.onChangeText({ nativeEvent: { text: 'Another place' } }));
    expect(store.load(scope)).toMatchObject({ parentAssetId: 'garage', parentQuery: 'Garage', title: 'Tent', description: 'Packed' });
    await h.run(() => new Promise(resolve => setTimeout(resolve, 350)));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    if (choice === 'permission') {
      const select = h.byLabel('Choose destination Shed')!.props.onPress;
      context.canAdd = false;
      await h.run(() => client.invalidateQueries());
      await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
      await h.run(select);
      expect(h.byLabel('Choose inventory top level')).toBeUndefined();
      expect(store.load(scope)).toMatchObject({ parentAssetId: 'garage', title: 'Tent' });
      expect(submissions).toEqual([]);
      return;
    }
    if (choice === 'cancel') await h.press(h.byLabel('Cancel location selection'));
    if (choice === 'existing') await h.press(h.byLabel('Choose destination Shed'));
    if (choice === 'top-level') await h.press(h.byLabel('Choose inventory top level'));
    if (choice === 'create') { await h.press(h.byLabel('New place')); await h.press(h.byLabel('Create place')); }
    expect(h.byLabel('Choose inventory top level')).toBeUndefined();
    await h.press(h.byLabel('Save item'));
    const parentAssetId = choice === 'cancel' ? 'garage' : choice === 'existing' ? 'shed' : choice === 'create' ? 'new-place' : undefined;
    expect(submissions.at(-1)).toMatchObject({ parentAssetId, title: 'Tent', description: 'Packed' });
    expect(store.load(scope)).toMatchObject({ parentAssetId, title: 'Tent', description: 'Packed' });
    if (choice === 'create') expect(submissions[0]).toMatchObject({ kind: 'location', title: 'Another place' });
  } finally { await h.unmount(); client.clear(); resetNavigation(); setScreenFocused(true); }
});
