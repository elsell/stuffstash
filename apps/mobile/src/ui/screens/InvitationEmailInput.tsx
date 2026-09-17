import { useRef } from 'react';
import { Platform } from 'react-native';
import { AppTextInput } from '../components/AppTextInput';
import type { InvitationEmailInputProps } from './InvitationEmailInput.types';

export function InvitationEmailInput({ email, ...props }: InvitationEmailInputProps) {
  const seed = useRef(email).current;
  return <AppTextInput {...props} {...(Platform.OS === 'ios' ? { defaultValue: seed } : { value: email })} accessibilityLabel="Invitee email"
    autoCapitalize="none" autoComplete="email" autoCorrect={false} spellCheck={false}
    keyboardType="email-address" placeholder="friend@example.com" />;
}
