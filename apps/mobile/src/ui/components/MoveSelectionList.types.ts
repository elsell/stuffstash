export type MoveSelectionRowModel = {
  readonly id: string; readonly label: string; readonly context: string;
  readonly kind: 'location' | 'container' | 'item' | 'root';
  readonly selected: boolean; readonly disabled?: boolean;
  readonly accessibilityLabel: string; readonly onPress: () => void;
};
export type MoveSelectionStatus = {
  readonly title?: string;
  readonly message: string;
  readonly retry?: { readonly label: string; readonly disabled?: boolean; readonly onPress: () => void };
};
export type MoveSelectionListProps = {
  readonly subjectLabel: string; readonly subject: string; readonly context: string;
  readonly title: string; readonly rows: readonly MoveSelectionRowModel[];
  readonly retainedSelection?: MoveSelectionRowModel; readonly statuses?: readonly MoveSelectionStatus[];
};
