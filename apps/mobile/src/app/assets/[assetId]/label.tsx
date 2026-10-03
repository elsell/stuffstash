import { LabelTaskHeader } from '../../../ui/labels/LabelTaskHeader';
import { useLocalSearchParams } from 'expo-router';
import { useEffect, useState } from 'react';
import { Text } from 'react-native';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import type { LabelScope } from '../../../application/labels/LabelWorkspace';
import { LabelOptionsScreen } from '../../../ui/labels/LabelOptionsScreen';
import { t } from '../../../presentation/localization';
export default function AssetLabelRoute() {
  const services = useAppServices(); const params = useLocalSearchParams<{ assetId: string }>();
  const [loaded, setLoaded] = useState<{ scope: LabelScope; owner: typeof services; assetId: string }>();
  const scope = loaded?.owner === services && loaded.assetId === params.assetId ? loaded.scope : undefined; const [error, setError] = useState(false);
  useEffect(() => {
    const controller = new AbortController(); setLoaded(undefined); setError(false);
    void services.currentInventoryScopeQuery.execute({ signal: controller.signal }).then(value => { if (!controller.signal.aborted) setLoaded({ scope: value, owner: services, assetId: params.assetId }); })
      .catch(() => { if (!controller.signal.aborted) setError(true); });
    return () => controller.abort();
  }, [services, params.assetId]);
  if (!scope || typeof params.assetId !== 'string') return <><LabelTaskHeader /><Text>{t(error ? 'labels.mobile.unavailable' : 'labels.mobile.loading')}</Text></>;
  return <><LabelTaskHeader /><LabelOptionsScreen key={`${services.serviceScopeId}:${scope.tenantId}:${scope.inventoryId}:${params.assetId}`} workspace={services.labels} scope={scope} assetId={params.assetId} /></>;
}
