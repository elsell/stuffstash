export type NativeReadStateButtonProps = {
  readonly read: boolean;
  readonly label: string;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};
