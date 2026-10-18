import { expect, it } from 'vitest';
import { canSaveEditAsset, hasDirtyEditAssetDraft, normalizedEditDraft } from './AssetDetailEditPresentation';
import { assetOverflowMenuActions } from './AssetLifecyclePresentation';

it('tracks kind-only edits for Save and discard without changing the loaded asset', () => {
  const asset = { kind: 'location' as const, title: 'Bin 60', description: '' };
  const draft = { ...asset, kind: 'container' as const };
  expect(canSaveEditAsset(asset, draft)).toBe(true);
  expect(hasDirtyEditAssetDraft(asset, draft)).toBe(true);
  expect(normalizedEditDraft(draft).kind).toBe('container');
  expect(canSaveEditAsset(asset, asset)).toBe(false);
});

it('hides irrelevant checkout history but preserves records after conversion', () => {
  const asset = { canArchive: true, canRestore: false, canDeletePermanently: false, canViewCheckoutHistory: false };
  expect(assetOverflowMenuActions(asset).map(action => action.id)).not.toContain('checkout-history');
  expect(assetOverflowMenuActions({ ...asset, canViewCheckoutHistory: true }).map(action => action.id)).toContain('checkout-history');
});
