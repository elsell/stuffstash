import { AssetTagSelectionTaskProvider, useAssetTagSelectionTask } from '../navigation/AssetTagSelectionTask';
import { scrollCommandsForTest } from '../../test-support/react-native';
import { setNativeHeaderHeight } from '../../test-support/react-navigation-elements';
import React from 'react';
import { NavigationOptionFeedback } from '../../test-support/NavigationOptionFeedback';
import { Platform, pressAlertButton, latestActionSheetCallback } from '../../test-support/react-native';
import { afterEach, expect, it } from 'vitest';
import { AddAssetScreen } from './AddAssetScreen';
import { AddAssetContextQuery } from '../../application/add/AddAssetContextQuery';
import { AddDraftScopeQuery } from '../../application/add/AddDraftScopeQuery';
import { InMemoryAddAssetDraftStore } from '../../application/add/AddAssetDraftStore';
import { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { navigationOptions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { AppFeedbackProvider } from '../feedback/AppFeedback';

it('removes the numbered draft photo while preserving the remaining selection and item draft', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const store = new InMemoryAddAssetDraftStore('scope');
  const scope = { tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' };
  const context = { ...scope, tenantName: 'Home', inventoryName: 'Home', canAdd: true, assetTags: [] };
  store.save(scope, { title: 'Keep this item', description: '', parentQuery: '', showDetails: false,
    selectedPhotos: ['first', 'second'].map(id => ({ id, uri: `file:///${id}.jpg`, fileName: `${id}.jpg`, contentType: 'image/jpeg' as const, sizeBytes: 100 })) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
      inventoryAssetTypesQuery={{ execute: async () => [] }} addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
      addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={store}
      createAssetCommand={{ execute: async () => { throw new Error('Must not save'); } }}
      parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })}
      photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} /></AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    expect(h.byLabel('Asset name')).toBeDefined();
    expect(h.byLabel('Remove photo 1')).toBeDefined();
    expect(h.byLabel('Remove photo 2')).toBeDefined();
    await h.press(h.byLabel('Remove photo 1'));
    expect(store.load(scope)?.selectedPhotos.map(photo => photo.id)).toEqual(['second']);
    expect(store.load(scope)?.title).toBe('Keep this item');
    expect(h.byLabel('Remove photo 2')).toBeUndefined();
    await h.press(h.byLabel('Remove photo 1'));
    expect(store.load(scope)?.selectedPhotos).toEqual([]);
    expect(store.load(scope)?.title).toBe('Keep this item');
  } finally { await h.unmount(); client.clear(); }
});

it.each([0, 1])('rejects an Add photo chooser from a departed visit, source=%s', async source => {
  resetNavigation(); let selections = 0;
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', canAdd: true, assetTags: [] };
  const select = async () => { selections++; return []; };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
      inventoryAssetTypesQuery={{ execute: async () => [] }} addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
      addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={new InMemoryAddAssetDraftStore('scope')}
      createAssetCommand={{ execute: async () => { throw new Error('Creation not requested'); } }} parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })}
      photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: select, captureFromCamera: select })} /></AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    const add = h.byLabel('Add photos');
    expect(add).toBeDefined();
    await h.press(add);
    const choose = latestActionSheetCallback();
    expect(choose).toBeTypeOf('function');
    await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true));
    await h.run(() => choose!(source));
    expect(selections).toBe(0);
    await h.press(add); await h.run(() => latestActionSheetCallback()!(source));
    expect(selections).toBe(1);
  } finally { await h.unmount(); client.clear(); resetNavigation(); }
});

