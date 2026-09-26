import { View } from 'react-native';
import { NativeCommandButton } from './NativeCommandButton.ios';
import type { NativeSheetActionsProps } from './NativeSheetActions.types';

/** Containers own safe areas and keyboard overlap; native commands measure their padded height. */
export function NativeSheetActions({ primaryLabel, primaryAccessibilityLabel = primaryLabel, secondaryLabel, secondaryAccessibilityLabel = secondaryLabel, disabled, secondaryDisabled = false, onApply, onBack }: NativeSheetActionsProps) {
  return <View style={{ width: '100%', gap: 8 }}>
    <NativeCommandButton label={primaryLabel} accessibilityLabel={primaryAccessibilityLabel}
      prominence="primary" fullWidth disabled={disabled} onPress={onApply} />
    <NativeCommandButton label={secondaryLabel} accessibilityLabel={secondaryAccessibilityLabel}
      prominence="standard" fullWidth disabled={secondaryDisabled} onPress={onBack} />
  </View>;
}
