import { expect, it } from 'vitest';
import { canSaveEditAsset, hasDirtyEditAssetDraft, normalizedEditDraft } from './AssetDetailEditPresentation';

it('tracks date-only edits, clearing, and initial type assignment', () => {
  const expiration = { date: '2028-02', precision: 'month' as const };
  const asset = { title: 'Medicine', description: '', expiration, customAssetTypeId: 'medicine' };
  const unchanged = { title: asset.title, description: '' };
  expect(hasDirtyEditAssetDraft(asset, unchanged)).toBe(false);
  expect(canSaveEditAsset(asset, { ...unchanged, expiration })).toBe(false);
  expect(canSaveEditAsset(asset, { ...unchanged, expiration: null })).toBe(true);
  expect(normalizedEditDraft({ ...unchanged, expiration: null }).expiration).toBeNull();
  const day = { date: '2028-02-29', precision: 'day' as const };
  expect(canSaveEditAsset(asset, { ...unchanged, expiration: day })).toBe(true);
  expect(normalizedEditDraft({ ...unchanged, expiration: day }).expiration).toEqual(day);
  expect(canSaveEditAsset(asset, { ...unchanged, expiration: day, expirationValid: false })).toBe(false);
  expect(hasDirtyEditAssetDraft(asset, { ...unchanged, expirationValid: false })).toBe(true);
  expect(canSaveEditAsset({ ...asset, customAssetTypeId: undefined }, { ...unchanged, customAssetTypeId: 'medicine' })).toBe(true);
});

it('preserves the date and type in the detail view model', async () => {
  const { toAssetDetailViewModel } = await import('../../application/assets/AssetViewModels');
  const { assetId } = await import('../../domain/assets/AssetSummary');
  const expiration = { date: '2028-02', precision: 'month' as const };
  const detail = toAssetDetailViewModel({ id: assetId('medicine'), title: 'Medicine', description: '',
    kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [],
    updatedAtLabel: '', hasPhoto: false, expiration, customAssetTypeId: 'type-medicine' });
  expect(detail).toMatchObject({ expiration, customAssetTypeId: 'type-medicine' });
});
