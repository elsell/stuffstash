import { useEffect, useState } from 'react';
import { Text, View } from 'react-native';
import { File } from 'expo-file-system';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { ExpoExportTemporaryFiles } from '../src/adapters/exports/ExpoExportTemporaryFiles';
import { ExpoExportFileShare } from '../src/adapters/exports/ExpoExportFileShare';
import { NativeExportFileDelivery } from '../src/adapters/exports/NativeExportFileDelivery';
import { ExportInventoryCommand } from '../src/application/exports/InventoryExport';
import { SettingsQuery } from '../src/application/settings/SettingsQuery';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { InventorySettingsScreen } from '../src/ui/screens/ScopedSettingsScreens';

// Real settings, cache files and system share sheet; only the network repository is fake.
export function InventoryExportFixture() {
  const [evidence, setEvidence] = useState('No export yet');
  const [fixture] = useState(() => {
    const observer = { record: () => setEvidence('Cleanup failed') };
    const files = new ExpoExportTemporaryFiles(observer);
    const sheet = new ExpoExportFileShare();
    const query = new SettingsQuery(
      { getCurrentPrincipal: async () => ({ id: 'export-owner', email: 'owner@example.invalid' }) },
      { getDiagnostics: () => ({ apiBaseUrl: 'https://inventory.example.invalid', appVersion: 'export-audit', authenticationMode: 'oidc-sso' }) },
      { getSelectedScope: async () => ({ tenant: { id: 'household', name: 'Household', permissions: ['configure'] }, inventory: { id: 'inventory', name: 'Main Inventory', permissions: ['view', 'share', 'configure'] } }) }
    );
    const command = new ExportInventoryCommand({ download: async (scope, format) => {
      if (scope.tenantId !== 'household' || scope.inventoryId !== 'inventory') throw new Error('Incorrect export scope');
      return format === 'json' ? '{"schemaVersion":1,"assets":[{"title":"Drill"}]}' : 'title\r\nDrill\r\n';
    } }, new NativeExportFileDelivery({
      sweep: () => files.sweep(),
      write: async file => {
        const temporary = await files.write(file);
        if (await new File(temporary.uri).text() !== file.content) throw new Error('Export file content differs');
        setEvidence(`${file.format}: file content verified`);
        return { uri: temporary.uri, finish: async retain => {
          await temporary.finish(retain);
          setEvidence(`${file.format}: ${new File(temporary.uri).exists ? 'file retained' : 'file removed'}`);
        } };
      }
    }, sheet, 'ios', observer));
    return { query, command, client: createMobileQueryClient() };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="export-audit" loadInventoryScope={async () => ({ tenantId: 'household', inventoryId: 'inventory' })}>
    <View style={{ flex: 1 }}><InventorySettingsScreen settingsQuery={fixture.query} exportCommand={fixture.command} onNavigate={() => {}} /><Text accessibilityLabel={evidence}>{evidence}</Text></View>
  </MobileServerStateProvider>;
}
