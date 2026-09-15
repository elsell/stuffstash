import { useRef } from 'react';
import { AppTextInput } from '../components/AppTextInput';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { onboardingStyles } from './OnboardingPresentation';
import type { OnboardingAddressInputProps } from './OnboardingAddressInput.types';

export function OnboardingAddressInput({ initialValue, onChangeText, disabled, onSubmit }: OnboardingAddressInputProps) {
  const seed = useRef(initialValue);
  const colors = useAppearanceAwarePalette();
  return <AppTextInput accessibilityLabel="Server address" defaultValue={seed.current}
    onChangeText={onChangeText} placeholder="https://stash.example.com" placeholderTextColor={colors.textMuted}
    autoCorrect={false} autoCapitalize="none" keyboardType="url" editable={!disabled}
    returnKeyType="go" onSubmitEditing={() => { if (!disabled) onSubmit(); }} style={onboardingStyles(colors).input} />;
}
