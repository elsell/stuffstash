import { mobileQueryKeys } from '../src/adapters/serverState/MobileQueryClient';
import { expect, it } from 'vitest';
import { createAssetEditJourney } from './AssetEditJourneyFixture';

it('persists edits through the real command for subsequent detail and editor reads', async () => {
  const journey = createAssetEditJourney();
  const before = await journey.core.execute('audit-edit-item');
  expect(before.view.title).toBe('Camping tent');
  await journey.update.execute({ assetId: 'audit-edit-item', title: 'Camping tent kit', description: 'Ready for the weekend',
    tagIds: [], newTags: [{ displayName: 'Camping' }], activeTags: [] });
  const after = await journey.core.execute('audit-edit-item');
  expect(after.view.title).toBe('Camping tent kit');
  expect(after.view.description).toBe('Ready for the weekend');
  expect(after.snapshot.asset.tags).toEqual([expect.objectContaining({ displayName: 'Camping' })]);
  expect(journey.writeCount()).toBe(1);
});

it('does not pretend to edit another fixture asset', async () => {
  const journey = createAssetEditJourney();
  await expect(journey.update.execute({ assetId: 'another-asset', title: 'Other', description: '' })).rejects.toThrow();
  expect(journey.writeCount()).toBe(0);
});


it('invalidates the detail cache through the production mutation observer', async () => {
  const journey = createAssetEditJourney();
  const key = mobileQueryKeys.assetCore('edit-journey', 'audit-tenant', 'audit-inventory', 'audit-edit-item');
  journey.client.setQueryData(key, await journey.core.execute('audit-edit-item'));
  await journey.update.execute({ assetId: 'audit-edit-item', title: 'Updated tent', description: '' });
  expect(journey.client.getQueryState(key)?.isInvalidated).toBe(true);
});

it('moves through the real command, preserving edits and refreshing placement for return', async () => {
  const journey = createAssetEditJourney();
  await journey.update.execute({ assetId: 'audit-edit-item', title: 'Weekend tent', description: 'Packed',
    tagIds: [], newTags: [{ displayName: 'Camping' }], activeTags: [] });
  const key = mobileQueryKeys.assetCore('edit-journey', 'audit-tenant', 'audit-inventory', 'audit-edit-item');
  journey.client.setQueryData(key, await journey.core.execute('audit-edit-item'));
  await journey.move.execute({ assetId: 'audit-edit-item', parentAssetId: 'journey-garage' });
  const moved = await journey.core.execute('audit-edit-item');
  expect(moved.snapshot.asset).toMatchObject({ title: 'Weekend tent', description: 'Packed',
    parentAssetId: 'journey-garage', parentLocationTrail: [{ id: 'journey-garage', title: 'Garage' }],
    tags: [expect.objectContaining({ displayName: 'Camping' })] });
  expect((await journey.placement.execute(moved.snapshot)).parentLocationTrailLabel).toBe('Garage');
  expect(journey.client.getQueryState(key)?.isInvalidated).toBe(true);
  await journey.move.execute({ assetId: 'audit-edit-item' });
  expect((await journey.core.execute('audit-edit-item')).snapshot.asset.parentAssetId).toBeUndefined();
  expect(journey.writeCount()).toBe(3);
});

it('rejects unknown destinations without changing the asset', async () => {
  const journey = createAssetEditJourney();
  await expect(journey.move.execute({ assetId: 'audit-edit-item', parentAssetId: 'missing' })).rejects.toThrow();
  expect(journey.writeCount()).toBe(0);
});

it('provides a populated container and real final metadata for tab clearance', async () => {
  const journey = createAssetEditJourney('populated-container');
  const core = await journey.core.execute('audit-edit-item');
  expect(core.view.updatedAtLabel).toBe('Updated today');
  expect(core.view.canContainAssets).toBe(true);
  const contents = await journey.contents.execute(core.snapshot);
  expect(contents.containedAssets).toHaveLength(12);
  expect(contents.containedAssets.every(child => !child.updatedAtLabel.startsWith('Updated '))).toBe(true);
});
