<script lang="ts">
  import { caseOperationLabel, caseOutcomeLabel } from '$lib/presentation/conversationCaseLabels';
  import { assetKindLabel } from '$lib/presentation/assetKindLabel';
  import { localization, t } from '$lib/presentation/localization';
  import { createQuery } from '@tanstack/svelte-query';
  import type { ConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import type { ConversationCaseRepository } from '$lib/ports/conversationCaseRepository';
  import type { EvaluationCasePin } from '$lib/domain/conversationWorkflow';
  import type { RunResult } from '$lib/domain/conversationRun';
  import * as Button from '$lib/components/ui/button/index.js';
  import CaseSummary from './CaseSummary.svelte';
  let { session, cases, pin, result }: { session: ConversationSession; cases: ConversationCaseRepository; pin: EvaluationCasePin; result: RunResult } = $props();
  let expanded = $state(false);
  const fixture = createQuery(() => ({ queryKey: conversationKey(session.scope, 'case', pin.caseId, pin.revisionId), enabled: expanded,
    staleTime: Infinity, queryFn: ({ signal }) => cases.get(session.scope.tenantId, pin.caseId, pin.revisionId, signal)
  }), () => session.client);
  const title = (id: string) => fixture.data?.definition.assets.find(asset => asset.id === id)?.title ?? t('web.RunResult.unknownFixture');
</script>
<p>{t('web.RunResult.modelCallsSecondsFull', { value: t(result.verdict.passed ? 'evaluation.passed' : 'evaluation.failed'), modelCalls: result.modelCalls, value3: (result.durationMilliseconds / 1000).toFixed(1) })}</p>
<Button.Root variant="outline" aria-expanded={expanded} onclick={() => { expanded = !expanded; }}>{expanded ? t('web.RunResult.hideResult') : t('web.RunResult.compareExpectedAndObserved')}</Button.Root>
{#if expanded}
  {#if fixture.isPending}<p role="status">{t('web.RunResult.loadingExpectedResult')}</p>{:else if fixture.isError}<p role="alert">{t('web.RunResult.couldNotLoadTheSavedExpectations')} <Button.Root onclick={() => fixture.refetch()}>{t('web.RunResult.retryExpectations')}</Button.Root></p>
  {:else if fixture.data}<div class="result-comparison"><section><h5>{t('web.RunResult.expected')}</h5><CaseSummary value={fixture.data.definition} /></section>
    <section><h5>{t('web.RunResult.observed')}</h5><p>{t('web.RunResult.outcomeFull', { kind: caseOutcomeLabel(result.observation.kind) })}</p>
      <p>{t('evaluation.referencedItems', { items: localization.list(result.observation.referencedAssets.map(title)) || t('web.RunResult.none') })}</p>
      <ul>{#each result.observation.locations as location}<li>{t('web.RunResult.insideFull', { value: title(location.assetId), value2: title(location.ancestorId) })}</li>{/each}</ul>
      <ul>{#each result.observation.proposals as proposal}<li>{caseOperationLabel(proposal.operation)}{#if proposal.newKind} ({assetKindLabel(proposal.newKind)}){/if}: {proposal.newTitle || title(proposal.targetId)}{#if proposal.destinationId} → {title(proposal.destinationId)}{/if}{#if proposal.details} · {proposal.details}{/if}</li>{/each}</ul>
      <p>{t('web.RunResult.executedOperationsFull', { value: localization.list(result.observation.executedOperations.map(caseOperationLabel)) || t('evaluation.none') })}</p>
      {#if !result.verdict.passed}<h5>{t('web.RunResult.differences')}</h5><ul>{#each result.verdict.failures as failure}<li>{failure.code.replaceAll('_', ' ')}{#if failure.fixtureId}: {title(failure.fixtureId)}{/if}{#if failure.operation} ({caseOperationLabel(failure.operation)}){/if}</li>{/each}</ul>{/if}
    </section></div>{/if}
{/if}
<style>.result-comparison { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 20rem), 1fr)); gap: 1rem; margin-top: 1rem; overflow-wrap: anywhere; } h5 { font-weight: 600; }</style>
