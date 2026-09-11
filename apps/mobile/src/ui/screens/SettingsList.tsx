import type { ReactNode } from 'react';
import { ActivityIndicator, Pressable, Switch, Text, useWindowDimensions, View } from 'react-native';
import { Check, ChevronRight } from 'lucide-react-native';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { createSettingsScreenStyles } from './SettingsScreen.styles';
import { settingsLayoutMode } from './SettingsScreenPresentation';

export function useSettingsListStyles() {
  const palette = useAppearancePalette();
  const { fontScale } = useWindowDimensions();
  const layout = settingsLayoutMode({ fontScale });
  return {
    layout,
    palette,
    styles: createSettingsScreenStyles(palette, { stacked: layout.stacksLabelValueRows })
  };
}

export function SettingsSection({
  children,
  footer,
  title
}: {
  readonly children: ReactNode;
  readonly footer?: string;
  readonly title?: string;
}) {
  const { styles } = useSettingsListStyles();
  return (
    <View style={styles.section}>
      {title ? <Text accessibilityRole="header" style={styles.sectionTitle}>{title}</Text> : null}
      <View style={styles.group}>{children}</View>
      {footer ? <Text style={styles.sectionFooter}>{footer}</Text> : null}
    </View>
  );
}

export function SettingsSeparator() {
  const { styles } = useSettingsListStyles();
  return <View accessibilityElementsHidden importantForAccessibility="no" style={styles.separator} />;
}

export function SettingsLoadingRow({ label }: { readonly label: string }) {
  const { palette, styles } = useSettingsListStyles();
  return <View accessibilityLiveRegion="polite" accessibilityRole="progressbar" style={styles.loadingRow}>
    <ActivityIndicator color={palette.action} />
    <Text style={styles.loadingText}>{label}</Text>
  </View>;
}

export function SettingsNavigationRow({
  accessibilityLabel,
  context,
  disabled = false,
  icon,
  label,
  onPress,
  value
}: {
  readonly accessibilityLabel: string;
  readonly context?: string;
  readonly disabled?: boolean;
  readonly icon?: ReactNode;
  readonly label: string;
  readonly onPress: () => void;
  readonly value?: string;
}) {
  const { palette, styles } = useSettingsListStyles();
  return (
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      accessibilityHint="Opens a settings screen"
      accessibilityValue={value ? { text: value } : undefined}
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [styles.navigationRow, pressed && styles.navigationRowPressed]}
    >
      <View style={styles.navigationRowContent}>
        {icon ? <View style={styles.rowIconFrame}>{icon}</View> : null}
        <View style={styles.rowText}>
          <Text style={styles.rowLabel}>{label}</Text>
          {context ? <Text style={styles.rowContext}>{context}</Text> : null}
        </View>
        <View style={styles.rowTrailing}>
          {value ? <Text style={styles.rowValue}>{value}</Text> : null}
          <ChevronRight color={palette.textMuted} size={18} />
        </View>
      </View>
    </Pressable>
  );
}

export function SettingsValueRow({
  label,
  value
}: {
  readonly label: string;
  readonly value: string;
}) {
  const { styles } = useSettingsListStyles();
  return (
    <View accessibilityLabel={`${label}, ${value}`} style={styles.navigationRow}>
      <View style={styles.navigationRowContent}>
        <Text style={styles.rowLabel}>{label}</Text>
        <Text selectable style={styles.rowValue}>{value}</Text>
      </View>
    </View>
  );
}

export function SettingsActionRow({
  accessibilityLabel,
  disabled = false,
  destructive = false,
  label,
  onPress
}: {
  readonly accessibilityLabel?: string;
  readonly disabled?: boolean;
  readonly destructive?: boolean;
  readonly label: string;
  readonly onPress: () => void;
}) {
  const { styles } = useSettingsListStyles();
  return (
    <Pressable
      accessibilityLabel={accessibilityLabel ?? label}
      accessibilityRole="button"
      accessibilityState={{ busy: disabled, disabled }}
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.actionRow,
        pressed && !disabled && styles.navigationRowPressed,
        disabled && { opacity: 0.55 }
      ]}
    >
      <Text style={destructive ? styles.dangerText : styles.actionText}>{label}</Text>
    </Pressable>
  );
}


export function SettingsSwitchRow({ label, value, disabled = false, onValueChange }: {
  readonly label: string; readonly value: boolean; readonly disabled?: boolean;
  readonly onValueChange: (value: boolean) => void;
}) {
  const { styles } = useSettingsListStyles();
  return <View style={styles.navigationRow}><View style={styles.navigationRowContent}>
    <Text style={[styles.rowLabel, styles.rowText]}>{label}</Text>
    <Switch accessibilityLabel={label} value={value} disabled={disabled} onValueChange={onValueChange} />
  </View></View>;
}

export function SettingsChoiceRow({ label, accessibilityLabel = label, selected, disabled = false, onPress }: {
  readonly label: string; readonly accessibilityLabel?: string; readonly selected: boolean;
  readonly disabled?: boolean; readonly onPress: () => void;
}) {
  const { palette, styles } = useSettingsListStyles();
  return <Pressable accessibilityRole="radio" accessibilityLabel={accessibilityLabel} accessibilityState={{ checked: selected, disabled }} disabled={disabled} onPress={onPress}
    style={({ pressed }) => [styles.choiceRow, pressed && !disabled && styles.navigationRowPressed, disabled && { opacity: 0.55 }]}>
    <View style={styles.navigationRowContent}><Text style={[styles.rowLabel, styles.rowText]}>{label}</Text>
      {selected ? <Check size={20} color={palette.action} /> : null}
    </View>
  </Pressable>;
}