it('keeps dirty Add parent/title across metadata refresh and exposes dismissal', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let contextReads = 0; let principals = 0; let dismissed = 0; let tags: readonly [] = []; let submittedTitle = '';
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Garage', canAdd: true, assetTags: tags };
  const store = new InMemoryAddAssetDraftStore('scope');
  const query = new AddAssetContextQuery({ getAddAssetContext: async () => ({ ...context, inventoryName: `Garage ${++contextReads}`, assetTags: tags }) });
  const scope = new AddDraftScopeQuery({ getCurrentPrincipal: async () => { principals++; return { id: 'principal' }; } });
  const settle = () => h.run(() => new Promise(r => setTimeout(r, 10)));
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen inventoryAssetTypesQuery={{ execute: async () => [] }} addAssetContextQuery={query} addDraftScopeQuery={scope} addAssetDraftStore={store} createAssetCommand={{ execute: async input => { submittedTitle = input.title; throw new Error('Keep draft for recovery'); } }} parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })} photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} initialParent={{ id: 'initial', title: 'Initial', kind: 'container', pathLabel: 'Initial', selectionHint: '', subtitle: '', willPromoteToContainer: false }} onDismiss={() => { dismissed++; }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await settle(); await settle();
    expect(navigationOptions().at(-1)).toMatchObject({ headerShown: true, title: 'Add item' });
    await h.changeText(h.byLabel('Asset name'), 'My dirty draft');
    // The draft store is a persistence port, not a live source of form state.
    store.save({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' }, { title: 'Stale stored draft', description: '', parentQuery: '', selectedPhotos: [], showDetails: false });
    await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.addContext('scope', 'tenant', 'inventory') })); await settle();
    await h.press(h.byLabel('Save item'));
    expect(submittedTitle).toBe('My dirty draft'); expect(principals).toBe(1);
    await h.press(h.byLabel('Close Add')); expect(dismissed).toBe(1);
  } finally { await h.unmount(); }
});


