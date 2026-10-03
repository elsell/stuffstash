<script lang="ts">
  import { localization, t } from '$lib/presentation/localization';
  import { createQuery } from '@tanstack/svelte-query';
  import type { ConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import { compareConversationRuns } from '$lib/application/conversationRunComparison';
  import type { EvaluationRun } from '$lib/domain/conversationRun';
  import type { ConversationRunRepository } from '$lib/ports/conversationRunRepository';
  import * as Button from '$lib/components/ui/button/index.js';
  let { session, runs, current }: { session: ConversationSession; runs: ConversationRunRepository; current: EvaluationRun } = $props();
  const seconds = (milliseconds: number) => localization.number(milliseconds / 1000, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  let expanded = $state(false); let cursor = $state<string | undefined>(); let baselineId = $state('');
  const heads = createQuery(() => ({ queryKey: conversationKey(session.scope, 'runs', cursor ?? ''), enabled: expanded,
    queryFn: ({ signal }) => runs.list(session.scope.tenantId, { limit: 20, cursor }, signal) }), () => session.client);
  const baseline = createQuery(() => ({ queryKey: conversationKey(session.scope, 'run', baselineId), enabled: expanded && !!baselineId,
    queryFn: ({ signal }) => runs.get(session.scope.tenantId, baselineId, signal) }), () => session.client);
  const comparison = $derived(baseline.data ? compareConversationRuns(baseline.data, current) : null);
  const reasons = { same: t('web.RunComparison.chooseADifferentRun'), incomplete: t('web.RunComparison.bothRunsMustFinishEveryCaseBeforeComparison'), cases: t('web.RunComparison.theCaseRevisionsDifferRunTheSameSavedCases'), providers: t('web.RunComparison.providerConfigurationDiffersTheseRunsCannotEstablishTheEffect') };
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
      <tr><th scope="row">{t('web.RunComparison.casesPassed')}</th><td>{localization.number(comparison.baseline.passedCases)}/{localization.number(comparison.cases.length)}</td><td>{localization.number(comparison.candidate.passedCases)}/{localization.number(comparison.cases.length)}</td></tr>
      <tr><th scope="row">{t('web.RunComparison.modelCalls')}</th><td>{localization.number(comparison.baseline.modelCalls)}</td><td>{localization.number(comparison.candidate.modelCalls)}</td></tr>
      <tr><th scope="row">{t('web.RunComparison.caseExecution')}</th><td>{t('web.RunComparison.sFull', { value: seconds(comparison.baseline.durationMilliseconds) })}</td><td>{t('web.RunComparison.sFull', { value: seconds(comparison.candidate.durationMilliseconds) })}</td></tr>
    </tbody></table>
      <p>{t('web.RunComparison.recordedCaseResultsExcludeQueueTimeAndAnyAttempts')}</p>
      <ul>{#each comparison.cases as value}<li><h5>{value.title}</h5><p>{t('evaluation.comparisonSelected', { verdict: t(value.baseline.passed ? 'evaluation.passed' : 'evaluation.failed'), count: value.baseline.modelCalls, seconds: seconds(value.baseline.durationMilliseconds) })}</p><p>{t('evaluation.comparisonCurrent', { verdict: t(value.candidate.passed ? 'evaluation.passed' : 'evaluation.failed'), count: value.candidate.modelCalls, seconds: seconds(value.candidate.durationMilliseconds) })}</p></li>{/each}</ul>
    {/if}
  {/if}
</section>{/if}
<style>.run-comparison, ul { display: grid; gap: .75rem; overflow-wrap: anywhere; } ul { list-style: none; padding: 0; } h4, h5 { font-weight: 600; } table { width: 100%; table-layout: fixed; border-collapse: collapse; } th, td { text-align: left; padding: .5rem; border-bottom: 1px solid var(--border); } caption { text-align: left; margin-bottom: .5rem; }</style>
