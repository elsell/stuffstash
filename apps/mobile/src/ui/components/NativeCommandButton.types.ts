export type NativeCommandButtonProps = {
  readonly prominence?: 'standard' | 'secondary' | 'primary';
  readonly role?: 'default' | 'destructive';
  readonly label: string;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};
