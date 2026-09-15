export type NativeCommandButtonProps = {
  readonly prominence?: 'standard' | 'primary';
  readonly role?: 'default' | 'destructive';
  readonly label: string;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};
