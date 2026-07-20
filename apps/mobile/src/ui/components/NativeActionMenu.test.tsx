import React from 'react';
import { describe, expect, it, vi } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeActionMenu } from './NativeActionMenu';

describe('NativeActionMenu contract', () => {
  it('exposes the trigger label and only runs enabled item callbacks', async () => {
    const harness = new MobileRenderHarness();
    const enabledPress = vi.fn();
    const disabledPress = vi.fn();
    await harness.render(<NativeActionMenu
      accessibilityLabel="Asset actions"
      groups={[{ id: 'actions', items: [
        { id: 'history', label: 'History', isSelected: true, onPress: enabledPress },
        { id: 'archive', label: 'Archive', disabled: true, isDestructive: true, onPress: disabledPress }
      ] }]}
    />);

    const trigger = harness.byLabel('Asset actions');
    expect(trigger?.props.accessibilityRole).toBe('button');
    expect(trigger?.props.accessibilityState).toEqual({ disabled: false, expanded: false });
    await harness.press(trigger);
    const historyItem = harness.all().find((node) => node.props.accessibilityRole === 'menuitem' && node.props.accessibilityState?.selected);
    const archiveItem = harness.all().find((node) => node.props.accessibilityRole === 'menuitem' && node.props.accessibilityState?.disabled);
    expect(historyItem?.props.accessibilityState).toEqual({ disabled: false, selected: true });
    await harness.press(archiveItem);
    expect(disabledPress).not.toHaveBeenCalled();
    await harness.press(historyItem);
    expect(enabledPress).toHaveBeenCalledOnce();
    expect(harness.byLabel('Asset actions')?.props.accessibilityState.expanded).toBe(false);
  });

  it('supports a disabled compact labeled trigger', async () => {
    const harness = new MobileRenderHarness();
    await harness.render(<NativeActionMenu accessibilityLabel="Change sort" disabled groups={[]} trigger={{ kind: 'label', label: 'Sort' }} />);

    expect(harness.byText('Sort')).toBeDefined();
    expect(harness.byLabel('Change sort')?.props.accessibilityState).toEqual({ disabled: true, expanded: false });
  });

  it('renders a semantic sort icon trigger instead of overflow dots', async () => {
    const harness = new MobileRenderHarness();
    await harness.render(<NativeActionMenu
      accessibilityLabel="Sort, recently changed"
      groups={[{ id: 'sort', items: [{ id: 'recent', label: 'Recently changed', onPress: vi.fn() }] }]}
      trigger={{ androidIcon: 'sort', kind: 'icon', systemImage: 'arrow.up.arrow.down' }}
    />);

    expect(harness.byText('⇅')).toBeDefined();
    expect(harness.byText('•••')).toBeUndefined();
    const trigger = harness.byLabel('Sort, recently changed');
    const styles = Array.isArray(trigger?.props.style) ? trigger.props.style : [trigger?.props.style];
    expect(styles).toContainEqual(expect.objectContaining({ height: 44, width: 44 }));
  });
});
