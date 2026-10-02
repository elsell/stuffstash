import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import { useAppServices } from '../ui/navigation/AppServicesContext';
import { InventoryArchiveScreen } from '../ui/screens/InventoryArchiveScreen';
import { DeniedSettingsState } from '../ui/screens/ScopedSettingsScreens';
import { returnToPreviousOrHome } from '../ui/navigation/returnToPreviousOrHome';
import { t } from '../presentation/localization';

export default function InventoryArchiveRoute() {
  const params = useLocalSearchParams<{ tenantId?: string; inventoryId?: string }>();
  const router = useRouter(), services = useAppServices();
  const tenantId = typeof params.tenantId === 'string' ? params.tenantId : undefined;
  const inventoryId = typeof params.inventoryId === 'string' ? params.inventoryId : undefined;
  if (!tenantId) return <DeniedSettingsState message={t('archive.denied')} />;
  return <>
    <Stack.Screen options={{ title: t(inventoryId ? 'archive.export' : 'archive.restore') }} />
    <InventoryArchiveScreen key={`${tenantId}:${inventoryId ?? ''}`} workspace={services.inventoryArchive} scope={{ tenantId, inventoryId }}
      onClose={() => returnToPreviousOrHome(router)} onOpen={async (id, signal) => {
        await services.selectInventoryCommand.execute(id, { signal });
        if (signal.aborted) return;
        router.dismissAll(); router.replace('/');
      }} />
  </>;
}
