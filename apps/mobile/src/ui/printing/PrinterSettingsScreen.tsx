import { useCallback } from 'react';
import { ScrollView, Text } from 'react-native';
import type { PrintScope, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { SettingsActionRow, SettingsLoadingRow, SettingsNavigationRow, SettingsSection, SettingsSeparator, useSettingsListStyles } from '../screens/SettingsList';
import { mediaSizeLabel, printerReadiness } from './PrintingStatus';
import { usePrintingTask } from './usePrintingTask';

export function PrinterSettingsScreen({ workspace, scope, onPrinter, onDefaults, onHistory }: {
  readonly workspace: PrintingWorkspace; readonly scope: PrintScope;
  readonly onPrinter: (id: string) => void; readonly onDefaults: () => void; readonly onHistory: () => void;
}) {
  const { palette, styles } = useSettingsListStyles();
  const load = useCallback((signal: AbortSignal) => workspace.repository.catalog(scope, signal), [workspace, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, `${scope.tenantId}/${scope.inventoryId}`);
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic">
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error ? <SettingsSection><Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text><SettingsActionRow label={t('printing.mobile.retry')} onPress={task.reload} /></SettingsSection> : null}
    {task.data ? <>
      <SettingsSection title={t('printing.mobile.title')}>
        {task.data.printers.length ? task.data.printers.map((printer, index) => <SettingsPrinterRow key={printer.id} first={index === 0} name={printer.name} size={mediaSizeLabel(printer.mediaName, printer.media)} state={printer.retired ? t('printing.mobile.retired') : printerReadiness(printer.readiness)} onPress={() => onPrinter(printer.id)} />)
          : <Text style={[styles.navigationRow, { color: palette.textMuted }]}>{t('printing.mobile.empty')}</Text>}
      </SettingsSection>
      <SettingsSection>
        <SettingsNavigationRow label={t('printing.mobile.defaults')} accessibilityLabel={t('printing.mobile.defaults')} onPress={onDefaults} />
        <SettingsSeparator />
        <SettingsNavigationRow label={t('printing.mobile.historyTitle')} accessibilityLabel={t('printing.mobile.historyTitle')} onPress={onHistory} />
      </SettingsSection>
      <SettingsSection><SettingsActionRow label={t('printing.mobile.refresh')} onPress={task.reload} /></SettingsSection>
    </> : null}
  </ScrollView>;
}

function SettingsPrinterRow({ name, size, state, first, onPress }: { readonly name: string; readonly size: string; readonly state: string; readonly first: boolean; readonly onPress: () => void }) {
  return <>{!first ? <SettingsSeparator /> : null}<SettingsNavigationRow label={name} accessibilityLabel={name} context={`${size} · ${state}`} onPress={onPress} /></>;
}
