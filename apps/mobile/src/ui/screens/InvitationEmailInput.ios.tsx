import { useRef } from 'react';
import { t } from '../../presentation/localization';
import { AppTextInput } from '../components/AppTextInput';
import type { InvitationEmailInputProps } from './InvitationEmailInput.types';

export function InvitationEmailInput({ email, editable, onChangeText, ...props }: InvitationEmailInputProps) {
  const seed = useRef(email).current;
  return <AppTextInput {...props} defaultValue={seed} editable={editable}
    accessibilityLabel={t('mobile.InvitationEmailInputios.inviteeEmail')}
    onChangeText={value => { if (editable) onChangeText(value); }}
    autoCapitalize="none" autoComplete="email" textContentType="emailAddress"
    autoCorrect={false} spellCheck={false} keyboardType="email-address"
    placeholder={t('mobile.InvitationEmailInputios.friendExampleCom')} />;
}
