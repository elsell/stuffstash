import { useCallback, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import { AppState, Text, View } from 'react-native';
import { ArchiveTask } from '../../application/archives/ArchiveTask';
import type { ArchiveJob, ArchivePreview, ArchiveScope, InventoryArchiveWorkspace } from '../../application/archives/InventoryArchive';
import { t } from '../../presentation/localization';
import { NativeFilterSheet } from '../components/NativeFilterSheet';
import { NativeActionMenu } from '../components/NativeActionMenu';
import { DraftTextField } from '../components/DraftTextField';
import { SettingsActionRow, SettingsLoadingRow, SettingsSection, SettingsSeparator, SettingsSwitchRow, useSettingsListStyles } from './SettingsList';

export function InventoryArchiveScreen({ workspace, scope, onClose, onOpen }: {
  workspace: InventoryArchiveWorkspace; scope: ArchiveScope; onClose(): void; onOpen(id: string, signal: AbortSignal): Promise<void>;
}) {
  const { styles } = useSettingsListStyles();
  const [photos, setPhotos] = useState(true), [otherFiles, setOtherFiles] = useState(true);
  const [jobs, setJobs] = useState<ArchiveJob[]>([]), [cursor, setCursor] = useState<string>();
  const [busy, setBusy] = useState(false), [loading, setLoading] = useState(true), [error, setError] = useState('');
  const [fileName, setFileName] = useState<string>();
  const [review, setReview] = useState<{ job: ArchiveJob; preview: ArchivePreview }>();
  const [name, setName] = useState('');
  const task = useRef<ArchiveTask | undefined>(undefined), working = useRef(false);
  const revision = useRef(0), currentJobs = useRef<ArchiveJob[]>([]);
  const store = (value: ArchiveJob[]) => { currentJobs.current = value; setJobs(value); };
  const merge = (values: ArchiveJob[]) => {
    const byId = new Map(currentJobs.current.map(job => [job.id, job]));
    values.forEach(job => byId.set(job.id, job));
    store([...byId.values()].sort((a, b) => b.id.localeCompare(a.id)));
  };
  const report = (caught: unknown) => {
    const status = (caught as { status?: number }).status;
    if (status === 401 || status === 403) { store([]); setReview(undefined); setCursor(undefined); }
    setError(t(status === 401 ? 'archive.signIn' : status === 403 ? 'archive.denied' : (status === 413 || status === 422) ? 'archive.tooLarge' : 'archive.error'));
  };
  useFocusEffect(useCallback(() => {
    const visit = new ArchiveTask(workspace, scope); task.current = visit;
    working.current = false; setBusy(false); store([]); setCursor(undefined); setReview(undefined); setFileName(undefined); setError(''); setLoading(true);
    let timer: ReturnType<typeof setTimeout> | undefined;
    let fetching = false;
    let initialized = false;
    const refresh = async (force = false) => {
      if (initialized && !force && !currentJobs.current.some(job => ['queued', 'running'].includes(job.state))) return;
      if (visit.signal.aborted || fetching || working.current || AppState.currentState !== 'active') return;
      fetching = true; const started = revision.current;
      try {
        const page = await workspace.repository.list(scope, undefined, visit.signal);
        const visible = new Set(page.jobs.map(job => job.id));
        const olderActive = currentJobs.current.filter(job => !visible.has(job.id) && ['queued', 'running'].includes(job.state));
        const older = await Promise.all(olderActive.map(job => workspace.repository.get(scope, job.id, visit.signal)));
        if (!visit.signal.aborted && started === revision.current) {
          const initial = !initialized;
          merge([...page.jobs, ...older]); if (initial) setCursor(page.nextCursor);
          initialized = true; setLoading(false);
        }
      } catch (caught) { if (!visit.signal.aborted && started === revision.current) { initialized = true; report(caught); setLoading(false); } }
      finally { fetching = false; }
    };
    const tick = async () => {
      await refresh();
      if (!visit.signal.aborted) timer = setTimeout(() => void tick(), 3000);
    };
    void tick();
    const subscription = AppState.addEventListener('change', state => { if (state === 'active') void refresh(true); });
    return () => { visit.close(); clearTimeout(timer); subscription.remove(); if (task.current === visit) task.current = undefined; };
  }, [workspace, scope.tenantId, scope.inventoryId]));
  const run = async (operation: (visit: ArchiveTask) => Promise<void>) => {
    const visit = task.current; if (!visit || working.current) return;
    working.current = true; revision.current++; setBusy(true); setError('');
    try { await operation(visit); }
    catch (caught) { if (!visit.signal.aborted) report(caught); }
    finally { if (task.current === visit) { revision.current++; working.current = false; setBusy(false); } }
  };
  const saveJob = (visit: ArchiveTask, job: ArchiveJob) => { if (!visit.signal.aborted) merge([job]); };
  const primary = () => void run(async visit => {
    if (review) {
      const job = await workspace.repository.approve(scope.tenantId, review.job.id, name.trim(), visit.signal);
      saveJob(visit, job); if (!visit.signal.aborted) setReview(undefined);
    } else if (scope.inventoryId) saveJob(visit, await visit.create({ photos, otherFiles }));
    else { saveJob(visit, await visit.upload()); if (!visit.signal.aborted) setFileName(undefined); }
  });
  return <NativeFilterSheet key={review ? 'review' : 'jobs'} footerTestID="archive-task-actions" title={scope.inventoryId ? t('archive.export') : t('archive.restore')} actions={{
    primaryLabel: review ? t('archive.restore') : scope.inventoryId ? t('archive.create') : t('archive.upload'),
    secondaryLabel: t(review ? 'archive.closeReview' : 'archive.close'), disabled: busy || (review ? !name.trim() : !scope.inventoryId && !fileName),
    onApply: primary, onBack: () => { if (review) setReview(undefined); else onClose(); }, secondaryDisabled: false
  }}>
    <View style={{ padding: 20 }}>
      <Text style={styles.detailSubtitle}>{t(review ? 'archive.newInventory' : scope.inventoryId ? 'archive.description' : 'archive.restoreDescription')}</Text>
      {review ? <SettingsSection footer={t('archive.omitted', { count: review.preview.omittedAttachments })}>
        <View style={styles.navigationRow}>
          <Text style={styles.rowLabel}>{t('archive.name')}</Text>
          <DraftTextField style={[styles.rowLabel, { minHeight: 48 }]} accessibilityLabel={t('archive.name')} value={name} onChangeText={setName} editable={!busy} />
          <Text style={styles.rowContext}>{t('archive.counts', { assets: review.preview.assets, tags: review.preview.tags, photos: review.preview.photos, files: review.preview.otherFiles })}</Text>
          <Text style={styles.rowContext}>{t('archive.definitions', { types: review.preview.customAssetTypes, fields: review.preview.customFields })}</Text>
          <Text style={styles.rowContext}>{t('archive.remappings', { count: review.preview.keyRemappings.length })}</Text>
        </View>
      </SettingsSection> : scope.inventoryId ? <SettingsSection footer={t(photos && otherFiles ? 'archive.metadata' : 'archive.partial')}>
        <SettingsSwitchRow label={t('archive.photos')} value={photos} disabled={busy} onValueChange={setPhotos} />
        <SettingsSeparator /><SettingsSwitchRow label={t('archive.files')} value={otherFiles} disabled={busy} onValueChange={setOtherFiles} />
      </SettingsSection> : <SettingsSection footer={fileName}>
        <SettingsActionRow label={t('archive.chooseFile')} disabled={busy} onPress={() => void run(async visit => { const selected = await visit.pick(); if (!visit.signal.aborted) setFileName(selected); })} />
      </SettingsSection>}
      {busy ? <SettingsLoadingRow label={t('archive.busy')} /> : null}
      {error ? <><Text accessibilityRole="alert" style={styles.errorMessage}>{error}</Text><SettingsActionRow label={t('archive.retry')} disabled={busy} onPress={() => void run(async visit => {
        const page = await workspace.repository.list(scope, undefined, visit.signal);
        if (!visit.signal.aborted) { store(page.jobs); setCursor(page.nextCursor); setLoading(false); }
      })} /></> : null}
      {!review ? <SettingsSection title={t('archive.jobs')} footer={t('archive.leave')}>
        {loading ? <SettingsLoadingRow label={t('archive.loading')} /> : null}
        {!loading && !jobs.length ? <Text style={styles.navigationRow}>{t('archive.noJobs')}</Text> : null}
        {jobs.map((job, index) => <View key={job.id}>
          {index ? <SettingsSeparator /> : null}
          <View style={styles.navigationRow}>
            <Text style={styles.rowLabel}>{t(job.phase === 'finalization' && ['queued', 'running'].includes(job.state) ? 'archive.restoring' : job.state === 'running' ? job.kind === 'export' ? 'archive.running' : job.phase === 'validation' ? 'archive.validating' : 'archive.restoring' : `archive.${job.state}`)}</Text>
            <Text style={styles.rowContext}>{t('archive.expires', { date: new Date(job.expiresAt).toLocaleString() })}</Text>
          </View>
          {job.state === 'ready' ? <SettingsActionRow disabled={busy} label={t(job.kind === 'export' ? 'archive.download' : 'archive.open')} onPress={() => void run(async visit => {
            if (job.kind === 'export') await workspace.files.share(() => workspace.repository.download(scope, job.id, visit.signal), visit.signal);
            else if (job.destinationInventoryId) await onOpen(job.destinationInventoryId, visit.signal);
          })} /> : null}
          {job.state === 'awaiting_approval' ? <SettingsActionRow disabled={busy} label={t('archive.review')} onPress={() => void run(async visit => {
            const preview = await workspace.repository.preview(scope.tenantId, job.id, visit.signal);
            if (!visit.signal.aborted) { setReview({ job, preview }); setName(preview.inventoryName); }
          })} /> : null}
          {(job.phase !== 'finalization' || job.state === 'failed') && ['queued', 'running', 'awaiting_approval', 'failed'].includes(job.state) ? <NativeActionMenu disabled={busy} accessibilityLabel={t('archive.jobs')} trigger={{ kind: 'ellipsis' }} groups={[{ id: 'job', items: job.state === 'failed' ? [{ id: 'retry', label: t('archive.retry'), onPress: () => void run(async visit => saveJob(visit, await workspace.repository.retry(scope, job.id, visit.signal))) }] : [{ id: 'cancel', label: t('archive.cancel'), isDestructive: true, onPress: () => void run(async visit => { saveJob(visit, await workspace.repository.cancel(scope, job.id, visit.signal)); }) }] }]} /> : null}
        </View>)}
        {cursor ? <SettingsActionRow label={t('archive.more')} disabled={busy} onPress={() => void run(async visit => { const page = await workspace.repository.list(scope, cursor, visit.signal); if (!visit.signal.aborted) { merge(page.jobs); setCursor(page.nextCursor); } })} /> : null}
      </SettingsSection> : null}
    </View>
  </NativeFilterSheet>;
}
