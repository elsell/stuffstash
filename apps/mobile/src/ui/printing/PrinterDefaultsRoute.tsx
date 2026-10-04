import { useCallback, useRef } from 'react';
import { Stack } from 'expo-router';
import { Text } from 'react-native';
import type { SettingsScopeRepository } from '../../application/settings/SettingsQuery';
import type { PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { SettingsActionRow, SettingsLoadingRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { PrinterDefaultsScreen, type PrinterDefaultsDraft } from './PrinterDefaultsScreen';
import { usePrintingTask } from './usePrintingTask';

/** The route owns the draft while focus-bound authorized reads hide the editor. */
export function PrinterDefaultsRoute({ workspace, query }: { readonly workspace: PrintingWorkspace; readonly query: SettingsScopeRepository }) {
  const { styles } = useSettingsListStyles();
  const load = useCallback((signal: AbortSignal) => query.getSelectedScope({ signal }), [query]);
  const task = usePrintingTask(load, 'print-defaults');
  const draft = useRef<{ owner: string; state: PrinterDefaultsDraft } | undefined>(undefined);
  const selected = task.data;
  const header = <Stack.Screen options={{ title: t('printing.mobile.defaults') }} />;
  if (!selected) return <>{header}{task.error ? <SettingsSection><Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text><SettingsActionRow label={t('printing.mobile.retry')} onPress={task.reload} /></SettingsSection> : <SettingsLoadingRow label={t('printing.mobile.loading')} />}</>;
  const canConfigure = selected.inventory.permissions.includes('configure');
  const owner = `${selected.tenant.id}/${selected.inventory.id}/${canConfigure}`;
  if (draft.current?.owner !== owner) draft.current = { owner, state: {} };
  return <>{header}<PrinterDefaultsScreen key={owner} workspace={workspace} scope={{ tenantId: selected.tenant.id, inventoryId: selected.inventory.id }} canConfigure={canConfigure} draftState={draft.current.state} /></>;
}
