import React, { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { actionableMenuGroups, pressNativeMenuItem } from './NativeActionMenuPresentation';
import type { NativeActionMenuProps } from './NativeActionMenu.types';

export type { NativeActionMenuGroup, NativeActionMenuItem, NativeActionMenuProps, NativeActionMenuTrigger } from './NativeActionMenu.types';

/** Non-native renderer used by tests and non-mobile targets. iOS and Android use their platform files. */
export function NativeActionMenu({ accessibilityLabel, disabled = false, groups, trigger = { kind: 'ellipsis' } }: NativeActionMenuProps) {
  const [expanded, setExpanded] = useState(false);
  const actionableGroups = actionableMenuGroups(groups);
  const menuDisabled = disabled || actionableGroups.length === 0;
  return <View style={styles.anchor}>
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      accessibilityState={{ disabled: menuDisabled, expanded }}
      disabled={menuDisabled}
      onPress={() => setExpanded((current) => !current)}
      style={[styles.trigger, trigger.kind === 'label' ? null : styles.compactTrigger]}
    >
      <Text>{trigger.kind === 'label' ? trigger.label : trigger.kind === 'icon' ? '⇅' : '•••'}</Text>
    </Pressable>
    {expanded ? <View accessibilityRole="menu" style={styles.menu}>
      {actionableGroups.map((group) => <View key={group.id}>
        {group.items.map((item) => <Pressable
          accessibilityRole="menuitem"
          accessibilityState={{ disabled: Boolean(item.disabled), selected: Boolean(item.isSelected) }}
          disabled={item.disabled}
          key={item.id}
          onPress={() => { setExpanded(false); pressNativeMenuItem(item); }}
          style={styles.item}
        >
          <Text>{item.label}</Text>
        </Pressable>)}
      </View>)}
    </View> : null}
  </View>;
}

const styles = StyleSheet.create({
  anchor: { position: 'relative' },
  compactTrigger: { height: 44, width: 44 },
  item: { minHeight: 44, paddingHorizontal: 16, paddingVertical: 10 },
  menu: { position: 'absolute', right: 0, top: 44, zIndex: 1 },
  trigger: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 44 }
});
