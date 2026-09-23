import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { MoveSelectionList } from './MoveSelectionList.ios';

it('keeps subject, selection and destination context distinct and rejects stale row actions', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  const render = (disabled = false, owner = 'first') => h.render(<MoveSelectionList
    subjectLabel="Moving" subject="Tent" context="Current location: Hall" title="Destinations"
    rows={[{ id: 'garage', label: 'Garage', context: 'House', kind: 'location', selected: true, disabled,
      accessibilityLabel: 'Choose destination Garage', onPress: () => calls.push(owner) }]} />);
  try {
    await render();
    expect(h.allText()).toContain('Tent'); expect(h.allText()).toContain('Current location: Hall');
    expect(h.allByType('SwiftUISection').map(node => node.props.title)).toEqual(['Moving', 'Destinations']);
    const row = h.byType('SwiftUIButton')!;
    expect(row.props.modifiers).toContainEqual({ type: 'accessibilityValue', value: 'Selected' });
    const press = row.props.onPress;
    await render(true); await h.run(press); expect(calls).toEqual([]);
    await render(false, 'current'); await h.run(press); expect(calls).toEqual(['current']);
    await h.unmount(); await h.run(press); expect(calls).toEqual(['current']);
  } finally { await h.unmount(); }
});
