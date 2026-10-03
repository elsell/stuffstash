import { useCallback, useEffect, useRef, useState } from 'react';
import { ScrollView, Text, View } from 'react-native';
import type { PrintScope, PrintSettings, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { AppSwitchField } from '../components/AppSwitchField';
import { SettingsLoadingRow, SettingsNavigationRow, SettingsSection, SettingsValueRow, useSettingsListStyles } from '../screens/SettingsList';
import { connectorState, printerReadiness, printJobStatus } from './PrintingStatus';
import { usePrintingTask } from './usePrintingTask';

export function PrinterSettingsScreen({ workspace, scope, canConfigure, onJob }: { readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly canConfigure: boolean; readonly onJob: (id: string) => void }) {
  const { palette, styles } = useSettingsListStyles();
  const load = useCallback(async (signal: AbortSignal) => ({ catalog: await workspace.repository.catalog(scope, signal), jobs: await workspace.repository.jobs(scope, signal) }), [workspace, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, `${scope.tenantId}/${scope.inventoryId}`);
  const [draft, setDraft] = useState<PrintSettings>(); const [saving, setSaving] = useState(false); const running = useRef(false);
  const [saveState, setSaveState] = useState<'idle' | 'saved' | 'failed'>('idle');
  useEffect(() => { setDraft(task.data?.catalog.settings); setSaveState('idle'); }, [task.data]);
  const catalog = task.data?.catalog;
  const editable = canConfigure && !saving;
  const change = (next: PrintSettings) => { if (editable && !running.current) { setDraft(next); setSaveState('idle'); } };
  const save = async () => {
    const owner = task.lifetime.current;
    if (!owner || owner.signal.aborted || !draft || !canConfigure || running.current) return;
    running.current = true; setSaving(true); setSaveState('idle');
    try { const saved = await workspace.repository.saveSettings(scope, draft); if (!owner.signal.aborted) { setDraft(saved); setSaveState('saved'); } }
    catch { if (!owner.signal.aborted) setSaveState('failed'); }
    finally { running.current = false; setSaving(false); }
  };
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content}>
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error ? <><Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text><NativeCommandButton label={t('printing.mobile.retry')} onPress={task.reload} /></> : null}
    {catalog && draft ? <>
      <SettingsSection title={t('printing.mobile.defaults')}>
        <SettingsPickerRow label={t('printing.mobile.printer')} accessibilityLabel={t('printing.mobile.printer')} value={draft.defaultPrinterId ?? ''} disabled={!editable}
          options={[{ value: '', label: t('printing.mobile.none') }, ...catalog.printers.filter(printer => !printer.retired).map(printer => ({ value: printer.id, label: `${printer.name} · ${printer.mediaName}` }))]}
          onChange={id => change({ ...draft, defaultPrinterId: id || null, printOnCreateDefault: !!id && draft.printOnCreateDefault })} />
        <SettingsPickerRow label={t('printing.mobile.template')} accessibilityLabel={t('printing.mobile.template')} value={`${draft.template.id}/${draft.template.version}`} disabled={!editable}
          options={catalog.templates.map(template => ({ value: `${template.id}/${template.version}`, label: template.name }))}
          onChange={id => { const template = catalog.templates.find(item => `${item.id}/${item.version}` === id); if (template) change({ ...draft, template: { id: template.id, version: template.version, showReference: template.showReference } }); }} />
        {catalog.templates.find(template => template.id === draft.template.id && template.version === draft.template.version)?.supportsReference ? <AppSwitchField label={t('printing.mobile.reference')} value={draft.template.showReference} disabled={!editable} onValueChange={showReference => change({ ...draft, template: { ...draft.template, showReference } })} /> : null}
        <AppSwitchField label={t('printing.mobile.auto')} value={draft.printOnCreateDefault} disabled={!editable || !draft.defaultPrinterId} onValueChange={printOnCreateDefault => change({ ...draft, printOnCreateDefault })} />
      </SettingsSection>
      {canConfigure ? <NativeCommandButton label={t('printing.mobile.save')} disabled={saving} onPress={() => void save()} /> : null}
      {saveState !== 'idle' ? <Text accessibilityLiveRegion="polite" style={{ color: saveState === 'failed' ? palette.danger : palette.text }}>{t(saveState === 'saved' ? 'printing.mobile.saved' : 'printing.mobile.saveFailed')}</Text> : null}
      {saveState === 'failed' ? <NativeCommandButton label={t('printing.mobile.reload')} disabled={saving} onPress={task.reload} /> : null}
      {!catalog.printers.length ? <Text style={{ color: palette.textMuted }}>{t('printing.mobile.empty')}</Text> : null}
      {catalog.printers.map(printer => <SettingsSection key={printer.id} title={printer.name} footer={printer.retired ? t('printing.mobile.retired') : printerReadiness(printer.readiness)}>
        <SettingsValueRow label={t('labels.mobile.size')} value={printer.mediaName} />
        {catalog.connectors.filter(connector => connector.printerIds.includes(printer.id)).map(connector => <View key={connector.id} style={{ padding: 16, gap: 4 }}>
          <Text style={{ color: palette.text }}>{t('printing.mobile.connector', { name: connector.name })}</Text>
          <Text style={{ color: palette.textMuted }}>{connectorState(connector.state)}</Text>
          <Text style={{ color: palette.textMuted }}>{connector.lastSeenAt ? t('printing.mobile.lastSeen', { time: new Date(connector.lastSeenAt).toLocaleString() }) : t('printing.mobile.neverSeen')}</Text>
        </View>)}
      </SettingsSection>)}
      <SettingsSection title={t('printing.mobile.jobs')}>
        {task.data!.jobs.length ? task.data!.jobs.map(job => <SettingsNavigationRow key={job.id} label={printJobStatus(job.status)} accessibilityLabel={`${printJobStatus(job.status)} · ${job.id}`} context={catalog.printers.find(printer => printer.id === job.printerId)?.name} onPress={() => onJob(job.id)} />) : <Text style={{ padding: 16, color: palette.textMuted }}>{t('printing.mobile.noJobs')}</Text>}
      </SettingsSection>
      <NativeCommandButton label={t('printing.mobile.refresh')} disabled={saving} onPress={task.reload} />
    </> : null}
  </ScrollView>;
}
