import React, { type ComponentProps } from 'react';
import { Button, Host, Image, Menu, Section } from '@expo/ui/swift-ui';
import {
  accessibilityLabel as nativeAccessibilityLabel,
  accessibilityValue as nativeAccessibilityValue,
  buttonStyle,
  controlSize,
  contentShape,
  frame,
  shapes,
  disabled as nativeDisabled,
  labelStyle,
  tint
} from '@expo/ui/swift-ui/modifiers';
import { StyleSheet, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { actionableMenuGroups, nativeMenuItemPresentation } from './NativeActionMenuPresentation';
import { useNativeMenuAction } from './useNativeMenuAction';
import type { NativeActionMenuProps } from './NativeActionMenu.types';

export type { NativeActionMenuGroup, NativeActionMenuItem, NativeActionMenuProps, NativeActionMenuTrigger } from './NativeActionMenu.types';

type SwiftButtonImage = ComponentProps<typeof Button>['systemImage'];
type SwiftImageName = ComponentProps<typeof Image>['systemName'];

export function NativeActionMenu({ accessibilityLabel, disabled = false, tone = 'standard', groups, trigger = { kind: 'ellipsis' } }: NativeActionMenuProps) {
  const palette = useAppearanceAwarePalette();
  const actionableGroups = actionableMenuGroups(groups);
  const menuDisabled = disabled || actionableGroups.length === 0;
  const { pressItem } = useNativeMenuAction(groups, menuDisabled);
  const menuLabel = trigger.kind === 'label'
    ? trigger.label
    : <Image
      color={tone === 'onDark' ? '#FFFFFF' : undefined}
      size={trigger.kind === 'icon' ? 16 : 20}
      systemName={(trigger.kind === 'icon' ? trigger.systemImage : 'ellipsis') as SwiftImageName}
      modifiers={trigger.kind === 'ellipsis' && tone !== 'onDark' ? [frame({ width: 44, height: 44 }), contentShape(shapes.rectangle())] : undefined}
    />;
  const compactTrigger = trigger.kind !== 'label';
  const onDarkSymbol = tone === 'onDark' && compactTrigger;

  return <View
    pointerEvents={menuDisabled ? 'none' : 'auto'}
    style={menuDisabled && styles.disabled}
  >
    <Host matchContents={!compactTrigger} style={onDarkSymbol ? styles.photoHost : compactTrigger ? styles.compactHost : styles.labelHost}>
      <Menu
        label={onDarkSymbol ? accessibilityLabel : menuLabel}
        systemImage={onDarkSymbol ? (trigger.kind === 'icon' ? trigger.systemImage : 'ellipsis') : undefined}
        modifiers={[
          nativeAccessibilityLabel(accessibilityLabel),
          ...(trigger.kind === 'label' || trigger.kind === 'icon' || tone === 'onDark'
            ? [buttonStyle('bordered'), controlSize(onDarkSymbol ? 'large' : trigger.kind === 'icon' ? 'small' : 'regular'), tint(tone === 'onDark' ? '#FFFFFF' : palette.action)]
            : []),
          ...(trigger.kind === 'icon' || onDarkSymbol ? [labelStyle('iconOnly')] : []),
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
                nativeDisabled(menuDisabled || !presentation.enabled)
              ]}
              onPress={() => pressItem(group.id, item.id)}
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
  photoHost: { height: 48, width: 54 },
  labelHost: { height: 44, minWidth: 44 }
});
