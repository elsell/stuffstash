import SegmentedControl from '@expo/ui/community/segmented-control';
import type { StyleProp, ViewStyle } from 'react-native';
import type { MobileColorPalette } from '../theme/tokens';

export type NativeSegment<Value extends string = string> = {
  readonly label: string;
  readonly value: Value;
};

export function NativeSegmentedControl<Value extends string>({
  colors,
  disabled = false,
  onChange,
  segments,
  style,
  value
}: {
  readonly colors: MobileColorPalette;
  readonly disabled?: boolean;
  readonly onChange: (value: Value) => void;
  readonly segments: readonly NativeSegment<Value>[];
  readonly style?: StyleProp<ViewStyle>;
  readonly value: Value;
}) {
  const selectedIndex = Math.max(0, segments.findIndex((segment) => segment.value === value));

  return <SegmentedControl
    enabled={!disabled}
    onValueChange={(label) => {
      const segment = segments.find((candidate) => candidate.label === label);
      if (segment) onChange(segment.value);
    }}
    selectedIndex={selectedIndex}
    style={[{ minHeight: 44 }, style]}
    tintColor={colors.selected}
    values={segments.map((segment) => segment.label)}
  />;
}
