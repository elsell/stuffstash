import { useCallback } from 'react';
import { Stack, useRouter } from 'expo-router';
import { Text } from 'react-native';
import { useAppServices } from '../navigation/AppServicesContext';
import { PrinterSettingsScreen } from './PrinterSettingsScreen';
import { AssetPrintScreen } from './AssetPrintScreen';
import { PrintJobScreen } from './PrintJobScreen';
import { usePrintingTask } from './usePrintingTask';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { t } from '../../presentation/localization';
export function PrintingRoute({ assetId, jobId }: { readonly assetId?: string; readonly jobId?: string }) {
  const services = useAppServices(); const router = useRouter();
  const load = useCallback((signal: AbortSignal) => services.settingsQuery.getSelectedScope({ signal }), [services]);
  const task = usePrintingTask(load, services.serviceScopeId); const selected = task.data;
  const header = <Stack.Screen options={{ title: t(assetId ? 'printing.mobile.print' : jobId ? 'printing.mobile.job' : 'printing.mobile.title') }} />;
  if (!selected) return <>{header}<Text>{t(task.error ? 'printing.mobile.unavailable' : 'printing.mobile.loading')}</Text>{task.error ? <NativeCommandButton label={t('printing.mobile.retry')} onPress={task.reload} /> : null}</>;
  const scope = { tenantId: selected.tenant.id, inventoryId: selected.inventory.id };
  const key = `${services.serviceScopeId}:${scope.tenantId}:${scope.inventoryId}:${assetId ?? jobId ?? 'settings'}`;
  const canPrint = selected.inventory.permissions.includes('edit_asset');
  const openJob = (id: string) => router.push({ pathname: '/print-jobs/[jobId]', params: { jobId: id } } as never);
  if (assetId && !canPrint) return <>{header}<Text>{t('printing.mobile.unavailable')}</Text></>;
  return <>{header}{assetId ? <AssetPrintScreen key={key} workspace={services.printing} scope={scope} assetId={assetId} onQueued={id => router.replace({ pathname: '/print-jobs/[jobId]', params: { jobId: id } } as never)} />
    : jobId ? <PrintJobScreen key={key} workspace={services.printing} scope={scope} jobId={jobId} canPrint={canPrint} />
    : <PrinterSettingsScreen key={key} workspace={services.printing} scope={scope} canConfigure={selected.inventory.permissions.includes('configure')} onJob={openJob} />}</>;
}
