import { LabelTaskHeader } from '../../../ui/labels/LabelTaskHeader';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useEffect, useState } from 'react';
import { Text } from 'react-native';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import type { LabelScope } from '../../../application/labels/LabelWorkspace';
import { LabelOptionsScreen } from '../../../ui/labels/LabelOptionsScreen';
import { t } from '../../../presentation/localization';
export default function AssetLabelRoute() {
  const services = useAppServices(); const router = useRouter(); const params = useLocalSearchParams<{ assetId: string }>();
  const [loaded, setLoaded] = useState<{ scope: LabelScope; owner: typeof services; assetId: string; canPrint: boolean }>();
  const scope = loaded?.owner === services && loaded.assetId === params.assetId ? loaded.scope : undefined; const [error, setError] = useState(false);
  useEffect(() => {
    const controller = new AbortController(); setLoaded(undefined); setError(false);
    void services.settingsQuery.getSelectedScope({ signal: controller.signal }).then(value => { if (!controller.signal.aborted) setLoaded({ scope: { tenantId: value.tenant.id, inventoryId: value.inventory.id }, owner: services, assetId: params.assetId, canPrint: value.inventory.permissions.includes('edit_asset') }); })
      .catch(() => { if (!controller.signal.aborted) setError(true); });
    return () => controller.abort();
  }, [services, params.assetId]);
  if (!scope || typeof params.assetId !== 'string') return <><LabelTaskHeader /><Text>{t(error ? 'labels.mobile.unavailable' : 'labels.mobile.loading')}</Text></>;
  return <><LabelTaskHeader /><LabelOptionsScreen key={`${services.serviceScopeId}:${scope.tenantId}:${scope.inventoryId}:${params.assetId}`} workspace={services.labels} scope={scope} assetId={params.assetId} onPrintOptions={loaded?.canPrint ? () => router.push(`/assets/${params.assetId}/print?options=1`) : undefined} /></>;
}
