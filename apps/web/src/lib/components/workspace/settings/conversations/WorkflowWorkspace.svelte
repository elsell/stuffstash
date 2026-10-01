<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { onDestroy } from 'svelte';
  import { createQuery } from '@tanstack/svelte-query';
  import { createConversationSession, type ConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import type { ConversationScope } from '$lib/domain/conversation';
  import type { WorkflowDefinition, WorkflowRevision } from '$lib/domain/conversationWorkflow';
  import type { ConversationWorkflowRepository } from '$lib/ports/conversationWorkflowRepository';
  import type { ConversationProviderRepository } from '$lib/ports/conversationProviderRepository';
  import * as Button from '$lib/components/ui/button/index.js';
  import WorkflowEditor from './WorkflowEditor.svelte';
  let { scope, session: sharedSession, workflows, providers, onNavigationBlockedChange = () => {}, onAccessLost = () => {} }: { scope: ConversationScope; session?: ConversationSession; workflows: ConversationWorkflowRepository; providers: ConversationProviderRepository; onNavigationBlockedChange?: (blocked: boolean) => void; onAccessLost?: () => void } = $props();
  let denied = $state(false);
  // svelte-ignore state_referenced_locally -- the parent keys this workspace by authenticated tenant scope.
  const session = sharedSession ?? createConversationSession(scope, () => { denied = true; editor = null; comparison = null; onAccessLost(); });
  // svelte-ignore state_referenced_locally -- session ownership is fixed for this mount.
  const ownsSession = sharedSession === undefined;
  onDestroy(() => { if (ownsSession) void session.dispose(); });
  $effect(() => { onNavigationBlockedChange(editor !== null || busy); });
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
    editor = { key: 'new', revision: null, definition: { name: '', providerProfileId: null, instructions: '',
      budget: { toolCalls: 6, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 } } };
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
        void session.client.invalidateQueries({ queryKey: key('workflow-history', saved.workflowId) });
        editor = { key: saved.id, revision: saved, definition: saved.definition }; comparison = null;
        message = `Draft revision ${saved.number} saved. Run test cases before activation.`;
      }); } finally { if (session.active) busy = false; }
  }
</script>

