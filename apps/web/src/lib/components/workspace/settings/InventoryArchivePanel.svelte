<script lang="ts">
  import { formatArchiveReview } from '@stuff-stash/localization';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { onMount } from 'svelte';
  import { t, localization } from '$lib/presentation/localization';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { ArchiveJob, ArchivePreview, ArchiveScope, InventoryArchiveWorkspace } from '$lib/ports/inventoryArchive';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';

  // The owning settings surface keys this task by principal, household and inventory.
  let { workspace, scope }: { workspace: InventoryArchiveWorkspace; scope: ArchiveScope } = $props();
  let photos = $state(true), otherFiles = $state(true), pending = $state(false), loading = $state(true);
  let jobs = $state<ArchiveJob[]>([]), cursor = $state<string | undefined>(), error = $state(''), message = $state('');
  let file = $state<File | undefined>(), review = $state<{ id: string; preview: ArchivePreview } | undefined>(), name = $state('');
  let fileInput = $state<HTMLInputElement | null>(null);
  let requestKey = '', selectionKey = '', disposed = false, generation = 0, loadedMore = false;
  const lifetime = new AbortController();
  let timer: ReturnType<typeof setTimeout> | undefined;
  function failure(caught: unknown) {
    if ((caught as { name?: string }).name === 'AbortError') return;
    if ((caught as { code?: string }).code === 'archive_streaming_save_required') { error = t('archive.streamingRequired'); return; }
    const status = (caught as { status?: number }).status;
    error = t(status === 401 ? 'archive.signIn' : status === 403 ? 'archive.denied' : status === 413 ? 'archive.tooLarge' : 'archive.error');
    if (status === 401 || status === 403) { jobs = []; review = undefined; }
  }
  async function refresh(more = false) {
    const version = generation;
    try {
      const page = await workspace.repository.list(scope, more ? cursor : undefined, lifetime.signal);
      if (disposed || version !== generation) return;
      const outstanding = jobs.filter(j => (j.state === 'queued' || j.state === 'running') && !page.jobs.some(current => current.id === j.id));
      const updated = await Promise.all(outstanding.map(j => workspace.repository.get(scope, j.id, lifetime.signal)));
      if (disposed || version !== generation) return;
      const combined = new Map(jobs.map(j => [j.id, j]));
      for (const current of [...page.jobs, ...updated]) combined.set(current.id, current);
      jobs = [...combined.values()].sort((a, b) => b.id.localeCompare(a.id));
      if (more) loadedMore = true;
      if (more || !loadedMore) cursor = page.nextCursor;
    } catch (caught) { if (!disposed && version === generation) failure(caught); }
    finally { if (!disposed) loading = false; }
  }
  function schedule() {
    timer = setTimeout(async () => {
      if (disposed) return;
      if (!document.hidden && !pending && !error && jobs.some(j => j.state === 'queued' || j.state === 'running')) await refresh();
      if (!disposed) schedule();
    }, 3000);
  }
  onMount(() => { void refresh(); schedule(); return () => { disposed = true; lifetime.abort(); clearTimeout(timer); }; });
  async function run(action: () => Promise<void>) {
    if (pending || disposed) return;
    generation++;
    pending = true; error = ''; message = '';
    try { await action(); } catch (caught) { if (!disposed) failure(caught); }
    finally { if (!disposed) pending = false; }
  }
  function remember(job: ArchiveJob) { if (!disposed) jobs = [job, ...jobs.filter(old => old.id !== job.id)]; }
  async function create() {
    const selection = JSON.stringify([photos, otherFiles]);
    if (!requestKey || selectionKey !== selection) { requestKey = crypto.randomUUID(); selectionKey = selection; }
    await run(async () => {
      const job = await workspace.repository.create({ tenantId: scope.tenantId, inventoryId: scope.inventoryId! }, requestKey, { photos, otherFiles }, lifetime.signal);
      if (!disposed) { remember(job); requestKey = ''; }
    });
  }
  function choose(event: Event) { file = (event.currentTarget as HTMLInputElement).files?.[0]; requestKey = crypto.randomUUID(); }
  async function upload() {
    if (!file) return;
    const selected = file;
    await run(async () => { const job = await workspace.repository.upload(scope.tenantId, requestKey, selected, lifetime.signal); if (!disposed) { remember(job); file = undefined; if (fileInput) fileInput.value = ''; requestKey = ''; } });
  }
  async function inspect(job: ArchiveJob) {
    await run(async () => { const preview = await workspace.repository.preview(scope.tenantId, job.id, lifetime.signal); if (!disposed) { review = { id: job.id, preview }; name = preview.inventoryName; } });
  }
  async function approve() {
    if (!review || !name.trim()) return;
    const id = review.id;
    await run(async () => { const job = await workspace.repository.approve(scope.tenantId, id, name.trim(), lifetime.signal); if (!disposed) { remember(job); review = undefined; } });
  }
  async function download(job: ArchiveJob) {
    await run(async () => { await workspace.files.save(() => workspace.repository.download(scope, job.id, lifetime.signal), lifetime.signal); if (!disposed) message = t('archive.downloadStarted'); });
  }
  function status(job: ArchiveJob) {
    if (job.phase === 'finalization' && ['queued', 'running'].includes(job.state)) return t('archive.restoring');
    if (job.state === 'running') return t(job.kind === 'export' ? 'archive.running' : job.phase === 'validation' ? 'archive.validating' : 'archive.restoring');
    return t(`archive.${job.state}`);
  }
  const reviewCopy = $derived(review ? formatArchiveReview(t, review.preview) : undefined);
