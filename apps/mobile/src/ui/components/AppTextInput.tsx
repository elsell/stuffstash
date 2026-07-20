import { forwardRef } from 'react';
import {
  Platform,
  TextInput,
  type ScrollViewProps,
  type TextInputProps
} from 'react-native';

export const AppTextInput = forwardRef<TextInput, TextInputProps>(function AppTextInput(
  props,
  ref
) {
  return <TextInput {...props} ref={ref} />;
});

export function appKeyboardDismissMode(): NonNullable<ScrollViewProps['keyboardDismissMode']> {
  return Platform.OS === 'ios' ? 'interactive' : 'on-drag';
}
