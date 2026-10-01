<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { onDestroy } from 'svelte';
  import { createQuery } from '@tanstack/svelte-query';
  import { createConversationSession, type ConversationSession } from '$lib/adapters/query/conversationSession';
  import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
  import type { ConversationScope } from '$lib/domain/conversation';
  import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
  import * as Button from '$lib/components/ui/button/index.js';
  import RunSetup from './RunSetup.svelte';
  import RunDetails from './RunDetails.svelte';
  let { scope, session: sharedSession, repositories, visible, onAccessLost = () => {}, onNavigationBlockedChange = () => {} }: { scope: ConversationScope; session?: ConversationSession; repositories: ConversationWorkspaceRepositories; visible: boolean; onAccessLost?: () => void; onNavigationBlockedChange?: (blocked: boolean) => void } = $props();
  let denied = $state(false); let creating = $state(false); let busy = $state(false); let selectedId = $state(''); let cursor = $state<string | undefined>();
  // svelte-ignore state_referenced_locally -- parent keys the workspace by authenticated scope.
  const session = sharedSession ?? createConversationSession(scope, () => { denied = true; creating = false; selectedId = ''; onAccessLost(); });
  // svelte-ignore state_referenced_locally -- session ownership is fixed for this mount.
  const ownsSession = sharedSession === undefined;
  onDestroy(() => { if (ownsSession) void session.dispose(); });
  $effect(() => { onNavigationBlockedChange(creating); });
  const heads = createQuery(() => ({ queryKey: conversationKey(session.scope, 'runs', cursor ?? ''), enabled: !denied,
    queryFn: ({ signal }) => repositories.runs.list(session.scope.tenantId, { limit: 20, cursor }, signal) }), () => session.client);
  const names = { queued: 'Queued', running: 'Running', succeeded: 'Completed', failed: 'Failed', cancelled: 'Cancelled' };
</script>
{#if denied}<section role="alert"><h2>{t('web.RunWorkspace.runsUnavailable')}</h2><p>{t('web.RunWorkspace.youNoLongerHaveAccessToConfigureThisTenant')}</p></section>
{:else}<section class="run-workspace" aria-labelledby="runs-title"><h2 id="runs-title">{t('web.RunWorkspace.runs')}</h2>
  {#if creating}<Button.Root variant="outline" disabled={busy} onclick={() => { if (!busy) creating = false; }}>{t('web.RunWorkspace.discardRunSetup')}</Button.Root>
    <RunSetup {session} {repositories} onBusy={value => { busy = value; }} onQueued={run => { creating = false; busy = false; selectedId = run.id; }} />
  {:else if selectedId}<Button.Root variant="outline" onclick={() => { selectedId = ''; void heads.refetch(); }}>{t('web.RunWorkspace.backToRuns')}</Button.Root>
    {#key selectedId}<RunDetails {session} runs={repositories.runs} cases={repositories.cases} workflows={repositories.workflows} runId={selectedId} {visible} />{/key}
  {:else}<Button.Root onclick={() => { creating = true; }}>{t('web.RunWorkspace.newRun')}</Button.Root>
    {#if heads.isPending}<p role="status">{t('web.RunWorkspace.loadingRuns')}</p>{:else if heads.isError}<p role="alert">{t('web.RunWorkspace.couldNotLoadRuns')} <Button.Root onclick={() => heads.refetch()}>{t('web.RunWorkspace.retryRuns')}</Button.Root></p>
    {:else}<ul>{#each heads.data.items as head (head.id)}<li><Button.Root variant="outline" onclick={() => { selectedId = head.id; }}>{t('web.RunWorkspace.casesFull', { value: names[head.state], completedCases: head.completedCases, totalCases: head.totalCases, value4: new Date(head.createdAt).toLocaleString() })}</Button.Root></li>{/each}</ul>
      {#if !heads.data.items.length}<p>{t('web.RunWorkspace.noEvaluationRunsYetStartWithASavedWorkflow')}</p>{/if}
      {#if heads.data.pagination.hasMore}<Button.Root onclick={() => { cursor = heads.data?.pagination.nextCursor ?? undefined; }}>{t('web.RunWorkspace.nextRuns')}</Button.Root>{/if}
      {#if cursor}<Button.Root variant="ghost" onclick={() => { cursor = undefined; }}>{t('web.RunWorkspace.firstRuns')}</Button.Root>{/if}
    {/if}
  {/if}
</section>{/if}
<style>.run-workspace, ul { display: grid; gap: 1rem; max-width: 56rem; } ul { list-style: none; padding: 0; } h2 { font-weight: 600; }</style>
