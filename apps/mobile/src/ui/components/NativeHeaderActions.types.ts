import type { StackScreenProps } from 'expo-router';
type NonFunction<T> = T extends (...args: any[]) => unknown ? never : T;
export type HeaderOptions = NonFunction<NonNullable<StackScreenProps['options']>>;
export type NativeHeaderAction = {
  readonly kind: 'notifications' | 'add' | 'account' | 'close' | 'save' | 'settings' | 'mark-read';
  readonly label: string;
  readonly disabled?: boolean;
  readonly badgeCount?: number;
  readonly onPress: () => void;
};
