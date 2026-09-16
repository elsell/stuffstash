import { NativeComposeHost as Host } from './NativeComposeHost.android';
import React, { Fragment, useState } from 'react';
import { DropdownMenu, DropdownMenuItem, HorizontalDivider, Icon, OutlinedButton, Text, TextButton } from '@expo/ui/jetpack-compose';
import { selectable, size } from '@expo/ui/jetpack-compose/modifiers';
import { StyleSheet, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { minimumTouchTargetSize } from '../theme/tokens';
import { actionableMenuGroups, nativeMenuItemPresentation, pressNativeMenuItem } from './NativeActionMenuPresentation';
import type { NativeActionMenuProps } from './NativeActionMenu.types';

export type { NativeActionMenuGroup, NativeActionMenuItem, NativeActionMenuProps, NativeActionMenuTrigger } from './NativeActionMenu.types';

export function NativeActionMenu({ accessibilityLabel, disabled = false, groups, trigger = { kind: 'ellipsis' } }: NativeActionMenuProps) {
  const palette = useAppearanceAwarePalette();
  const actionableGroups = actionableMenuGroups(groups);
  const menuDisabled = disabled || actionableGroups.length === 0;
  const [expanded, setExpanded] = useState(false);
  const TriggerButton = trigger.kind === 'ellipsis' ? TextButton : OutlinedButton;
  const compactTrigger = trigger.kind !== 'label';

  return <View
    accessible
    accessibilityLabel={accessibilityLabel}
    accessibilityRole="button"
    accessibilityState={{ disabled: menuDisabled, expanded }}
    onAccessibilityTap={() => {
      if (!menuDisabled) setExpanded(true);
    }}
    pointerEvents={menuDisabled ? 'none' : 'auto'}
    style={[styles.wrapper, menuDisabled && styles.disabled]}
  >
    <Host matchContents={!compactTrigger} style={compactTrigger ? styles.compactHost : styles.labelHost}>
      <DropdownMenu expanded={expanded} onDismissRequest={() => setExpanded(false)}>
        <DropdownMenu.Trigger>
          <TriggerButton
            colors={{ contentColor: palette.action, disabledContentColor: palette.textMuted }}
            contentPadding={{ start: 12, top: 10, end: 12, bottom: 10 }}
            enabled={!menuDisabled}
            modifiers={trigger.kind === 'icon' ? [size(minimumTouchTargetSize, minimumTouchTargetSize)] : undefined}
            onClick={() => setExpanded(true)}
          >
            {trigger.kind === 'icon'
              ? <Icon size={18} source={require('./android-icons/sort-arrows.xml')} tint={palette.action} />
              : <Text style={{ fontSize: trigger.kind === 'ellipsis' ? 24 : 14, fontWeight: '600' }}>
                {trigger.kind === 'label' ? trigger.label : '⋮'}
              </Text>}
          </TriggerButton>
        </DropdownMenu.Trigger>
        <DropdownMenu.Items>
          {actionableGroups.map((group, groupIndex) => <Fragment key={group.id}>
            {groupIndex > 0 ? <HorizontalDivider /> : null}
            {group.items.map((item) => {
              const presentation = nativeMenuItemPresentation(item);
              const chooseItem = () => { setExpanded(false); pressNativeMenuItem(item); };
              return <DropdownMenuItem
                elementColors={{
                  disabledTextColor: palette.textMuted,
                  textColor: item.isDestructive ? palette.danger : palette.text
                }}
                enabled={presentation.enabled}
                key={item.id}
                modifiers={item.isSelected === undefined
                  ? undefined
                  : [selectable(item.isSelected, chooseItem, 'radioButton')]}
                onClick={chooseItem}
              >
                <DropdownMenuItem.Text><Text>{item.label}</Text></DropdownMenuItem.Text>
                {item.isSelected ? <DropdownMenuItem.TrailingIcon><Text color={palette.action}>✓</Text></DropdownMenuItem.TrailingIcon> : null}
              </DropdownMenuItem>;
            })}
          </Fragment>)}
        </DropdownMenu.Items>
      </DropdownMenu>
    </Host>
  </View>;
}

const styles = StyleSheet.create({
  wrapper: { alignSelf: 'flex-start' },
  disabled: { opacity: 0.5 },
  compactHost: { height: minimumTouchTargetSize, width: minimumTouchTargetSize },
  labelHost: { height: minimumTouchTargetSize, minWidth: minimumTouchTargetSize }
});
