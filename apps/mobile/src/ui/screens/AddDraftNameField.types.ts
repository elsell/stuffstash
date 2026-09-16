import type { TextInputProps } from 'react-native';

export type AddDraftNameFieldProps = Pick<TextInputProps, 'editable' | 'placeholder' | 'placeholderTextColor' | 'style'> & {
  readonly accessibilityLabel: string;
  readonly value: string;
  readonly onChangeText: (value: string) => void;
};
