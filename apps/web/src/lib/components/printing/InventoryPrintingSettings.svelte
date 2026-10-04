<script lang="ts">
import { onMount, tick, untrack } from 'svelte';
import { t } from '$lib/presentation/localization';
import { timestampLabel } from '$lib/presentation/timestamp';
import { labelMediaName, connectorAvailabilityLabel, readinessReasonLabel, readinessLabel, connectorStatusLabel, printingFailureMessage } from '$lib/presentation/printing';
import { PrintingFailure, type ReportedPrintOutcome, type PrintScope, type PrintDefaults, type RegisteredPrinter, type PrintConnector, type LabelTemplate, type PrintJob, type PrinterMediaChoice } from '$lib/domain/printing';
import type { PrintingRepository, PrintIntents } from '$lib/ports/printingRepository';
import * as Button from '$lib/components/ui/button/index.js';
import { Checkbox } from '$lib/components/ui/checkbox/index.js';
import { Label } from '$lib/components/ui/label/index.js';
import PairingChoice from './PairingChoice.svelte';
import PrintStatusPolling from './PrintStatusPolling.svelte';
import {printJobNeedsPolling} from '$lib/application/printing/polling';
import type {PrintPollingRuntime} from '$lib/ports/printPolling';
import PrintJobList from './PrintJobList.svelte';
import PrinterSettingsDialog from './PrinterSettingsDialog.svelte';
import PrinterTestDialog from './PrinterTestDialog.svelte';
import PrinterSetupDialog from './PrinterSetupDialog.svelte';
import type {TextClipboard} from '$lib/ports/textClipboard';
import AssetPrintDialog from './AssetPrintDialog.svelte';
let { scope, repository, intents, pollingRuntime, canConfigure, canPrint, view = 'settings', apiBaseUrl, clipboard, historyHref, onNavigate }: {
    scope: PrintScope;
    repository: PrintingRepository;
    intents?: PrintIntents;
    pollingRuntime?:PrintPollingRuntime;
    canConfigure: boolean;
    canPrint: boolean;
    view?:'settings'|'history';
    apiBaseUrl?:string;
    clipboard?:TextClipboard;
    historyHref?:string;
    onNavigate?:(href:string)=>void;
} = $props();
let titleElement:HTMLHeadingElement;
let lastView=untrack(()=>view);
$effect(()=>{if(view!==lastView){lastView=view;void tick().then(()=>titleElement?.focus());}});
let setupOpen=$state(false);let setupTrigger:HTMLElement|undefined;
$effect(()=>{if(!canConfigure)setupOpen=false;});
let mediaProfiles=$state<PrinterMediaChoice[]>([]);
function mediaName(printer:RegisteredPrinter){return labelMediaName(mediaProfiles.find(p=>p.adapterId===printer.adapterId&&p.media.presetId===printer.media.presetId&&p.media.version===printer.media.version)?.media??printer.media);}
let printers = $state<RegisteredPrinter[]>([]), connectors = $state<PrintConnector[]>([]), templates = $state<LabelTemplate[]>([]), jobs = $state<PrintJob[]>([]);
const activeJobs=$derived(jobs.filter(job=>printJobNeedsPolling(job)));
let draft = $state<PrintDefaults | null>(null), busy = $state(true), error = $state(''), saved = $state(false), conflict = $state(false), nextCursor = $state<string | undefined>();
let savedDraft=$state('');
const savedUnchanged=$derived(saved&&JSON.stringify(draft)===savedDraft);
let editing=$state<RegisteredPrinter|null>(null);
let editTrigger:HTMLElement|undefined;
function editPrinter(printer:RegisteredPrinter,trigger:HTMLElement){if(!canConfigure||busy)return;editTrigger=trigger;editing=printer;}
$effect(()=>{if(!canConfigure)editing=null;});
let testPrinterId=$state('');
let testTrigger:HTMLElement|undefined;
function openTest(printer:RegisteredPrinter,trigger:HTMLElement){if(!canPrint||busy||printer.retired||!intents)return;testTrigger=trigger;testPrinterId=printer.id;}
let reprint = $state<PrintJob | null>(null);
let reprintTrigger: HTMLElement | undefined;
function openReprint(job: PrintJob, trigger: HTMLElement) {
    if (!canPrint || busy || !intents || !job.assetId || !['completed','failed','canceled'].includes(job.status)) return;
    reprintTrigger=trigger; reprint=job;
}
$effect(()=>{ if(!canPrint){reprint=null;testPrinterId='';} });
const activePrinters = $derived(printers.filter(p => !p.retired));
const selectedTemplate = $derived(draft ? `${draft.templateId}:${draft.templateVersion}` : '');
const missingDefault = $derived(Boolean(draft?.defaultPrinterId && !activePrinters.some(p => p.id === draft?.defaultPrinterId)));
onMount(() => { void load(); });
async function load() { busy = true; error = ''; saved = false; try {
    const values = await Promise.all([repository.printers(scope), repository.connectors(scope), repository.templates(scope), repository.settings(scope), repository.jobs(scope), repository.mediaProfiles(scope).catch(()=>[])]);
    [printers, connectors, templates, draft] = values;
    jobs = values[4].items;
    mediaProfiles=values[5];
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
    pollError=false;
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
    savedDraft = JSON.stringify(draft);
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
    const merged=new Map(jobs.map(job=>[job.id,job]));
    for(const job of page.items){const prior=merged.get(job.id);if(!prior||prior.revision<=job.revision)merged.set(job.id,job);}
    jobs=[...merged.values()];
    nextCursor = page.nextCursor;
    pollError=false;
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
let pollError=$state(false);
async function pollStatus(signal:AbortSignal){
    const capturedScope={...scope};const displayedJobs=[...jobs];
    const [currentPrinters,currentConnectors,page]=await Promise.all([repository.printers(capturedScope,signal),repository.connectors(capturedScope,signal),repository.jobs(capturedScope,undefined,signal)]);
    signal.throwIfAborted();
    const ids=new Set(page.items.map(j=>j.id));
    const older=await Promise.all(displayedJobs.filter(j=>!ids.has(j.id)&&printJobNeedsPolling(j)).map(j=>repository.job(capturedScope,j.id,signal)));
    return {currentPrinters,currentConnectors,page,older};
}
function receiveStatus(value:Awaited<ReturnType<typeof pollStatus>>){
    if(busy)return;pollError=false;
    printers=value.currentPrinters.map(p=>{const current=printers.find(c=>c.id===p.id);return current&&current.revision>p.revision?current:p;});
    connectors=value.currentConnectors;
    const old=new Map(jobs.map(j=>[j.id,j]));
    const refreshed=[...value.page.items,...value.older].map(j=>{const current=old.get(j.id);return current&&current.revision>j.revision?current:j;});
    const ids=new Set(refreshed.map(j=>j.id));jobs=[...refreshed,...jobs.filter(j=>!ids.has(j.id))];
}
</script>
<PrintStatusPolling identity={JSON.stringify([scope.tenantId,scope.inventoryId])} enabled={!busy&&Boolean(draft)&&!editing&&!testPrinterId&&!reprint&&!setupOpen} read={pollStatus} publish={receiveStatus} failed={()=>{if(!busy)pollError=true;}} runtime={pollingRuntime}/>
<section class="printing-settings" aria-labelledby="printing-settings-title" aria-busy={busy}>
 <header><h1 bind:this={titleElement} tabindex="-1" id="printing-settings-title">{t(view==='history'?'web.Printing.jobs':'web.Printing.title')}</h1><p>{t(view==='history'?'web.Printing.historyOrder':'web.Printing.description')}</p></header>
 {#if error}<p role="alert">{error}</p>{/if}
 {#if pollError}<p role="status">{t('web.Printing.statusRefreshFailed')}</p>{/if}
 {#if !draft}{#if busy}<p role="status">{t('web.Printing.loading')}</p>{:else}<Button.Root onclick={()=>void load()}>{t('web.Printing.reload')}</Button.Root>{/if}
 {:else}

  {#if view==='history'}
   <div><Button.Root variant="outline" disabled={busy} onclick={()=>void refresh()}>{t('web.Printing.refresh')}</Button.Root></div>
   <PrintJobList {jobs} {printers} {scope} {canPrint} {busy} onCancel={cancel} onResolve={resolve} onReprint={intents?openReprint:undefined}/>
   {#if nextCursor}<div><Button.Root variant="outline" disabled={busy} onclick={()=>void moreJobs()}>{t('web.Printing.moreJobs')}</Button.Root></div>{/if}
  {:else}
  <div class="settings-grid">
  <section class="printers-panel" aria-labelledby="registered-printers-title">
   <div class="section-heading"><h2 id="registered-printers-title">{t('web.Printing.registered')}</h2><Button.Root variant="ghost" disabled={busy} onclick={()=>void refresh()}>{t('web.Printing.refresh')}</Button.Root></div>
   {#if printers.length===0}<p class="empty-state">{t('web.Printing.emptyPrinters')}</p>{/if}
   <ul class="printer-list">{#each printers as printer(printer.id)}
    <li class="printer-row">
     <div class="printer-heading"><h3>{printer.name}</h3><span class="readiness" class:ready={!printer.retired&&printer.readiness==='ready'}>{printer.retired?t('web.Printing.retired'):readinessLabel(printer.readiness)}</span></div>
     <p>{mediaName(printer)}</p>
     {#if printer.readinessReason}<p class="attention">{readinessReasonLabel(printer.readinessReason)}</p>{/if}
     <div class="printer-actions">{#if canPrint&&intents&&!printer.retired}<Button.Root variant="outline" disabled={busy} onclick={event=>openTest(printer,event.currentTarget)}>{t('web.Printing.testLabel')}</Button.Root>{/if}{#if canConfigure}<Button.Root variant="ghost" disabled={busy} onclick={event=>editPrinter(printer,event.currentTarget)}>{t('web.Printing.editPrinter')}</Button.Root>{/if}</div>
     {#if printer.reportedAt}<details class="diagnostics"><summary>{t('web.Printing.printerDetails')}</summary><p>{t('web.Printing.reportedAt',{time:timestampLabel(printer.reportedAt)})}</p></details>{/if}
    </li>
   {/each}</ul>
   {#if canConfigure&&apiBaseUrl}<div class="setup-action"><Button.Root onclick={event=>{setupTrigger=event.currentTarget;setupOpen=true;}}>{t('web.Printing.addPrinter')}</Button.Root></div>{/if}
   {#if historyHref}<div class="history-action"><Button.Root variant="link" href={historyHref} onclick={event=>{if(onNavigate&&!event.metaKey&&!event.ctrlKey&&!event.shiftKey&&!event.altKey&&event.button===0){event.preventDefault();onNavigate(historyHref!);}}}>{t('web.Printing.viewHistory')}</Button.Root></div>{/if}
   {#if activeJobs.length}<section class="active-jobs"><h2>{t('web.Printing.activeJobs')}</h2><PrintJobList jobs={activeJobs} {printers} {scope} {canPrint} {busy} onCancel={cancel} onResolve={resolve}/></section>{/if}
   {#each connectors.filter(connector=>connector.availability!=='online'||connector.authorizationPending||connector.state!=='active') as connector(connector.id)}<p class="attention">{connector.name} · {connectorAvailabilityLabel(connector)} · {connectorStatusLabel(connector)}</p>{/each}
   <details class="computers"><summary>{t('web.Printing.connectors')} <span class="count">{connectors.length}</span></summary>
    {#if connectors.length===0}<p>{t('web.Printing.emptyConnectors')}</p>{/if}
    <ul>{#each connectors as connector(connector.id)}<li><strong>{connector.name}</strong><p>{connectorStatusLabel(connector)} · {connectorAvailabilityLabel(connector)}</p><p>{t('web.Printing.lastSeen')}: {connector.lastSeenAt?timestampLabel(connector.lastSeenAt):t('web.Printing.neverSeen')}</p></li>{/each}</ul>
   </details>
  </section>
  <section class="defaults-panel"><h2>{t('web.Printing.defaults')}</h2>
   {#if !canConfigure}<p>{t('web.Printing.readOnly')}</p>{/if}
   {#if missingDefault}<p role="status">{t('web.Printing.missingDefault')}</p>{/if}
   <form onsubmit={save} class="defaults-form">
    <PairingChoice id="default-print-destination" label={t('web.Printing.defaultPrinter')} value={draft.defaultPrinterId??''} options={[{value:'',label:t('web.Printing.noDefault')},...activePrinters.map(p=>({value:p.id,label:`${p.name} — ${mediaName(p)}`}))]} disabled={!canConfigure||busy||conflict} onChange={value=>{if(draft){draft.defaultPrinterId=value||null;if(!value)draft.printOnCreateDefault=false;saved=false;}}}/>
    <PairingChoice id="default-print-template" label={t('web.Printing.template')} value={selectedTemplate} options={templates.map(template=>({value:`${template.id}:${template.version}`,label:template.name}))} disabled={!canConfigure||busy||conflict} onChange={value=>{const template=templates.find(t=>`${t.id}:${t.version}`===value);if(draft&&template){draft.templateId=template.id;draft.templateVersion=template.version;saved=false;}}}/>
    <div class="check-row"><Checkbox id="default-print-reference" aria-describedby="default-reference-help" bind:checked={draft.showReference} disabled={!canConfigure||busy||conflict}/><div><Label for="default-print-reference">{t('web.Printing.showReference')}</Label><p class="field-hint" id="default-reference-help">{t('web.Printing.referenceHelp')}</p></div></div>
    <div class="check-row"><Checkbox id="default-auto-print" bind:checked={draft.printOnCreateDefault} disabled={!canConfigure||busy||conflict||!draft.defaultPrinterId||missingDefault}/><Label for="default-auto-print">{t('web.Printing.autoPrint')}</Label></div>
    {#if canConfigure}<div class="save-action"><Button.Root type="submit" disabled={busy||conflict||missingDefault}>{t('web.Printing.save')}</Button.Root></div>{/if}
    {#if conflict}<Button.Root variant="outline" disabled={busy} onclick={()=>void load()}>{t('web.Printing.reload')}</Button.Root>{/if}
    {#if savedUnchanged}<p role="status">{t('web.Printing.saved')}</p>{/if}
   </form>
  </section>
  </div>
  {/if}
 {/if}
</section>
{#if setupOpen&&canConfigure&&apiBaseUrl}<PrinterSetupDialog {apiBaseUrl} {clipboard} onClose={()=>{setupOpen=false;}} onRestoreFocus={()=>setupTrigger?.focus()}/>{/if}
{#if reprint?.assetId && intents && canPrint}
 {#key reprint.id}<AssetPrintDialog {scope} assetId={reprint.assetId} predecessor={reprint.id} {repository} {intents} onClose={()=>{reprint=null;}} onRestoreFocus={()=>{if(reprintTrigger?.isConnected)reprintTrigger.focus();else titleElement?.focus();}}/>{/key}
{/if}
{#if testPrinterId && intents && canPrint}
 {#key testPrinterId}<PrinterTestDialog {scope} printerId={testPrinterId} {repository} {intents} onClose={()=>{testPrinterId='';}} onRestoreFocus={()=>testTrigger?.focus()}/>{/key}
{/if}
{#if editing && canConfigure}
 {#key editing.id}<PrinterSettingsDialog {scope} printer={editing} {repository} onSaved={updated=>{printers=printers.map(p=>p.id===updated.id?updated:p);}} onClose={()=>{editing=null;}} onRestoreFocus={()=>editTrigger?.focus()}/>{/key}
{/if}
<style>
.printing-settings{display:grid;gap:var(--space-6);width:100%;max-width:70rem;min-width:0}
header h1{margin:0;font-size:var(--text-title-size);letter-spacing:-.025em}header p{margin-top:var(--space-2)}
h2{font-size:var(--text-section-size);font-weight:600;margin:0 0 var(--space-4)}h3{font-size:var(--text-label-size);font-weight:600;margin:0}
.settings-grid{display:grid;grid-template-columns:minmax(0,1.2fr) minmax(0,1fr);gap:var(--space-6);align-items:start}
.printers-panel,.defaults-panel{min-width:0}.defaults-panel{background:var(--muted);border-radius:var(--radius-lg);padding:var(--space-4)}
.defaults-form{display:grid;gap:var(--space-4);min-width:0}.save-action{display:flex;margin-top:var(--space-2)}
.field-hint{font-size:var(--text-metadata-size);margin:var(--space-1) 0 0}.check-row{display:flex;gap:var(--space-3);align-items:start}.check-row :global(label){line-height:1.5}
.section-heading,.printer-heading{display:flex;flex-wrap:wrap;gap:var(--space-3);align-items:center;justify-content:space-between}.section-heading h2{margin:0}.section-heading{margin-bottom:var(--space-3)}
ul{list-style:none;padding:0;margin:0;display:grid;gap:var(--space-3)}li{min-width:0;padding-block:var(--space-3)}.printer-row{border-bottom:1px solid var(--border)}
.printer-actions{display:flex;flex-wrap:wrap;gap:var(--space-2);margin-top:var(--space-3)}
.readiness{font-size:var(--text-metadata-size);font-weight:600;background:var(--muted);padding:.25rem .625rem;border-radius:var(--radius-pill)}.readiness.ready{color:var(--foreground)}
p{color:var(--muted-foreground);margin-block:var(--space-2);line-height:1.5}strong,p,h3,summary{overflow-wrap:anywhere}
details{font-size:var(--text-metadata-size)}summary{cursor:pointer;min-height:44px;align-content:center;font-weight:500}summary:focus-visible{outline:2px solid var(--ring);outline-offset:3px;border-radius:var(--radius-sm)}
.diagnostics{margin-top:var(--space-2);color:var(--muted-foreground)}.setup-action,.history-action,.active-jobs,.computers{margin-top:var(--space-3)}.computers{border-top:1px solid var(--border);padding-top:var(--space-2)}.count{color:var(--muted-foreground);margin-left:.5rem}.attention{color:var(--foreground)}.empty-state{padding-block:var(--space-4)}
@media(max-width:1024px){.settings-grid{grid-template-columns:minmax(0,1fr)}.defaults-panel{padding:var(--space-4)}header h1{font-size:var(--text-title-size)}}
</style>
