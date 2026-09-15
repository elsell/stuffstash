export type NativeConversationButtonProps = {
  readonly kind: 'record' | 'send' | 'cancel';
  readonly label: string;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};