it('creates with a month expiration and retains the draft when saving fails', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', canAdd: true, assetTags: [] };
  const store = new InMemoryAddAssetDraftStore('scope');
  const submissions: unknown[] = [];
  store.save({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' }, { title: 'Restored name', description: '', parentQuery: '', selectedPhotos: [], showDetails: false });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
      inventoryAssetTypesQuery={{ execute: async () => [{ kind: 'asset-type', id: 'medicine', key: 'medicine', displayName: 'Medicine', description: '', tenantId: 'tenant', inventoryId: 'inventory', scope: 'inventory', lifecycle: 'active', expirationEnabled: true }] }}
      addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
      addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={store}
      createAssetCommand={{ execute: async (input) => { submissions.push(input); if (submissions.length === 1) throw new Error('Save failed'); return { id: 'created', title: input.title, message: 'Saved' }; } }}
      parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })} photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} /></AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    expect(h.byLabel('Asset name')?.props.defaultValue).toBe('Restored name');
    await h.changeText(h.byLabel('Asset name'), 'Tylenol');
    await h.press(h.byLabel('Item type'));
    await h.press(h.byLabel('Medicine'));
    await h.press(h.byLabel('Expiration'));
    await h.change(h.byType('NativeSegmentedControl'), 'Month and year');
    await h.press(h.byLabel('Expiration month'));
    await h.press(h.byLabel('February'));
    await h.changeText(h.byLabel('Expiration year'), '2028');
    await h.press(h.byLabel('Save item'));
    expect(submissions).toEqual([expect.objectContaining({ customAssetTypeId: 'medicine', expiration: { date: '2028-02', precision: 'month' } })]);
    expect(store.load({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' })).toMatchObject({ customAssetTypeId: 'medicine', expiration: { date: '2028-02', precision: 'month' } });
    expect(h.byLabel('Expiration year')?.props.value).toBe('2028');
    await h.press(h.byLabel('Save item'));
    expect(submissions).toHaveLength(2);
    expect(h.byLabel('Asset name')?.props.defaultValue).toBe('');
    expect(store.load({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' })?.expiration).toBeUndefined();
    expect(h.byLabel('Expiration year')).toBeUndefined();
  } finally { await h.unmount(); }
});

it('preserves the submitted draft and prevents duplicate saves while saving is pending', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', canAdd: true, assetTags: [] };
  const store = new InMemoryAddAssetDraftStore('scope');
  let rejectSave!: (error: Error) => void; let saves = 0; let dismissed = 0;
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
      inventoryAssetTypesQuery={{ execute: async () => [] }}
      addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
      addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={store}
      createAssetCommand={{ execute: async () => { saves++; return new Promise((_resolve, reject) => { rejectSave = reject; }); } }}
      parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })}
      photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })}
      onDismiss={() => { dismissed++; }} /></AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    await h.changeText(h.byLabel('Asset name'), 'Submitted name');
    const save = h.byLabel('Save item')!;
    await h.run(() => { void save.props.onPress(); void save.props.onPress(); });
    expect(saves).toBe(1);
    expect(h.byLabel('Asset name')?.props.editable).toBe(false);
    await h.changeText(h.byLabel('Asset name'), 'Late edit');
    expect(store.load({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' })?.title).toBe('Submitted name');
    await h.press(h.byLabel('Close Add'));
    expect(dismissed).toBe(0);
    await h.run(() => rejectSave(new Error('Save unavailable')));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
    expect(h.byText('Could not save asset')).toBeDefined();
    const failure = h.byText('Save unavailable');
    expect(failure).toBeDefined();
    // Feedback must belong to the presented form, not the root overlay.
    let ancestor = failure?.parent;
    while (ancestor && ancestor.type !== 'ScrollView') ancestor = ancestor.parent;
    expect(ancestor?.type).toBe('ScrollView');
    const revealError = failure?.parent?.props.onLayout;
    expect(revealError).toBeTypeOf('function');
    await h.run(() => setNativeHeaderHeight(72));
    await h.run(() => h.byText('Save unavailable')?.parent?.props.onLayout());
    expect(scrollCommandsForTest().at(-1)).toEqual({ y: -72, animated: false });
    await h.run(() => setNativeHeaderHeight(96));
    await h.run(() => h.byText('Save unavailable')?.parent?.props.onLayout());
    expect(scrollCommandsForTest().at(-1)).toEqual({ y: -96, animated: false });
    expect(ancestor?.props.scrollToOverflowEnabled).toBe(true);
    expect(h.byLabel('Asset name')?.props.editable).toBe(true);
    expect(store.load({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' })?.title).toBe('Submitted name');
    await h.changeText(h.byLabel('Asset name'), 'Retry name');
    expect(store.load({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' })?.title).toBe('Retry name');
  } finally { await h.unmount(); client.clear(); }
});

for (const operation of ['parent', 'photo', 'library-failure', 'camera-failure'] as const) {
  it(`preserves the draft during pending ${operation} selection and restores editing after failure or cancellation`, async () => {
    const h = new MobileRenderHarness(); const client = createMobileQueryClient();
    const originalPlatform = Platform.OS; Platform.OS = 'android';
    const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', canAdd: true, assetTags: [] };
    let finish!: () => void; let submissions = 0; let dismissed = 0;
    try {
      await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
        inventoryAssetTypesQuery={{ execute: async () => [] }}
        addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
        addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })}
        addAssetDraftStore={new InMemoryAddAssetDraftStore('scope')}
        createAssetCommand={{ execute: async () => { submissions++; return new Promise((_resolve, reject) => { finish = () => reject(new Error('Parent unavailable')); }); } }}
        parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })}
        photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: () => new Promise((resolve, reject) => { finish = () => operation === 'library-failure' ? reject(new Error('Library unavailable')) : resolve([]); }), captureFromCamera: () => new Promise((_resolve, reject) => { finish = () => reject(new Error('Camera unavailable')); }) })}
        onDismiss={() => { dismissed++; }} /></AppFeedbackProvider></MobileServerStateProvider>);
      await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
      await h.changeText(h.byLabel('Asset name'), 'Keep this draft');
      if (operation === 'parent') {
        await h.press(h.byText('No parent')?.parent?.parent ?? undefined);
        await h.changeText(h.byLabel('Search parent'), 'New parent');
        await h.run(() => new Promise(resolve => setTimeout(resolve, 400)));
        await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
        const create = h.byLabel('Create "New parent" as a place');
        expect(create).toBeDefined();
        await h.run(() => { void create!.props.onPress(); });
        expect(h.byLabel('Creating place…')?.props.accessibilityState).toMatchObject({ disabled: true });
      } else {
        await h.press(h.all().find(node => node.props.accessibilityHint === 'Choose camera or photo library'));
        await h.run(() => { void pressAlertButton(operation === 'camera-failure' ? 'Take Photo' : 'Choose from Library'); });
      }
      expect(finish).toBeDefined();
      expect(h.byLabel('Asset name')?.props.editable).toBe(false);
      await h.changeText(h.byLabel('Asset name'), 'Late edit');
      await h.press(h.byLabel('Save item'));
      await h.press(h.byLabel('Close Add'));
      expect(submissions).toBe(operation === 'parent' ? 1 : 0);
      expect(dismissed).toBe(0);
      expect(h.byLabel('Asset name')?.props.value).toBe('Keep this draft');
      await h.run(() => finish());
      await h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
      if (operation !== 'photo') {
        const message = operation === 'parent' ? 'Parent unavailable' : operation === 'library-failure' ? 'Library unavailable' : 'Camera unavailable';
        let owner = h.byText(message)?.parent;
        while (owner && owner.type !== 'ScrollView') owner = owner.parent;
        expect(owner?.type).toBe('ScrollView');
      }
      expect(h.byLabel('Asset name')?.props.editable).toBe(true);
      await h.changeText(h.byLabel('Asset name'), 'Recovered draft');
      expect(h.byLabel('Asset name')?.props.value).toBe('Recovered draft');
      await h.press(h.byLabel('Close Add'));
      expect(dismissed).toBe(1);
    } finally { await h.unmount(); client.clear(); Platform.OS = originalPlatform; }
  });
}


