<script lang="ts">
  import { onDestroy } from 'svelte';
  import { createQuery } from '@tanstack/svelte-query';
  import { createConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import type { ConversationScope } from '$lib/domain/conversation';
  import type { CaseDefinition, CaseRevision } from '$lib/domain/conversationCase';
  import type { ConversationCaseRepository } from '$lib/ports/conversationCaseRepository';
  import * as Button from '$lib/components/ui/button/index.js';
  import CaseEditor from './CaseEditor.svelte';
  import CaseSummary from './CaseSummary.svelte';
  let { scope, cases, onEditingChange = () => {}, onAccessLost = () => {} }: { scope: ConversationScope; cases: ConversationCaseRepository; onEditingChange?: (editing: boolean) => void; onAccessLost?: () => void } = $props();
  let denied = $state(false); let editor = $state<{ revision: CaseRevision | null; definition: CaseDefinition; key: string } | null>(null);
  let comparison = $state<CaseRevision | null>(null); let busy = $state(false); let message = $state(''); let cursor = $state<string | undefined>();
  // svelte-ignore state_referenced_locally -- the parent keys this component by authenticated tenant scope.
  const session = createConversationSession(scope, () => { denied = true; editor = null; comparison = null; onAccessLost(); });
  onDestroy(() => { void session.dispose(); });
  $effect(() => { onEditingChange(editor !== null); });
  const key = (...parts: string[]) => conversationKey(session.scope, ...parts);
  const heads = createQuery(() => ({ queryKey: key('cases', cursor ?? ''), enabled: !denied,
    queryFn: ({ signal }) => cases.list(session.scope.tenantId, { limit: 20, cursor }, signal) }), () => session.client);
  function startNew() {
    if (busy) return;
    editor = { key: 'new', revision: null, definition: { title: '', utterance: '', assets: [], expectations: { kind: 'answer', referencedAssets: [], locations: [], proposals: [], forbiddenOperations: [] } } };
    comparison = null; message = '';
  }
  async function load(caseId: string, compare = false) {
    if (busy) return; busy = true; message = '';
    try {
      const revision = await session.client.fetchQuery({ queryKey: key('case', caseId, 'latest'), staleTime: compare ? 0 : 30_000,
        queryFn: ({ signal }) => cases.get(session.scope.tenantId, caseId, undefined, signal) });
      if (!session.active) return;
      if (compare) comparison = revision; else editor = { key: revision.id, revision, definition: revision.definition };
    } catch { if (session.active) message = 'Could not load the test case. Try again.'; }
    finally { if (session.active) busy = false; }
  }
  async function save(definition: CaseDefinition) {
    const revision = editor?.revision; busy = true;
    try { await session.mutate(() => revision ? cases.append(session.scope.tenantId, revision.caseId, revision.number, definition)
      : cases.create(session.scope.tenantId, definition), saved => {
        session.client.setQueryData(key('case', saved.caseId, 'latest'), saved);
        session.client.setQueryData(key('case', saved.caseId, saved.id), saved);
        void session.client.invalidateQueries({ queryKey: key('cases') });
        editor = { key: saved.id, revision: saved, definition: saved.definition }; comparison = null; message = `Test case revision ${saved.number} saved.`;
      }); } finally { if (session.active) busy = false; }
  }
</script>
{#if denied}<section role="alert"><h2>Test cases unavailable</h2><p>You no longer have access to configure this tenant.</p></section>
{:else}<section class="case-workspace" aria-labelledby="saved-cases-title"><header><h2 id="saved-cases-title">Test cases</h2><p>Check realistic requests against a controlled test inventory.</p></header>
  {#if editor}
    <Button.Root variant="outline" disabled={busy} onclick={() => { editor = null; comparison = null; }}>Close editor and discard unsaved edits</Button.Root>
    {#key editor.key}<CaseEditor initial={editor.definition} onSave={save} onReload={editor.revision ? () => { void load(editor!.revision!.caseId, true); } : undefined} />{/key}
    {#if comparison}<aside aria-label="Latest saved case"><h3>Latest saved revision {comparison.number}</h3><CaseSummary value={comparison.definition} />
      <Button.Root variant="outline" disabled={busy} onclick={() => { if (!busy && comparison) { editor = { key: comparison.id, revision: comparison, definition: comparison.definition }; comparison = null; } }}>Replace my edits with this revision</Button.Root></aside>{/if}
  {:else}<Button.Root disabled={busy} onclick={startNew}>New test case</Button.Root>
    {#if heads.isPending}<p role="status">Loading test cases…</p>{:else if heads.isError}<p role="alert">Could not load test cases. <Button.Root onclick={() => heads.refetch()}>Retry cases</Button.Root></p>
    {:else}<ul>{#each heads.data?.items ?? [] as head (head.id)}<li><Button.Root variant="outline" disabled={busy} onclick={() => load(head.id)}>{head.title} · Revision {head.latestRevision}</Button.Root></li>{/each}</ul>
      {#if !heads.data?.items.length}<p>No saved test cases yet.</p>{/if}
      {#if heads.data?.pagination.hasMore}<Button.Root variant="outline" onclick={() => { cursor = heads.data?.pagination.nextCursor ?? undefined; }}>Next cases</Button.Root>{/if}
      {#if cursor}<Button.Root variant="ghost" onclick={() => { cursor = undefined; }}>Back to first cases</Button.Root>{/if}
    {/if}
  {/if}<p role="status" aria-live="polite">{message}</p>
</section>{/if}
<style>.case-workspace { display: grid; gap: 1rem; max-width: 56rem; } h2, h3 { font-weight: 600; } ul { display: grid; gap: .75rem; list-style: none; padding: 0; } aside { border: 1px solid var(--border); border-radius: var(--radius); padding: 1rem; }</style>
