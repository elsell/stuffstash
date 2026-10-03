<script lang="ts">
import { onMount, onDestroy, untrack } from 'svelte';
import * as Dialog from '$lib/components/ui/dialog/index.js';
import { Button } from '$lib/components/ui/button/index.js';
import { Input } from '$lib/components/ui/input/index.js';
import { Label } from '$lib/components/ui/label/index.js';
import { Checkbox } from '$lib/components/ui/checkbox/index.js';
import { t } from '$lib/presentation/localization';
import { printingFailureMessage } from '$lib/presentation/printing';
import type { PrintingRepository, PrintIntents } from '$lib/ports/printingRepository';
import type { PrintScope, RegisteredPrinter, LabelTemplate, PrintJob } from '$lib/domain/printing';
import PairingChoice from './PairingChoice.svelte';
import PrintJobList from './PrintJobList.svelte';
let { scope, assetId, repository, intents, onClose, onRestoreFocus }: {
    scope: PrintScope;
    assetId: string;
    repository: PrintingRepository;
    intents: PrintIntents;
    onClose: () => void;
    onRestoreFocus?: () => void;
} = $props();
let request = $state(untrack(() => intents.forAsset(scope, assetId)));
let printers = $state<RegisteredPrinter[]>([]), templates = $state<LabelTemplate[]>([]), printerId = $state(''), templateKey = $state(''), showReference = $state(true), copies = $state(1);
let busy = $state(true), locked = $state(untrack(() => request.locked)), error = $state(''), previewUrl = $state(''), rotation = $state(0), job = $state<PrintJob | null>(untrack(() => request.result)), alive = true;
const previewLifetime = new AbortController();
const printer = $derived(printers.find(p => p.id === printerId));
const template = $derived(templates.find(p => `${p.id}:${p.version}` === templateKey));
onMount(() => { void load(); });
onDestroy(() => { alive = false; previewLifetime.abort(); releasePreview(); });
function releasePreview() { if (previewUrl)
    URL.revokeObjectURL(previewUrl); previewUrl = ''; }
function changed() { if (locked)
    return; request.invalidate(); releasePreview(); error = ''; }
