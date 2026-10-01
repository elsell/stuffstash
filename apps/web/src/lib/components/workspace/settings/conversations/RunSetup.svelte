<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { createQuery } from '@tanstack/svelte-query';
  import type { ConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
  import type { CaseHead } from '$lib/domain/conversationCase';
  import type { WorkflowHead } from '$lib/domain/conversationWorkflow';
  import type { EvaluationRun } from '$lib/domain/conversationRun';
  import { ConversationFailure } from '$lib/domain/conversation';
  import * as Button from '$lib/components/ui/button/index.js';
  let { session, repositories, onQueued, onBusy = () => {} }: { session: ConversationSession; repositories: ConversationWorkspaceRepositories; onQueued: (run: EvaluationRun) => void; onBusy?: (busy: boolean) => void } = $props();
  let workflowCursor = $state<string | undefined>(); let caseCursor = $state<string | undefined>(); let historyCursor = $state<string | undefined>();
  let selectedWorkflow = $state<WorkflowHead | null>(null); let revisionId = $state(''); let selectedCases = $state<CaseHead[]>([]);
  let busy = $state(false); let message = $state('');
  $effect(() => { onBusy(busy); });
  const key = (...parts: string[]) => conversationKey(session.scope, ...parts);
  const workflows = createQuery(() => ({ queryKey: key('workflows', workflowCursor ?? ''), queryFn: ({ signal }) => repositories.workflows.list(session.scope.tenantId, { limit: 20, cursor: workflowCursor }, signal) }), () => session.client);
  const cases = createQuery(() => ({ queryKey: key('cases', caseCursor ?? ''), queryFn: ({ signal }) => repositories.cases.list(session.scope.tenantId, { limit: 20, cursor: caseCursor }, signal) }), () => session.client);
  const profiles = createQuery(() => ({ queryKey: key('models'), queryFn: ({ signal }) => repositories.providers.list(session.scope.tenantId, signal) }), () => session.client);
  const history = createQuery(() => ({ queryKey: key('workflow-history', selectedWorkflow?.id ?? '', historyCursor ?? ''), enabled: !!selectedWorkflow,
    queryFn: ({ signal }) => repositories.workflows.history(session.scope.tenantId, selectedWorkflow!.id, { limit: 20, cursor: historyCursor }, signal) }), () => session.client);
  const revision = createQuery(() => ({ queryKey: key('workflow', selectedWorkflow?.id ?? '', revisionId), enabled: !!selectedWorkflow && !!revisionId, staleTime: Infinity,
    queryFn: ({ signal }) => repositories.workflows.get(session.scope.tenantId, selectedWorkflow!.id, revisionId, signal) }), () => session.client);
  function chooseWorkflow(value: WorkflowHead) { if (busy) return; selectedWorkflow = value; revisionId = value.latestRevisionId; historyCursor = undefined; }
  function toggleCase(value: CaseHead) {
    if (busy) return;
    if (selectedCases.some(item => item.id === value.id)) selectedCases = selectedCases.filter(item => item.id !== value.id);
    else if (selectedCases.length < 100) selectedCases = [...selectedCases, value];
  }
  async function queue() {
    if (busy || !selectedWorkflow || !revision.data || revision.isError || !profiles.isSuccess || selectedCases.length === 0) return;
    const input = { workflowId: selectedWorkflow.id, revisionId: revision.data.id, cases: selectedCases.map(value => ({ caseId: value.id, revisionId: value.latestRevisionId })) };
    busy = true; message = '';
    try { await session.mutate(() => repositories.runs.queue(session.scope.tenantId, input), run => { session.client.setQueryData(key('run', run.id), run); void session.client.invalidateQueries({ queryKey: key('runs') }); onQueued(run); }); }
    catch (error) { if (session.active) message = error instanceof ConversationFailure && (error.kind === 'invalid' || error.kind === 'precondition') ? 'This setup is not ready to run. Check the selected revisions and provider configuration.' : 'Could not queue this run. Your selections are still here.'; }
    finally { if (session.active) busy = false; }
  }
</script>
<section class="run-setup" aria-label={t('web.RunSetup.setUpEvaluation')}>
  <h3>{t('web.RunSetup.setUpATestRun')}</h3><p>{t('web.RunSetup.testYourConfiguredModelsUsingSavedCasesHouseholdInventory')}</p>
  <section aria-label={t('web.RunSetup.chooseWorkflow')}><h4>{t('web.RunSetup.workflow')}</h4>
    {#if workflows.isPending}<p role="status">{t('web.RunSetup.loadingWorkflows')}</p>{:else if workflows.isError}<p role="alert">{t('web.RunSetup.couldNotLoadWorkflows')} <Button.Root onclick={() => workflows.refetch()}>{t('web.RunSetup.retryWorkflows')}</Button.Root></p>
    {:else}<ul>{#each workflows.data.items as value (value.id)}<li><Button.Root variant="outline" disabled={busy} aria-pressed={selectedWorkflow?.id === value.id} onclick={() => chooseWorkflow(value)}>{t('web.RunSetup.revisionFull', { name: value.name, latestRevision: value.latestRevision })}</Button.Root></li>{/each}</ul>
      {#if !workflows.data.items.length}<p>{t('web.RunSetup.saveAWorkflowBeforeRunningCases')}</p>{/if}
      {#if workflows.data.pagination.hasMore}<Button.Root disabled={busy} onclick={() => { workflowCursor = workflows.data?.pagination.nextCursor ?? undefined; }}>{t('web.RunSetup.nextWorkflows')}</Button.Root>{/if}
      {#if workflowCursor}<Button.Root disabled={busy} onclick={() => { workflowCursor = undefined; }}>{t('web.RunSetup.firstWorkflows')}</Button.Root>{/if}
    {/if}
    {#if selectedWorkflow}<p>{t('web.RunSetup.selectedFull', { name: selectedWorkflow.name })}</p>
      <details><summary>{t('web.RunSetup.chooseASavedRevision')}</summary>
        {#if history.isPending}<p role="status">{t('web.RunSetup.loadingRevisions')}</p>{:else if history.isError}<p role="alert">{t('web.RunSetup.couldNotLoadRevisions')} <Button.Root onclick={() => history.refetch()}>{t('web.RunSetup.retryRevisions')}</Button.Root></p>
        {:else}<ul>{#each history.data?.items ?? [] as value (value.id)}<li><Button.Root variant="outline" disabled={busy} aria-pressed={revisionId === value.id} onclick={() => { revisionId = value.id; }}>{t('web.RunSetup.revisionFull2', { number: value.number, name: value.definition.name })}</Button.Root></li>{/each}</ul>
          {#if history.data?.pagination.hasMore}<Button.Root disabled={busy} onclick={() => { historyCursor = history.data?.pagination.nextCursor ?? undefined; }}>{t('web.RunSetup.nextRevisions')}</Button.Root>{/if}
          {#if historyCursor}<Button.Root disabled={busy} onclick={() => { historyCursor = undefined; }}>{t('web.RunSetup.firstRevisions')}</Button.Root>{/if}
        {/if}
      </details>
    {/if}
  </section>
  <section aria-label={t('web.RunSetup.chooseTestCases')}><h4>{t('web.RunSetup.testCases')}</h4>
    {#if cases.isPending}<p role="status">{t('web.RunSetup.loadingCases')}</p>{:else if cases.isError}<p role="alert">{t('web.RunSetup.couldNotLoadCases')} <Button.Root onclick={() => cases.refetch()}>{t('web.RunSetup.retryCases')}</Button.Root></p>
    {:else}<ul>{#each cases.data.items as value (value.id)}{@const selected = selectedCases.some(item => item.id === value.id)}<li><Button.Root variant="outline" disabled={busy || (!selected && selectedCases.length >= 100)} aria-pressed={selected} onclick={() => toggleCase(value)}>{t('web.RunSetup.revisionFull3', { title: value.title, latestRevision: value.latestRevision })}</Button.Root></li>{/each}</ul>
      {#if !cases.data.items.length}<p>{t('web.RunSetup.saveATestCaseBeforeStartingARun')}</p>{/if}
      {#if cases.data.pagination.hasMore}<Button.Root disabled={busy} onclick={() => { caseCursor = cases.data?.pagination.nextCursor ?? undefined; }}>{t('web.RunSetup.nextCases')}</Button.Root>{/if}
      {#if caseCursor}<Button.Root disabled={busy} onclick={() => { caseCursor = undefined; }}>{t('web.RunSetup.firstCases')}</Button.Root>{/if}
    {/if}
    <h4>{t('web.RunSetup.selectedUpTo100Full', { length: selectedCases.length })}</h4><ul>{#each selectedCases as value (value.id)}<li>{t('web.RunSetup.revisionFull3', { title: value.title, latestRevision: value.latestRevision })} <Button.Root variant="ghost" disabled={busy} onclick={() => toggleCase(value)}>{t('web.RunSetup.removeFull', { title: value.title })}</Button.Root></li>{/each}</ul>
  </section>
  {#if selectedWorkflow && revision.isPending}<p role="status">{t('web.RunSetup.loadingSelectedWorkflow')}</p>{:else if revision.isError}<p role="alert">{t('web.RunSetup.couldNotLoadTheSelectedRevision')} <Button.Root onclick={() => revision.refetch()}>{t('web.RunSetup.retrySelectedRevision')}</Button.Root></p>
  {:else if revision.data}<section aria-label={t('web.RunSetup.runUsage')}><h4>{t('web.RunSetup.revisionFull4', { name: revision.data.definition.name, number: revision.data.number })}</h4>
    <p>{t('web.RunSetup.perAttemptUpToModelCallsAcrossCasesFull', { length: revision.data.definition.budget.modelCalls * selectedCases.length, length2: selectedCases.length, elapsedSeconds: revision.data.definition.budget.elapsedSeconds })}</p>
    {#if profiles.isPending}<p role="status">{t('web.RunSetup.loadingModelChoices')}</p>{:else if profiles.isError}<p role="alert">{t('web.RunSetup.couldNotLoadConfiguredModels')} <Button.Root onclick={() => profiles.refetch()}>{t('web.RunSetup.retryModels')}</Button.Root></p>
    {:else}<p>{t('web.RunSetup.model')} {revision.data.definition.providerProfileId ? profiles.data.find(profile => profile.id === revision.data?.definition.providerProfileId)?.name ?? 'Selected profile unavailable' : 'Tenant default model'}</p>{/if}
    <p>{t('web.RunSetup.textOnlyCoverageSpeechInputAndPlaybackNeedSeparate')}</p>
  </section>{/if}
  <Button.Root disabled={busy || !revision.data || revision.isError || !profiles.isSuccess || selectedCases.length === 0} onclick={queue}>{busy ? 'Queueing…' : 'Run selected cases'}</Button.Root><p role="status">{message}</p>
</section>
<style>.run-setup { display: grid; gap: 1.25rem; max-width: 56rem; overflow-wrap: anywhere; } ul { display: grid; gap: .5rem; list-style: none; padding: 0; } h3, h4 { font-weight: 600; } section section { display: grid; gap: .75rem; }</style>
