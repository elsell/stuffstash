<script lang="ts">
  import { t } from '$lib/presentation/localization';
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
  import ValidationMessage from './ValidationMessage.svelte';
  import { validationMessages, validationAttributes } from './validationPresentation';
  let { initial, onSave, onReload, disabled = false }: { initial: CaseDefinition; disabled?: boolean; onSave: (value: CaseDefinition) => Promise<void>; onReload?: () => void } = $props();
  // svelte-ignore state_referenced_locally -- the parent keys this editor by immutable case revision.
  let draft = $state<CaseDefinition>(structuredClone($state.snapshot(initial)));
  let issues = $state<CaseDraftIssue[]>([]);
  const errors = $derived(validationMessages(issues)); let message = $state(''); let saving = $state(false); let conflict = $state(false);
  let form: HTMLFormElement; let summary = $state<HTMLDivElement>(); let alive = true;
  onDestroy(() => { alive = false; });
  const labels: Record<string, string> = { 'case-title': t('web.CaseEditor.caseTitle'), 'case-utterance': t('web.CaseEditor.request'), 'case-fixtures-title': t('web.CaseEditor.testInventory'), 'case-expectations-title': t('web.CaseEditor.expectedResults'), 'expected-outcome': t('web.CaseEditor.expectedOutcome') };
  function fieldLabel(field: string) {
    if (labels[field]) return labels[field];
    if (field.startsWith('fixture-')) {
      const asset = draft.assets.find(asset => field.endsWith(`-${asset.id}`));
      return asset?.title ? t('conversation.fixtureNamed', { title: asset.title }) : t('conversation.fixtureSettings');
    }
    return field.startsWith('location-') ? t('web.CaseEditor.expectedLocation') : t('web.CaseEditor.proposedChange');
  }
  function focusField(event: MouseEvent, field: string) {
    event.preventDefault();
    const control = form.elements.namedItem(field) ?? document.getElementById(field);
    if (control instanceof HTMLElement) { if (!control.matches('input,textarea,button,select,a')) control.tabIndex = -1; control.focus(); }
  }
  async function showIssues() { await tick(); if (alive) summary?.focus(); }
  async function save(event: SubmitEvent) {
    event.preventDefault(); if (saving || disabled) return;
    const prepared = prepareCaseDraft($state.snapshot(draft));
    issues = prepared.issues; message = ''; conflict = false;
    if (issues.length) { await showIssues(); return; }
    saving = true;
    try { await onSave(prepared.definition); if (alive) message = t('web.CaseEditor.testCaseRevisionSaved'); }
    catch (error) {
      if (!alive) return;
      conflict = error instanceof ConversationFailure && error.kind === 'conflict';
      message = conflict ? t('web.CaseEditor.aNewerRevisionExistsYourEditsAreStillHere')
        : error instanceof ConversationFailure && ['forbidden', 'unauthenticated'].includes(error.kind)
          ? t('web.CaseEditor.youNoLongerHaveAccessToSaveThisCase')
          : t('web.CaseEditor.couldNotSaveTheCaseYourEditsAreStill');
      if (error instanceof ConversationFailure && error.kind === 'invalid') {
        issues = [{ field: 'case-title', message: t('web.CaseEditor.theServerRejectedThisCaseCheckItsFixtureAnd') }];
        await showIssues();
      }
    } finally { if (alive) saving = false; }
  }
</script>
<form bind:this={form} class="case-editor" onsubmit={save} novalidate>
  <header><h2>{t('web.CaseEditor.testCase')}</h2><p>{t('web.CaseEditor.describeARealisticRequestATestInventoryAndWhat')}</p></header>
  {#if issues.length}<div bind:this={summary} role="alert" tabindex="-1" class="validation-summary"><h3>{t('web.CaseEditor.checkTheseDetails')}</h3><ul>
    {#each issues as issue}<li><a href={`#${encodeURIComponent(issue.field)}`} onclick={event => focusField(event, issue.field)}>{fieldLabel(issue.field)}: {issue.message}</a></li>{/each}
  </ul></div>{/if}
  <Label.Root class="grid gap-2">{t('web.CaseEditor.caseTitle')}<Input.Root id="case-title" name="case-title" bind:value={draft.title} disabled={saving || disabled} required {...validationAttributes(errors['case-title'], 'case-title')} /></Label.Root>
  <ValidationMessage field="case-title" message={errors['case-title']} />
  <Label.Root class="grid gap-2">{t('web.CaseEditor.whatTheUserSays')}<Textarea.Root id="case-utterance" name="case-utterance" bind:value={draft.utterance} disabled={saving || disabled} required rows={3} {...validationAttributes(errors['case-utterance'], 'case-utterance')} /></Label.Root>
  <ValidationMessage field="case-utterance" message={errors['case-utterance']} />
  <CaseFixtures value={draft} {errors} disabled={saving || disabled} onChange={value => { draft = value; }} />
  <CaseExpectations value={draft} {errors} disabled={saving || disabled} onChange={value => { draft = value; }} />
  <div class="actions"><Button.Root type="submit" disabled={saving || disabled}>{saving ? t('web.CaseEditor.saving') : t('web.CaseEditor.saveTestCase')}</Button.Root>
    {#if conflict && onReload}<Button.Root type="button" variant="outline" disabled={saving || disabled} onclick={onReload}>{t('web.CaseEditor.loadLatestToCompare')}</Button.Root>{/if}
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
