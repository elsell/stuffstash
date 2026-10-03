<script lang="ts">
import { onMount } from 'svelte';
import { t } from '$lib/presentation/localization';
import { timestampLabel } from '$lib/presentation/timestamp';
import { readinessLabel, connectorStatusLabel, printingFailureMessage } from '$lib/presentation/printing';
import { PrintingFailure, type ReportedPrintOutcome, type PrintScope, type PrintDefaults, type RegisteredPrinter, type PrintConnector, type LabelTemplate, type PrintJob } from '$lib/domain/printing';
import type { PrintingRepository } from '$lib/ports/printingRepository';
import * as Button from '$lib/components/ui/button/index.js';
import { Checkbox } from '$lib/components/ui/checkbox/index.js';
import { Label } from '$lib/components/ui/label/index.js';
import PairingChoice from './PairingChoice.svelte';
import PrintJobList from './PrintJobList.svelte';
let { scope, repository, canConfigure, canPrint }: {
    scope: PrintScope;
    repository: PrintingRepository;
    canConfigure: boolean;
    canPrint: boolean;
} = $props();
let printers = $state<RegisteredPrinter[]>([]), connectors = $state<PrintConnector[]>([]), templates = $state<LabelTemplate[]>([]), jobs = $state<PrintJob[]>([]);
let draft = $state<PrintDefaults | null>(null), busy = $state(true), error = $state(''), saved = $state(false), conflict = $state(false), nextCursor = $state<string | undefined>();
const activePrinters = $derived(printers.filter(p => !p.retired));
const selectedTemplate = $derived(draft ? `${draft.templateId}:${draft.templateVersion}` : '');
const missingDefault = $derived(Boolean(draft?.defaultPrinterId && !activePrinters.some(p => p.id === draft?.defaultPrinterId)));
onMount(() => { void load(); });
async function load() { busy = true; error = ''; saved = false; try {
    const values = await Promise.all([repository.printers(scope), repository.connectors(scope), repository.templates(scope), repository.settings(scope), repository.jobs(scope)]);
    [printers, connectors, templates, draft] = values;
    jobs = values[4].items;
    nextCursor = values[4].nextCursor;
    conflict = false;
}
catch (caught) {
    error = printingFailureMessage(caught);
}
finally {
    busy = false;
} }
async function refresh() { busy = true; error = ''; try {
    [printers, connectors] = await Promise.all([repository.printers(scope), repository.connectors(scope)]);
    const page = await repository.jobs(scope);
    jobs = page.items;
    nextCursor = page.nextCursor;
}
catch (caught) {
    error = printingFailureMessage(caught);
}
finally {
    busy = false;
} }
async function save(event: SubmitEvent) { event.preventDefault(); if (!draft || !canConfigure || busy || conflict)
    return; busy = true; error = ''; saved = false; try {
    draft = await repository.saveSettings(scope, { ...draft });
    saved = true;
}
catch (caught) {
    error = printingFailureMessage(caught);
    conflict = caught instanceof PrintingFailure && caught.kind === 'conflict';
}
finally {
    busy = false;
} }
async function moreJobs() { if (!nextCursor || busy)
    return; busy = true; try {
    const page = await repository.jobs(scope, nextCursor);
    jobs = [...jobs, ...page.items];
    nextCursor = page.nextCursor;
}
catch (caught) {
    error = printingFailureMessage(caught);
}
finally {
    busy = false;
} }
async function cancel(job: PrintJob) { if (!canPrint || busy)
    return; busy = true; error = ''; try {
    const updated = await repository.cancel(scope, job);
    jobs = jobs.map(j => j.id === updated.id ? updated : j);
}
catch (caught) {
    error = printingFailureMessage(caught);
}
finally {
    busy = false;
} }
async function resolve(job:PrintJob,outcome:ReportedPrintOutcome){if(!canPrint||busy)return;busy=true;error='';try{const updated=await repository.resolve(scope,job,outcome);jobs=jobs.map(current=>current.id===updated.id?updated:current);}catch(caught){error=printingFailureMessage(caught);}finally{busy=false;}}
</script>
<section class="printing-settings" aria-labelledby="printing-settings-title" aria-busy={busy}>
 <header><h1 id="printing-settings-title">{t('web.Printing.title')}</h1><p>{t('web.Printing.description')}</p></header>
 {#if error}<p role="alert">{error}</p>{/if}
 {#if !draft}{#if busy}<p role="status">{t('web.Printing.loading')}</p>{:else}<Button.Root onclick={()=>void load()}>{t('web.Printing.reload')}</Button.Root>{/if}
 {:else}
  <section><h2>{t('web.Printing.defaults')}</h2>
   {#if !canConfigure}<p>{t('web.Printing.readOnly')}</p>{/if}
   {#if missingDefault}<p role="status">{t('web.Printing.missingDefault')}</p>{/if}
   <form onsubmit={save} class="defaults-form">
    <PairingChoice id="default-print-destination" label={t('web.Printing.defaultPrinter')} value={draft.defaultPrinterId??''} options={[{value:'',label:t('web.Printing.noDefault')},...activePrinters.map(p=>({value:p.id,label:`${p.name} — ${p.media.name}`}))]} disabled={!canConfigure||busy||conflict} onChange={value=>{if(draft){draft.defaultPrinterId=value||null;if(!value)draft.printOnCreateDefault=false;saved=false;}}}/>
    <PairingChoice id="default-print-template" label={t('web.Printing.template')} value={selectedTemplate} options={templates.map(template=>({value:`${template.id}:${template.version}`,label:template.name}))} disabled={!canConfigure||busy||conflict} onChange={value=>{const template=templates.find(t=>`${t.id}:${t.version}`===value);if(draft&&template){draft.templateId=template.id;draft.templateVersion=template.version;saved=false;}}}/>
    <div class="check-row"><Checkbox id="default-print-reference" bind:checked={draft.showReference} disabled={!canConfigure||busy||conflict}/><Label for="default-print-reference">{t('web.Printing.showReference')}</Label></div>
    <div class="check-row"><Checkbox id="default-auto-print" bind:checked={draft.printOnCreateDefault} disabled={!canConfigure||busy||conflict||!draft.defaultPrinterId||missingDefault}/><Label for="default-auto-print">{t('web.Printing.autoPrint')}</Label></div>
    {#if canConfigure}<Button.Root type="submit" disabled={busy||conflict||missingDefault}>{t('web.Printing.save')}</Button.Root>{/if}
    {#if conflict}<Button.Root variant="outline" disabled={busy} onclick={()=>void load()}>{t('web.Printing.reload')}</Button.Root>{/if}
    {#if saved}<p role="status">{t('web.Printing.saved')}</p>{/if}
   </form>
  </section>
  <section><div class="section-heading"><h2>{t('web.Printing.registered')}</h2><Button.Root variant="outline" disabled={busy} onclick={()=>void refresh()}>{t('web.Printing.refresh')}</Button.Root></div>
   {#if printers.length===0}<p>{t('web.Printing.emptyPrinters')}</p>{/if}
   <ul>{#each printers as printer(printer.id)}<li><strong>{printer.name}</strong><p>{printer.media.name} · {printer.retired?t('web.Printing.retired'):readinessLabel(printer.readiness)}</p></li>{/each}</ul>
   {#if canConfigure}<p>{t('web.Printing.registrationHelp')}</p>{/if}
  </section>
  <section><h2>{t('web.Printing.connectors')}</h2>{#if connectors.length===0}<p>{t('web.Printing.emptyConnectors')}</p>{/if}<ul>{#each connectors as connector(connector.id)}<li><strong>{connector.name}</strong><p>{connectorStatusLabel(connector)}</p><p>{t('web.Printing.lastSeen')}: {connector.lastSeenAt?timestampLabel(connector.lastSeenAt):t('web.Printing.neverSeen')}</p></li>{/each}</ul></section>
  <section><h2>{t('web.Printing.jobs')}</h2><PrintJobList {jobs} {printers} {scope} {canPrint} {busy} onCancel={cancel} onResolve={resolve}/>{#if nextCursor}<Button.Root variant="outline" disabled={busy} onclick={()=>void moreJobs()}>{t('web.Printing.moreJobs')}</Button.Root>{/if}</section>
 {/if}
</section>
<style>.printing-settings{display:grid;gap:var(--space-6);max-width:48rem}.defaults-form{display:grid;gap:var(--space-3);max-width:32rem}.check-row{display:flex;gap:var(--space-3);align-items:center}.section-heading{display:flex;flex-wrap:wrap;gap:var(--space-3);align-items:center;justify-content:space-between}ul{list-style:none;padding:0;display:grid;gap:var(--space-3)}li{border-bottom:1px solid var(--border);padding-block:var(--space-2)}p{color:var(--muted-foreground);margin-block:var(--space-2)}strong,p{overflow-wrap:anywhere}</style>
