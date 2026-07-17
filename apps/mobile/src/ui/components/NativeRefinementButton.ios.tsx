import React, { type ComponentProps } from 'react';
import { Button, Host } from '@expo/ui/swift-ui';
import {
  accessibilityLabel as nativeAccessibilityLabel,
  accessibilityValue as nativeAccessibilityValue,
  buttonStyle,
  controlSize,
  disabled as nativeDisabled,
  labelStyle,
  tint
} from '@expo/ui/swift-ui/modifiers';
import { StyleSheet, Text, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { NativeRefinementButtonProps } from './NativeRefinementButton.types';

export type { NativeRefinementButtonProps } from './NativeRefinementButton.types';

type SwiftButtonImage = ComponentProps<typeof Button>['systemImage'];

export function NativeRefinementButton({
  accessibilityLabel,
  accessibilityState,
  badgeCount,
  disabled = false,
  iconOnly = false,
  label,
  onPress,
  systemImage
}: NativeRefinementButtonProps) {
  const palette = useAppearanceAwarePalette();
  return <View pointerEvents={disabled ? 'none' : 'auto'} style={iconOnly ? styles.iconRoot : undefined}>
    <Host matchContents={!iconOnly} style={[styles.host, iconOnly ? styles.iconHost : null]}>
      <Button
        label={label}
        modifiers={[
          nativeAccessibilityLabel(accessibilityLabel),
          ...(accessibilityState?.expanded === undefined
            ? []
            : [nativeAccessibilityValue(accessibilityState.expanded ? 'Expanded' : 'Collapsed')]),
          buttonStyle('bordered'),
          controlSize('regular'),
          ...(iconOnly ? [labelStyle('iconOnly')] : []),
          tint(palette.action),
          nativeDisabled(disabled)
        ]}
        onPress={onPress}
        systemImage={systemImage as SwiftButtonImage}
      />
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
