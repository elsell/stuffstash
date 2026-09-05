<script lang="ts">
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
<section class="run-setup" aria-label="Set up evaluation">
  <h3>Set up a test run</h3><p>Test your configured models using saved cases. Household inventory is not changed.</p>
  <section aria-label="Choose workflow"><h4>Workflow</h4>
    {#if workflows.isPending}<p role="status">Loading workflows…</p>{:else if workflows.isError}<p role="alert">Could not load workflows. <Button.Root onclick={() => workflows.refetch()}>Retry workflows</Button.Root></p>
    {:else}<ul>{#each workflows.data.items as value (value.id)}<li><Button.Root variant="outline" disabled={busy} aria-pressed={selectedWorkflow?.id === value.id} onclick={() => chooseWorkflow(value)}>{value.name} · Revision {value.latestRevision}</Button.Root></li>{/each}</ul>
      {#if !workflows.data.items.length}<p>Save a workflow before running cases.</p>{/if}
      {#if workflows.data.pagination.hasMore}<Button.Root disabled={busy} onclick={() => { workflowCursor = workflows.data?.pagination.nextCursor ?? undefined; }}>Next workflows</Button.Root>{/if}
      {#if workflowCursor}<Button.Root disabled={busy} onclick={() => { workflowCursor = undefined; }}>First workflows</Button.Root>{/if}
    {/if}
    {#if selectedWorkflow}<p>Selected: {selectedWorkflow.name}</p>
      <details><summary>Choose a saved revision</summary>
        {#if history.isPending}<p role="status">Loading revisions…</p>{:else if history.isError}<p role="alert">Could not load revisions. <Button.Root onclick={() => history.refetch()}>Retry revisions</Button.Root></p>
        {:else}<ul>{#each history.data?.items ?? [] as value (value.id)}<li><Button.Root variant="outline" disabled={busy} aria-pressed={revisionId === value.id} onclick={() => { revisionId = value.id; }}>Revision {value.number} · {value.definition.name}</Button.Root></li>{/each}</ul>
          {#if history.data?.pagination.hasMore}<Button.Root disabled={busy} onclick={() => { historyCursor = history.data?.pagination.nextCursor ?? undefined; }}>Next revisions</Button.Root>{/if}
          {#if historyCursor}<Button.Root disabled={busy} onclick={() => { historyCursor = undefined; }}>First revisions</Button.Root>{/if}
        {/if}
      </details>
    {/if}
  </section>
  <section aria-label="Choose test cases"><h4>Test cases</h4>
    {#if cases.isPending}<p role="status">Loading cases…</p>{:else if cases.isError}<p role="alert">Could not load cases. <Button.Root onclick={() => cases.refetch()}>Retry cases</Button.Root></p>
    {:else}<ul>{#each cases.data.items as value (value.id)}{@const selected = selectedCases.some(item => item.id === value.id)}<li><Button.Root variant="outline" disabled={busy || (!selected && selectedCases.length >= 100)} aria-pressed={selected} onclick={() => toggleCase(value)}>{value.title} · Revision {value.latestRevision}</Button.Root></li>{/each}</ul>
      {#if !cases.data.items.length}<p>Save a test case before starting a run.</p>{/if}
      {#if cases.data.pagination.hasMore}<Button.Root disabled={busy} onclick={() => { caseCursor = cases.data?.pagination.nextCursor ?? undefined; }}>Next cases</Button.Root>{/if}
      {#if caseCursor}<Button.Root disabled={busy} onclick={() => { caseCursor = undefined; }}>First cases</Button.Root>{/if}
    {/if}
    <h4>{selectedCases.length} selected (up to 100)</h4><ul>{#each selectedCases as value (value.id)}<li>{value.title} · Revision {value.latestRevision} <Button.Root variant="ghost" disabled={busy} onclick={() => toggleCase(value)}>Remove {value.title}</Button.Root></li>{/each}</ul>
  </section>
  {#if selectedWorkflow && revision.isPending}<p role="status">Loading selected workflow…</p>{:else if revision.isError}<p role="alert">Could not load the selected revision. <Button.Root onclick={() => revision.refetch()}>Retry selected revision</Button.Root></p>
  {:else if revision.data}<section aria-label="Run usage"><h4>{revision.data.definition.name} · Revision {revision.data.number}</h4>
    <p>Per attempt, up to {revision.data.definition.budget.modelCalls * selectedCases.length} model calls across {selectedCases.length} cases; each case has a {revision.data.definition.budget.elapsedSeconds}-second budget. Recovering an interrupted run may add model calls.</p>
    {#if profiles.isPending}<p role="status">Loading model choices…</p>{:else if profiles.isError}<p role="alert">Could not load configured models. <Button.Root onclick={() => profiles.refetch()}>Retry models</Button.Root></p>
    {:else}<ul>{#each revision.data.definition.steps as step (step.kind)}<li>{step.kind === 'interpret' ? 'Understand' : step.kind === 'assess' ? 'Look up and assess' : 'Respond'}: {step.kind === 'respond' && revision.data.definition.response === 'grounded' ? 'Grounded answer; no model call' : step.providerProfileId ? profiles.data.find(profile => profile.id === step.providerProfileId)?.name ?? 'Selected profile unavailable' : 'Tenant default model'}</li>{/each}</ul>{/if}
    <p>Text-only coverage. Speech input and playback need separate testing.</p>
  </section>{/if}
  <Button.Root disabled={busy || !revision.data || revision.isError || !profiles.isSuccess || selectedCases.length === 0} onclick={queue}>{busy ? 'Queueing…' : 'Run selected cases'}</Button.Root><p role="status">{message}</p>
</section>
<style>.run-setup { display: grid; gap: 1.25rem; max-width: 56rem; overflow-wrap: anywhere; } ul { display: grid; gap: .5rem; list-style: none; padding: 0; } h3, h4 { font-weight: 600; } section section { display: grid; gap: .75rem; }</style>
