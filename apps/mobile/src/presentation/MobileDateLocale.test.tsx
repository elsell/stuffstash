import React from 'react';
import { afterAll, expect, it } from 'vitest';
import type { Asset } from '@stuff-stash/api-client';
import type { InventoryArchiveWorkspace } from '../application/archives/InventoryArchive';
import { assetId } from '../domain/assets/AssetSummary';

// Use the real localization bootstrap, not a mocked translator or device Intl.
const priorLocale = process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = 'de';
afterAll(() => {
  if (priorLocale === undefined) delete process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
  else process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = priorLocale;
});
const { updatedAtLabel } = await import('../adapters/inventories/InventoryAssetMapping');
const { toAssetDetailViewModel } = await import('../application/assets/AssetViewModels');
const { InventoryArchiveScreen } = await import('../ui/screens/InventoryArchiveScreen');
const { MobileRenderHarness } = await import('../test-support/render');
const timestamp = '2026-10-03T12:00:00Z';
const date = new Intl.DateTimeFormat('de', { month: 'short', day: 'numeric', year: 'numeric' }).format(new Date(timestamp));

it('formats asset update dates in the configured locale and preserves fallbacks', () => {
  expect(updatedAtLabel({ updatedAt: timestamp } as Asset)).toBe('Updated ' + date);
  expect(updatedAtLabel({ createdAt: timestamp } as Asset)).toBe('Updated ' + date);
  expect(updatedAtLabel({ updatedAt: 'invalid' } as Asset)).toBe('Loaded from API');
  expect(updatedAtLabel({} as Asset)).toBe('Loaded from API');
});
it('formats checkout metadata without changing availability or invalid-date recovery', () => {
  const asset = { id: assetId('test'), title: 'Test', kind: 'item' as const, lifecycleState: 'active' as const,
    locationLabel: '', locationTrail: [], parentLocationTrail: [], description: '', updatedAtLabel: '', hasPhoto: false };
  const checkout = { id: 'checkout', state: 'open' as const, checkedOutAt: timestamp, checkedOutByPrincipalId: 'user' };
  expect(toAssetDetailViewModel({ ...asset, currentCheckout: checkout }).checkoutLabel).toBe('Checked out ' + date);
  expect(toAssetDetailViewModel({ ...asset, currentCheckout: { ...checkout, checkedOutAt: 'invalid' } }).checkoutLabel).toBe('Checked out');
  expect(toAssetDetailViewModel(asset).checkoutLabel).toBe('Available');
});
it('renders archive expiry in the configured locale', async () => {
  const workspace = { newRequestKey: () => 'key', files: {}, repository: {
    list: async () => ({ jobs: [{ id: 'job', kind: 'export', state: 'queued', phase: 'export', createdAt: timestamp, expiresAt: timestamp, photos: true, otherFiles: true }] })
  } } as unknown as InventoryArchiveWorkspace;
  const h = new MobileRenderHarness();
  try {
    await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'test' }} onClose={() => {}} onOpen={async () => {}} />);
    expect(h.allText().join(' ')).toContain(new Date(timestamp).toLocaleString('de'));
  } finally { await h.unmount(); }
});
