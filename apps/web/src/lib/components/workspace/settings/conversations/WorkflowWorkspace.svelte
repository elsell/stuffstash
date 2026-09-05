<script lang="ts">
  import { onDestroy } from 'svelte';
  import { createQuery } from '@tanstack/svelte-query';
  import { createConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import type { ConversationScope } from '$lib/domain/conversation';
  import type { WorkflowDefinition, WorkflowRevision } from '$lib/domain/conversationWorkflow';
  import type { ConversationWorkflowRepository } from '$lib/ports/conversationWorkflowRepository';
  import type { ConversationProviderRepository } from '$lib/ports/conversationProviderRepository';
  import * as Button from '$lib/components/ui/button/index.js';
  import WorkflowEditor from './WorkflowEditor.svelte';
  let { scope, workflows, providers }: { scope: ConversationScope; workflows: ConversationWorkflowRepository; providers: ConversationProviderRepository } = $props();
  let denied = $state(false);
  // svelte-ignore state_referenced_locally -- the parent keys this workspace by authenticated tenant scope.
  const session = createConversationSession(scope, () => { denied = true; editor = null; comparison = null; });
  onDestroy(() => { void session.dispose(); });
  let cursor = $state<string | undefined>();
  let editor = $state<{ revision: WorkflowRevision | null; definition: WorkflowDefinition; key: string } | null>(null);
  let comparison = $state<WorkflowRevision | null>(null);
  let busy = $state(false);
  let message = $state('');
  const key = (...parts: string[]) => conversationKey(session.scope, ...parts);
  const heads = createQuery(() => ({ queryKey: key('workflows', cursor ?? ''), enabled: !denied,
    queryFn: ({ signal }) => workflows.list(session.scope.tenantId, { limit: 20, cursor }, signal) }), () => session.client);
  const selection = createQuery(() => ({ queryKey: key('selection'), enabled: !denied,
    queryFn: ({ signal }) => workflows.selection(session.scope.tenantId, signal) }), () => session.client);
  const models = createQuery(() => ({ queryKey: key('models'), enabled: !denied,
    queryFn: ({ signal }) => providers.list(session.scope.tenantId, signal) }), () => session.client);
  function startNew() {
    if (busy) return;
    editor = { key: 'new', revision: null, definition: { name: '', retrieval: 'precise_first', response: 'generated_with_grounded_fallback',
      budget: { evidenceRounds: 3, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 }, steps: [
        { kind: 'interpret', attempts: 1, instructions: '', providerProfileId: null },
        { kind: 'assess', attempts: 1, instructions: '', providerProfileId: null },
        { kind: 'respond', attempts: 1, instructions: '', providerProfileId: null }] } };
    message = ''; comparison = null;
  }
  async function load(workflowId: string, compare = false) {
    if (busy) return;
    busy = true; message = '';
    try {
      const revision = await session.client.fetchQuery({ queryKey: key('workflow', workflowId, 'latest'), staleTime: compare ? 0 : 30_000,
        queryFn: ({ signal }) => workflows.get(session.scope.tenantId, workflowId, undefined, signal) });
      if (!session.active) return;
      if (compare) comparison = revision;
      else editor = { key: revision.id, revision, definition: revision.definition };
    } catch { if (session.active) message = 'Could not load the workflow. Try again.'; }
    finally { if (session.active) busy = false; }
  }
  async function save(definition: WorkflowDefinition) {
    const revision = editor?.revision;
    busy = true;
    try { await session.mutate(() => revision
      ? workflows.append(session.scope.tenantId, revision.workflowId, revision.number, definition)
      : workflows.create(session.scope.tenantId, definition), saved => {
        session.client.setQueryData(key('workflow', saved.workflowId, 'latest'), saved);
        session.client.setQueryData(key('workflow', saved.workflowId, saved.id), saved);
        void session.client.invalidateQueries({ queryKey: key('workflows') });
        editor = { key: saved.id, revision: saved, definition: saved.definition }; comparison = null;
        message = `Draft revision ${saved.number} saved. Run test cases before activation.`;
      }); } finally { if (session.active) busy = false; }
  }
</script>

{#if denied}
  <section role="alert"><h2>Conversation settings unavailable</h2><p>Your account no longer has access to configure this tenant.</p></section>
{:else}
  <section class="workflow-workspace" aria-labelledby="conversation-workflows-title">
    <header><h1 id="conversation-workflows-title">Conversations</h1><p>Tune how your configured models work with your inventory.</p></header>
    {#if selection.isError}<p role="alert">Could not load the active workflow. <Button.Root variant="outline" onclick={() => selection.refetch()}>Retry active workflow</Button.Root></p>
    {:else if selection.isPending}<p role="status">Loading active workflow…</p>
    {:else}<p>{selection.data ? `Active workflow: ${heads.data?.items.find(head => head.id === selection.data?.workflowId)?.name ?? 'Saved workflow'}` : 'Using the default conversation workflow.'}</p>{/if}
    {#if editor}
      <Button.Root variant="outline" disabled={busy} onclick={() => { editor = null; comparison = null; }}>Close editor and discard unsaved edits</Button.Root>
      {#if models.isPending}<p role="status">Loading configured models…</p>
      {:else if models.isError}<p role="alert">Could not load configured models. <Button.Root onclick={() => models.refetch()}>Retry models</Button.Root></p>
      {:else}
        {#key editor.key}<WorkflowEditor initial={editor.definition} providers={models.data ?? []} onSave={save}
          onReload={editor.revision ? () => { void load(editor!.revision!.workflowId, true); } : undefined} />{/key}
      {/if}
      {#if comparison}
        <aside aria-label="Latest revision comparison"><h3>Latest saved revision {comparison.number}</h3><p>{comparison.definition.name}</p>
          <dl><dt>Search</dt><dd>{comparison.definition.retrieval === 'precise_first' ? 'Precise matches first' : 'Broader discovery'}</dd>
            <dt>Answer style</dt><dd>{comparison.definition.response === 'grounded' ? 'Grounded facts' : 'Model answer with grounded recovery'}</dd>
            <dt>Shared limits</dt><dd>{comparison.definition.budget.evidenceRounds} searches · {comparison.definition.budget.modelCalls} model calls · {comparison.definition.budget.elapsedSeconds} seconds · {comparison.definition.budget.followUpTurns} follow-ups</dd></dl>
          {#each comparison.definition.steps as step}<section><h4>{step.kind === 'interpret' ? 'Understand' : step.kind === 'assess' ? 'Look up and assess' : 'Respond'}</h4>
            <p>{models.data?.find(model => model.id === step.providerProfileId)?.name ?? (step.providerProfileId ? 'Saved model profile' : 'Tenant default model')} · {step.attempts} attempts</p>
            <p class="instructions">{step.instructions || 'No additional instructions'}</p></section>{/each}
          <Button.Root variant="outline" disabled={busy} onclick={() => { if (!busy && comparison) { editor = { key: comparison.id, revision: comparison, definition: comparison.definition }; comparison = null; } }}>Replace my edits with this revision</Button.Root>
        </aside>
      {/if}
    {:else}
      <Button.Root disabled={busy} onclick={startNew}>New workflow</Button.Root>
      {#if heads.isPending}<p role="status">Loading workflows…</p>
      {:else if heads.isError}<p role="alert">Could not load workflows. <Button.Root onclick={() => heads.refetch()}>Retry workflows</Button.Root></p>
      {:else}
        <ul>{#each heads.data?.items ?? [] as head (head.id)}<li><Button.Root variant="outline" disabled={busy} onclick={() => load(head.id)}>{head.name} · Revision {head.latestRevision}</Button.Root></li>{/each}</ul>
        {#if !heads.data?.items.length}<p>No saved workflows yet. Create one to start tuning your conversations.</p>{/if}
        {#if heads.data?.pagination.hasMore}<Button.Root variant="outline" onclick={() => { cursor = heads.data?.pagination.nextCursor ?? undefined; }}>Next workflows</Button.Root>{/if}
        {#if cursor}<Button.Root variant="ghost" onclick={() => { cursor = undefined; }}>Back to first workflows</Button.Root>{/if}
      {/if}
    {/if}
    <p role="status" aria-live="polite">{message}</p>
  </section>
{/if}
<style>
  .workflow-workspace { display: grid; gap: 1rem; max-width: 56rem; }
  h1 { font-size: 1.6rem; font-weight: 650; } header p { color: var(--muted-foreground); }
  ul { display: grid; gap: .75rem; list-style: none; padding: 0; }
  .instructions { white-space: pre-wrap; overflow-wrap: anywhere; }
  dt, h4 { font-weight: 600; }
  aside { padding: 1rem; border: 1px solid var(--border); border-radius: var(--radius); }
</style>
