export type NativeCommandButtonProps = {
  readonly prominence?: 'standard' | 'primary';
  readonly label: string;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};
