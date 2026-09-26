import React, { type ComponentProps } from 'react';
import { Button, Host, HStack, Spacer, Text, Image, Menu, Section } from '@expo/ui/swift-ui';
import {
  accessibilityLabel as nativeAccessibilityLabel,
  accessibilityValue as nativeAccessibilityValue,
  buttonStyle,
  controlSize,
  contentShape,
  frame,
  foregroundStyle,
  shapes,
  disabled as nativeDisabled,
  labelStyle,
  tint
} from '@expo/ui/swift-ui/modifiers';
import { StyleSheet, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { actionableMenuGroups, nativeMenuItemPresentation } from './NativeActionMenuPresentation';
import { useNativeMenuAction } from './useNativeMenuAction';
import { NativePhotoSymbol, nativePhotoControlModifiers } from './NativePhotoSymbol.ios';
import type { NativeActionMenuProps } from './NativeActionMenu.types';

export type { NativeActionMenuGroup, NativeActionMenuItem, NativeActionMenuProps, NativeActionMenuTrigger } from './NativeActionMenu.types';

type SwiftButtonImage = ComponentProps<typeof Button>['systemImage'];
type SwiftImageName = ComponentProps<typeof Image>['systemName'];

export function NativeActionMenu({ accessibilityLabel, disabled = false, tone = 'standard', groups, trigger = { kind: 'ellipsis' } }: NativeActionMenuProps) {
  const palette = useAppearanceAwarePalette();
  const actionableGroups = actionableMenuGroups(groups);
  const menuDisabled = disabled || actionableGroups.length === 0;
  const { pressItem } = useNativeMenuAction(groups, menuDisabled);
  const onDarkSymbol = tone === 'onDark' && trigger.kind !== 'label' && trigger.kind !== 'row';
  const menuLabel = onDarkSymbol
    ? <NativePhotoSymbol symbol={(trigger.kind === 'icon' ? trigger.systemImage : 'ellipsis') as SwiftImageName} />
    : trigger.kind === 'row' ? <HStack spacing={8} modifiers={[frame({ minHeight: 48, maxWidth: Infinity }), contentShape(shapes.rectangle())]}><Text modifiers={[foregroundStyle(palette.text)]}>{trigger.label}</Text><Spacer />{trigger.value ? <Text modifiers={[foregroundStyle(palette.textMuted)]}>{trigger.value}</Text> : null}<Image systemName="chevron.up.chevron.down" size={12} /></HStack>
    : trigger.kind === 'label'
    ? trigger.label
    : <Image
      size={trigger.kind === 'icon' ? 16 : 20}
      systemName={(trigger.kind === 'icon' ? trigger.systemImage : 'ellipsis') as SwiftImageName}
      modifiers={trigger.kind === 'ellipsis' ? [frame({ width: 44, height: 44 }), contentShape(shapes.rectangle())] : undefined}
    />;
  const compactTrigger = trigger.kind !== 'label' && trigger.kind !== 'row';

  return <View
    pointerEvents={menuDisabled ? 'none' : 'auto'}
    style={menuDisabled && styles.disabled}
  >
    <Host matchContents={trigger.kind === 'row' ? { vertical: true } : onDarkSymbol || !compactTrigger} style={trigger.kind === 'row' ? { width: '100%', minHeight: 48 } : onDarkSymbol ? undefined : compactTrigger ? styles.compactHost : styles.labelHost}>
      <Menu
        label={menuLabel}
        modifiers={[
          nativeAccessibilityLabel(accessibilityLabel),
          ...(trigger.kind === 'row' ? [buttonStyle('plain')] : onDarkSymbol ? nativePhotoControlModifiers
            : trigger.kind === 'label' || trigger.kind === 'icon'
              ? [buttonStyle('bordered'), controlSize(trigger.kind === 'icon' ? 'small' : 'regular'), tint(tone === 'onDark' ? '#FFFFFF' : palette.action)]
              : []),
          ...(trigger.kind === 'icon' && !onDarkSymbol ? [labelStyle('iconOnly')] : []),
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
  labelHost: { height: 44, minWidth: 44 }
});