it('settles navigation updates while header actions use the latest Add draft', async () => {
  resetNavigation();
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', canAdd: true, assetTags: [] };
  const submissions: string[] = [];
  let dismissed = 0;
  const props: React.ComponentProps<typeof AddAssetScreen> = {
    inventoryAssetTypesQuery: { execute: async () => [] },
    addAssetContextQuery: new AddAssetContextQuery({ getAddAssetContext: async () => context }),
    addDraftScopeQuery: new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) }),
    addAssetDraftStore: new InMemoryAddAssetDraftStore('scope'),
    createAssetCommand: { execute: async input => { submissions.push(input.title); throw new Error('Save failed'); } },
    parentLookupQuery: new ParentLookupQuery({ listParentCandidates: async () => [] }),
    photoSelectionQuery: new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] }),
    onDismiss: () => { dismissed++; }
  };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider>
      <NavigationOptionFeedback render={() => <AddAssetScreen {...props} />} />
    </AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    await h.changeText(h.byLabel('Asset name'), 'First name');
    await h.changeText(h.byLabel('Asset name'), 'Latest name');
    await h.press(h.byLabel('Save item'));
    expect(submissions).toEqual(['Latest name']);
    await h.changeText(h.byLabel('Asset name'), 'Retry name');
    await h.press(h.byLabel('Save item'));
    expect(submissions).toEqual(['Latest name', 'Retry name']);
    await h.press(h.byLabel('Close Add'));
    expect(dismissed).toBe(1);
  } finally { await h.unmount(); client.clear(); resetNavigation(); }
});

afterEach(() => setNativeHeaderHeight(144));

it('recovers unavailable item types without claiming loading or replacing the Add draft', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  client.setDefaultOptions({ queries: { retry: false } });
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', canAdd: true, assetTags: [] };
  let reads = 0;
  const store = new InMemoryAddAssetDraftStore('scope');
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
      inventoryAssetTypesQuery={{ execute: async () => { if (++reads !== 2) throw new Error('Types unavailable'); return [{ kind: 'asset-type', id: 'medicine', key: 'medicine', displayName: 'Medicine', description: '', tenantId: 'tenant', inventoryId: 'inventory', scope: 'inventory', lifecycle: 'active', expirationEnabled: true }]; } }}
      addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
      addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })}
      addAssetDraftStore={store} createAssetCommand={{ execute: async () => { throw new Error('No save requested'); } }}
      parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })}
      photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })}
    /></AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    await h.changeText(h.byLabel('Asset name'), 'Retained medicine');
    expect(h.byLabel('Retry asset types')).toBeDefined();
    expect(h.byText('Loading expiration settings…')).toBeUndefined();
    await h.press(h.byLabel('Retry asset types'));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    expect(reads).toBe(2);
    expect(h.byLabel('Retry asset types')).toBeUndefined();
    await h.press(h.byLabel('Item type')); await h.press(h.byLabel('Medicine'));
    expect(store.load({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' })).toMatchObject({ title: 'Retained medicine', customAssetTypeId: 'medicine' });
    expect(h.byLabel('Expiration')).toBeDefined();
    await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.customization('scope', 'tenant', 'inventory', 'inventory', 'asset-type-choices', 'active') }));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    expect(reads).toBe(3);
    expect(h.byLabel('Retry asset types')).toBeDefined();
    expect(h.byLabel('Expiration')).toBeDefined();
    expect(h.byLabel('Item type')?.props.accessibilityValue).toEqual({ text: 'Medicine' });
    expect(h.byText('Loading expiration settings…')).toBeUndefined();
    expect(store.load({ tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' })?.title).toBe('Retained medicine');
  } finally { await h.unmount(); client.clear(); }
});


