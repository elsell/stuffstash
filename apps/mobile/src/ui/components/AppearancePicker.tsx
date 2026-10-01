import { t } from '../../presentation/localization';
import { useCallback, useRef } from 'react';
import { useFocusEffect } from 'expo-router';
import { appearancePreferences, isAppearancePreference } from '../../application/settings/AppearancePreference';
import { useAppFeedback } from '../feedback/AppFeedback';
import { appearanceLabel } from '../screens/SettingsScreenPresentation';
import { useAppearance } from '../theme/AppearanceContext';
import { NativeChoicePicker } from './NativeChoicePicker';

const options = appearancePreferences.map(value => ({ value, label: appearanceLabel(value) }));

export function AppearancePicker() {
  const { preference, setPreference } = useAppearance();
  const feedback = useAppFeedback();
  const focused = useRef(false);
  const selection = useRef(0);
  useFocusEffect(useCallback(() => {
    focused.current = true;
    return () => { focused.current = false; selection.current++; };
  }, []));
  async function select(value: string) {
    if (!focused.current || !isAppearancePreference(value) || value === preference) return;
    const request = ++selection.current;
    try { await setPreference(value); }
    catch {
      if (!focused.current || selection.current !== request) return;
      feedback.showNotice({ tone: 'error', title: t('mobile.AppearancePicker.appearanceNotSaved'), message: t('mobile.AppearancePicker.stuffStashCouldNotSaveTheAppearanceSetting') });
    }
  }
  return <NativeChoicePicker label={t('mobile.AppearancePicker.appearance')} accessibilityLabel={t('mobile.AppearancePicker.chooseAppearance')}
    includeEmptyOption={false} value={preference} options={options} onChange={value => void select(value)} />;
}
