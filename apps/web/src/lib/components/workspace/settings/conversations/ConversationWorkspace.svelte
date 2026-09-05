<script lang="ts">
  import type { ConversationScope } from '$lib/domain/conversation';
  import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
  import * as Button from '$lib/components/ui/button/index.js';
  import WorkflowWorkspace from './WorkflowWorkspace.svelte';
  import CaseWorkspace from './CaseWorkspace.svelte';
  let { scope, repositories }: { scope: ConversationScope; repositories: ConversationWorkspaceRepositories } = $props();
  let section = $state<'workflows' | 'cases'>('workflows'); let visitedCases = $state(false);
  let editing = $state(false); let denied = $state(false);
</script>
{#if denied}<section role="alert"><h2>Conversation settings unavailable</h2><p>Your account no longer has access to configure this tenant.</p></section>
{:else}
  <nav aria-label="Conversation settings sections" class="conversation-sections">
    <Button.Root variant="outline" aria-pressed={section === 'workflows'} disabled={editing && section !== 'workflows'} onclick={() => { if (!editing) section = 'workflows'; }}>Workflows</Button.Root>
    <Button.Root variant="outline" aria-pressed={section === 'cases'} disabled={editing && section !== 'cases'} onclick={() => { if (!editing) { visitedCases = true; section = 'cases'; } }}>Test cases</Button.Root>
  </nav>
  {#if editing}<p>Close the editor before switching sections to keep your draft safe.</p>{/if}
  <div hidden={section !== 'workflows'}><WorkflowWorkspace {scope} workflows={repositories.workflows} providers={repositories.providers} onEditingChange={value => { editing = value; }} onAccessLost={() => { denied = true; }} /></div>
  {#if visitedCases}<div hidden={section !== 'cases'}><CaseWorkspace {scope} cases={repositories.cases} onEditingChange={value => { editing = value; }} onAccessLost={() => { denied = true; }} /></div>{/if}
{/if}
<style>.conversation-sections { display: flex; gap: .5rem; flex-wrap: wrap; margin-bottom: 1rem; }</style>
