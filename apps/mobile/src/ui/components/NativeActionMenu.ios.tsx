import React, { type ComponentProps } from 'react';
import { Button, Host, Image, Menu, Section } from '@expo/ui/swift-ui';
import {
  accessibilityLabel as nativeAccessibilityLabel,
  accessibilityValue as nativeAccessibilityValue,
  buttonStyle,
  controlSize,
  disabled as nativeDisabled,
  labelStyle,
  tint
} from '@expo/ui/swift-ui/modifiers';
import { StyleSheet, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { actionableMenuGroups, nativeMenuItemPresentation, pressNativeMenuItem } from './NativeActionMenuPresentation';
import type { NativeActionMenuProps } from './NativeActionMenu.types';

export type { NativeActionMenuGroup, NativeActionMenuItem, NativeActionMenuProps, NativeActionMenuTrigger } from './NativeActionMenu.types';

type SwiftButtonImage = ComponentProps<typeof Button>['systemImage'];
type SwiftImageName = ComponentProps<typeof Image>['systemName'];

export function NativeActionMenu({ accessibilityLabel, disabled = false, groups, trigger = { kind: 'ellipsis' } }: NativeActionMenuProps) {
  const palette = useAppearanceAwarePalette();
  const actionableGroups = actionableMenuGroups(groups);
  const menuDisabled = disabled || actionableGroups.length === 0;
  const menuLabel = trigger.kind === 'label'
    ? trigger.label
    : <Image size={trigger.kind === 'icon' ? 16 : 20} systemName={(trigger.kind === 'icon' ? trigger.systemImage : 'ellipsis') as SwiftImageName} />;
  const compactTrigger = trigger.kind !== 'label';

  return <View
    pointerEvents={menuDisabled ? 'none' : 'auto'}
    style={menuDisabled && styles.disabled}
  >
    <Host matchContents={!compactTrigger} style={compactTrigger ? styles.compactHost : styles.labelHost}>
      <Menu
        label={menuLabel}
        modifiers={[
          nativeAccessibilityLabel(accessibilityLabel),
          ...(trigger.kind === 'label' || trigger.kind === 'icon'
            ? [buttonStyle('bordered'), controlSize(trigger.kind === 'icon' ? 'small' : 'regular'), tint(palette.action)]
            : []),
          ...(trigger.kind === 'icon' ? [labelStyle('iconOnly')] : []),
          nativeDisabled(menuDisabled)
        ]}
      >
        {actionableGroups.map((group) => <Section key={group.id}>
          {group.items.map((item) => {
            const presentation = nativeMenuItemPresentation(item);
            return <Button
              key={item.id}
              label={item.label}
              modifiers={[
                ...(presentation.selectionAccessibilityValue
                  ? [nativeAccessibilityValue(presentation.selectionAccessibilityValue)]
                  : []),
                nativeDisabled(!presentation.enabled)
              ]}
              onPress={() => pressNativeMenuItem(item)}
              role={presentation.role}
              systemImage={presentation.systemImage as SwiftButtonImage}
            />;
          })}
        </Section>)}
      </Menu>
    </Host>
  </View>;
}

const styles = StyleSheet.create({
  disabled: { opacity: 0.5 },
  compactHost: { height: 44, width: 44 },
  labelHost: { height: 44, minWidth: 44 }
});
