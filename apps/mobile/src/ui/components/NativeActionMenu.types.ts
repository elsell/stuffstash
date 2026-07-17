export type NativeActionMenuTrigger =
  | { readonly kind: 'ellipsis' }
  | { readonly kind: 'label'; readonly label: string }
  | {
      readonly androidIcon: 'sort';
      readonly kind: 'icon';
      readonly systemImage: string;
    };

export type NativeActionMenuItem = {
  readonly id: string;
  readonly label: string;
  readonly systemImage?: string;
  readonly isDestructive?: boolean;
  readonly isSelected?: boolean;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};

export type NativeActionMenuGroup = {
  readonly id: string;
  readonly items: readonly NativeActionMenuItem[];
};

export type NativeActionMenuProps = {
  readonly accessibilityLabel: string;
  readonly disabled?: boolean;
  readonly groups: readonly NativeActionMenuGroup[];
  readonly trigger?: NativeActionMenuTrigger;
};
