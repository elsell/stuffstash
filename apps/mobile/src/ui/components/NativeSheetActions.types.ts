export interface NativeSheetActionsProps {
  /** Fixed for this host's lifetime; container means it already applies keyboard overlap. */
  readonly keyboardAvoidance?: 'native' | 'container';
  readonly primaryLabel: string;
  readonly primaryAccessibilityLabel?: string;
  readonly secondaryAccessibilityLabel?: string;
  readonly secondaryLabel: string;
  readonly disabled: boolean;
  readonly onApply: () => void;
  readonly onBack: () => void;
}
