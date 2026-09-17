import type { TextInputProps } from 'react-native';

export type InvitationEmailInputProps = Pick<TextInputProps, 'style' | 'placeholderTextColor'> & {
  readonly email: string;
  readonly editable: boolean;
  readonly onChangeText: (email: string) => void;
};
