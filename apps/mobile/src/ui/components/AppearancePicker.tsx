import { appearancePreferences, isAppearancePreference } from '../../application/settings/AppearancePreference';
import { useAppFeedback } from '../feedback/AppFeedback';
import { appearanceLabel } from '../screens/SettingsScreenPresentation';
import { useAppearance } from '../theme/AppearanceContext';
import { NativeChoicePicker } from './NativeChoicePicker';

const options = appearancePreferences.map(value => ({ value, label: appearanceLabel(value) }));

export function AppearancePicker() {
  const { preference, setPreference } = useAppearance();
  const feedback = useAppFeedback();
  async function select(value: string) {
    if (!isAppearancePreference(value) || value === preference) return;
    try { await setPreference(value); }
    catch {
      feedback.showNotice({ tone: 'error', title: 'Appearance not saved', message: 'Stuff Stash could not save the appearance setting.' });
    }
  }
  return <NativeChoicePicker label="Appearance" accessibilityLabel="Choose appearance"
    includeEmptyOption={false} value={preference} options={options} onChange={value => void select(value)} />;
}
