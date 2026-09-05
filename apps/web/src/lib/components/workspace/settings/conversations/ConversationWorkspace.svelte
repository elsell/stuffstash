<script lang="ts">
  import type { ConversationScope } from '$lib/domain/conversation';
  import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
  import * as Button from '$lib/components/ui/button/index.js';
  import WorkflowWorkspace from './WorkflowWorkspace.svelte';
  import CaseWorkspace from './CaseWorkspace.svelte';
  import RunWorkspace from './RunWorkspace.svelte';
  let { scope, repositories }: { scope: ConversationScope; repositories: ConversationWorkspaceRepositories } = $props();
  let section = $state<'workflows' | 'cases' | 'runs'>('workflows'); let visitedCases = $state(false); let visitedRuns = $state(false);
  let blocked = $state({ workflows: false, cases: false, runs: false });
  const editing = $derived(blocked.workflows || blocked.cases || blocked.runs); let denied = $state(false);
</script>
{#if denied}<section role="alert"><h2>Conversation settings unavailable</h2><p>Your account no longer has access to configure this tenant.</p></section>
{:else}
  <nav aria-label="Conversation settings sections" class="conversation-sections">
    <Button.Root variant="outline" aria-pressed={section === 'workflows'} disabled={editing && section !== 'workflows'} onclick={() => { if (!editing) section = 'workflows'; }}>Workflows</Button.Root>
    <Button.Root variant="outline" aria-pressed={section === 'cases'} disabled={editing && section !== 'cases'} onclick={() => { if (!editing) { visitedCases = true; section = 'cases'; } }}>Test cases</Button.Root>
    <Button.Root variant="outline" aria-pressed={section === 'runs'} disabled={editing && section !== 'runs'} onclick={() => { if (!editing) { visitedRuns = true; section = 'runs'; } }}>Runs</Button.Root>
  </nav>
  {#if editing}<p>Finish loading or close the editor before switching sections.</p>{/if}
  <div hidden={section !== 'workflows'}><WorkflowWorkspace {scope} workflows={repositories.workflows} providers={repositories.providers} onNavigationBlockedChange={value => { blocked.workflows = value; }} onAccessLost={() => { denied = true; }} /></div>
  {#if visitedCases}<div hidden={section !== 'cases'}><CaseWorkspace {scope} cases={repositories.cases} onNavigationBlockedChange={value => { blocked.cases = value; }} onAccessLost={() => { denied = true; }} /></div>{/if}
  {#if visitedRuns}<div hidden={section !== 'runs'}><RunWorkspace {scope} {repositories} visible={section === 'runs'} onNavigationBlockedChange={value => { blocked.runs = value; }} onAccessLost={() => { denied = true; }} /></div>{/if}
{/if}
<style>.conversation-sections { display: flex; gap: .5rem; flex-wrap: wrap; margin-bottom: 1rem; }</style>
