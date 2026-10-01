import { useLayoutEffect, useRef, useState } from 'react';
import { Text } from 'react-native';
import { useIsFocused } from '@react-navigation/native';
import type { ExportInventoryCommand, InventoryExportFormat, InventoryExportScope } from '../../application/exports/InventoryExport';
import { NativeActionMenu } from '../components/NativeActionMenu';
import { SettingsActionRow, SettingsLoadingRow, SettingsSection, SettingsSeparator, useSettingsListStyles } from './SettingsList';

export function InventoryExportAction({ command, scope }: { readonly command: ExportInventoryCommand; readonly scope: InventoryExportScope }) {
  const { styles } = useSettingsListStyles();
  const focused = useIsFocused();
  const active = useRef<AbortController | undefined>(undefined);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const lastFormat = useRef<InventoryExportFormat>('json');
  useLayoutEffect(() => {
    setPending(false); setError('');
    return () => { active.current?.abort(); active.current = undefined; };
  }, [command, focused, scope.tenantId, scope.inventoryId]);
  const cancel = () => { active.current?.abort(); active.current = undefined; setPending(false); };
  const run = async (format: InventoryExportFormat) => {
    if (!focused || active.current) return;
    const request = new AbortController(); active.current = request;
    lastFormat.current = format; setPending(true); setError('');
    try { await command.execute(scope, format, request.signal); }
    catch (caught) {
      if (!request.signal.aborted) {
        const status = (caught as { status?: number }).status;
        setError(status === 401 ? 'Sign in again to export this inventory.' : status === 403 ? 'You no longer have access to export this inventory.' : status === 422 ? 'This inventory exceeds the server’s export limit. Ask your administrator to increase it.' : 'Could not export this inventory. Try again.');
      }
    } finally { if (active.current === request) { active.current = undefined; setPending(false); } }
  };
  return <SettingsSection footer="Includes archived items and attachment details. Photo and file contents are not included.">
    <NativeActionMenu accessibilityLabel="Export inventory" disabled={pending || !focused} trigger={{ kind: 'row', label: 'Export inventory' }} groups={[{ id: 'formats', items: [
      { id: 'json', label: 'JSON — complete inventory data', systemImage: 'doc', onPress: () => void run('json') },
      { id: 'csv', label: 'CSV — spreadsheet rows', systemImage: 'tablecells', onPress: () => void run('csv') }
    ] }]} />
    {pending ? <><SettingsSeparator /><SettingsLoadingRow label="Preparing export…" /><SettingsActionRow label="Cancel export" onPress={cancel} /></> : null}
    {error ? <><Text accessibilityRole="alert" accessibilityLiveRegion="assertive" style={styles.errorMessage}>{error}</Text><SettingsActionRow label="Retry export" onPress={() => void run(lastFormat.current)} /></> : null}
  </SettingsSection>;
}
