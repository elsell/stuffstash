import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Stack, useNavigation } from 'expo-router';
import { usePreventRemove, type NavigationAction } from '@react-navigation/native';
import { Alert, ScrollView, Text } from 'react-native';
import type { PrintScope, PrintSettings, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
import { t } from '../../presentation/localization';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { SettingsActionRow, SettingsLoadingRow, SettingsSection, SettingsSeparator, SettingsSwitchRow, useSettingsListStyles } from '../screens/SettingsList';
import { mediaSizeLabel } from './PrintingStatus';
import { usePrintingTask } from './usePrintingTask';

export interface PrinterDefaultsDraft { draft?: PrintSettings; committed?: PrintSettings }

export function PrinterDefaultsScreen({ workspace, scope, canConfigure, draftState }: { readonly workspace: PrintingWorkspace; readonly scope: PrintScope; readonly canConfigure: boolean; readonly draftState?: PrinterDefaultsDraft }) {
  const { palette, styles } = useSettingsListStyles();
  const navigation = useNavigation();
  const load = useCallback((signal: AbortSignal) => workspace.repository.catalog(scope, signal), [workspace, scope.tenantId, scope.inventoryId]);
  const task = usePrintingTask(load, `${scope.tenantId}/${scope.inventoryId}`);
  const [draft, setDraft] = useState<PrintSettings | undefined>(draftState?.draft);
  const [committed, setCommitted] = useState<PrintSettings | undefined>(draftState?.committed);
  const [saving, setSaving] = useState(false);
  const running = useRef(false);
  const [exitAction, setExitAction] = useState<NavigationAction>();
  const [saveState, setSaveState] = useState<'idle' | 'saved' | 'failed'>('idle');
  const catalog = task.data;
  useLayoutEffect(() => { if (draftState) { draftState.draft = draft; draftState.committed = committed; } }, [draftState, draft, committed]);
  const dirty = !!draft && !!committed && settingsChanged(draft, committed);
  useEffect(() => {
    if (catalog && !dirty) { setDraft(catalog.settings); setCommitted(catalog.settings); setSaveState('idle'); }
  }, [catalog?.settings]);
  const editable = canConfigure && !saving;
  const change = (next: PrintSettings) => { if (editable && !running.current) { setDraft(next); setSaveState('idle'); } };
  const save = async () => {
    const owner = task.lifetime.current;
    if (!owner || owner.signal.aborted || !catalog || !draft || !dirty || !canConfigure || running.current) return;
    running.current = true; setSaving(true); setSaveState('idle');
    try { const saved = await workspace.repository.saveSettings(scope, draft); if (!owner.signal.aborted) { setDraft(saved); setCommitted(saved); setSaveState('saved'); } }
    catch { if (!owner.signal.aborted) setSaveState('failed'); }
    finally { running.current = false; setSaving(false); }
  };
  const header = useNativeHeaderActionOptions(canConfigure && catalog ? [{ kind: 'save', label: t('printing.mobile.saveHeader'), emphasis: 'primary', disabled: !dirty || saving, onPress: () => void save() }] : []);
  usePreventRemove(dirty && canConfigure && !exitAction, ({ data }) => {
    const owner = task.lifetime.current;
    if (!owner || owner.signal.aborted || running.current) return;
    Alert.alert(t('printing.mobile.discardTitle'), t('printing.mobile.discardMessage'), [
      { text: t('printing.mobile.keepEditing'), style: 'cancel' },
      { text: t('printing.mobile.discard'), style: 'destructive', onPress: () => { if (!owner.signal.aborted && !running.current) setExitAction(data.action); } }
    ]);
  });
  useEffect(() => { if (exitAction) navigation.dispatch(exitAction); }, [exitAction, navigation]);
  const reload = () => { if (!running.current) { setDraft(undefined); setCommitted(undefined); setSaveState('idle'); task.reload(); } };
  return <><Stack.Screen options={header} /><ScrollView style={styles.shell} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic">
    {task.loading ? <SettingsLoadingRow label={t('printing.mobile.loading')} /> : null}
    {task.error ? <SettingsSection><Text accessibilityRole="alert" style={styles.errorMessage}>{t('printing.mobile.unavailable')}</Text><SettingsActionRow label={t('printing.mobile.retry')} onPress={task.reload} /></SettingsSection> : null}
    {catalog && draft ? <>
      <SettingsSection footer={t('labels.mobile.referenceHelp')}>
        <SettingsPickerRow label={t('printing.mobile.printer')} accessibilityLabel={t('printing.mobile.printer')} value={draft.defaultPrinterId ?? ''} disabled={!editable}
          options={[{ value: '', label: t('printing.mobile.none') }, ...catalog.printers.filter(printer => !printer.retired).map(printer => ({ value: printer.id, label: `${printer.name} · ${mediaSizeLabel(printer.mediaName, printer.media)}` }))]}
          onChange={id => change({ ...draft, defaultPrinterId: id || null, printOnCreateDefault: !!id && draft.printOnCreateDefault })} />
        <SettingsSeparator />
        <SettingsPickerRow label={t('printing.mobile.template')} accessibilityLabel={t('printing.mobile.template')} value={`${draft.template.id}/${draft.template.version}`} disabled={!editable}
          options={catalog.templates.map(template => ({ value: `${template.id}/${template.version}`, label: template.name }))}
          onChange={id => { const template = catalog.templates.find(item => `${item.id}/${item.version}` === id); if (template) change({ ...draft, template: { id: template.id, version: template.version, showReference: template.showReference } }); }} />
        <SettingsSeparator />
        {catalog.templates.find(template => template.id === draft.template.id && template.version === draft.template.version)?.supportsReference ? <SettingsSwitchRow label={t('printing.mobile.reference')} value={draft.template.showReference} disabled={!editable} onValueChange={showReference => change({ ...draft, template: { ...draft.template, showReference } })} /> : null}
        <SettingsSeparator />
        <SettingsSwitchRow label={t('printing.mobile.auto')} value={draft.printOnCreateDefault} disabled={!editable || !draft.defaultPrinterId} onValueChange={printOnCreateDefault => change({ ...draft, printOnCreateDefault })} />
      </SettingsSection>
      {saveState !== 'idle' ? <Text accessibilityRole={saveState === 'failed' ? 'alert' : undefined} accessibilityLiveRegion="polite" style={[styles.sectionFooter, { color: saveState === 'failed' ? palette.danger : palette.text }]}>{t(saveState === 'saved' ? 'printing.mobile.saved' : 'printing.mobile.saveFailed')}</Text> : null}
      {saveState === 'failed' ? <SettingsSection><SettingsActionRow label={t('printing.mobile.reload')} disabled={saving} onPress={reload} /></SettingsSection> : null}
    </> : null}
  </ScrollView></>;
}

function settingsChanged(a: PrintSettings, b: PrintSettings) {
  return a.defaultPrinterId !== b.defaultPrinterId || a.printOnCreateDefault !== b.printOnCreateDefault || a.template.id !== b.template.id || a.template.version !== b.template.version || a.template.showReference !== b.template.showReference;
}
