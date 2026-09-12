<script lang="ts">
 import * as Dialog from '$lib/components/ui/dialog/index.js';
 import * as Button from '$lib/components/ui/button/index.js';
 import { Input } from '$lib/components/ui/input/index.js';
 import type { ExpirationFilter, ExpirationChoices } from '$lib/ports/expirationRepository';
 let { open = $bindable(false), filter, choices, loading, error, onApply, onRetry }: { open?: boolean; filter: ExpirationFilter; choices?: ExpirationChoices; loading: boolean; error?: string; onApply: (filter: ExpirationFilter) => void; onRetry: () => void } = $props();
 let draft = $state<ExpirationFilter>({mode:'all'});
 $effect(() => { if (open) draft = { ...filter, tagIds: [...(filter.tagIds ?? [])] }; });
 let reversed = $derived(!!draft.fromDate && !!draft.throughDate && draft.fromDate > draft.throughDate);
</script>
<Dialog.Root bind:open><Dialog.Content class="max-h-[85dvh] overflow-y-auto sm:max-w-lg">
 <Dialog.Header><Dialog.Title>Expiration filters</Dialog.Title><Dialog.Description>Filter recorded dates in this inventory. Month-only dates are compared at the end of the month.</Dialog.Description></Dialog.Header>
 {#if loading}<p role="status">Loading filter choices…</p>{/if}
 {#if error}<p role="alert">{error}</p><Button.Root variant="outline" onclick={onRetry}>Retry choices</Button.Root>{/if}
 <div class="fields">
  <label>Kind<select bind:value={draft.kind}><option value="">Any kind</option><option value="item">Items</option><option value="container">Containers</option><option value="location">Places</option></select></label>
  <label>Availability<select bind:value={draft.checkoutState}><option value="">Any availability</option><option value="available">Available</option><option value="checked_out">Checked out</option></select></label>
  <label>Type<select bind:value={draft.typeId}><option value="">Any type</option>{#each choices?.types ?? [] as choice}<option value={choice.id}>{choice.title}</option>{/each}</select></label>
  <label>Location<select bind:value={draft.locationId}><option value="">Anywhere</option>{#each choices?.locations ?? [] as choice}<option value={choice.id}>{choice.title}</option>{/each}</select></label>
  <fieldset><legend>Tags · match all selected</legend><div class="tags">{#each choices?.tags ?? [] as choice}<label><input type="checkbox" value={choice.id} bind:group={draft.tagIds} />{choice.title}</label>{/each}{#if choices && !choices.tags.length}<p>No tags in this inventory.</p>{/if}</div></fieldset>
  <div class="dates"><label>From<Input type="date" bind:value={draft.fromDate} /></label><label>Through<Input type="date" bind:value={draft.throughDate} /></label></div>
  {#if reversed}<p role="alert">Through must be on or after From.</p>{/if}
 </div>
 <Dialog.Footer><Button.Root variant="ghost" onclick={() => { draft = {mode:filter.mode}; }}>Clear filters</Button.Root><Button.Root variant="outline" onclick={() => { open = false; }}>Cancel</Button.Root><Button.Root disabled={reversed || loading || !!error} onclick={() => { onApply(draft); open = false; }}>Apply filters</Button.Root></Dialog.Footer>
</Dialog.Content></Dialog.Root>
<style>
 .fields {display:grid;gap:1.25rem;padding:.5rem 0;} label {display:grid;gap:.4rem;font-size:var(--text-body-size);} select {min-height:2.75rem;padding:.5rem;border:1px solid var(--border);border-radius:.5rem;background:var(--background);color:var(--foreground);width:100%;} .dates {display:grid;grid-template-columns:1fr 1fr;gap:1rem;} .dates label {min-width:0;} .tags {display:grid;max-height:12rem;overflow:auto;} .tags label {display:flex;align-items:center;gap:.65rem;min-height:2.75rem;} input[type=checkbox] {width:1.15rem;height:1.15rem;} legend {margin-bottom:.5rem;} @media(max-width:380px){.dates{grid-template-columns:1fr;}}
</style>
