<script lang="ts">
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
<Button.Root variant="outline" aria-expanded={expanded} onclick={() => { expanded = !expanded; }}>{expanded ? 'Hide run comparison' : 'Compare with another run'}</Button.Root>
{#if expanded}<section aria-label="Compare runs" class="run-comparison"><h4>Choose an earlier run</h4>
  {#if heads.isPending}<p role="status">Loading runs…</p>{:else if heads.isError}<p role="alert">Could not load runs. <Button.Root onclick={() => heads.refetch()}>Retry comparison runs</Button.Root></p>
  {:else}<ul>{#each heads.data.items.filter(value => value.id !== current.id) as head (head.id)}<li><Button.Root variant="outline" aria-pressed={baselineId === head.id} onclick={() => { baselineId = head.id; }}>{new Date(head.createdAt).toLocaleString()} · {head.passedCases}/{head.totalCases} passed</Button.Root></li>{/each}</ul>
    {#if !heads.data.items.some(value => value.id !== current.id)}<p>No other runs on this page.</p>{/if}
    {#if heads.data.pagination.hasMore}<Button.Root onclick={() => { cursor = heads.data?.pagination.nextCursor ?? undefined; }}>Next comparison runs</Button.Root>{/if}
    {#if cursor}<Button.Root variant="ghost" onclick={() => { cursor = undefined; }}>First comparison runs</Button.Root>{/if}
  {/if}
  {#if baselineId && baseline.isPending}<p role="status">Loading comparison…</p>{:else if baseline.isError}<p role="alert">Could not load the selected run. <Button.Root onclick={() => baseline.refetch()}>Retry selected run</Button.Root></p>
  {:else if comparison}
    {#if !comparison.compatible}<p role="status">{reasons[comparison.reason]}</p>{#if comparison.reason === 'incomplete'}<Button.Root variant="outline" onclick={() => baseline.refetch()}>Refresh selected run</Button.Root>{/if}
    {:else}<table><caption>Recorded results for the same cases and providers</caption><thead><tr><th scope="col">Measure</th><th scope="col">Selected run</th><th scope="col">This run</th></tr></thead><tbody>
      <tr><th scope="row">Cases passed</th><td>{comparison.baseline.passedCases}/{comparison.cases.length}</td><td>{comparison.candidate.passedCases}/{comparison.cases.length}</td></tr>
      <tr><th scope="row">Model calls</th><td>{comparison.baseline.modelCalls}</td><td>{comparison.candidate.modelCalls}</td></tr>
      <tr><th scope="row">Case execution</th><td>{(comparison.baseline.durationMilliseconds / 1000).toFixed(2)} s</td><td>{(comparison.candidate.durationMilliseconds / 1000).toFixed(2)} s</td></tr>
    </tbody></table>
      <p>Recorded case results exclude queue time and any attempts that did not produce a saved result. Text-only coverage does not measure speech quality.</p>
      <ul>{#each comparison.cases as value}<li><h5>{value.title}</h5><p>Selected run: {value.baseline.passed ? 'Passed' : 'Failed'} · {value.baseline.modelCalls} calls · {(value.baseline.durationMilliseconds / 1000).toFixed(2)} s</p><p>This run: {value.candidate.passed ? 'Passed' : 'Failed'} · {value.candidate.modelCalls} calls · {(value.candidate.durationMilliseconds / 1000).toFixed(2)} s</p></li>{/each}</ul>
    {/if}
  {/if}
</section>{/if}
<style>.run-comparison, ul { display: grid; gap: .75rem; overflow-wrap: anywhere; } ul { list-style: none; padding: 0; } h4, h5 { font-weight: 600; } table { width: 100%; table-layout: fixed; border-collapse: collapse; } th, td { text-align: left; padding: .5rem; border-bottom: 1px solid var(--border); } caption { text-align: left; margin-bottom: .5rem; }</style>
