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
    expect(h.allText()).not.toContain('Moving'); expect(h.allText()).toContain('Destinations');
    const row = h.byType('SwiftUIButton')!;
    expect(row.props.modifiers).toContainEqual({ type: 'accessibilityValue', value: 'Selected' });
    const press = row.props.onPress;
    await render(true); await h.run(press); expect(calls).toEqual([]);
    await render(false, 'current'); await h.run(press); expect(calls).toEqual(['current']);
    await h.unmount(); await h.run(press); expect(calls).toEqual(['current']);
  } finally { await h.unmount(); }
});

it('keeps recovery inside the native list and retires removed retry actions', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  const render = (owner?: string) => h.render(<MoveSelectionList subjectLabel="Destination" subject="Garage"
    context="Choose an item" title="Items" rows={[]} statuses={owner ? [{ message: 'Suggestions unavailable',
      retry: { label: 'Retry suggestions', onPress: () => calls.push(owner) } }] : []} />);
  try {
    await render('first');
    const retry = h.byType('SwiftUIButton'); expect(retry).toBeDefined();
    let parent = retry!.parent;
    while (parent && parent.type !== 'SwiftUIList') parent = parent.parent;
    expect(parent).toBeDefined();
    expect(h.allText()).toContain('Suggestions unavailable');
    const press = retry!.props.onPress;
    await render('current'); await h.run(press); expect(calls).toEqual(['current']);
    await render(); await h.run(press); expect(calls).toEqual(['current']);
  } finally { await h.unmount(); }
});


it('keeps filtered-out selection available beside destination context without committing it', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  try {
    await h.render(<MoveSelectionList subjectLabel="Moving" subject="Tent"
      context="Current location: Hall" title="Destinations" rows={[]}
      retainedSelection={{ id: 'shed', label: 'Shed', context: 'Garden', kind: 'location',
        selected: true, accessibilityLabel: 'Choose destination Shed', onPress: () => calls.push('shed') }} />);
    expect(h.allText()).toContain('Tent');
    expect(h.allText()).toContain('Current location: Hall');
    expect(h.allText()).toContain('Shed');
    expect(calls).toEqual([]);
    await h.run(h.byType('SwiftUIButton')!.props.onPress);
    expect(calls).toEqual(['shed']);
  } finally { await h.unmount(); }
});


it('keeps the Move to summary stable when search hides the selected destination', async () => {
  const h = new MobileRenderHarness();
  const row = { id: 'shed', label: 'Shed', context: 'Garden', kind: 'location' as const,
    selected: true, accessibilityLabel: 'Choose destination Shed', onPress: () => {} };
  const render = (filtered: boolean) => h.render(<MoveSelectionList subjectLabel="Moving" subject="Tent"
    context="Current location: Hall" destinationLabel="Garden / Shed" title="Destinations"
    rows={filtered ? [] : [row]} retainedSelection={filtered ? row : undefined} />);
  try {
    await render(false);
    expect(h.allText()).toContain('Move to: Garden / Shed');
    expect(h.allText()).toContain('Move to: Garden / Shed');
    await render(true);
    expect(h.allText()).toContain('Move to: Garden / Shed');
    expect(h.allText()).toContain('Move to: Garden / Shed');
    expect(h.allText()).not.toContain('Selected');
  } finally { await h.unmount(); }
});
