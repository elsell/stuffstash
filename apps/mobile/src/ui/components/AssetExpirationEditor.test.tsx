import React from 'react';
import { expect, it } from 'vitest';
import { latestAlert } from '../../test-support/react-native';
import { setScreenFocused } from '../../test-support/navigation';
import { MobileRenderHarness } from '../../test-support/render';
import { AssetExpirationEditor } from './AssetExpirationEditor';
import type { EditDraft } from '../screens/AssetDetailEditPresentation';

it('explains an empty type search and restores choices without changing the draft', async () => {
  const h = new MobileRenderHarness(); const changes: EditDraft[] = [];
  try {
    await h.render(<AssetExpirationEditor asset={{ id: 'new-item', title: 'Medicine', description: '' }}
      draft={{ title: 'Medicine', description: '', customAssetTypeId: 'medicine', expiration: { date: '2028-02', precision: 'month' } }}
      types={[{ kind: 'asset-type', id: 'medicine', key: 'medicine', displayName: 'Medicine', description: '', tenantId: 'tenant', inventoryId: 'inventory', scope: 'inventory', lifecycle: 'active', expirationEnabled: true }]}
      disabled={false} onChange={draft => changes.push(draft)} />);
    await h.press(h.byLabel('Item type'));
    await h.changeText(h.byLabel('Search item types'), 'No such type');
    expect(h.byText('No matching item types.')).toBeDefined();
    expect(h.byLabel('Medicine')).toBeUndefined();
    await h.changeText(h.byLabel('Search item types'), '');
    expect(h.byText('No matching item types.')).toBeUndefined();
    expect(h.byLabel('Medicine')?.props.accessibilityState.checked).toBe(true);
    expect(changes).toEqual([]);
  } finally { await h.unmount(); }
});

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

it.each(['disabled', 'draft', 'asset', 'settings', 'visit', 'unmount'] as const)('ignores a type-change confirmation after %s changes', async change => {
  const h = new MobileRenderHarness();
  const results: EditDraft[] = [];
  const asset = { id: 'item', title: 'Medicine', description: '', expiration: { date: '2028-02', precision: 'month' as const } };
  const type = { kind: 'asset-type' as const, id: 'medicine', key: 'medicine', displayName: 'Medicine', description: '', tenantId: 'tenant', scope: 'inventory' as const, inventoryId: 'inventory', lifecycle: 'active' as const, expirationEnabled: true };
  let draft = { title: 'Keep name', description: 'Keep notes', tagIds: ['tag'] };
  const render = () => h.render(<AssetExpirationEditor asset={change === 'asset' && changed ? { ...asset, id: 'replacement' } : asset} draft={draft} types={change === 'settings' && changed ? [] : [type]} disabled={change === 'disabled' && changed} onChange={value => results.push(value)} />);
  let changed = false;
  try {
    await render();
    await h.press(h.byLabel('Item type'));
    await h.press(h.byLabel('Medicine'));
    const confirm = latestAlert()?.buttons.find(button => button.text === 'Change type')?.onPress;
    expect(confirm).toBeTypeOf('function');
    changed = true;
    if (change === 'draft') draft = { ...draft, title: 'Newer name' };
    if (change === 'unmount') await h.unmount();
    else if (change === 'visit') { await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); }
    else await render();
    await h.run(() => confirm?.());
    expect(results).toEqual([]);
  } finally { await h.unmount(); setScreenFocused(true); }
});

it('applies a current type confirmation once and preserves unrelated edits', async () => {
  const h = new MobileRenderHarness(); const results: EditDraft[] = [];
  try {
    await h.render(<AssetExpirationEditor asset={{ id: 'item', title: 'Original', description: '', expiration: { date: '2028-02', precision: 'month' } }} draft={{ title: 'Keep name', description: 'Keep notes', tagIds: ['tag'] }} types={[{ kind: 'asset-type', id: 'medicine', key: 'medicine', displayName: 'Medicine', description: '', tenantId: 'tenant', scope: 'inventory', inventoryId: 'inventory', lifecycle: 'active', expirationEnabled: true }]} disabled={false} onChange={value => results.push(value)} />);
    await h.press(h.byLabel('Item type')); await h.press(h.byLabel('Medicine'));
    const confirm = latestAlert()?.buttons.find(button => button.text === 'Change type')?.onPress;
    expect(confirm).toBeTypeOf('function');
    await h.run(() => confirm?.()); await h.run(() => confirm?.());
    expect(results).toEqual([{ title: 'Keep name', description: 'Keep notes', tagIds: ['tag'], customAssetTypeId: 'medicine', expiration: null, expirationValid: true }]);
  } finally { await h.unmount(); }
});
