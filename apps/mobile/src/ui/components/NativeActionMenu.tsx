import React, { useEffect, useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { actionableMenuGroups } from './NativeActionMenuPresentation';
import { useNativeMenuAction } from './useNativeMenuAction';
import type { NativeActionMenuProps } from './NativeActionMenu.types';

export type { NativeActionMenuGroup, NativeActionMenuItem, NativeActionMenuProps, NativeActionMenuTrigger } from './NativeActionMenu.types';

/** Non-native renderer used by tests and non-mobile targets. iOS and Android use their platform files. */
export function NativeActionMenu({ accessibilityLabel, disabled = false, groups, trigger = { kind: 'ellipsis' } }: NativeActionMenuProps) {
  const [expanded, setExpanded] = useState(false);
  const actionableGroups = actionableMenuGroups(groups);
  const menuDisabled = disabled || actionableGroups.length === 0;
  const { pressItem, trigger: openMenu } = useNativeMenuAction(groups, menuDisabled);
  useEffect(() => { if (menuDisabled) setExpanded(false); }, [menuDisabled]);
  return <View style={styles.anchor}>
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      accessibilityState={{ disabled: menuDisabled, expanded: expanded && !menuDisabled }}
      disabled={menuDisabled}
      onPress={() => openMenu(() => setExpanded((current) => !current))}
      style={[styles.trigger, trigger.kind === 'label' || trigger.kind === 'row' ? null : styles.compactTrigger]}
    >
      <Text>{trigger.kind === 'row' ? `${trigger.label} ${trigger.value ?? ''}` : trigger.kind === 'label' ? trigger.label : trigger.kind === 'icon' ? '⇅' : '•••'}</Text>
    </Pressable>
    {expanded && !menuDisabled ? <View accessibilityRole="menu" style={styles.menu}>
      {actionableGroups.map((group) => <View key={group.id}>
        {group.items.map((item) => <Pressable
          accessibilityRole="menuitem"
          accessibilityState={{ disabled: menuDisabled || Boolean(item.disabled), selected: Boolean(item.isSelected) }}
          disabled={menuDisabled || item.disabled}
          key={item.id}
          onPress={() => pressItem(group.id, item.id, () => setExpanded(false))}
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
