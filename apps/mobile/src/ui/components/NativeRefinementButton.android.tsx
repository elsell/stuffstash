import React from 'react';
import { Host, Icon, OutlinedButton, Text as ComposeText } from '@expo/ui/jetpack-compose';
import { size } from '@expo/ui/jetpack-compose/modifiers';
import { StyleSheet, Text, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { NativeRefinementButtonProps } from './NativeRefinementButton.types';

export type { NativeRefinementButtonProps } from './NativeRefinementButton.types';

export function NativeRefinementButton({
  accessibilityLabel,
  accessibilityState,
  badgeCount,
  disabled = false,
  iconOnly = false,
  label,
  onPress
}: NativeRefinementButtonProps) {
  const palette = useAppearanceAwarePalette();
  return <View
    accessible
    accessibilityLabel={accessibilityLabel}
    accessibilityRole="button"
    accessibilityState={{ ...accessibilityState, disabled }}
    onAccessibilityTap={() => {
      if (!disabled) onPress();
    }}
    pointerEvents={disabled ? 'none' : 'auto'}
    style={iconOnly ? styles.iconRoot : undefined}
  >
    <Host matchContents={!iconOnly} style={[styles.host, iconOnly ? styles.iconHost : null]}>
      <OutlinedButton
        colors={{ contentColor: palette.action, disabledContentColor: palette.textMuted }}
        contentPadding={{ start: 12, top: 10, end: 12, bottom: 10 }}
        enabled={!disabled}
        modifiers={iconOnly ? [size(44, 44)] : undefined}
        onClick={onPress}
      >
        {iconOnly
          ? <Icon size={18} source={require('./android-icons/filter-list.xml')} tint={palette.action} />
          : <ComposeText style={{ fontSize: 14, fontWeight: '600' }}>{label}</ComposeText>}
      </OutlinedButton>
    </Host>
    {badgeCount && badgeCount > 0 ? <View
      accessible={false}
      importantForAccessibility="no-hide-descendants"
      pointerEvents="none"
      style={[styles.badge, { backgroundColor: palette.accent }]}
    >
      <Text style={styles.badgeText}>{badgeCount > 9 ? '9+' : badgeCount.toString()}</Text>
    </View> : null}
  </View>;
}

const styles = StyleSheet.create({
  badge: {
    alignItems: 'center',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 16,
    minWidth: 16,
    paddingHorizontal: 3,
    position: 'absolute',
    right: -3,
    top: -3
  },
  badgeText: { color: '#FFFFFF', fontSize: 10, fontWeight: '700', lineHeight: 12 },
  host: { height: 44, minWidth: 44 },
  iconHost: { width: 44 },
  iconRoot: { height: 44, position: 'relative', width: 44 }
});
