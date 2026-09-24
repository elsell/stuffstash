export type NativeCommandButtonProps = {
  readonly prominence?: 'standard' | 'secondary' | 'primary';
  readonly role?: 'default' | 'destructive';
  readonly label: string;
  readonly accessibilityLabel?: string;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};
