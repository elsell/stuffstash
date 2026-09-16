import type { TextInputProps } from 'react-native';

export type AddAssetNameFieldProps = Pick<TextInputProps, 'editable' | 'placeholder' | 'placeholderTextColor' | 'style'> & {
  readonly accessibilityLabel: string;
  readonly value: string;
  readonly onChangeText: (value: string) => void;
};
