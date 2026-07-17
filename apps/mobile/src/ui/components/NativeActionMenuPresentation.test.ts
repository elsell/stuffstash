import { describe, expect, it, vi } from 'vitest';
import { actionableMenuGroups, nativeMenuItemPresentation, pressNativeMenuItem } from './NativeActionMenuPresentation';
import type { NativeActionMenuGroup, NativeActionMenuItem } from './NativeActionMenu.types';

function item(overrides: Partial<NativeActionMenuItem> = {}): NativeActionMenuItem {
  return { id: 'edit', label: 'Edit', onPress: vi.fn(), ...overrides };
}

describe('NativeActionMenu presentation', () => {
  it('preserves group and item order while omitting empty groups', () => {
    const groups: readonly NativeActionMenuGroup[] = [
      { id: 'primary', items: [item({ id: 'history', label: 'History' }), item()] },
      { id: 'empty', items: [] },
      { id: 'danger', items: [item({ id: 'archive', label: 'Archive', isDestructive: true })] }
    ];

    expect(actionableMenuGroups(groups).map((group) => [group.id, ...group.items.map((entry) => entry.id)])).toEqual([
      ['primary', 'history', 'edit'],
      ['danger', 'archive']
    ]);
  });

  it('maps selected and destructive semantics to native presentation', () => {
    expect(nativeMenuItemPresentation(item({ isSelected: true, isDestructive: true, systemImage: 'clock' }))).toEqual({
      enabled: true,
      role: 'destructive',
      selectionAccessibilityValue: 'Selected',
      systemImage: 'checkmark'
    });
    expect(nativeMenuItemPresentation(item({ isSelected: false, systemImage: 'clock' }))).toEqual({
      enabled: true,
      role: 'default',
      selectionAccessibilityValue: 'Not selected',
      systemImage: 'clock'
    });
    expect(nativeMenuItemPresentation(item({ disabled: true, systemImage: 'clock' }))).toEqual({
      enabled: false,
      role: 'default',
      selectionAccessibilityValue: undefined,
      systemImage: 'clock'
    });
  });

  it('does not invoke disabled items and invokes enabled items once', () => {
    const onPress = vi.fn();
    pressNativeMenuItem(item({ disabled: true, onPress }));
    expect(onPress).not.toHaveBeenCalled();

    pressNativeMenuItem(item({ onPress }));
    expect(onPress).toHaveBeenCalledOnce();
  });
});