</script>

<section class="settings-resource-group archive-task" aria-label={t(scope.inventoryId ? 'archive.export' : 'archive.restore')}>
  <h2>{t(scope.inventoryId ? 'archive.export' : 'archive.restore')}</h2>
  <p>{t(scope.inventoryId ? 'archive.description' : 'archive.restoreDescription')}</p>
  {#if scope.inventoryId}
    <fieldset disabled={pending}>
      <legend>{t('archive.metadata')}</legend>
      <Label class="flex flex-wrap items-center gap-2"><Checkbox bind:checked={photos} /> {t('archive.photos')}</Label>
      <Label class="flex flex-wrap items-center gap-2"><Checkbox bind:checked={otherFiles} /> {t('archive.files')}</Label>
    </fieldset>
    {#if !photos || !otherFiles}<p>{t('archive.partial')}</p>{/if}
    <div><Button.Root disabled={pending} onclick={create}>{t(pending ? 'archive.busy' : 'archive.create')}</Button.Root></div>
  {:else}
    <Label class="flex flex-wrap items-center gap-2">{t('archive.chooseFile')}<Input bind:ref={fileInput} type="file" accept=".zip,application/zip" onchange={choose} disabled={pending} /></Label>
    <div><Button.Root disabled={pending || !file} onclick={upload}>{t(pending ? 'archive.busy' : 'archive.upload')}</Button.Root></div>
  {/if}
  {#if error}<p role="alert">{error}</p><div><Button.Root variant="outline" disabled={pending} onclick={() => run(() => refresh())}>{t('archive.retry')}</Button.Root></div>{/if}
  {#if message}<p role="status">{message}</p>{/if}
  {#if review}
    <section class="archive-review" aria-label={t('archive.review')}>
      <h3>{t('archive.review')}</h3><p>{t('archive.newInventory')}</p>
      <p>{reviewCopy?.content}</p>
      <p>{reviewCopy?.schema}</p>
      {#if review.preview.omittedAttachments}<p>{reviewCopy?.omitted}</p>{/if}
      {#if review.preview.keyRemappings.length}<p>{reviewCopy?.remappings}</p>{/if}
      <Label class="flex flex-wrap items-center gap-2">{t('archive.name')}<Input name="inventoryName" bind:value={name} maxlength={200} disabled={pending} /></Label>
      <div class="archive-actions"><Button.Root disabled={pending || !name.trim()} onclick={approve}>{t('archive.restore')}</Button.Root><Button.Root variant="outline" disabled={pending} onclick={() => { review = undefined; }}>{t('archive.closeReview')}</Button.Root></div>
    </section>
  {/if}
  <h3>{t('archive.jobs')}</h3>
  {#if loading}<p role="status">{t('archive.loading')}</p>{:else if !jobs.length}<p>{t('archive.noJobs')}</p>{/if}
  <ul class="archive-jobs">
    {#each jobs as job (job.id)}
      <li>
        <div><strong>{status(job)}</strong><p>{t('archive.expires', { date: new Date(job.expiresAt).toLocaleString(localization.locale) })}</p></div>
        <div class="archive-actions">
          {#if job.state === 'ready' && job.kind === 'export'}<Button.Root variant="outline" disabled={pending} onclick={() => download(job)}>{t('archive.download')}</Button.Root>{/if}
          {#if job.state === 'ready' && job.kind === 'restore' && job.destinationInventoryId}<Button.Root href={workspaceRouteHref({ mode: 'home' }, scope.tenantId, job.destinationInventoryId)}>{t('archive.open')}</Button.Root>{/if}
          {#if job.state === 'awaiting_approval'}<Button.Root variant="outline" disabled={pending} onclick={() => inspect(job)}>{t('archive.review')}</Button.Root>{/if}
          {#if job.state === 'failed'}<Button.Root variant="outline" disabled={pending} onclick={() => run(async () => remember(await workspace.repository.retry(scope, job.id, lifetime.signal)))}>{t('archive.retry')}</Button.Root>{/if}
          {#if job.phase !== 'finalization' && ['queued', 'running', 'awaiting_approval', 'failed'].includes(job.state)}<Button.Root variant="outline" disabled={pending} onclick={() => run(async () => { remember(await workspace.repository.cancel(scope, job.id, lifetime.signal)); if (review?.id === job.id) review = undefined; })}>{t('archive.cancel')}</Button.Root>{/if}
        </div>
      </li>
    {/each}
  </ul>
  {#if cursor}<Button.Root variant="outline" disabled={pending} onclick={() => run(() => refresh(true))}>{t('archive.more')}</Button.Root>{/if}
  <p>{t('archive.leave')}</p>
</section>
<style>
  .archive-task { display: grid; gap: 1rem; }
  .archive-task h2, .archive-task h3, .archive-task p { margin: 0; }
  fieldset { border: 0; padding: 0; display: grid; gap: .75rem; }
  legend { margin-bottom: .75rem; }
  .archive-jobs { list-style: none; padding: 0; margin: 0; }
  .archive-jobs li { display: flex; justify-content: space-between; gap: 1rem; flex-wrap: wrap; padding: 1rem 0; border-top: 1px solid var(--border); }
  .archive-actions { display: flex; gap: .5rem; flex-wrap: wrap; align-items: center; }
  .archive-review { display: grid; gap: .75rem; padding: 1rem; border: 1px solid var(--border); border-radius: .5rem; }
</style>
