<script lang="ts">
import {onMount,onDestroy,untrack} from 'svelte';
import * as Dialog from '$lib/components/ui/dialog/index.js';
import {Button} from '$lib/components/ui/button/index.js';
import {Input} from '$lib/components/ui/input/index.js';
import {Label} from '$lib/components/ui/label/index.js';
import {Checkbox} from '$lib/components/ui/checkbox/index.js';
import {t} from '$lib/presentation/localization';
import {labelMediaName,printingFailureMessage} from '$lib/presentation/printing';
import {PrintingFailure,type RegisteredPrinter,type PrinterMediaChoice,type PrintScope} from '$lib/domain/printing';
import type {PrintingRepository} from '$lib/ports/printingRepository';
import PairingChoice from './PairingChoice.svelte';
let {scope,printer,repository,onSaved,onClose,onRestoreFocus}:{scope:PrintScope;printer:RegisteredPrinter;repository:PrintingRepository;onSaved:(printer:RegisteredPrinter)=>void;onClose:()=>void;onRestoreFocus:()=>void}=$props();
let current=$state(untrack(()=>({...printer}))),name=$state(untrack(()=>printer.name)),retired=$state(untrack(()=>printer.retired));
let media=$state<PrinterMediaChoice[]>([]),mediaKey=$state(untrack(()=>`${printer.media.presetId}:${printer.media.version}`));
let busy=$state(true),mustReload=$state(false),error=$state('');let alive=true;
const selected=$derived(media.find(m=>`${m.media.presetId}:${m.media.version}`===mediaKey));
onDestroy(()=>{alive=false;});onMount(()=>{void load(false);});
async function load(refresh:boolean){
 if(!alive)return;busy=true;error='';
 try{
  const choices=await repository.mediaProfiles(scope);
  const actual=refresh?(await repository.printers(scope)).find(p=>p.id===current.id):current;
  if(!alive)return;if(!actual)throw new PrintingFailure('invalid');
  current=actual;media=choices.filter(choice=>choice.adapterId===actual.adapterId);
  name=actual.name;retired=actual.retired;mediaKey=`${actual.media.presetId}:${actual.media.version}`;mustReload=false;
 }catch(caught){if(alive){error=printingFailureMessage(caught);mustReload=true;}}
 finally{if(alive)busy=false;}
}
async function save(event:SubmitEvent){
 event.preventDefault();if(busy||mustReload||!selected||!name.trim())return;busy=true;error='';
 try{
  const updated=await repository.updatePrinter(scope,current,name.trim(),retired,selected.media);
  if(alive){onSaved(updated);onClose();}
 }catch(caught){if(alive){error=printingFailureMessage(caught);mustReload=!(caught instanceof PrintingFailure&&caught.kind==='invalid');}}
 finally{if(alive)busy=false;}
}
</script>
<Dialog.Root open onOpenChange={open=>{if(!open)onClose();}}>
 <Dialog.Content onCloseAutoFocus={event=>{event.preventDefault();onRestoreFocus();}}>
  <Dialog.Header><Dialog.Title>{t('web.Printing.editPrinter')}</Dialog.Title><Dialog.Description>{t('web.Printing.editPrinterHelp')}</Dialog.Description></Dialog.Header>
  {#if error}<p role="alert">{error}</p>{/if}
  <form onsubmit={save}>
   <div><Label for="edit-printer-name">{t('web.Printing.printerName')}</Label><Input id="edit-printer-name" bind:value={name} disabled={busy||mustReload} required/></div>
   <PairingChoice id="edit-printer-media" label={t('web.Printing.registeredMedia')} value={mediaKey} options={media.map(choice=>({value:`${choice.media.presetId}:${choice.media.version}`,label:labelMediaName(choice.media)}))} disabled={busy||mustReload} onChange={value=>{mediaKey=value;}}/>
   <div class="check-row"><Checkbox id="edit-printer-retired" bind:checked={retired} disabled={busy||mustReload}/><Label for="edit-printer-retired">{t('web.Printing.retirePrinter')}</Label></div>
   <Dialog.Footer>{#if mustReload}<Button variant="outline" disabled={busy} onclick={()=>void load(true)}>{t('web.Printing.reloadPrinter')}</Button>{/if}<Button type="submit" disabled={busy||mustReload||!selected||!name.trim()}>{t('web.Printing.savePrinter')}</Button></Dialog.Footer>
  </form>
 </Dialog.Content>
</Dialog.Root>
<style>form{display:grid;gap:var(--space-4)}.check-row{display:flex;align-items:center;gap:var(--space-3)}p{overflow-wrap:anywhere}</style>
