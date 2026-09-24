import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { AssetCard } from './AssetCard';

it('reserves peer status space without announcing a false checkout', async () => {
 const h = new MobileRenderHarness();
 const asset = { id: 'plain', title: 'Tent', kindLabel: 'Item', description: '', hasPhoto: false,
   locationTrailLabel: '', parentLocationTrail: [], updatedAtLabel: '', photoLabel: '', imagePlaceholderLabel: 'Item' };
 const render = (checkedOutLabel?: string) => <AssetCard asset={{...asset, checkedOutLabel}}
   reservedCheckoutLabel="Checked out" onPress={()=>{}} onParentLocationPress={()=>{}} />;
 try {
  await h.render(render());
  const slot = h.allByType('Text').find(node=>node.children.includes('Checked out'));
  expect(slot).toBeDefined();
  expect(slot?.props.accessibilityElementsHidden).toBe(true);
  expect(slot?.props.importantForAccessibility).toBe('no-hide-descendants');
  await h.render(render('Checked out'));
  const status = h.allByType('Text').find(node=>node.children.includes('Checked out'));
  expect(status?.props.accessibilityElementsHidden).not.toBe(true);
 } finally { await h.unmount(); }
});
