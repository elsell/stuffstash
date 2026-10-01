import { t } from '../../presentation/localization';
import { useRef } from 'react';
import { Host, TextField } from '@expo/ui/swift-ui';
import {
  accessibilityLabel, autocorrectionDisabled, disabled as nativeDisabled, keyboardType,
  onSubmit as nativeSubmit, submitLabel, textFieldStyle, textInputAutocapitalization
} from '@expo/ui/swift-ui/modifiers';
import type { OnboardingAddressInputProps } from './OnboardingAddressInput.types';

export function OnboardingAddressInput({ initialValue, onChangeText, disabled, onSubmit }: OnboardingAddressInputProps) {
  const seed = useRef(initialValue);
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 54 }}>
    <TextField defaultValue={seed.current} placeholder={t('mobile.OnboardingAddressInputios.httpsStashExampleCom')} onValueChange={onChangeText}
      modifiers={[
        accessibilityLabel(t('mobile.OnboardingAddressInputios.serverAddress')), keyboardType('url'), autocorrectionDisabled(),
        textInputAutocapitalization('never'), textFieldStyle('roundedBorder'), nativeDisabled(disabled),
        submitLabel('go'), nativeSubmit(() => { if (!disabled) onSubmit(); })
      ]} />
  </Host>;
}
