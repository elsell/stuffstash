import { Building2, Package } from 'lucide-react-native';
import type {
  StyleProp,
  TextStyle,
  ViewStyle
} from 'react-native';
import {
  StyleSheet,
  Text,
  View
} from 'react-native';
import { spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';

export type IdentityKind = 'tenant' | 'inventory';

type IdentityIconSize = 'xs' | 'sm' | 'md';

type IdentityIconProps = {
  readonly kind: IdentityKind;
  readonly size?: IdentityIconSize;
  readonly style?: StyleProp<ViewStyle>;
};

type IdentityLabelProps = {
  readonly kind: IdentityKind;
  readonly label: string;
  readonly iconSize?: IdentityIconSize;
  readonly numberOfLines?: number;
  readonly style?: StyleProp<ViewStyle>;
  readonly textStyle?: StyleProp<TextStyle>;
};

const iconSizes: Record<IdentityIconSize, { readonly frame: number; readonly icon: number; readonly radius: number }> = {
  xs: { frame: 22, icon: 13, radius: 11 },
  sm: { frame: 26, icon: 15, radius: 13 },
  md: { frame: 32, icon: 17, radius: 16 }
};

export function IdentityIcon({
  kind,
  size = 'md',
  style
}: IdentityIconProps) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const metrics = iconSizes[size];
  const Icon = kind === 'tenant' ? Building2 : Package;
  const color = kind === 'tenant' ? colors.onAction : colors.accentStrong;

  return (
    <View
      style={[
        styles.iconFrame,
        {
          borderRadius: metrics.radius,
          height: metrics.frame,
          width: metrics.frame
        },
        kind === 'tenant' ? styles.tenantIcon : styles.inventoryIcon,
        style
      ]}
    >
      <Icon color={color} size={metrics.icon} strokeWidth={2.4} />
    </View>
  );
}

export function IdentityLabel({
  kind,
  label,
  iconSize = 'sm',
  numberOfLines = 1,
  style,
  textStyle
}: IdentityLabelProps) {
  const styles = createStyles(useAppearanceAwarePalette());
  return (
    <View style={[styles.labelRow, style]}>
      <IdentityIcon kind={kind} size={iconSize} />
      <Text numberOfLines={numberOfLines} style={[styles.labelText, textStyle]}>
        {label}
      </Text>
    </View>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  iconFrame: {
    alignItems: 'center',
    justifyContent: 'center'
  },
  tenantIcon: {
    backgroundColor: colors.brandCharcoal
  },
  inventoryIcon: {
    backgroundColor: colors.surfaceMuted,
    borderColor: colors.border,
    borderWidth: 1
  },
  labelRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs,
    minWidth: 0
  },
  labelText: {
    flexShrink: 1,
    letterSpacing: 0,
    minWidth: 0
  }
  });
}