{#if denied}
  <section role="alert"><h2>{t('web.WorkflowWorkspace.conversationSettingsUnavailable')}</h2><p>{t('web.WorkflowWorkspace.yourAccountNoLongerHasAccessToConfigureThis')}</p></section>
{:else}
  <section class="workflow-workspace" aria-labelledby="conversation-workflows-title">
    <header><h1 id="conversation-workflows-title">{t('web.WorkflowWorkspace.conversations')}</h1><p>{t('web.WorkflowWorkspace.tuneHowYourConfiguredModelsWorkWithYourInventory')}</p></header>
    {#if selection.isError}<p role="alert">{t('web.WorkflowWorkspace.couldNotLoadTheActiveWorkflow')} <Button.Root variant="outline" onclick={() => selection.refetch()}>{t('web.WorkflowWorkspace.retryActiveWorkflow')}</Button.Root></p>
    {:else if selection.isPending}<p role="status">{t('web.WorkflowWorkspace.loadingActiveWorkflow')}</p>
    {:else}<p>{selection.data ? t('web.WorkflowWorkspace.activeWorkflow', { value: String(heads.data?.items.find(head => head.id === selection.data?.workflowId)?.name ?? 'Saved workflow') }) : t('web.WorkflowWorkspace.usingTheDefaultConversationWorkflow')}</p>{/if}
    {#if editor}
      {#if editor.revision?.settingsMigration}<p role="status">{t('web.WorkflowWorkspace.thisRevisionWasConvertedFromThePreviousWorkflowFormat')}</p>{/if}
      <Button.Root variant="outline" disabled={busy} onclick={() => { editor = null; comparison = null; }}>{t('web.WorkflowWorkspace.closeEditorAndDiscardUnsavedEdits')}</Button.Root>
      {#if models.isPending}<p role="status">{t('web.WorkflowWorkspace.loadingConfiguredModels')}</p>
      {:else if models.isError}<p role="alert">{t('web.WorkflowWorkspace.couldNotLoadConfiguredModels')} <Button.Root onclick={() => models.refetch()}>{t('web.WorkflowWorkspace.retryModels')}</Button.Root></p>
      {:else}
        {#key editor.key}<WorkflowEditor disabled={busy} initial={editor.definition} providers={models.data ?? []} onSave={save}
          onReload={editor.revision ? () => { void load(editor!.revision!.workflowId, true); } : undefined} />{/key}
      {/if}
      {#if comparison}
        <aside aria-label={t('web.WorkflowWorkspace.latestRevisionComparison')}><h3>{t('web.WorkflowWorkspace.latestSavedRevisionFull', { number: comparison.number })}</h3><p>{comparison.definition.name}</p>
          <dl><dt>{t('web.WorkflowWorkspace.model')}</dt><dd>{models.data?.find(model => model.id === comparison?.definition.providerProfileId)?.name ?? (comparison.definition.providerProfileId ? t('web.WorkflowWorkspace.savedModelProfile') : t('web.WorkflowWorkspace.tenantDefaultModel'))}</dd>
            <dt>{t('web.WorkflowWorkspace.perTurnLimits')}</dt><dd>{t('web.WorkflowWorkspace.toolCallsModelCallsSecondsFull', { toolCalls: comparison.definition.budget.toolCalls, modelCalls: comparison.definition.budget.modelCalls, elapsedSeconds: comparison.definition.budget.elapsedSeconds })}</dd>
            <dt>{t('web.WorkflowWorkspace.followUps')}</dt><dd>{comparison.definition.budget.followUpTurns}</dd></dl>
          <p class="instructions">{comparison.definition.instructions || t('web.WorkflowWorkspace.noAdditionalInstructions')}</p>
          <Button.Root variant="outline" disabled={busy} onclick={() => { if (!busy && comparison) { editor = { key: comparison.id, revision: comparison, definition: comparison.definition }; comparison = null; } }}>{t('web.WorkflowWorkspace.replaceMyEditsWithThisRevision')}</Button.Root>
        </aside>
      {/if}
    {:else}
      <Button.Root disabled={busy} onclick={startNew}>{t('web.WorkflowWorkspace.newWorkflow')}</Button.Root>
      {#if heads.isPending}<p role="status">{t('web.WorkflowWorkspace.loadingWorkflows')}</p>
      {:else if heads.isError}<p role="alert">{t('web.WorkflowWorkspace.couldNotLoadWorkflows')} <Button.Root onclick={() => heads.refetch()}>{t('web.WorkflowWorkspace.retryWorkflows')}</Button.Root></p>
      {:else}
        <ul>{#each heads.data?.items ?? [] as head (head.id)}<li><Button.Root variant="outline" disabled={busy} onclick={() => load(head.id)}>{t('web.WorkflowWorkspace.revisionFull', { name: head.name, latestRevision: head.latestRevision })}</Button.Root></li>{/each}</ul>
        {#if !heads.data?.items.length}<p>{t('web.WorkflowWorkspace.noSavedWorkflowsYetCreateOneToStartTuning')}</p>{/if}
        {#if heads.data?.pagination.hasMore}<Button.Root variant="outline" onclick={() => { cursor = heads.data?.pagination.nextCursor ?? undefined; }}>{t('web.WorkflowWorkspace.nextWorkflows')}</Button.Root>{/if}
        {#if cursor}<Button.Root variant="ghost" onclick={() => { cursor = undefined; }}>{t('web.WorkflowWorkspace.backToFirstWorkflows')}</Button.Root>{/if}
      {/if}
    {/if}
    <p role="status" aria-live="polite">{message}</p>
  </section>
{/if}
<style>
  .workflow-workspace { display: grid; gap: 1rem; max-width: 56rem; }
  h1 { font-size: 1.6rem; font-weight: 600; } header p { color: var(--muted-foreground); }
  ul { display: grid; gap: .75rem; list-style: none; padding: 0; }
  .instructions { white-space: pre-wrap; overflow-wrap: anywhere; }
  dt, h4 { font-weight: 600; }
  aside { padding: 1rem; border: 1px solid var(--border); border-radius: var(--radius); }
</style>
