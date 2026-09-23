import type { TextInputProps } from 'react-native';

export type DraftTextFieldProps = Pick<TextInputProps, 'editable' | 'placeholder' | 'placeholderTextColor' | 'style'> & {
  readonly accessibilityLabel: string;
  readonly value: string;
  readonly onChangeText: (value: string) => void;
};
