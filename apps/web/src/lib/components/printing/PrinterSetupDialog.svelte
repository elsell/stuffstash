<script lang="ts">
 import * as Dialog from '$lib/components/ui/dialog/index.js';
 import {Button} from '$lib/components/ui/button/index.js';
 import {Input} from '$lib/components/ui/input/index.js';
 import {Label} from '$lib/components/ui/label/index.js';
 import {t} from '$lib/presentation/localization';
 import type {TextClipboard} from '$lib/ports/textClipboard';
 import {printerRegistrationCommand} from '$lib/application/printing/registrationCommand';
 let {apiBaseUrl,clipboard,onClose,onRestoreFocus}:{apiBaseUrl:string;clipboard?:TextClipboard;onClose:()=>void;onRestoreFocus:()=>void}=$props();
 let computerName=$state(t('web.Printing.defaultComputerName')),copied=$state(''),copyFailed=$state(false);
 const command=$derived(printerRegistrationCommand(apiBaseUrl,computerName));
 async function copy(){if(!computerName.trim()||!clipboard)return;copyFailed=false;const requestedCommand=command;try{await clipboard.write(requestedCommand);copied=requestedCommand;}catch{copyFailed=true;}}
</script>
<Dialog.Root open onOpenChange={open=>{if(!open)onClose();}}>
 <Dialog.Content class="sm:max-w-xl" onCloseAutoFocus={event=>{event.preventDefault();onRestoreFocus();}}>
  <Dialog.Header><Dialog.Title>{t('web.Printing.addPrinter')}</Dialog.Title><Dialog.Description>{t('web.Printing.setupIntro')}</Dialog.Description></Dialog.Header>
  <ol class="setup-steps">
   <li><h3>{t('web.Printing.setupInstall')}</h3><p>{t('web.Printing.setupInstallHelp')}</p><a href="https://stuffstash.org/cli-downloads/" target="_blank" rel="noopener noreferrer">{t('web.Printing.downloadCLI')}</a> · <a href="https://stuffstash.org/printing/setup/" target="_blank" rel="noopener noreferrer">{t('web.Printing.setupGuide')}</a></li>
   <li><h3>{t('web.Printing.setupRegister')}</h3><Label for="printer-computer-name">{t('web.Printing.computerName')}</Label><Input id="printer-computer-name" bind:value={computerName} placeholder={t('web.Printing.computerNameExample')}/>
    <p>{t('web.Printing.setupCommandHelp')}</p><pre aria-label={t('web.Printing.registrationCommand')}><code>{command}</code></pre>
    <Button variant="outline" disabled={!computerName.trim()||!clipboard} onclick={()=>void copy()}>{t('web.Printing.copyCommand')}</Button>
    {#if copied===command}<p role="status">{t('web.Printing.commandCopied')}</p>{/if}{#if copyFailed}<p role="alert">{t('web.Printing.copyCommandFailed')}</p>{/if}
   </li>
   <li><h3>{t('web.Printing.setupApprove')}</h3><p>{t('web.Printing.setupApproveHelp')}</p></li>
   <li><h3>{t('web.Printing.setupRun')}</h3><p>{t('web.Printing.setupRunHelp')}</p></li>
  </ol>
  <Dialog.Footer><Button onclick={onClose}>{t('web.Printing.close')}</Button></Dialog.Footer>
 </Dialog.Content>
</Dialog.Root>
<style>
.setup-steps{display:grid;gap:var(--space-4);padding-left:var(--space-4);margin:0}li{padding-left:var(--space-2);min-width:0}h3{font-size:var(--text-label-size);font-weight:600;margin:0 0 var(--space-2)}p{color:var(--muted-foreground);margin:var(--space-2) 0}a{text-decoration:underline;text-underline-offset:.2em}pre{white-space:pre-wrap;overflow-wrap:anywhere;background:var(--muted);padding:var(--space-3);border-radius:var(--radius-control);font-size:var(--text-metadata-size);margin-block:var(--space-3)}
</style>
