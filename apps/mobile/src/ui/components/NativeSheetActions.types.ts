export interface NativeSheetActionsProps {
  readonly primaryLabel: string;
  readonly primaryAccessibilityLabel?: string;
  readonly secondaryAccessibilityLabel?: string;
  readonly secondaryLabel: string;
  readonly disabled: boolean;
  readonly secondaryDisabled?: boolean;
  readonly onApply: () => void;
  readonly onBack: () => void;
}
