import { useCallback, useRef } from 'react';
import { ScrollView, Text, View } from 'react-native';
import type { LabelScope, LabelWorkspace } from '../../application/labels/LabelWorkspace';
import type { PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { AssetPrintScreen, type AssetPrintDraft } from '../printing/AssetPrintScreen';
import { usePrintingTask } from '../printing/usePrintingTask';
import { SettingsActionRow, SettingsLoadingRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { LabelOptionsScreen } from './LabelOptionsScreen';

type Props = { labels: LabelWorkspace; printing?: PrintingWorkspace; scope: LabelScope; assetId: string; onQueued(id: string): void };
/** Permission is supplied by the scoped route; export-only readers never query printing. */
export function AssetLabelTask(props: Props) {
  return props.printing ? <RegisteredLabelTask {...props} printing={props.printing} /> : <LabelOptionsScreen workspace={props.labels} scope={props.scope} assetId={props.assetId} />;
}
function RegisteredLabelTask({ labels, printing, scope, assetId, onQueued }: Props & { printing: PrintingWorkspace }) {
  const { styles } = useSettingsListStyles();
  const draft = useRef<AssetPrintDraft>({});
  const load = useCallback((signal: AbortSignal) => printing.repository.catalog(scope, signal), [printing, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, assetId);
  if (!task.data) return <ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic"><SettingsSection>
    {task.error ? <><View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.dangerText}>{t('printing.mobile.unavailable')}</Text></View><SettingsActionRow label={t('printing.mobile.retry')} onPress={task.reload} /></> : <SettingsLoadingRow label={t('printing.mobile.loading')} />}
  </SettingsSection></ScrollView>;
  if (!task.data.printers.some(printer => !printer.retired)) return <LabelOptionsScreen workspace={labels} scope={scope} assetId={assetId} noPrinter />;
  return <AssetPrintScreen workspace={printing} labelWorkspace={labels} scope={scope} assetId={assetId} draftState={draft.current} onQueued={onQueued} />;
}
