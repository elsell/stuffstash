<script lang="ts">
  import { onMount } from 'svelte';
  import { createQuery } from '@tanstack/svelte-query';
  import type { ConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey, runPollInterval } from '$lib/adapters/query/conversationQueryClient';
  import { ConversationFailure } from '$lib/domain/conversation';
  import type { ConversationRunRepository } from '$lib/ports/conversationRunRepository';
  import type { ConversationCaseRepository } from '$lib/ports/conversationCaseRepository';
  import * as Button from '$lib/components/ui/button/index.js';
  import RunResult from './RunResult.svelte';
  import RunActivation from './RunActivation.svelte';
  import RunComparison from './RunComparison.svelte';
  import { conversationRunHasCompleteResults } from '$lib/application/conversationRunComparison';
  import type { ConversationWorkflowRepository } from '$lib/ports/conversationWorkflowRepository';
  let { session, runs, cases, workflows, runId, visible }: { session: ConversationSession; runs: ConversationRunRepository; cases: ConversationCaseRepository; workflows?: ConversationWorkflowRepository; runId: string; visible: boolean } = $props();
  let documentVisible = $state(false); let cancelling = $state(false); let message = $state('');
  onMount(() => { const update = () => { documentVisible = document.visibilityState !== 'hidden'; }; update();
    document.addEventListener('visibilitychange', update); return () => document.removeEventListener('visibilitychange', update); });
  const key = () => conversationKey(session.scope, 'run', runId);
  let failedPolls = 0;
  const run = createQuery(() => ({ queryKey: key(), retry: false, queryFn: async ({ signal }) => {
    try { const result = await runs.get(session.scope.tenantId, runId, signal); if (!signal.aborted) failedPolls = 0; return result; }
    catch (error) { if (!signal.aborted) failedPolls = Math.min(4, failedPolls + 1); throw error; }
  },
    refetchInterval: query => query.state.data ? runPollInterval(query.state.data.state, failedPolls, visible && documentVisible && !cancelling) : false
  }), () => session.client);
  const names = { queued: 'Queued', running: 'Running', succeeded: 'Completed', failed: 'Failed', cancelled: 'Cancelled' };
  const pending = $derived(run.data?.state === 'queued' || run.data?.state === 'running');
  async function cancel() {
    if (!run.data || !pending || cancelling) return;
    const version = run.data.version; const requestedId = runId; const requestedKey = key(); cancelling = true; message = '';
    try {
      await session.client.cancelQueries({ queryKey: requestedKey, exact: true });
      await session.mutate(() => runs.cancel(session.scope.tenantId, requestedId, version), value => { session.client.setQueryData(requestedKey, value); });
    } catch (error) {
      if (!session.active) return;
      message = error instanceof ConversationFailure && error.kind === 'conflict' ? 'The run changed before cancellation. Refreshing its current status.' : 'Could not cancel the run. Check its current status and try again.';
      void run.refetch();
    } finally { if (session.active) cancelling = false; }
  }
</script>
{#if run.isPending}<p role="status">Loading run…</p>
{:else if !run.data}<p role="alert">Could not load this run. <Button.Root onclick={() => run.refetch()}>Retry run</Button.Root></p>
{:else}<section class="run-details" aria-label="Evaluation run">
  <h3 aria-live="polite">{names[run.data.state]}</h3>
  <p>Text-only evaluation. This does not test recording, transcription or spoken playback.</p>
  <p>{run.data.completedCases} of {run.data.totalCases} cases completed · {run.data.passedCases} passed</p>
  {#if run.isError}<p role="alert">Status could not be refreshed. Showing the last received result.</p>{/if}
  {#if pending}<Button.Root variant="outline" disabled={cancelling} onclick={cancel}>{cancelling ? 'Cancelling…' : 'Cancel run'}</Button.Root>{/if}
  {#if run.data.failureCode}<p role="alert">The run could not finish. Check the configured provider and try again. <span class="failure-code">Reference: {run.data.failureCode}</span></p>{/if}
  <ul>{#each run.data.cases as pin (pin.revisionId)}{@const result = run.data.results.find(value => value.caseRevisionId === pin.revisionId)}
    <li><h4>{pin.title}</h4>{#if result}<RunResult {session} {cases} {pin} {result} />{:else}<p>{pending ? 'Not run yet' : 'Not completed'} — no passing result recorded.</p>{/if}</li>
  {/each}</ul>
  {#if conversationRunHasCompleteResults(run.data)}<RunComparison {session} {runs} current={run.data} />{/if}
  {#if workflows}<RunActivation {session} {workflows} run={run.data} />{/if}
  <p role="status">{message}</p>
</section>{/if}
<style>.run-details, ul { display: grid; gap: 1rem; } ul { padding: 0; list-style: none; } li { border: 1px solid var(--border); border-radius: var(--radius); padding: 1rem; overflow-wrap: anywhere; } h3, h4 { font-weight: 600; } .failure-code { display: block; }</style>