it('retains unfinished Add tag input through disclosure and scoped draft restoration', async () => {
  const client = createMobileQueryClient(); const store = new InMemoryAddAssetDraftStore('scope');
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', canAdd: true, assetTags: Array.from({ length: 14 }, (_, index) => ({ id: `tag-${index + 1}`, key: `tag-${index + 1}`, displayName: `Tag ${index + 1}` })).reverse() };
  const draftContext = { tenantId: 'tenant', inventoryId: 'inventory', principalId: 'principal' };
  const saved: unknown[] = [];
  const render = (h: MobileRenderHarness) => h.render(<AssetTagSelectionTaskProvider><SelectionContent /><MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
    inventoryAssetTypesQuery={{ execute: async () => [] }} addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
    addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={store}
    createAssetCommand={{ execute: async input => { saved.push(input); return { id: 'created', title: input.title, message: 'Saved' }; } }}
    parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async () => [] })} photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} /></AppFeedbackProvider></MobileServerStateProvider></AssetTagSelectionTaskProvider>);
  let h = new MobileRenderHarness();
  const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
  try {
    await render(h); await settle();
    await h.changeText(h.byLabel('Asset name'), 'Tent');
    await h.press(h.byText('More details')?.parent ?? undefined);
    expect(h.byText('Tag 1')).toBeUndefined();
    await h.press(h.byLabel('Choose tags'));
    await h.press(h.byLabel('Select tag Tag 14'));
    await h.press(h.byLabel('Done selecting tags'));
    expect(store.load(draftContext)?.selectedTagIds).toEqual(['tag-14']);
    expect(h.byLabel('New tag name')).toBeUndefined();
    await h.press(h.byLabel('New tag'));
    const cancelNewTag = h.byLabel('Cancel new tag')?.props.onPress;
    await h.changeText(h.byLabel('New tag name'), 'Temporary');
    await h.changeText(h.byLabel('Asset name'), 'Camping tent');
    await h.run(() => cancelNewTag?.());
    expect(h.byLabel('New tag name')).toBeUndefined();
    expect(store.load(draftContext)?.title).toBe('Camping tent');
    expect(store.load(draftContext)?.selectedTagIds).toEqual(['tag-14']);
    await h.press(h.byLabel('New tag'));
    const overlongName = 'Camping'.repeat(20);
    await h.changeText(h.byLabel('New tag name'), overlongName);
    expect(h.byText('Use a shorter tag name.')).toBeDefined();
    await h.press(h.byLabel('Add tag'));
    expect(store.load(draftContext)?.inlineTag?.name).toBe(overlongName);
    await h.changeText(h.byLabel('New tag name'), 'Camping');
    expect(h.byText('Use a shorter tag name.')).toBeUndefined();
    const tagField = () => h.allByType('TextInput').find(node => node.props.accessibilityLabel === 'New tag name');
    expect(tagField()?.props.defaultValue).toBe('');
    expect(tagField()?.props.value).toBeUndefined();
    await h.press(h.byText('More details')?.parent ?? undefined);
    expect(h.byText('Open More details to add or clear the unfinished tag before saving.')).toBeDefined();
    await h.press(h.byText('More details')?.parent ?? undefined);
    expect(store.load(draftContext)?.inlineTag?.name).toBe('Camping');
    expect(tagField()?.props.defaultValue).toBe('Camping');
    expect(store.load(draftContext)).toMatchObject({ inlineTag: { name: 'Camping', color: '' } });
    expect(store.load({ ...draftContext, inventoryId: 'other' })).toBeUndefined();
    await h.press(h.byLabel('Save item')); expect(saved).toEqual([]);
    expect(h.byText('Add this tag or clear its name and color before saving.')).toBeDefined();
    await h.unmount(); h = new MobileRenderHarness(); await render(h); await settle();
    expect(store.load(draftContext)?.inlineTag?.name).toBe('Camping');
    await h.press(h.byLabel('Choose Blue tag color'));
    expect(store.load(draftContext)?.inlineTag?.color).toBe('#2F80ED');
    await h.changeText(h.byLabel('Asset name'), 'Draft to clear');
    expect(store.load(draftContext)?.title).toBe('Draft to clear');
    expect(h.byLabel('Clear draft')).toBeDefined();
    await h.press(h.byLabel('Clear draft'));
    expect(store.load(draftContext)?.title).toBe('');
    expect(store.load(draftContext)?.inlineTag?.name ?? '').toBe('');
    expect(store.load(draftContext)?.inlineTag?.color ?? '').toBe('');
    await h.changeText(h.byLabel('Asset name'), 'Tent');
    await h.press(h.byText('More details')?.parent ?? undefined);
    await h.press(h.byLabel('New tag'));
    await h.changeText(h.byLabel('New tag name'), 'Camping');
    await h.press(h.byLabel('Add tag'));
    expect(store.load(draftContext)?.inlineTag?.name).toBe('');
    expect(tagField()?.props.defaultValue).toBe('');
    await h.changeText(h.byLabel('New tag name'), 'Other');
    await h.press(h.byLabel('Add tag'));
    await h.changeText(h.byLabel('New tag name'), 'Unfinished');
    expect(h.byLabel('Remove new tag Other')).toBeDefined();
    await h.press(h.byLabel('Remove new tag Other'));
    expect(store.load(draftContext)?.newTags).toEqual([{ displayName: 'Camping' }]);
    expect(store.load(draftContext)?.inlineTag?.name).toBe('Unfinished');
    expect(tagField()?.props.defaultValue).toBe('');
    await h.changeText(h.byLabel('New tag name'), '');
    await h.press(h.byLabel('Save item')); await settle();
    expect(saved).toEqual([expect.objectContaining({ title: 'Tent', newTags: [{ displayName: 'Camping' }] })]);
    expect(store.load(draftContext)?.inlineTag?.name ?? '').toBe('');
  } finally { await h.unmount(); }
});


