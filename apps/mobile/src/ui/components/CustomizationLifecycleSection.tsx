import { t } from '../../presentation/localization';
import type { CustomizationKind, CustomizationLifecycle } from '../../domain/customization/Customization';
import { SettingsSection, SettingsSeparator } from '../screens/SettingsList';
import { NativeCommandButton } from './NativeCommandButton';

export function CustomizationLifecycleSection({
  busy,
  kind,
  lifecycle,
  onAction
}: {
  readonly busy: boolean;
  readonly kind: CustomizationKind;
  readonly lifecycle: CustomizationLifecycle;
  readonly onAction: (action: 'archive' | 'restore' | 'delete') => void;
}) {
  if (kind === 'tag' && lifecycle === 'archived') return null;
  return <SettingsSection>
    {lifecycle === 'active'
      ? <NativeCommandButton prominence="standard" role="destructive" disabled={busy} label={busy ? t('mobile.CustomizationLifecycleSection.working') : t('mobile.CustomizationLifecycleSection.archive')} onPress={() => onAction('archive')} />
      : <><NativeCommandButton prominence="standard" disabled={busy} label={busy ? t('mobile.CustomizationLifecycleSection.working') : t('mobile.CustomizationLifecycleSection.restore')} onPress={() => onAction('restore')} /><SettingsSeparator /><NativeCommandButton prominence="standard" role="destructive" disabled={busy} label={t('mobile.CustomizationLifecycleSection.deletePermanently')} onPress={() => onAction('delete')} /></>}
  </SettingsSection>;
}
