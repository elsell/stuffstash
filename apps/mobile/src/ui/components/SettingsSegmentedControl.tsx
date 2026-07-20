import { useAppearancePalette } from '../theme/AppearanceContext';
import { NativeSegmentedControl, type NativeSegment } from './NativeSegmentedControl';

export type SettingsSegment = NativeSegment;

export function SettingsSegmentedControl({
  disabled = false,
  onChange,
  segments,
  value
}: {
  readonly disabled?: boolean;
  readonly onChange: (value: string) => void;
  readonly segments: readonly SettingsSegment[];
  readonly value: string;
}) {
  const colors = useAppearancePalette();
  return <NativeSegmentedControl
    colors={colors}
    disabled={disabled}
    onChange={onChange}
    segments={segments}
    value={value}
  />;
}
