import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { VoiceResponseEntityText } from './VoiceResponseEntityText';

it('renders formatted list text with native asset links and content-sized message layout', async () => {
  const h = new MobileRenderHarness();
  const reference = { type: 'asset_reference' as const, assetId: 'baby-clothes', title: 'Newborn Clothes', assetKind: 'item' as const };
  const opened: string[] = [];
  try {
    await h.render(<VoiceResponseEntityText markdown enabled text={'Found:\n* **Newborn** Clothes: Bin 58.'} references={[reference]} onOpen={asset => opened.push(asset.assetId)} />);
    expect(h.allByType('Text').some(node => node.props.selectable === true)).toBe(true);
    expect(h.allText().join('')).not.toContain('**');
    expect(h.allText()).toContain('• ');
    await h.press(h.byLabel('Open Newborn Clothes'));
    expect(opened).toEqual(['baby-clothes']);
    const group = h.byType('View');
    expect(group?.props.style.flex).toBeUndefined();
    expect(group?.props.style.flexGrow ?? 0).toBe(0);
    await h.render(<VoiceResponseEntityText enabled showFallbackReferences={false} text="Where are my baby clothes?" references={[reference]} onOpen={asset => opened.push(asset.assetId)} />);
    expect(h.allByType('Pressable')).toHaveLength(0);
    expect(h.allText().join('')).toContain('Where are my baby clothes?');
  } finally { await h.unmount(); }
});
