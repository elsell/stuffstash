<script lang="ts">
  import { onDestroy, tick } from 'svelte';
  import { ConversationFailure } from '$lib/domain/conversation';
  import type { CaseDefinition } from '$lib/domain/conversationCase';
  import { prepareCaseDraft, type CaseDraftIssue } from '$lib/application/caseDraftValidation';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Input from '$lib/components/ui/input/index.js';
  import * as Textarea from '$lib/components/ui/textarea/index.js';
  import * as Label from '$lib/components/ui/label/index.js';
  import CaseFixtures from './CaseFixtures.svelte';
  import CaseExpectations from './CaseExpectations.svelte';
  let { initial, onSave, onReload }: { initial: CaseDefinition; onSave: (value: CaseDefinition) => Promise<void>; onReload?: () => void } = $props();
  // svelte-ignore state_referenced_locally -- the parent keys this editor by immutable case revision.
  let draft = $state<CaseDefinition>(structuredClone($state.snapshot(initial)));
  let issues = $state<CaseDraftIssue[]>([]); let message = $state(''); let saving = $state(false); let conflict = $state(false);
  let form: HTMLFormElement; let summary: HTMLDivElement | undefined; let alive = true;
  onDestroy(() => { alive = false; });
  const labels: Record<string, string> = { 'case-title': 'Case title', 'case-utterance': 'Request', 'case-fixtures-title': 'Test inventory', 'case-expectations-title': 'Expected results', 'expected-outcome': 'Expected outcome' };
  function fieldLabel(field: string) {
    if (labels[field]) return labels[field];
    if (field.startsWith('fixture-')) {
      const asset = draft.assets.find(asset => field.endsWith(`-${asset.id}`));
      return `Fixture ${asset?.title || 'settings'}`;
    }
    return field.startsWith('location-') ? 'Expected location' : 'Proposed change';
  }
  function focusField(event: MouseEvent, field: string) {
    event.preventDefault();
    const control = form.elements.namedItem(field) ?? document.getElementById(field);
    if (control instanceof HTMLElement) { if (!control.matches('input,textarea,button,select,a')) control.tabIndex = -1; control.focus(); }
  }
  async function showIssues() { await tick(); if (alive) summary?.focus(); }
  async function save(event: SubmitEvent) {
    event.preventDefault(); if (saving) return;
    const prepared = prepareCaseDraft($state.snapshot(draft));
    issues = prepared.issues; message = ''; conflict = false;
    if (issues.length) { await showIssues(); return; }
    saving = true;
    try { await onSave(prepared.definition); if (alive) message = 'Test case revision saved.'; }
    catch (error) {
      if (!alive) return;
      conflict = error instanceof ConversationFailure && error.kind === 'conflict';
      message = conflict ? 'A newer revision exists. Your edits are still here; load the latest revision to compare.'
        : error instanceof ConversationFailure && ['forbidden', 'unauthenticated'].includes(error.kind)
          ? 'You no longer have access to save this case.'
          : 'Could not save the case. Your edits are still here.';
      if (error instanceof ConversationFailure && error.kind === 'invalid') {
        issues = [{ field: 'case-title', message: 'The server rejected this case. Check its fixture and expectation limits.' }];
        await showIssues();
      }
    } finally { if (alive) saving = false; }
  }
</script>
<form bind:this={form} class="case-editor" onsubmit={save} novalidate>
  <header><h2>Test case</h2><p>Describe a realistic request, a test inventory and what a good response should do. This case evaluates text interactions.</p></header>
  {#if issues.length}<div bind:this={summary} role="alert" tabindex="-1" class="validation-summary"><h3>Check these details</h3><ul>
    {#each issues as issue}<li><a href={`#${encodeURIComponent(issue.field)}`} onclick={event => focusField(event, issue.field)}>{fieldLabel(issue.field)}: {issue.message}</a></li>{/each}
  </ul></div>{/if}
  <Label.Root class="grid gap-2">Case title<Input.Root id="case-title" name="case-title" bind:value={draft.title} disabled={saving} required aria-invalid={issues.some(issue => issue.field === 'case-title')} /></Label.Root>
  <Label.Root class="grid gap-2">What the user says<Textarea.Root id="case-utterance" name="case-utterance" bind:value={draft.utterance} disabled={saving} required rows={3} aria-invalid={issues.some(issue => issue.field === 'case-utterance')} /></Label.Root>
  <CaseFixtures value={draft} disabled={saving} onChange={value => { draft = value; }} />
  <CaseExpectations value={draft} disabled={saving} onChange={value => { draft = value; }} />
  <div class="actions"><Button.Root type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save test case'}</Button.Root>
    {#if conflict && onReload}<Button.Root type="button" variant="outline" onclick={onReload}>Load latest to compare</Button.Root>{/if}
  </div>
  <p role="status" aria-live="polite">{message}</p>
</form>
<style>
  .case-editor { display: grid; gap: 1.25rem; max-width: 56rem; }
  h2, h3 { font-weight: 600; } header p { color: var(--muted-foreground); }
  .actions { display: flex; flex-wrap: wrap; gap: .75rem; }
  .validation-summary { border: 1px solid var(--destructive); border-radius: var(--radius); padding: 1rem; }
  .validation-summary a { text-decoration: underline; }
</style>
