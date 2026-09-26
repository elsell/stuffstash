import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeActionRow } from './NativeActionRow';
import { ContainedSpatialActions } from './AssetContainedWorkspace';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';

it('uses current row command and rejects retained callbacks after disabling or removal', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  const render = (name: string, disabled = false) => h.render(<NativeActionRow label="Reset all" disabled={disabled} onPress={() => calls.push(name)} />);
  try {
    await render('old'); const press = h.byLabel('Reset all')!.props.onPress;
    await render('current'); await h.run(press); expect(calls).toEqual(['current']);
    await render('disabled', true); await h.run(press); expect(calls).toEqual(['current']);
    await h.unmount(); press(); expect(calls).toEqual(['current']);
  } finally { await h.unmount(); }
});

it('offers both contents commands through one Add menu and preserves disabled guards', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  const asset = { kind: 'container', canContainAssets: true, canAddContainedAssets: true } as AssetDetailViewModel;
  const render = (pending = false) => h.render(<ContainedSpatialActions asset={asset} isActionPending={pending}
    onAddHere={() => calls.push('new')} onMoveThingsHere={() => calls.push('move')} />);
  try {
    await render(); expect(h.byText('Add item here')).toBeUndefined();
    await h.press(h.byLabel('Add to contents')); await h.press(h.byText('Add item here')?.parent ?? undefined);
    await h.press(h.byLabel('Add to contents')); const move = h.byText('Move items here')?.parent ?? undefined;
    expect(move).toBeDefined(); await h.press(move); expect(calls).toEqual(['new', 'move']);
    await render(true); await h.press(move); expect(calls).toEqual(['new', 'move']);
  } finally { await h.unmount(); }
});
