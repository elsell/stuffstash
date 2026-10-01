<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { createQuery } from '@tanstack/svelte-query';
  import type { ConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import type { ConversationWorkflowRepository } from '$lib/ports/conversationWorkflowRepository';
  import type { EvaluationRun } from '$lib/domain/conversationRun';
  import { ConversationFailure } from '$lib/domain/conversation';
  import * as Button from '$lib/components/ui/button/index.js';
  let { session, workflows, run }: { session: ConversationSession; workflows: ConversationWorkflowRepository; run: EvaluationRun } = $props();
  let busy = $state(false); let message = $state(''); let conflict = $state(false);
  const eligible = $derived(run.state === 'succeeded' && run.totalCases > 0 && run.totalCases === run.passedCases && run.cases.length === run.totalCases && run.results.length === run.totalCases && run.results.every(result => result.verdict.passed));
  const key = (...parts: string[]) => conversationKey(session.scope, ...parts);
  const selection = createQuery(() => ({ queryKey: key('selection'), enabled: eligible, staleTime: 0,
    queryFn: ({ signal }) => workflows.selection(session.scope.tenantId, signal) }), () => session.client);
  const revision = createQuery(() => ({ queryKey: key('workflow', run.workflowId, run.revisionId), enabled: eligible, staleTime: Infinity,
    queryFn: ({ signal }) => workflows.get(session.scope.tenantId, run.workflowId, run.revisionId, signal) }), () => session.client);
  const active = $derived(selection.data?.workflowId === run.workflowId && selection.data?.revisionId === run.revisionId);
  async function activate() {
    if (!eligible || busy || conflict || active || !selection.isSuccess || selection.isFetching || !revision.isSuccess) return;
    const workflowId = run.workflowId;
    const input = { revisionId: run.revisionId, runId: run.id, cases: run.cases.map(pin => ({ caseId: pin.caseId, revisionId: pin.revisionId })), expected: selection.data };
    busy = true; message = '';
    try { await session.mutate(async () => {
      const value = await workflows.activate(session.scope.tenantId, workflowId, input);
      await session.client.cancelQueries({ queryKey: key('selection'), exact: true });
      return value;
    }, value => {
      session.client.setQueryData(key('selection'), { workflowId: value.workflowId, revisionId: value.id });
      void session.client.invalidateQueries({ queryKey: key('workflows') }); message = 'Workflow activated.';
    }); } catch (error) { if (session.active) {
      conflict = error instanceof ConversationFailure && error.kind === 'conflict';
      message = conflict ? 'The active selection changed. Check the current selection before trying again.'
        : error instanceof ConversationFailure && error.kind === 'precondition' ? 'This run no longer meets the current quality gate. Run the cases again after checking the workflow and provider configuration.'
        : 'Could not activate this revision. The current selection has not been confirmed changed.';
    } } finally { if (session.active) busy = false; }
  }
  async function refreshSelection() {
    if (busy) return;
    const result = await selection.refetch(); if (session.active && result.isSuccess) { conflict = false; message = 'Current selection refreshed. Review before activating.'; }
  }
</script>
{#if eligible}<section class="run-activation" aria-label={t('web.RunActivation.workflowActivation')}><h4>{t('web.RunActivation.useThisTestedRevision')}</h4>
  {#if revision.isPending || selection.isPending}<p role="status">{t('web.RunActivation.checkingTheCurrentWorkflowSelection')}</p>
  {:else if revision.isError || selection.isError}<p role="alert">{t('web.RunActivation.couldNotCheckTheWorkflowSelection')} <Button.Root onclick={() => { void revision.refetch(); void selection.refetch(); }}>{t('web.RunActivation.retryActivationCheck')}</Button.Root></p>
  {:else if revision.data}<p>{t('web.RunActivation.revisionFull', { name: revision.data.definition.name, number: revision.data.number })}</p>
    {#if active}<p role="status">{t('web.RunActivation.thisRevisionIsActive')}</p>{:else}<p>{t('web.RunActivation.theServerChecksCurrentCasesProvidersAndLimitsFull', { value: selection.data ? 'This will replace the current custom workflow.' : 'This will replace the default conversation workflow.' })}</p>
      <Button.Root disabled={busy || conflict || selection.isFetching} onclick={activate}>{busy ? 'Activating…' : t('web.RunActivation.activateTestedRevision')}</Button.Root>
    {/if}
  {/if}
  {#if conflict}<Button.Root variant="outline" disabled={busy || selection.isFetching} onclick={refreshSelection}>{t('web.RunActivation.checkCurrentSelection')}</Button.Root>{/if}
  <p role="status">{message}</p>
</section>{:else if run.state !== 'queued' && run.state !== 'running'}<p>{t('web.RunActivation.everySelectedCaseMustPassInACompletedRun')}</p>{/if}
<style>.run-activation { display: grid; gap: .75rem; border: 1px solid var(--border); border-radius: var(--radius); padding: 1rem; } h4 { font-weight: 600; }</style>
