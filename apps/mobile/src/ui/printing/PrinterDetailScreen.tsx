import { useCallback } from 'react';
import { Stack } from 'expo-router';
import { ScrollView, Text, View } from 'react-native';
import type { PrintScope, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { SettingsActionRow, SettingsLoadingRow, SettingsSection, SettingsSeparator, SettingsValueRow, useSettingsListStyles } from '../screens/SettingsList';
import { connectorAvailability, connectorState, mediaSizeLabel, printerAttention, printerReadiness } from './PrintingStatus';
import { PrinterMediaSettings } from './PrinterMediaSettings';
import { PrinterTestCommand } from './PrinterTestCommand';
import { usePrintingTask } from './usePrintingTask';

export function PrinterDetailScreen({ workspace, scope, printerId, canConfigure, canPrint, onJob }: {
  readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly printerId: string;
  readonly canConfigure: boolean; readonly canPrint: boolean; readonly onJob: (id: string) => void;
}) {
  const { palette, styles } = useSettingsListStyles();
  const load = useCallback((signal: AbortSignal) => workspace.repository.catalog(scope, signal), [workspace, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, `${scope.tenantId}/${scope.inventoryId}/${printerId}`);
  const printer = task.data?.printers.find(value => value.id === printerId);
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic">
    <Stack.Screen options={{ title: printer?.name ?? t('printing.mobile.printer') }} />
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error || task.data && !printer ? <SettingsSection><Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text><SettingsActionRow label={t('printing.mobile.retry')} onPress={task.reload} /></SettingsSection> : null}
    {printer && task.data ? <>
      <SettingsSection footer={printerAttention(printer.readinessReason) || undefined}>
        <SettingsValueRow label={t('printing.mobile.statusLabel')} value={printer.retired ? t('printing.mobile.retired') : printerReadiness(printer.readiness)} />
        <SettingsSeparator />
        {canConfigure && !printer.retired ? <PrinterMediaSettings workspace={workspace} scope={scope} printer={printer} onReload={task.reload} onSaved={next => task.setData(current => current ? { ...current, printers: current.printers.map(value => value.id === next.id ? next : value) } : current)} />
          : <SettingsValueRow label={t('labels.mobile.size')} value={mediaSizeLabel(printer.mediaName, printer.media)} />}
        {canPrint ? <><SettingsSeparator /><PrinterTestCommand workspace={workspace} scope={scope} printer={printer} template={task.data.settings.template} lifetime={task.lifetime} onQueued={onJob} /></> : null}
      </SettingsSection>
      <SettingsSection title={t('printing.mobile.connectionDetails')}>
        {printer.reportedAt ? <Text style={[styles.navigationRow, { color: palette.textMuted }]}>{t('printing.mobile.reportedAt', { time: new Date(printer.reportedAt).toLocaleString() })}</Text> : null}
        {task.data.connectors.filter(connector => connector.printerIds.includes(printer.id)).map(connector => <View key={connector.id} style={styles.navigationRow}>
          <Text style={styles.rowLabel}>{connector.name}</Text>
          <Text style={styles.rowContext}>{connectorAvailability(connector.availability)} · {connectorState(connector.state)}</Text>
          <Text style={styles.rowContext}>{connector.lastSeenAt ? t('printing.mobile.lastSeen', { time: new Date(connector.lastSeenAt).toLocaleString() }) : t('printing.mobile.neverSeen')}</Text>
        </View>)}
        <SettingsActionRow label={t('printing.mobile.refresh')} onPress={task.reload} />
      </SettingsSection>
    </> : null}
  </ScrollView>;
}