async function load() { try {
    const [destinations, layouts, defaults] = await Promise.all([repository.printers(scope), repository.templates(scope), repository.settings(scope)]);
    if (!alive)
        return;
    printers = destinations.filter(p => !p.retired);
    templates = layouts;
    printerId = printers.find(p => p.id === defaults.defaultPrinterId)?.id ?? '';
    templateKey = `${defaults.templateId}:${defaults.templateVersion}`;
    showReference = defaults.showReference;
    const selection = request.selection;
    if (selection) {
        printerId = selection.printerId;
        templateKey = `${selection.templateId}:${selection.templateVersion}`;
        showReference = selection.showReference;
        copies = selection.copies;
    }
    if (request.rendered) {
        rotation = request.rendered.displayRotation;
        previewUrl = URL.createObjectURL(request.rendered.bytes);
    }
}
catch (caught) {
    if (alive)
        error = printingFailureMessage(caught);
}
finally {
    if (alive)
        busy = false;
} }
async function preview() { if (!printer || !template || busy || locked)
    return; busy = true; error = ''; releasePreview(); try {
    const rendered = await request.preview({ printerId, expectedMediaFingerprint: printer.mediaFingerprint, templateId: template.id, templateVersion: template.version, showReference, copies }, printer.media, previewLifetime.signal);
    if (alive) {
        rotation = rendered.displayRotation;
        previewUrl = URL.createObjectURL(rendered.bytes);
    }
}
catch (caught) {
    if (alive)
        error = printingFailureMessage(caught);
}
finally {
    if (alive)
        busy = false;
} }
async function submit() { if (busy)
    return; busy = true; error = ''; try {
    const result = await request.submit();
    if (alive)
        job = result;
}
catch (caught) {
    if (alive)
        error = printingFailureMessage(caught);
}
finally {
    if (alive) {
        locked = request.locked;
        busy = false;
    }
} }
function another() { request = intents.startAnother(scope, assetId); job = null; locked = false; releasePreview(); error = ""; }
async function refresh() { if (!job || busy)
    return; busy = true; error = ''; try {
    const result = await repository.job(scope, job.id);
    if (alive)
        job = result;
}
catch (caught) {
    if (alive)
        error = printingFailureMessage(caught);
}
finally {
    if (alive)
        busy = false;
} }
async function cancel(current: PrintJob) { if (busy)
    return; busy = true; error = ''; try {
    const result = await repository.cancel(scope, current);
    if (alive)
        job = result;
}
catch (caught) {
    if (alive)
        error = printingFailureMessage(caught);
}
finally {
    if (alive)
        busy = false;
} }
</script>
<Dialog.Root open onOpenChange={open=>{if(!open)onClose();}}>
 <Dialog.Content onCloseAutoFocus={event=>{if(onRestoreFocus){event.preventDefault();onRestoreFocus();}}}><Dialog.Header><Dialog.Title>{t('web.Printing.printLabel')}</Dialog.Title><Dialog.Description>{t('web.Printing.description')}</Dialog.Description></Dialog.Header>
 {#if error}<p role="alert">{error}</p>{/if}
 {#if job}{#if job.status==='queued'}<p role="status">{t('web.Printing.requestQueued')}</p>{/if}<PrintJobList jobs={[job]} {printers} {scope} canPrint={true} {busy} onCancel={cancel}/><Button variant="outline" disabled={busy} onclick={()=>void refresh()}>{t('web.Printing.refresh')}</Button>{#if ['completed','failed','canceled'].includes(job.status)}<Button onclick={another}>{t('web.Printing.printAgain')}</Button>{/if}
 {:else}
 <PairingChoice id="label-printer" label={t('web.Printing.printer')} value={printerId} options={printers.map(p=>({value:p.id,label:`${p.name} — ${p.media.name}`}))} disabled={busy||locked} onChange={value=>{printerId=value;changed();}}/>
 <PairingChoice id="label-template" label={t('web.Printing.template')} value={templateKey} options={templates.map(p=>({value:`${p.id}:${p.version}`,label:p.name}))} disabled={busy||locked} onChange={value=>{templateKey=value;changed();}}/>
 <div class="check-row"><Checkbox id="label-reference" bind:checked={showReference} disabled={busy||locked} onchange={changed}/><Label for="label-reference">{t('web.Printing.showReference')}</Label></div>
 <div><Label for="label-copies">{t('web.Printing.copies')}</Label><Input id="label-copies" type="number" min="1" step="1" bind:value={copies} oninput={changed} disabled={busy||locked}/></div>
 {#if previewUrl && printer}<figure><svg role="img" aria-label={t('web.Printing.previewAlt')} viewBox={`0 0 ${rotation%180?printer.media.rasterHeight:printer.media.rasterWidth} ${rotation%180?printer.media.rasterWidth:printer.media.rasterHeight}`}><image href={previewUrl} width={printer.media.rasterWidth} height={printer.media.rasterHeight} transform={rotation===90?`translate(${printer.media.rasterHeight} 0) rotate(90)`:rotation===180?`translate(${printer.media.rasterWidth} ${printer.media.rasterHeight}) rotate(180)`:rotation===270?`translate(0 ${printer.media.rasterWidth}) rotate(270)`:undefined}/></svg></figure>{/if}
 {#if locked}<p role="status">{t('web.Printing.requestUnknown')}</p>{/if}
 <Dialog.Footer>{#if !locked}<Button variant="outline" disabled={busy||!printer||!template||!Number.isInteger(copies)||copies<1} onclick={()=>void preview()}>{t('web.Printing.preview')}</Button>{/if}<Button disabled={busy||!previewUrl} onclick={()=>void submit()}>{t(locked?'web.Printing.retryRequest':'web.Printing.printLabel')}</Button></Dialog.Footer>
 {/if}
 </Dialog.Content>
</Dialog.Root>
<style>.check-row{display:flex;gap:var(--space-3);align-items:center}figure{margin:0;display:grid;place-items:center;background:white;padding:var(--space-3);border:1px solid var(--border)}svg{width:100%;max-height:18rem}p{overflow-wrap:anywhere}</style>
