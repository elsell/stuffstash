<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { createQuery } from '@tanstack/svelte-query';
  import type { ConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import { compareConversationRuns } from '$lib/application/conversationRunComparison';
  import type { EvaluationRun } from '$lib/domain/conversationRun';
  import type { ConversationRunRepository } from '$lib/ports/conversationRunRepository';
  import * as Button from '$lib/components/ui/button/index.js';
  let { session, runs, current }: { session: ConversationSession; runs: ConversationRunRepository; current: EvaluationRun } = $props();
  let expanded = $state(false); let cursor = $state<string | undefined>(); let baselineId = $state('');
  const heads = createQuery(() => ({ queryKey: conversationKey(session.scope, 'runs', cursor ?? ''), enabled: expanded,
    queryFn: ({ signal }) => runs.list(session.scope.tenantId, { limit: 20, cursor }, signal) }), () => session.client);
  const baseline = createQuery(() => ({ queryKey: conversationKey(session.scope, 'run', baselineId), enabled: expanded && !!baselineId,
    queryFn: ({ signal }) => runs.get(session.scope.tenantId, baselineId, signal) }), () => session.client);
  const comparison = $derived(baseline.data ? compareConversationRuns(baseline.data, current) : null);
  const reasons = { same: 'Choose a different run.', incomplete: 'Both runs must finish every case before comparison.', cases: 'The case revisions differ. Run the same saved cases for a controlled comparison.', providers: 'Provider configuration differs. These runs cannot establish the effect of the workflow change alone.' };
</script>
<Button.Root variant="outline" aria-expanded={expanded} onclick={() => { expanded = !expanded; }}>{expanded ? t('web.RunComparison.hideRunComparison') : t('web.RunComparison.compareWithAnotherRun')}</Button.Root>
{#if expanded}<section aria-label={t('web.RunComparison.compareRuns')} class="run-comparison"><h4>{t('web.RunComparison.chooseAnEarlierRun')}</h4>
  {#if heads.isPending}<p role="status">{t('web.RunComparison.loadingRuns')}</p>{:else if heads.isError}<p role="alert">{t('web.RunComparison.couldNotLoadRuns')} <Button.Root onclick={() => heads.refetch()}>{t('web.RunComparison.retryComparisonRuns')}</Button.Root></p>
  {:else}<ul>{#each heads.data.items.filter(value => value.id !== current.id) as head (head.id)}<li><Button.Root variant="outline" aria-pressed={baselineId === head.id} onclick={() => { baselineId = head.id; }}>{t('web.RunComparison.passedFull', { value: new Date(head.createdAt).toLocaleString(), passedCases: head.passedCases, totalCases: head.totalCases })}</Button.Root></li>{/each}</ul>
    {#if !heads.data.items.some(value => value.id !== current.id)}<p>{t('web.RunComparison.noOtherRunsOnThisPage')}</p>{/if}
    {#if heads.data.pagination.hasMore}<Button.Root onclick={() => { cursor = heads.data?.pagination.nextCursor ?? undefined; }}>{t('web.RunComparison.nextComparisonRuns')}</Button.Root>{/if}
    {#if cursor}<Button.Root variant="ghost" onclick={() => { cursor = undefined; }}>{t('web.RunComparison.firstComparisonRuns')}</Button.Root>{/if}
  {/if}
  {#if baselineId && baseline.isPending}<p role="status">{t('web.RunComparison.loadingComparison')}</p>{:else if baseline.isError}<p role="alert">{t('web.RunComparison.couldNotLoadTheSelectedRun')} <Button.Root onclick={() => baseline.refetch()}>{t('web.RunComparison.retrySelectedRun')}</Button.Root></p>
  {:else if comparison}
    {#if !comparison.compatible}<p role="status">{reasons[comparison.reason]}</p>{#if comparison.reason === 'incomplete'}<Button.Root variant="outline" onclick={() => baseline.refetch()}>{t('web.RunComparison.refreshSelectedRun')}</Button.Root>{/if}
    {:else}<table><caption>{t('web.RunComparison.recordedResultsForTheSameCasesAndProviders')}</caption><thead><tr><th scope="col">{t('web.RunComparison.measure')}</th><th scope="col">{t('web.RunComparison.selectedRun')}</th><th scope="col">{t('web.RunComparison.thisRun')}</th></tr></thead><tbody>
      <tr><th scope="row">{t('web.RunComparison.casesPassed')}</th><td>{comparison.baseline.passedCases}/{comparison.cases.length}</td><td>{comparison.candidate.passedCases}/{comparison.cases.length}</td></tr>
      <tr><th scope="row">{t('web.RunComparison.modelCalls')}</th><td>{comparison.baseline.modelCalls}</td><td>{comparison.candidate.modelCalls}</td></tr>
      <tr><th scope="row">{t('web.RunComparison.caseExecution')}</th><td>{t('web.RunComparison.sFull', { value: (comparison.baseline.durationMilliseconds / 1000).toFixed(2) })}</td><td>{t('web.RunComparison.sFull', { value: (comparison.candidate.durationMilliseconds / 1000).toFixed(2) })}</td></tr>
    </tbody></table>
      <p>{t('web.RunComparison.recordedCaseResultsExcludeQueueTimeAndAnyAttempts')}</p>
      <ul>{#each comparison.cases as value}<li><h5>{value.title}</h5><p>{t('web.RunComparison.selectedRunCallsSFull', { value: value.baseline.passed ? 'Passed' : 'Failed', modelCalls: value.baseline.modelCalls, value3: (value.baseline.durationMilliseconds / 1000).toFixed(2) })}</p><p>{t('web.RunComparison.thisRunCallsSFull', { value: value.candidate.passed ? 'Passed' : 'Failed', modelCalls: value.candidate.modelCalls, value3: (value.candidate.durationMilliseconds / 1000).toFixed(2) })}</p></li>{/each}</ul>
    {/if}
  {/if}
</section>{/if}
<style>.run-comparison, ul { display: grid; gap: .75rem; overflow-wrap: anywhere; } ul { list-style: none; padding: 0; } h4, h5 { font-weight: 600; } table { width: 100%; table-layout: fixed; border-collapse: collapse; } th, td { text-align: left; padding: .5rem; border-bottom: 1px solid var(--border); } caption { text-align: left; margin-bottom: .5rem; }</style>
