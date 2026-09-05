<script lang="ts">
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
  const title = (id: string) => fixture.data?.definition.assets.find(asset => asset.id === id)?.title ?? 'Unknown fixture';
</script>
<p>{result.verdict.passed ? 'Passed' : 'Failed'} · {result.modelCalls} model calls · {(result.durationMilliseconds / 1000).toFixed(1)} seconds</p>
<Button.Root variant="outline" aria-expanded={expanded} onclick={() => { expanded = !expanded; }}>{expanded ? 'Hide result' : 'Compare expected and observed'}</Button.Root>
{#if expanded}
  {#if fixture.isPending}<p role="status">Loading expected result…</p>{:else if fixture.isError}<p role="alert">Could not load the saved expectations. <Button.Root onclick={() => fixture.refetch()}>Retry expectations</Button.Root></p>
  {:else if fixture.data}<div class="result-comparison"><section><h5>Expected</h5><CaseSummary value={fixture.data.definition} /></section>
    <section><h5>Observed</h5><p>Outcome: {result.observation.kind}</p>
      <p>Referenced items: {result.observation.referencedAssets.map(title).join(', ') || 'None'}</p>
      <ul>{#each result.observation.locations as location}<li>{title(location.assetId)} inside {title(location.ancestorId)}</li>{/each}</ul>
      <ul>{#each result.observation.proposals as proposal}<li>{proposal.operation}: {proposal.newTitle || title(proposal.targetId)}{#if proposal.destinationId} → {title(proposal.destinationId)}{/if}{#if proposal.details} · {proposal.details}{/if}</li>{/each}</ul>
      <p>Executed operations: {result.observation.executedOperations.join(', ') || 'None'}</p>
      {#if !result.verdict.passed}<h5>Differences</h5><ul>{#each result.verdict.failures as failure}<li>{failure.code.replaceAll('_', ' ')}{#if failure.fixtureId}: {title(failure.fixtureId)}{/if}{#if failure.operation} ({failure.operation}){/if}</li>{/each}</ul>{/if}
    </section></div>{/if}
{/if}
<style>.result-comparison { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 20rem), 1fr)); gap: 1rem; margin-top: 1rem; overflow-wrap: anywhere; } h5 { font-weight: 600; }</style>
