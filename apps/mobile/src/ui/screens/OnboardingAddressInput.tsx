import { t } from '../../presentation/localization';
import { useRef } from 'react';
import { AppTextInput } from '../components/AppTextInput';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { onboardingStyles } from './OnboardingPresentation';
import type { OnboardingAddressInputProps } from './OnboardingAddressInput.types';

export function OnboardingAddressInput({ initialValue, onChangeText, disabled, onSubmit }: OnboardingAddressInputProps) {
  const seed = useRef(initialValue);
  const colors = useAppearanceAwarePalette();
  return <AppTextInput accessibilityLabel={t('mobile.OnboardingAddressInput.serverAddress')} defaultValue={seed.current}
    onChangeText={onChangeText} placeholder={t('mobile.OnboardingAddressInput.httpsStashExampleCom')} placeholderTextColor={colors.textMuted}
    autoCorrect={false} autoCapitalize="none" keyboardType="url" editable={!disabled}
    returnKeyType="go" onSubmitEditing={() => { if (!disabled) onSubmit(); }} style={onboardingStyles(colors).input} />;
}
