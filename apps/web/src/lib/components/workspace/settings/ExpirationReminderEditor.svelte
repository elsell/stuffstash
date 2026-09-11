<script lang="ts">
  import SegmentedControl from '../SegmentedControl.svelte';
  import { untrack } from 'svelte';
  import type { ExpirationReminderPolicy } from '$lib/domain/notification';
  import { safeWorkspaceErrorMessage } from '$lib/application/workspaceSafeError';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Button from '$lib/components/ui/button/index.js';

  let { initialPolicy, inheritedPolicy, onSave }: {
    initialPolicy: ExpirationReminderPolicy | null;
    inheritedPolicy?: ExpirationReminderPolicy;
    onSave: (policy: ExpirationReminderPolicy | null) => Promise<void>;
  } = $props();
  const id = $props.id();
  const initial = untrack(() => initialPolicy ?? inheritedPolicy);
  if (!initial) throw new Error('A reminder policy is required.');
  let inherit = $state(untrack(() => initialPolicy === null));
  let draft = $state({ ...initial });
  let days = $state(String(initial.advanceDays));
  let saving = $state(false);
  let error = $state('');
  let saved = $state(false);
  let dirty = $state(false);
  $effect(() => {
    const policy = initialPolicy ?? inheritedPolicy;
    if (untrack(() => dirty || saving) || !policy) return;
    inherit = initialPolicy === null; draft = { ...policy }; days = String(policy.advanceDays);
  });
  let displayed = $derived(inherit && inheritedPolicy ? inheritedPolicy : draft);
  let validDays = $derived(/^\d+$/.test(days) && Number(days) <= 3650);
  let controlsDisabled = $derived(saving || inherit);
  let editingDays = $state(false);
  let mode = $derived(inherit ? 'defaults' : draft.enabled ? 'custom' : 'off');

  function changed() { dirty = true; saved = false; error = ''; }
  async function save(event?: SubmitEvent) {
    event?.preventDefault();
    if (saving || (!inherit && !validDays)) return;
    saving = true; error = ''; saved = false;
    try {
      await onSave(inherit ? null : { ...draft, advanceDays: Number(days) });
      draft = { ...draft, advanceDays: Number(days) }; dirty = false; saved = true;
    } catch (caught) {
      error = safeWorkspaceErrorMessage(caught, 'Reminders could not be saved. Your changes are still here. Try again.');
    } finally { saving = false; }
  }
</script>

<form onsubmit={save} aria-label="Expiration reminders">
  {#if inheritedPolicy}
    <SegmentedControl label="Reminder policy" value={mode} options={[{value:'defaults',label:'Use defaults',disabled:saving},{value:'custom',label:'Custom',disabled:saving},{value:'off',label:'Off',disabled:saving}]} onSelect={value => { inherit = value === 'defaults'; draft.enabled = value === 'custom'; changed(); void save(); }} />
  {:else}
    <Label class="setting"><Checkbox checked={draft.enabled} disabled={saving} onchange={event => { draft.enabled = event.currentTarget.checked; changed(); void save(); }} />Default reminders</Label>
    <p>Types with custom reminders can override these defaults.</p>
  {/if}
  {#if !inherit}
  {#if draft.enabled}
  <Button.Root variant="ghost" type="button" aria-expanded={editingDays} onclick={() => { editingDays = !editingDays; }}>Before expiration <span>{draft.upcoming ? `${draft.advanceDays} days` : 'Off'}</span></Button.Root>
  {#if editingDays}
  <Label class="setting"><Checkbox checked={displayed.upcoming} disabled={controlsDisabled} onchange={(event) => { draft.upcoming = event.currentTarget.checked; changed(); void save(); }} />Notify before expiration</Label>
  {#if draft.upcoming}<div class="days">
    <Label for={`${id}-days`}>Days before expiration</Label>
    <Input id={`${id}-days`} type="number" min={0} max={3650} step={1} value={inherit ? displayed.advanceDays : days} disabled={controlsDisabled}
      aria-invalid={!inherit && !validDays} aria-describedby={`${id}-help`}
      oninput={(event) => { days = event.currentTarget.value; changed(); }} />
    <p id={`${id}-help`}>{!inherit && !validDays ? 'Enter a whole number from 0 to 3650.' : 'Calendar days before the expiration date ends.'}</p>
  </div>{/if}{/if}
  <Label class="setting"><Checkbox checked={displayed.expired} disabled={controlsDisabled} onchange={(event) => { draft.expired = event.currentTarget.checked; changed(); void save(); }} />When expired</Label>{/if}
  {:else}<p>{displayed.enabled ? `${displayed.upcoming ? `${displayed.advanceDays} days before expiration` : ''}${displayed.expired ? ' and when expired' : ''}` : 'Default reminders are off'}</p>{/if}
  {#if error}<p role="alert">{error}</p>{/if}
  {#if saved}<p role="status">Reminders saved.</p>{/if}
  {#if dirty && !saving}<Button.Root type="button" variant="ghost" onclick={() => { const policy = initialPolicy ?? inheritedPolicy!; draft = {...policy}; days = String(policy.advanceDays); inherit = initialPolicy === null; dirty = false; error = ''; editingDays = false; }}>Discard changes</Button.Root>{/if}
  {#if dirty || saving}<Button.Root type="submit" disabled={saving || (!inherit && !validDays)}>{saving ? 'Saving…' : 'Save reminders'}</Button.Root>{/if}
</form>

<style>
  form { display: grid; gap: 1rem; }
  form :global(.setting) { display: flex; align-items: center; gap: 0.75rem; min-height: 2.75rem; cursor: pointer; }
  .days { display: grid; gap: 0.5rem; }
  p { color: var(--muted-foreground); font-size: 0.875rem; }
  [role='alert'] { color: var(--destructive); }
</style>
