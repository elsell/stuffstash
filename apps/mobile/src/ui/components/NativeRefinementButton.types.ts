import type { AccessibilityState } from 'react-native';

export type NativeRefinementButtonProps = {
  readonly accessibilityLabel: string;
  readonly accessibilityState?: AccessibilityState;
  readonly badgeCount?: number;
  readonly disabled?: boolean;
  readonly iconOnly?: boolean;
  readonly label: string;
  readonly onPress: () => void;
  /** SF Symbol used by the iOS native button. */
  readonly systemImage?: string;
};
