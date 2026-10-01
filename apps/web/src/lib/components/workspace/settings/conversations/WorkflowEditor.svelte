<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import * as Label from '$lib/components/ui/label/index.js';
  import { onDestroy, tick } from 'svelte';
  import { ConversationFailure } from '$lib/domain/conversation';
  import type { WorkflowDefinition } from '$lib/domain/conversationWorkflow';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Input from '$lib/components/ui/input/index.js';
  import * as Textarea from '$lib/components/ui/textarea/index.js';

  import type { ConversationModelChoice } from '$lib/domain/conversationProvider';
  import WorkflowSelect from './WorkflowSelect.svelte';

  let { initial, providers, onSave, onReload, disabled = false }: {
    initial: WorkflowDefinition;
    disabled?: boolean;
    providers: ConversationModelChoice[];
    onSave: (definition: WorkflowDefinition) => Promise<void>;
    onReload?: () => void;
  } = $props();
  // The owner keys this editor by revision. Query refreshes never overwrite a draft.
  // svelte-ignore state_referenced_locally -- each keyed editor owns a draft seeded by its immutable revision.
  let draft = $state<WorkflowDefinition>(structuredClone($state.snapshot(initial)));
  let saving = $state(false);
  let message = $state('');
  let conflict = $state(false);
  let errorSummary: HTMLParagraphElement | undefined;
  let invalid = $state(false);
  let alive = true;
  onDestroy(() => { alive = false; });
  const budgets = [
    { key: 'toolCalls', label: t('web.WorkflowEditor.toolCallsPerTurn') }, { key: 'modelCalls', label: t('web.WorkflowEditor.modelCallsPerTurn') },
    { key: 'elapsedSeconds', label: t('web.WorkflowEditor.processingTimePerTurnSeconds') }, { key: 'followUpTurns', label: t('web.WorkflowEditor.followUpTurns') }
  ] as const;
  async function save(event: SubmitEvent) {
    event.preventDefault();
    if (saving || disabled) return;
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
  <header><h2>{t('web.WorkflowEditor.workflowSettings')}</h2><p>{t('web.WorkflowEditor.chooseAModelAndGiveItGuidanceForYour')}</p></header>
  <Label.Root class="grid gap-2 text-sm">{t('web.WorkflowEditor.workflowName')}<Input.Root name="name" bind:value={draft.name} required disabled={saving || disabled} /></Label.Root>
  <WorkflowSelect id="provider-model" label={t('web.WorkflowEditor.modelProfile')} value={draft.providerProfileId ?? ''}
    disabled={saving || disabled}
    options={[{ value: '', label: 'Tenant default model' }, ...providers.map(provider => ({ value: provider.id, label: provider.name })),
      ...(draft.providerProfileId && !providers.some(provider => provider.id === draft.providerProfileId) ? [{ value: draft.providerProfileId, label: 'Saved profile (currently unavailable)' }] : [])]}
    onChange={value => { draft.providerProfileId = value || null; }} />
  <Label.Root class="grid gap-2 text-sm">{t('web.WorkflowEditor.additionalInstructions')}<Textarea.Root name="instructions" bind:value={draft.instructions} disabled={saving || disabled} rows={3} /></Label.Root>
  <details><summary>{t('web.WorkflowEditor.conversationLimits')}</summary>
    <div class="budget-grid">{#each budgets as budget}<Label.Root class="grid gap-2 text-sm">{budget.label}<Input.Root name={budget.key} disabled={saving || disabled} type="number" min={1} step={1} required bind:value={draft.budget[budget.key]} /></Label.Root>{/each}</div>
    <p class="help">{t('web.WorkflowEditor.callAndProcessingLimitsApplyToEachTurnTime')}</p>
  </details>
  <div class="editor-actions"><Button.Root type="submit" disabled={saving || disabled}>{saving ? t('web.WorkflowEditor.saving') : t('web.WorkflowEditor.saveDraft')}</Button.Root>
    {#if conflict && onReload}<Button.Root type="button" variant="outline" disabled={saving || disabled} onclick={onReload}>{t('web.WorkflowEditor.loadLatestToCompare')}</Button.Root>{/if}
  </div>
  <p bind:this={errorSummary} tabindex="-1" role={invalid ? "alert" : "status"} aria-live="polite">{message}</p>
</form>

<style>
  .conversation-editor { display: grid; gap: 1.25rem; max-width: 52rem; }
  header h2 { font-size: 1.25rem; font-weight: 600; }
  header p, .help { color: var(--muted-foreground); font-size: .9rem; }
  summary:focus-visible { outline: 2px solid var(--ring); outline-offset: 2px; }
  .budget-grid { margin-top: 1rem; }
  summary { cursor: pointer; font-weight: 600; }
  .budget-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
  .editor-actions { display: flex; flex-wrap: wrap; gap: .75rem; }
  @media (max-width: 40rem) { .budget-grid { grid-template-columns: 1fr; } }
</style>
