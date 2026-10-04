import { PrinterDefaultsRoute } from './PrinterDefaultsRoute';
import { PrinterDetailScreen } from './PrinterDetailScreen';
import { PrintHistoryScreen } from './PrintHistoryScreen';
import { QuickPrintScreen } from './QuickPrintScreen';
import { ReprintLabelScreen } from './ReprintLabelScreen';
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
type PrintingRouteProps = { readonly assetId?: string; readonly jobId?: string; readonly reprint?: boolean; readonly options?: boolean; readonly settingsView?: 'defaults' | 'history' | 'printer'; readonly printerId?: string };
export function PrintingRoute(props: PrintingRouteProps) {
  const services = useAppServices();
  return props.settingsView === 'defaults'
    ? <PrinterDefaultsRoute key={services.serviceScopeId} workspace={services.printing} query={services.settingsQuery} />
    : <PrintingContent {...props} services={services} />;
}
function PrintingContent({ services, assetId, jobId, reprint = false, options = false, settingsView, printerId }: PrintingRouteProps & { readonly services: ReturnType<typeof useAppServices> }) {
  const router = useRouter();
  const load = useCallback((signal: AbortSignal) => services.settingsQuery.getSelectedScope({ signal }), [services]);
  const task = usePrintingTask(load, services.serviceScopeId); const selected = task.data;
  const header = <Stack.Screen options={{ title: t(reprint ? 'printing.mobile.reprint' : assetId ? 'printing.mobile.print' : jobId ? 'printing.mobile.job' : settingsView === 'defaults' ? 'printing.mobile.defaults' : settingsView === 'history' ? 'printing.mobile.historyTitle' : 'printing.mobile.title') }} />;
  if (!selected) return <>{header}<Text>{t(task.error ? 'printing.mobile.unavailable' : 'printing.mobile.loading')}</Text>{task.error ? <NativeCommandButton label={t('printing.mobile.retry')} onPress={task.reload} /> : null}</>;
  const scope = { tenantId: selected.tenant.id, inventoryId: selected.inventory.id };
  const key = `${services.serviceScopeId}:${scope.tenantId}:${scope.inventoryId}:${assetId ?? jobId ?? printerId ?? settingsView ?? 'settings'}:${reprint}`;
  const canPrint = selected.inventory.permissions.includes('edit_asset');
  const openJob = (id: string) => router.push({ pathname: '/print-jobs/[jobId]', params: { jobId: id } } as never);
  const canConfigure = selected.inventory.permissions.includes('configure');
  if (settingsView === 'history') return <>{header}<PrintHistoryScreen key={key} workspace={services.printing} scope={scope} onJob={openJob} /></>;
  if (settingsView === 'printer' && printerId) return <>{header}<PrinterDetailScreen key={key} workspace={services.printing} scope={scope} printerId={printerId} canConfigure={canConfigure} canPrint={canPrint} onJob={openJob} /></>;
  if ((assetId || reprint) && !canPrint) return <>{header}<Text>{t('printing.mobile.unavailable')}</Text></>;
  return <>{header}{jobId && reprint ? <ReprintLabelScreen key={key} workspace={services.printing} scope={scope} jobId={jobId} onQueued={id => router.replace({ pathname: '/print-jobs/[jobId]', params: { jobId: id } } as never)} /> : assetId && !options ? <QuickPrintScreen key={key} workspace={services.printing} scope={scope} assetId={assetId} onQueued={id => router.replace({ pathname: '/print-jobs/[jobId]', params: { jobId: id } } as never)} /> : assetId ? <AssetPrintScreen key={key} workspace={services.printing} scope={scope} assetId={assetId} onQueued={id => router.replace({ pathname: '/print-jobs/[jobId]', params: { jobId: id } } as never)} />
    : jobId ? <PrintJobScreen key={key} workspace={services.printing} scope={scope} jobId={jobId} canPrint={canPrint} onReprint={id => router.push({ pathname: '/print-jobs/[jobId]', params: { jobId: id, action: 'reprint' } } as never)} />
    : <PrinterSettingsScreen key={key} workspace={services.printing} scope={scope} onPrinter={id => router.push({ pathname: '/settings/inventory/printer-details', params: { printerId: id } } as never)} onDefaults={() => router.push('/settings/inventory/print-defaults' as never)} onHistory={() => router.push('/settings/inventory/print-history' as never)} />}</>;
}
