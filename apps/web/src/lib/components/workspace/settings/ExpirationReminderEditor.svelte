<script lang="ts">
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
  let displayed = $derived(inherit && inheritedPolicy ? inheritedPolicy : draft);
  let validDays = $derived(/^\d+$/.test(days) && Number(days) <= 3650);
  let controlsDisabled = $derived(saving || inherit);

  function changed() { saved = false; error = ''; }
  async function save(event: SubmitEvent) {
    event.preventDefault();
    if (saving || (!inherit && !validDays)) return;
    saving = true; error = ''; saved = false;
    try {
      await onSave(inherit ? null : { ...draft, advanceDays: Number(days) });
      saved = true;
    } catch (caught) {
      error = safeWorkspaceErrorMessage(caught, 'Reminders could not be saved. Your changes are still here. Try again.');
    } finally { saving = false; }
  }
</script>

<form onsubmit={save} aria-label="Expiration reminders">
  {#if inheritedPolicy}
    <Label class="setting"><Checkbox checked={inherit} disabled={saving} onchange={(event) => { inherit = event.currentTarget.checked; changed(); }} />Use inventory defaults</Label>
    <p>{inherit ? 'Inherited from your inventory settings.' : 'Custom settings for this asset type.'}</p>
  {:else}
    <p>Your defaults for this inventory. Asset types with custom settings can still send reminders when these defaults are off.</p>
  {/if}
  <Label class="setting"><Checkbox checked={displayed.enabled} disabled={controlsDisabled} onchange={(event) => { draft.enabled = event.currentTarget.checked; changed(); }} />Enable expiration reminders</Label>
  <Label class="setting"><Checkbox checked={displayed.upcoming} disabled={controlsDisabled} onchange={(event) => { draft.upcoming = event.currentTarget.checked; changed(); }} />Notify before expiration</Label>
  <div class="days">
    <Label for={`${id}-days`}>Days before expiration</Label>
    <Input id={`${id}-days`} type="number" min={0} max={3650} step={1} value={inherit ? displayed.advanceDays : days} disabled={controlsDisabled}
      aria-invalid={!inherit && !validDays} aria-describedby={`${id}-help`}
      oninput={(event) => { days = event.currentTarget.value; changed(); }} />
    <p id={`${id}-help`}>{!inherit && !validDays ? 'Enter a whole number from 0 to 3650.' : 'Calendar days before the expiration date ends.'}</p>
  </div>
  <Label class="setting"><Checkbox checked={displayed.expired} disabled={controlsDisabled} onchange={(event) => { draft.expired = event.currentTarget.checked; changed(); }} />Notify when expired</Label>
  {#if error}<p role="alert">{error}</p>{/if}
  {#if saved}<p role="status">Reminders saved.</p>{/if}
  <Button.Root type="submit" disabled={saving || (!inherit && !validDays)}>{saving ? 'Saving…' : 'Save reminders'}</Button.Root>
</form>

<style>
  form { display: grid; gap: 1rem; }
  form :global(.setting) { display: flex; align-items: center; gap: 0.75rem; min-height: 2.75rem; cursor: pointer; }
  .days { display: grid; gap: 0.5rem; }
  p { color: var(--muted-foreground); font-size: 0.875rem; }
  [role='alert'] { color: var(--destructive); }
</style>
