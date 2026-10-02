import { t } from '../../presentation/localization';
import { forwardRef, useImperativeHandle, useRef } from 'react';
import { Platform, type TextInput } from 'react-native';
import { AppTextInput } from '../components/AppTextInput';
import type { InvitationEmailInputHandle, InvitationEmailInputProps } from './InvitationEmailInput.types';

export const InvitationEmailInput = forwardRef<InvitationEmailInputHandle, InvitationEmailInputProps>(function InvitationEmailInput({ email, ...props }, ref) {
  const seed = useRef(email).current;
  const field = useRef<TextInput>(null);
  useImperativeHandle(ref, () => ({ blur: () => field.current?.blur() }), []);
  return <AppTextInput ref={field} {...props} {...(Platform.OS === 'ios' ? { defaultValue: seed } : { value: email })} accessibilityLabel={t('mobile.InvitationEmailInput.inviteeEmail')}
    autoCapitalize="none" autoComplete="email" autoCorrect={false} spellCheck={false}
    keyboardType="email-address" placeholder={t('mobile.InvitationEmailInput.friendExampleCom')} />;
});
