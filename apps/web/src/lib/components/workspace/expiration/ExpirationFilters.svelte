<script lang="ts">
 import * as Dialog from '$lib/components/ui/dialog/index.js';
 import * as Button from '$lib/components/ui/button/index.js';
 import { Label } from '$lib/components/ui/label/index.js';
 import { Checkbox } from '$lib/components/ui/checkbox/index.js';
 import ExpirationChoice from './ExpirationChoice.svelte';
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
  <ExpirationChoice label="Kind" emptyLabel="Any kind" value={draft.kind} choices={[{id:'item',title:'Items'},{id:'container',title:'Containers'},{id:'location',title:'Places'}]} onChange={value=>{draft.kind=value as ExpirationFilter['kind'];}} />
  <ExpirationChoice label="Availability" emptyLabel="Any availability" value={draft.checkoutState} choices={[{id:'available',title:'Available'},{id:'checked_out',title:'Checked out'}]} onChange={value=>{draft.checkoutState=value as ExpirationFilter['checkoutState'];}} />
  <ExpirationChoice label="Type" emptyLabel="Any type" value={draft.typeId} choices={choices?.types??[]} onChange={value=>{draft.typeId=value;}} />
  <ExpirationChoice label="Location" emptyLabel="Anywhere" value={draft.locationId} choices={choices?.locations??[]} onChange={value=>{draft.locationId=value;}} />
  <fieldset><legend>Tags · match all selected</legend><div class="tags">{#each choices?.tags ?? [] as choice}<Label><Checkbox checked={draft.tagIds?.includes(choice.id)??false} onchange={event=>{draft.tagIds=event.currentTarget.checked?[...(draft.tagIds??[]),choice.id]:draft.tagIds?.filter(id=>id!==choice.id);}} />{choice.title}</Label>{/each}{#if choices && !choices.tags.length}<p>No tags in this inventory.</p>{/if}</div></fieldset>
  <div class="dates"><Label>From<Input type="date" bind:value={draft.fromDate} /></Label><Label>Through<Input type="date" bind:value={draft.throughDate} /></Label></div>
  {#if reversed}<p role="alert">Through must be on or after From.</p>{/if}
 </div>
 <Dialog.Footer><Button.Root variant="ghost" onclick={() => { draft = {mode:filter.mode}; }}>Clear filters</Button.Root><Button.Root variant="outline" onclick={() => { open = false; }}>Cancel</Button.Root><Button.Root disabled={reversed || loading || !!error} onclick={() => { onApply(draft); open = false; }}>Apply filters</Button.Root></Dialog.Footer>
</Dialog.Content></Dialog.Root>
<style>
 .fields {display:grid;gap:1.25rem;padding:.5rem 0;} .fields :global(label) {display:grid;gap:.4rem;font-size:var(--text-body-size);} .dates {display:grid;grid-template-columns:1fr 1fr;gap:1rem;} .dates :global(label) {min-width:0;} .tags {display:grid;max-height:12rem;overflow:auto;} .tags :global(label) {display:flex;align-items:center;gap:.65rem;min-height:2.75rem;} .tags :global(input[type=checkbox]) {width:1.15rem;height:1.15rem;} legend {margin-bottom:.5rem;} @media(max-width:380px){.dates{grid-template-columns:1fr;}}
</style>
