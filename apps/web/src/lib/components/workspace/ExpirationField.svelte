<script lang="ts">
  import { untrack } from 'svelte';
  import { validExpirationInput } from '$lib/domain/expiration';
  import type { AssetExpiration } from '$lib/domain/inventory';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import SegmentedControl from './SegmentedControl.svelte';

  let { id, initialValue, onChange }: {
    id: string;
    initialValue?: AssetExpiration;
    onChange: (value: AssetExpiration | undefined, valid: boolean) => void;
  } = $props();
  const initial = untrack(() => initialValue);
  let precision = $state<AssetExpiration['precision']>(initial?.precision ?? 'day');
  let day = $state(initial?.precision === 'day' ? initial.date : '');
  let month = $state(initial?.precision === 'month' ? initial.date : '');
  let invalid = $state(false);
  let value = $derived(precision === 'day' ? day : month);

  function publish() {
    onChange(value ? { date: value, precision } : undefined, !invalid);
  }
  function select(next: string) {
    precision = next as AssetExpiration['precision'];
    invalid = !validExpirationInput(value, precision);
    publish();
  }
  function input(event: Event) {
    const target = event.currentTarget as HTMLInputElement;
    if (precision === 'day') day = target.value;
    else month = target.value;
    invalid = !target.validity.valid || !validExpirationInput(target.value, precision);
    publish();
  }
  function clear() {
    if (precision === 'day') day = '';
    else month = '';
    invalid = false;
    publish();
  }
</script>

<div class="expiration-field">
  <Label for={id}>Expiration (optional)</Label>
  <SegmentedControl label="Expiration precision" value={precision}
    options={[{ value: 'day', label: 'Exact date' }, { value: 'month', label: 'Month and year' }]}
    onSelect={select} />
  <Input {id} type={precision === 'day' ? 'date' : 'month'} {value}
    oninput={input} aria-invalid={invalid} aria-describedby={`${id}-help`} />
  <p id={`${id}-help`}>
    {#if invalid}Enter a complete, valid expiration date.
    {:else if precision === 'month'}Tracked through the end of this month.
    {:else}Tracked through the end of this day.{/if}
  </p>
  {#if value || invalid}<Button.Root type="button" variant="ghost" onclick={clear}>Clear expiration</Button.Root>{/if}
</div>

<style>
  .expiration-field { display: grid; gap: 0.5rem; }
  p { color: var(--muted-foreground); font-size: 0.875rem; }
</style>
