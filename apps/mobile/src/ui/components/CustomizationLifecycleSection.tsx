import type { CustomizationKind, CustomizationLifecycle } from '../../domain/customization/Customization';
import { SettingsActionRow, SettingsSection, SettingsSeparator } from '../screens/SettingsList';

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
  return <SettingsSection title="Lifecycle">
    {lifecycle === 'active'
      ? <SettingsActionRow destructive disabled={busy} label={busy ? 'Working…' : 'Archive'} onPress={() => onAction('archive')} />
      : <><SettingsActionRow disabled={busy} label={busy ? 'Working…' : 'Restore'} onPress={() => onAction('restore')} /><SettingsSeparator /><SettingsActionRow destructive disabled={busy} label="Delete permanently" onPress={() => onAction('delete')} /></>}
  </SettingsSection>;
}
