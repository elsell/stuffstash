import { t } from '../../presentation/localization';
import { useEffect, useRef, useState } from 'react';
import { Text, View } from 'react-native';
import type { CreateWorkspace, CreatedHousehold, CreatedInventory } from '../../application/inventories/CreateWorkspace';
import { DraftTextField } from '../components/DraftTextField';
import { NativeFilterSheet } from '../components/NativeFilterSheet';
import { SettingsSection, useSettingsListStyles } from './SettingsList';

export type WorkspaceCreationTask = { readonly kind: 'household' } | { readonly kind: 'inventory'; readonly household: { readonly id: string; readonly name: string } };
export function WorkspaceCreationForm({ task, command, onCancel, onCreated, onBusy }: {
  readonly task: WorkspaceCreationTask;
  readonly command: CreateWorkspace;
  readonly onCancel: () => void;
  readonly onCreated: (result: { kind: 'household'; value: CreatedHousehold } | { kind: 'inventory'; value: CreatedInventory }) => void;
  readonly onBusy: (busy: boolean) => void;
}) {
  const { palette, styles } = useSettingsListStyles();
  const [name, setName] = useState(''); const [error, setError] = useState(''); const [saving, setSaving] = useState(false);
  const pending = useRef(false); const alive = useRef(false);
  useEffect(() => { alive.current = true; return () => { alive.current = false; }; }, []);
  async function create() {
    if (pending.current || !alive.current || !name.trim()) return;
    pending.current = true; setSaving(true); setError(''); onBusy(true);
    try {
      const result = task.kind === 'household'
        ? { kind: 'household' as const, value: await command.household(name) }
        : { kind: 'inventory' as const, value: await command.inventory(task.household.id, name) };
      if (alive.current) onCreated(result);
    } catch {
      if (alive.current) setError(t(`workspace.creationFailed.${task.kind}`));
    } finally {
      pending.current = false;
      if (alive.current) { setSaving(false); onBusy(false); }
    }
  }
  const label = task.kind === 'household' ? t('mobile.WorkspaceCreationForm.householdName') : t('mobile.WorkspaceCreationForm.inventoryName');
  return <NativeFilterSheet title={task.kind === 'household' ? t('mobile.WorkspaceCreationForm.newHousehold') : t('mobile.WorkspaceCreationForm.newInventory')} footerTestID="workspace-creation-actions"
    actions={{ primaryLabel: task.kind === 'household' ? t('mobile.WorkspaceCreationForm.createHousehold') : t('mobile.WorkspaceCreationForm.createInventory'), secondaryLabel: t('mobile.WorkspaceCreationForm.cancel'),
      secondaryAccessibilityLabel: t('mobile.WorkspaceCreationForm.cancelCreation'), disabled: saving || !name.trim(), secondaryDisabled: saving,
      onApply: () => { void create(); }, onBack: onCancel }}>
    <SettingsSection title={label} footer={task.kind === 'inventory' ? t('mobile.WorkspaceCreationForm.in', { name: String(task.household.name) }) : t('mobile.WorkspaceCreationForm.aHouseholdHasItsOwnInventoriesAndSharing')}>
      <View style={styles.navigationRow}><DraftTextField style={[styles.rowLabel, { flex: 1, minHeight: 48 }]} placeholderTextColor={palette.textMuted} accessibilityLabel={label} placeholder={label}
        value={name} onChangeText={setName} editable={!saving} /></View>
    </SettingsSection>
    {saving ? <Text accessibilityLiveRegion="polite" style={styles.sectionFooter}>{t('mobile.WorkspaceCreationForm.creating')}</Text> : null}
    {error ? <Text accessibilityRole="alert" style={styles.sectionFooter}>{error}</Text> : null}
  </NativeFilterSheet>;
}
