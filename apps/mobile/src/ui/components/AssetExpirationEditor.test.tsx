import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { AssetExpirationEditor } from './AssetExpirationEditor';
import type { EditDraft } from '../screens/AssetDetailEditPresentation';

it('clears a stored date without dropping the other edit fields', async () => {
  const harness = new MobileRenderHarness();
  let result: EditDraft | undefined;
  try {
    await harness.render(<AssetExpirationEditor asset={{ id: 'item', title: 'Medicine', description: '', customAssetTypeId: 'medicine', expiration: { date: '2028-02', precision: 'month' } }}
      draft={{ title: 'Edited name', description: 'Notes', tagIds: ['tag'] }}
      types={[{ kind: 'asset-type', id: 'medicine', key: 'medicine', displayName: 'Medicine', description: '', tenantId: 'tenant', scope: 'inventory', inventoryId: 'inventory', lifecycle: 'active', expirationEnabled: true }]}
      disabled={false} onChange={(draft) => { result = draft; }} />);
    await harness.press(harness.byLabel('Clear expiration'));
    expect(result).toMatchObject({ title: 'Edited name', description: 'Notes', tagIds: ['tag'], expiration: null });
  } finally { await harness.unmount(); }
});

it('can discard an invalid date when refreshed settings disable tracking', async () => {
  const harness = new MobileRenderHarness();
  let result: EditDraft | undefined;
  const asset = { id: 'item', title: 'Medicine', description: '', customAssetTypeId: 'medicine' };
  const type = { kind: 'asset-type' as const, id: 'medicine', key: 'medicine', displayName: 'Medicine', description: '', tenantId: 'tenant', scope: 'inventory' as const, inventoryId: 'inventory', lifecycle: 'active' as const };
  const draft = { title: 'Changed', description: 'Keep', expiration: null, expirationValid: false };
  try {
    await harness.render(<AssetExpirationEditor asset={asset} draft={draft} types={[{ ...type, expirationEnabled: true }]} disabled={false} onChange={(value) => { result = value; }} />);
    await harness.render(<AssetExpirationEditor asset={asset} draft={draft} types={[{ ...type, expirationEnabled: false }]} disabled={false} onChange={(value) => { result = value; }} />);
    await harness.press(harness.byLabel('Clear expiration'));
    expect(result).toMatchObject({ title: 'Changed', description: 'Keep', expiration: null, expirationValid: true });
  } finally { await harness.unmount(); }
});
