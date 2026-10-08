import { formatArchiveReview } from '@stuff-stash/localization';
import { useCallback, useLayoutEffect, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import { AppState, ScrollView, Text, View } from 'react-native';
import { ArchiveTask } from '../../application/archives/ArchiveTask';
import type { ArchiveJob, ArchivePreview, ArchiveScope, InventoryArchiveWorkspace } from '../../application/archives/InventoryArchive';
import { localization, t } from '../../presentation/localization';
import { useIsFocused } from '@react-navigation/native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import type { ExportInventoryCommand } from '../../application/exports/InventoryExport';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { NativeChoicePicker } from '../components/NativeChoicePicker';
import { SettingsSegmentedControl } from '../components/SettingsSegmentedControl';
import { NativeActionMenu } from '../components/NativeActionMenu';
import { DraftTextField } from '../components/DraftTextField';
import { SettingsActionRow, SettingsLoadingRow, SettingsSection, SettingsSeparator, SettingsSwitchRow, useSettingsListStyles } from './SettingsList';

export function InventoryArchiveScreen({ workspace, scope, exportCommand, onOpen }: {
  workspace: InventoryArchiveWorkspace; scope: ArchiveScope; exportCommand?: ExportInventoryCommand; onClose(): void; onOpen(id: string, signal: AbortSignal): Promise<void>;
}) {
  const { styles } = useSettingsListStyles();
  const focused = useIsFocused(), insets = useSafeAreaInsets();
  const household = { tenantId: scope.tenantId };
  const [mode, setMode] = useState(scope.inventoryId ? 'export' : 'import');
  const [format, setFormat] = useState('archive');
  const direct = useRef<AbortController | undefined>(undefined);
  const [exporting, setExporting] = useState(false);
  const [photos, setPhotos] = useState(true), [otherFiles, setOtherFiles] = useState(true);
  const [jobs, setJobs] = useState<ArchiveJob[]>([]), [cursor, setCursor] = useState<string>();
  const [busy, setBusy] = useState(false), [loading, setLoading] = useState(true), [error, setError] = useState('');
  const [errorIsActivity, setErrorIsActivity] = useState(false);
  const [fileName, setFileName] = useState<string>();
  const [review, setReview] = useState<{ job: ArchiveJob; preview: ArchivePreview }>();
  const [name, setName] = useState('');
  const task = useRef<ArchiveTask | undefined>(undefined), working = useRef(false);
  const revision = useRef(0), currentJobs = useRef<ArchiveJob[]>([]);
  const store = (value: ArchiveJob[]) => { currentJobs.current = value; setJobs(value); };
  const merge = (values: ArchiveJob[]) => {
    const byId = new Map(currentJobs.current.map(job => [job.id, job]));
    values.forEach(job => byId.set(job.id, job));
    store([...byId.values()].sort((a, b) => Number(isActive(b)) - Number(isActive(a)) || b.createdAt.localeCompare(a.createdAt) || b.id.localeCompare(a.id)));
  };
  const report = (caught: unknown, activity = false) => {
    setErrorIsActivity(activity);
    const status = (caught as { status?: number }).status;
    if (status === 401 || status === 403) { store([]); setReview(undefined); setCursor(undefined); }
    setError(t(status === 401 ? 'archive.signIn' : status === 403 ? 'archive.denied' : (status === 413 || status === 422) ? 'archive.tooLarge' : 'archive.error'));
  };
  useLayoutEffect(() => () => { task.current?.close(); direct.current?.abort(); }, [workspace, exportCommand, focused, scope.tenantId, scope.inventoryId]);
  useFocusEffect(useCallback(() => {
    const visit = new ArchiveTask(workspace, scope); task.current = visit;
    working.current = false; setBusy(false); setExporting(false); setMode(scope.inventoryId ? 'export' : 'import'); setFormat('archive'); store([]); setCursor(undefined); setReview(undefined); setFileName(undefined); setError(''); setLoading(true);
    let timer: ReturnType<typeof setTimeout> | undefined;
    let fetching = false;
    let initialized = false;
    const refresh = async (force = false) => {
      if (initialized && !force && !currentJobs.current.some(isRunning)) return;
      if (visit.signal.aborted || fetching || working.current || AppState.currentState !== 'active') return;
      fetching = true; const started = revision.current;
      try {
        const page = await workspace.repository.list(household, undefined, visit.signal);
        const visible = new Set(page.jobs.map(job => job.id));
        const olderActive = currentJobs.current.filter(job => !visible.has(job.id) && isRunning(job));
        const older = await Promise.all(olderActive.map(job => workspace.repository.get(household, job.id, visit.signal)));
        if (!visit.signal.aborted && started === revision.current) {
          const initial = !initialized;
          merge([...page.jobs, ...older]); if (initial) setCursor(page.nextCursor);
          initialized = true; setLoading(false);
        }
      } catch (caught) { if (!visit.signal.aborted && started === revision.current) { initialized = true; report(caught, true); setLoading(false); } }
      finally { fetching = false; }
    };
    const tick = async () => {
      await refresh();
      if (!visit.signal.aborted) timer = setTimeout(() => void tick(), 3000);
    };
    void tick();
    const subscription = AppState.addEventListener('change', state => { if (state === 'active') void refresh(true); });
    return () => { visit.close(); clearTimeout(timer); subscription.remove(); if (task.current === visit) task.current = undefined; };
  }, [workspace, exportCommand, scope.tenantId, scope.inventoryId]));
  const run = async (operation: (visit: ArchiveTask) => Promise<void>, activity = false) => {
    const visit = task.current; if (!focused || !visit || visit.signal.aborted || working.current) return;
    working.current = true; revision.current++; setBusy(true); setError('');
    try { await operation(visit); }
    catch (caught) { if (!visit.signal.aborted) report(caught, activity); }
    finally { if (task.current === visit) { revision.current++; working.current = false; setBusy(false); setExporting(false); } }
  };
  const saveJob = (visit: ArchiveTask, job: ArchiveJob) => { if (!visit.signal.aborted) merge([job]); };
  const primary = () => void run(async visit => {
    if (review) {
      const job = await workspace.repository.approve(scope.tenantId, review.job.id, name.trim(), visit.signal);
      saveJob(visit, job); if (!visit.signal.aborted) setReview(undefined);
    } else if (mode === 'export' && scope.inventoryId) {
      if (format === 'archive') saveJob(visit, await visit.create({ photos, otherFiles }));
      else if (exportCommand && (format === 'json' || format === 'csv')) {
        const request = new AbortController(); direct.current = request; setExporting(true);
        try { await exportCommand.execute({ tenantId: scope.tenantId, inventoryId: scope.inventoryId }, format, request.signal); }
        catch (caught) {
          if (!request.signal.aborted) {
            if ((caught as { status?: number }).status === 422) { setErrorIsActivity(false); setError(t('mobile.InventoryExportAction.thisInventoryExceedsTheServerSExportLimitAsk')); }
            else throw caught;
          }
        }
        finally { if (direct.current === request) { direct.current = undefined; if (!visit.signal.aborted) setExporting(false); } }
      }
    }
    else { saveJob(visit, await visit.upload()); if (!visit.signal.aborted) setFileName(undefined); }
  });
  const reviewCopy = review ? formatArchiveReview(t, review.preview) : undefined;
  return <ScrollView style={styles.shell} contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled" contentContainerStyle={[styles.content, { paddingBottom: Math.max(insets.bottom, 20) }]}>
      {!review && scope.inventoryId ? <View style={[styles.contentBlock, { marginTop: 16 }]}><SettingsSegmentedControl disabled={busy} value={mode} onChange={setMode} segments={[{ value: 'export', label: t('archive.exportMode') }, { value: 'import', label: t('archive.importMode') }]} /></View> : null}
      {!review && mode === 'export' ? <SettingsSection footer={t(format === 'archive' ? 'archive.description' : format === 'json' ? 'archive.jsonDescription' : 'archive.csvDescription')}>
        <View style={styles.navigationRow}><NativeChoicePicker label={t('archive.format')} includeEmptyOption={false} disabled={busy} value={format} onChange={setFormat} options={[{ value: 'archive', label: t('archive.backup') }, ...(exportCommand ? [{ value: 'json', label: t('archive.json') }, { value: 'csv', label: t('archive.csv') }] : [])]} /></View>
      </SettingsSection> : <Text style={[styles.detailSubtitle, styles.contentBlock]}>{t(review ? 'archive.newInventory' : 'archive.restoreDescription')}</Text>}
      {review ? <SettingsSection footer={reviewCopy?.omitted}>
        <View style={styles.navigationRow}>
          <Text style={styles.rowLabel}>{t('archive.name')}</Text>
          <DraftTextField style={[styles.rowLabel, { minHeight: 48 }]} accessibilityLabel={t('archive.name')} value={name} onChangeText={setName} editable={!busy} />
          <Text style={styles.rowContext}>{reviewCopy?.content}</Text>
          <Text style={styles.rowContext}>{reviewCopy?.schema}</Text>
          <Text style={styles.rowContext}>{reviewCopy?.remappings}</Text>
        </View>
      </SettingsSection> : mode === 'export' ? format === 'archive' ? <SettingsSection footer={t(photos && otherFiles ? 'archive.metadata' : 'archive.partial')}>
        <SettingsSwitchRow label={t('archive.photos')} value={photos} disabled={busy} onValueChange={setPhotos} />
        <SettingsSeparator /><SettingsSwitchRow label={t('archive.files')} value={otherFiles} disabled={busy} onValueChange={setOtherFiles} />
      </SettingsSection> : null : <SettingsSection footer={fileName}>
        <SettingsActionRow label={t('archive.chooseFile')} disabled={busy} onPress={() => void run(async visit => { const selected = await visit.pick(); if (!visit.signal.aborted) setFileName(selected); })} />
      </SettingsSection>}
      <View testID="archive-task-actions" style={styles.contentBlock}>
        <NativeCommandButton prominence="primary" label={t(review ? 'archive.restore' : mode === 'import' ? 'archive.upload' : format === 'archive' ? 'archive.create' : format === 'json' ? 'archive.exportJSON' : 'archive.exportCSV')} disabled={busy || !focused || (review ? !name.trim() : mode === 'import' && !fileName)} onPress={primary} />
        {review ? <NativeCommandButton label={t('archive.backToActivity')} disabled={busy} onPress={() => setReview(undefined)} /> : null}
        <Text style={styles.sectionFooter}>{t(review ? 'archive.newInventory' : mode === 'import' ? 'archive.localTransfer' : format === 'archive' ? 'archive.acceptedLeave' : 'archive.directTransfer')}</Text>
        {exporting ? <NativeCommandButton label={t('mobile.InventoryExportAction.cancelExport')} onPress={() => direct.current?.abort()} /> : null}
      </View>
      {busy ? <SettingsLoadingRow label={t('archive.localBusy')} /> : null}
      {error ? <><Text accessibilityRole="alert" style={styles.errorMessage}>{error}</Text>{errorIsActivity ? <SettingsActionRow label={t('archive.retry')} disabled={busy} onPress={() => void run(async visit => {
        const page = await workspace.repository.list(household, undefined, visit.signal);
        if (!visit.signal.aborted) { merge(page.jobs); setCursor(page.nextCursor); setLoading(false); }
      }, true)} /> : null}</> : null}
      {!review ? <SettingsSection title={t('archive.householdActivity')} footer={t(jobs.length ? 'archive.activityLeave' : 'archive.activityEmpty')}>
        {loading ? <SettingsLoadingRow label={t('archive.loading')} /> : null}
        {!loading && !jobs.length ? <Text style={styles.navigationRow}>{t('archive.noJobs')}</Text> : null}
        {jobs.map((job, index) => <View key={job.id} testID={`archive-job-${job.id}`}>
          {index ? <SettingsSeparator /> : null}
          <View style={styles.navigationRow}>
            <Text style={styles.rowLabel}>{t(job.kind === 'export' ? 'archive.backup' : 'archive.restore')}</Text>
            <Text style={styles.rowContext}>{t('archive.created', { date: new Date(job.createdAt).toLocaleString(localization.locale) })}</Text>
            <Text style={styles.rowLabel}>{t(job.phase === 'finalization' && ['queued', 'running'].includes(job.state) ? 'archive.restoring' : job.state === 'running' ? job.kind === 'export' ? 'archive.running' : job.phase === 'validation' ? 'archive.validating' : 'archive.restoring' : `archive.${job.state}`)}</Text>
            {job.kind === 'export' && job.state === 'ready' ? <Text style={styles.rowContext}>{t('archive.expires', { date: new Date(job.expiresAt).toLocaleString(localization.locale) })}</Text> : null}
          </View>
          {job.state === 'ready' ? <SettingsActionRow disabled={busy} label={t(job.kind === 'export' ? 'archive.download' : 'archive.open')} onPress={() => void run(async visit => {
            if (job.kind === 'export') await workspace.files.share(() => workspace.repository.download(household, job.id, visit.signal), visit.signal);
            else if (job.destinationInventoryId) await onOpen(job.destinationInventoryId, visit.signal);
          })} /> : null}
          {job.state === 'awaiting_approval' ? <SettingsActionRow disabled={busy} label={t('archive.review')} onPress={() => void run(async visit => {
            const preview = await workspace.repository.preview(scope.tenantId, job.id, visit.signal);
            if (!visit.signal.aborted) { setReview({ job, preview }); setName(preview.inventoryName); }
          })} /> : null}
          {(job.phase !== 'finalization' || job.state === 'failed') && ['queued', 'running', 'awaiting_approval', 'failed'].includes(job.state) ? <NativeActionMenu disabled={busy} accessibilityLabel={t('archive.jobs')} trigger={{ kind: 'ellipsis' }} groups={[{ id: 'job', items: job.state === 'failed' ? [{ id: 'retry', label: t('archive.retry'), onPress: () => void run(async visit => saveJob(visit, await workspace.repository.retry(household, job.id, visit.signal))) }] : [{ id: 'cancel', label: t('archive.cancel'), isDestructive: true, onPress: () => void run(async visit => { saveJob(visit, await workspace.repository.cancel(household, job.id, visit.signal)); }) }] }]} /> : null}
        </View>)}
        {cursor ? <SettingsActionRow label={t('archive.more')} disabled={busy} onPress={() => void run(async visit => { const page = await workspace.repository.list(household, cursor, visit.signal); if (!visit.signal.aborted) { merge(page.jobs); setCursor(page.nextCursor); } })} /> : null}
      </SettingsSection> : null}
  </ScrollView>;
}

function isActive(job: ArchiveJob) { return ['queued', 'running', 'awaiting_approval'].includes(job.state); }

function isRunning(job: ArchiveJob) { return ['queued', 'running'].includes(job.state); }
