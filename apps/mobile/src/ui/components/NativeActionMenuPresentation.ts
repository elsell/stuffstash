import type { NativeActionMenuGroup, NativeActionMenuItem } from './NativeActionMenu.types';

export function actionableMenuGroups(groups: readonly NativeActionMenuGroup[]): readonly NativeActionMenuGroup[] {
  return groups.filter((group) => group.items.length > 0);
}

export function nativeMenuItemPresentation(item: NativeActionMenuItem): {
  readonly enabled: boolean;
  readonly role: 'default' | 'destructive';
  readonly selectionAccessibilityValue: 'Selected' | 'Not selected' | undefined;
  readonly systemImage?: string;
} {
  return {
    enabled: !item.disabled,
    role: item.isDestructive ? 'destructive' : 'default',
    selectionAccessibilityValue: item.isSelected === undefined
      ? undefined
      : item.isSelected ? 'Selected' : 'Not selected',
    systemImage: item.isSelected ? 'checkmark' : item.systemImage
  };
}

export function pressNativeMenuItem(item: NativeActionMenuItem): void {
  if (!item.disabled) item.onPress();
}
