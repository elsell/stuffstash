<script lang="ts">
  import { onDestroy } from 'svelte';
  import { createQuery } from '@tanstack/svelte-query';
  import { createConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import type { ConversationScope } from '$lib/domain/conversation';
  import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
  import * as Button from '$lib/components/ui/button/index.js';
  import RunSetup from './RunSetup.svelte';
  import RunDetails from './RunDetails.svelte';
  let { scope, repositories, visible, onAccessLost = () => {}, onNavigationBlockedChange = () => {} }: { scope: ConversationScope; repositories: ConversationWorkspaceRepositories; visible: boolean; onAccessLost?: () => void; onNavigationBlockedChange?: (blocked: boolean) => void } = $props();
  let denied = $state(false); let creating = $state(false); let busy = $state(false); let selectedId = $state(''); let cursor = $state<string | undefined>();
  // svelte-ignore state_referenced_locally -- parent keys the workspace by authenticated scope.
  const session = createConversationSession(scope, () => { denied = true; creating = false; selectedId = ''; onAccessLost(); });
  onDestroy(() => { void session.dispose(); });
  $effect(() => { onNavigationBlockedChange(creating); });
  const heads = createQuery(() => ({ queryKey: conversationKey(session.scope, 'runs', cursor ?? ''), enabled: !denied,
    queryFn: ({ signal }) => repositories.runs.list(session.scope.tenantId, { limit: 20, cursor }, signal) }), () => session.client);
  const names = { queued: 'Queued', running: 'Running', succeeded: 'Completed', failed: 'Failed', cancelled: 'Cancelled' };
</script>
{#if denied}<section role="alert"><h2>Runs unavailable</h2><p>You no longer have access to configure this tenant.</p></section>
{:else}<section class="run-workspace" aria-labelledby="runs-title"><h2 id="runs-title">Runs</h2>
  {#if creating}<Button.Root variant="outline" disabled={busy} onclick={() => { if (!busy) creating = false; }}>Discard run setup</Button.Root>
    <RunSetup {session} {repositories} onBusy={value => { busy = value; }} onQueued={run => { creating = false; busy = false; selectedId = run.id; }} />
  {:else if selectedId}<Button.Root variant="outline" onclick={() => { selectedId = ''; void heads.refetch(); }}>Back to runs</Button.Root>
    {#key selectedId}<RunDetails {session} runs={repositories.runs} cases={repositories.cases} runId={selectedId} {visible} />{/key}
  {:else}<Button.Root onclick={() => { creating = true; }}>New run</Button.Root>
    {#if heads.isPending}<p role="status">Loading runs…</p>{:else if heads.isError}<p role="alert">Could not load runs. <Button.Root onclick={() => heads.refetch()}>Retry runs</Button.Root></p>
    {:else}<ul>{#each heads.data.items as head (head.id)}<li><Button.Root variant="outline" onclick={() => { selectedId = head.id; }}>{names[head.state]} · {head.completedCases}/{head.totalCases} cases · {new Date(head.createdAt).toLocaleString()}</Button.Root></li>{/each}</ul>
      {#if !heads.data.items.length}<p>No evaluation runs yet. Start with a saved workflow and test case.</p>{/if}
      {#if heads.data.pagination.hasMore}<Button.Root onclick={() => { cursor = heads.data?.pagination.nextCursor ?? undefined; }}>Next runs</Button.Root>{/if}
      {#if cursor}<Button.Root variant="ghost" onclick={() => { cursor = undefined; }}>First runs</Button.Root>{/if}
    {/if}
  {/if}
</section>{/if}
<style>.run-workspace, ul { display: grid; gap: 1rem; max-width: 56rem; } ul { list-style: none; padding: 0; } h2 { font-weight: 600; }</style>
