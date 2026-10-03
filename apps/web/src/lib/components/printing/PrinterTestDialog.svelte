<script lang="ts">
import {onMount,onDestroy,untrack} from 'svelte';
import * as Dialog from '$lib/components/ui/dialog/index.js';
import {Button} from '$lib/components/ui/button/index.js';
import {Checkbox} from '$lib/components/ui/checkbox/index.js';
import {Label} from '$lib/components/ui/label/index.js';
import {t} from '$lib/presentation/localization';
import {labelMediaName,printingFailureMessage} from '$lib/presentation/printing';
import {PrintingFailure,type PrintScope,type RegisteredPrinter,type LabelTemplate,type PrintJob,type ReportedPrintOutcome} from '$lib/domain/printing';
import type {PrintingRepository,PrintIntents} from '$lib/ports/printingRepository';
import PairingChoice from './PairingChoice.svelte';
import PrintJobList from './PrintJobList.svelte';
let {scope,printerId,repository,intents,onClose,onRestoreFocus}:{scope:PrintScope;printerId:string;repository:PrintingRepository;intents:PrintIntents;onClose:()=>void;onRestoreFocus:()=>void}=$props();
let request=$state(untrack(()=>intents.forPrinterTest(scope,printerId)));
let printer=$state<RegisteredPrinter|null>(null),templates=$state<LabelTemplate[]>([]),templateKey=$state(''),showReference=$state(true);
let busy=$state(true),locked=$state(untrack(()=>request.locked)),job=$state<PrintJob|null>(untrack(()=>request.result)),error=$state('');
let alive=true;
const template=$derived(templates.find(t=>`${t.id}:${t.version}`===templateKey));
onDestroy(()=>{alive=false;});
onMount(()=>{void load();});
async function load(){
 try{
  const [printers,layouts,defaults]=await Promise.all([repository.printers(scope),repository.templates(scope),repository.settings(scope)]);
  if(!alive)return;
  printer=printers.find(p=>p.id===printerId)??null;templates=layouts;
  const selection=request.selection;
  templateKey=selection?`${selection.templateId}:${selection.templateVersion}`:`${defaults.templateId}:${defaults.templateVersion}`;
  showReference=selection?.showReference??defaults.showReference;
  if(!printer)throw new PrintingFailure('invalid');
 }catch(caught){if(alive)error=printingFailureMessage(caught);}finally{if(alive)busy=false;}
}
async function run(operation:()=>Promise<PrintJob>){
 if(busy)return;busy=true;error='';
 try{const result=await operation();if(alive)job=result;}
 catch(caught){if(alive)error=printingFailureMessage(caught);}
 finally{if(alive){busy=false;locked=request.locked;}}
}
function submit(){
 if(busy)return;
 const retained=request.locked?request.selection:null;
 if(!retained&&(!printer||printer.retired||!template))return;
 const selection=retained??{printerId,expectedMediaFingerprint:printer!.mediaFingerprint,templateId:template!.id,templateVersion:template!.version,showReference,copies:1};
 void run(()=>request.submit(selection));
}
function another(){
 if(!job||busy||!printer||printer.retired)return;
 request=intents.startAnotherPrinterTest(scope,printerId,job);job=null;locked=false;error='';
}
async function resolve(current:PrintJob,outcome:ReportedPrintOutcome){await run(()=>repository.resolve(scope,current,outcome));}
async function cancel(current:PrintJob){await run(()=>repository.cancel(scope,current));}
</script>
<Dialog.Root open onOpenChange={open=>{if(!open)onClose();}}>
 <Dialog.Content onCloseAutoFocus={event=>{event.preventDefault();onRestoreFocus();}}>
  <Dialog.Header><Dialog.Title>{t('web.Printing.testLabel')}</Dialog.Title><Dialog.Description>{t('web.Printing.testDescription')}</Dialog.Description></Dialog.Header>
  {#if error}<p role="alert">{error}</p>{/if}
  {#if printer}<p><strong>{printer.name}</strong> · {labelMediaName(printer.media)}</p>{/if}
  {#if job}
   {#if job.status==='queued'}<p role="status">{t('web.Printing.requestQueued')}</p>{/if}
   <PrintJobList jobs={[job]} printers={printer?[printer]:[]} {scope} canPrint={true} {busy} onCancel={cancel} onResolve={resolve}/>
   <Button variant="outline" disabled={busy} onclick={()=>{const current=job;if(current)void run(()=>repository.job(scope,current.id));}}>{t('web.Printing.refresh')}</Button>
   {#if ['completed','failed','canceled'].includes(job.status)&&printer&&!printer.retired}<Button disabled={busy} onclick={another}>{t('web.Printing.anotherTest')}</Button>{/if}
  {:else}
   <PairingChoice id="test-label-template" label={t('web.Printing.template')} value={templateKey} options={templates.map(t=>({value:`${t.id}:${t.version}`,label:t.name}))} disabled={busy||locked} onChange={value=>{templateKey=value;}}/>
   <div class="check-row"><Checkbox id="test-label-reference" bind:checked={showReference} disabled={busy||locked}/><Label for="test-label-reference">{t('web.Printing.showReference')}</Label></div>
   {#if locked}<p role="status">{t('web.Printing.requestUnknown')}</p>{/if}
   <Dialog.Footer><Button disabled={busy||(!locked&&(!printer||printer.retired||!template))} onclick={submit}>{t(locked?'web.Printing.retryRequest':'web.Printing.printOneTest')}</Button></Dialog.Footer>
  {/if}
 </Dialog.Content>
</Dialog.Root>
<style>.check-row{display:flex;align-items:center;gap:var(--space-3)}p{overflow-wrap:anywhere}</style>
