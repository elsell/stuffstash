export const photoViewerActions = {
  close: { label: 'Close photo viewer', symbol: 'xmark' },
  previous: { label: 'Previous photo', symbol: 'chevron.left' },
  next: { label: 'Next photo', symbol: 'chevron.right' },
  remove: { label: 'Remove photo', symbol: 'trash' }
} as const;

export type PhotoViewerActionButtonProps = {
  readonly action: keyof typeof photoViewerActions;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};
