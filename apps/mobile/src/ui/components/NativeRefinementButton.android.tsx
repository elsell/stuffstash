import { NativeComposeHost as Host } from './NativeComposeHost.android';
import React from 'react';
import { Icon, OutlinedButton, Text as ComposeText } from '@expo/ui/jetpack-compose';
import { size } from '@expo/ui/jetpack-compose/modifiers';
import { StyleSheet, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { minimumTouchTargetSize } from '../theme/tokens';
import { RefinementCountBadge } from './RefinementCountBadge';
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
        modifiers={iconOnly ? [size(minimumTouchTargetSize, minimumTouchTargetSize)] : undefined}
        onClick={onPress}
      >
        {iconOnly
          ? <Icon size={18} source={require('./android-icons/filter-list.xml')} tint={palette.action} />
          : <ComposeText style={{ fontSize: 14, fontWeight: '600' }}>{label}</ComposeText>}
      </OutlinedButton>
    </Host>
    <RefinementCountBadge count={badgeCount} />
  </View>;
}

const styles = StyleSheet.create({
  host: { height: minimumTouchTargetSize, minWidth: minimumTouchTargetSize },
  iconHost: { width: minimumTouchTargetSize },
  iconRoot: { height: minimumTouchTargetSize, position: 'relative', width: minimumTouchTargetSize }
});
