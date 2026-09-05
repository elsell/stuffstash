<script lang="ts">
  import { onDestroy, tick } from 'svelte';
  import { ConversationFailure } from '$lib/domain/conversation';
  import type { WorkflowDefinition, WorkflowStepKind } from '$lib/domain/conversationWorkflow';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Input from '$lib/components/ui/input/index.js';
  import * as Textarea from '$lib/components/ui/textarea/index.js';

  import type { ConversationModelChoice } from '$lib/domain/conversationProvider';
  import WorkflowSelect from './WorkflowSelect.svelte';

  let { initial, providers, onSave, onReload }: {
    initial: WorkflowDefinition;
    providers: ConversationModelChoice[];
    onSave: (definition: WorkflowDefinition) => Promise<void>;
    onReload?: () => void;
  } = $props();
  // The owner keys this editor by revision. Query refreshes never overwrite a draft.
  let draft = $state<WorkflowDefinition>(structuredClone($state.snapshot(initial)));
  let saving = $state(false);
  let message = $state('');
  let conflict = $state(false);
  let errorSummary: HTMLParagraphElement | undefined;
  let invalid = $state(false);
  let alive = true;
  onDestroy(() => { alive = false; });
  const labels: Record<WorkflowStepKind, string> = { interpret: 'Understand', assess: 'Look up and assess', respond: 'Respond' };
  const budgets = [
    { key: 'evidenceRounds', label: 'Search rounds' }, { key: 'modelCalls', label: 'Model calls' },
    { key: 'elapsedSeconds', label: 'Time limit (seconds)' }, { key: 'followUpTurns', label: 'Follow-up turns' }
  ] as const;
  async function save(event: SubmitEvent) {
    event.preventDefault();
    if (saving) return;
    saving = true; message = ''; conflict = false; invalid = false;
    try {
      await onSave($state.snapshot(draft));
      if (alive) message = 'Draft saved. Run test cases before activating it.';
    } catch (error) {
      if (!alive) return;
      invalid = error instanceof ConversationFailure && error.kind === 'invalid';
      conflict = error instanceof ConversationFailure && error.kind === 'conflict';
      message = conflict ? 'A newer revision exists. Your edits are still here. Load the latest revision to compare.'
        : error instanceof ConversationFailure && ['forbidden', 'unauthenticated'].includes(error.kind)
          ? 'You no longer have access to save this workflow.'
          : error instanceof ConversationFailure && error.kind === 'invalid'
            ? 'These settings could not be saved. Check the values against your server’s configured limits.'
            : 'Could not save the draft. Your edits are still here; try again.';
      if (invalid) { await tick(); if (alive) errorSummary?.focus(); }
    } finally { if (alive) saving = false; }
  }
</script>

<form class="conversation-editor" onsubmit={save}>
  <header><h2>Workflow settings</h2><p>Choose how your configured models understand requests and find answers. Saving creates a draft.</p></header>
  <label>Workflow name<Input.Root name="name" bind:value={draft.name} required disabled={saving} /></label>
  {#each draft.steps as step, index (step.kind)}
    <fieldset disabled={saving}>
      <legend>{index + 1}. {labels[step.kind]}</legend>
      <WorkflowSelect id={`provider-${step.kind}`} label="Model profile" value={step.providerProfileId ?? ''}
        disabled={saving || (step.kind === 'respond' && draft.response === 'grounded')}
        options={[{ value: '', label: 'Tenant default model' }, ...providers.map(provider => ({ value: provider.id, label: provider.name })),
          ...(step.providerProfileId && !providers.some(provider => provider.id === step.providerProfileId) ? [{ value: step.providerProfileId, label: 'Saved profile (currently unavailable)' }] : [])]}
        onChange={value => { step.providerProfileId = value || null; }} />
      {#if step.kind === 'respond' && draft.response === 'grounded'}<p class="help">Grounded answers use verified inventory facts without another model call. This step’s model and instructions are unused.</p>{/if}
      <label>Additional instructions<Textarea.Root name={`instructions-${step.kind}`} bind:value={step.instructions} disabled={step.kind === 'respond' && draft.response === 'grounded'} rows={3} /></label>
      <label>Maximum attempts<Input.Root name={`attempts-${step.kind}`} type="number" min={1} step={1} required bind:value={step.attempts} /></label>
    </fieldset>
  {/each}
  <details><summary>Advanced: retrieval, response and limits</summary>
    <WorkflowSelect id="retrieval" label="Search strategy" value={draft.retrieval} disabled={saving}
      options={[{ value: 'precise_first', label: 'Precise matches first' }, { value: 'expanded', label: 'Broader discovery' }]}
      onChange={value => { if (value === 'precise_first' || value === 'expanded') draft.retrieval = value; }} />
    <WorkflowSelect id="response" label="Answer style" value={draft.response} disabled={saving}
      options={[{ value: 'generated_with_grounded_fallback', label: 'Model answer with grounded recovery' }, { value: 'grounded', label: 'Grounded facts' }]}
      onChange={value => { if (value === 'generated_with_grounded_fallback' || value === 'grounded') draft.response = value; }} />
    <div class="budget-grid">{#each budgets as budget}<label>{budget.label}<Input.Root name={budget.key} disabled={saving} type="number" min={1} step={1} required bind:value={draft.budget[budget.key]} /></label>{/each}</div>
    <p class="help">Limits are shared across the conversation and must fit your server’s configured maximums.</p>
  </details>
  <div class="editor-actions"><Button.Root type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save draft'}</Button.Root>
    {#if conflict && onReload}<Button.Root type="button" variant="outline" onclick={onReload}>Load latest to compare</Button.Root>{/if}
  </div>
  <p bind:this={errorSummary} tabindex="-1" role={invalid ? "alert" : "status"} aria-live="polite">{message}</p>
</form>

<style>
  .conversation-editor { display: grid; gap: 1.25rem; max-width: 52rem; }
  header h2 { font-size: 1.25rem; font-weight: 650; }
  header p, .help { color: var(--muted-foreground); font-size: .9rem; }
  label { display: grid; gap: .4rem; font-size: .9rem; font-weight: 500; }
  fieldset { display: grid; gap: 1rem; border: 1px solid var(--border); border-radius: var(--radius); padding: 1rem; min-width: 0; }
  legend { font-weight: 650; padding-inline: .4rem; }
  summary:focus-visible { outline: 2px solid var(--ring); outline-offset: 2px; }
  details > label, .budget-grid { margin-top: 1rem; }
  summary { cursor: pointer; font-weight: 600; }
  .budget-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
  .editor-actions { display: flex; flex-wrap: wrap; gap: .75rem; }
  @media (max-width: 40rem) { .budget-grid { grid-template-columns: 1fr; } }
</style>