it('waits for known parent suggestions before offering quick creation in Add', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let attempts = 0;
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', canAdd: true, assetTags: [] };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => context}><AppFeedbackProvider><AddAssetScreen
      inventoryAssetTypesQuery={{ execute: async () => [] }} addAssetContextQuery={new AddAssetContextQuery({ getAddAssetContext: async () => context })}
      addDraftScopeQuery={new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'principal' }) })} addAssetDraftStore={new InMemoryAddAssetDraftStore('scope')}
      createAssetCommand={{ execute: async () => { throw new Error('Creation not requested'); } }}
      parentLookupQuery={new ParentLookupQuery({ listParentCandidates: async input => { if (input === 'New parent' && ++attempts === 1) throw new Error('Lookup unavailable'); return []; } })}
      photoSelectionQuery={new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })} /></AppFeedbackProvider></MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    await h.press(h.byText('No parent')?.parent?.parent ?? undefined);
    await h.changeText(h.byLabel('Search parent'), 'New parent');
    expect(h.allByType('Text').find(node => node.children.join('') === 'Create "New parent" as a place')).toBeUndefined();
    expect(h.byText('Not selected yet')).toBeDefined();
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400)));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    expect(h.byText('Suggestions could not be loaded.')).toBeDefined();
    expect(h.allByType('Text').find(node => node.children.join('') === 'Create "New parent" as a place')).toBeUndefined();
    expect(h.byLabel('Retry suggestions')).toBeDefined();
    await h.press(h.byLabel('Retry suggestions'));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    expect(h.allByType('Text').find(node => node.children.join('') === 'Create "New parent" as a place')).toBeDefined();
    expect(h.byLabel('Search parent')?.props.value).toBe('New parent');
    await h.changeText(h.byLabel('Search parent'), '');
    expect(h.byText('Not selected yet')).toBeUndefined();
    expect(h.byText('Top level in this inventory')).toBeDefined();
  } finally { await h.unmount(); }
});

function SelectionContent() { const task = useAssetTagSelectionTask(); return <>{task?.content}</>; }
