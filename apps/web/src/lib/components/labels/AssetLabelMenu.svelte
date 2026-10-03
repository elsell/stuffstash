<script lang="ts">
  import {getContext} from 'svelte';
  import Ellipsis from '@lucide/svelte/icons/ellipsis';
  import {t} from '$lib/presentation/localization';
  import {labelWorkspaceContext,type LabelWorkspace} from '$lib/ports/labels';
  import {printingWorkspaceContext,type PrintingWorkspace} from '$lib/ports/printingRepository';
  import type {LabelScope} from '$lib/domain/label';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import {Button} from '$lib/components/ui/button/index.js';
  import WorkspaceTaskSheet from '../workspace/action-surface/WorkspaceTaskSheet.svelte';
  import AssetPrintDialog from '../printing/AssetPrintDialog.svelte';
  import LabelOptions from './LabelOptions.svelte';
  let {scope,canPrint,disabled=false,recentJobId}:{scope:LabelScope;canPrint:boolean;disabled?:boolean;recentJobId?:string}=$props();
  const getLabels=getContext<(()=>LabelWorkspace|null)|undefined>(labelWorkspaceContext);
  const printing=getContext<PrintingWorkspace|undefined>(printingWorkspaceContext);
  const labels=$derived(getLabels?.()??null);
  const printable=$derived(canPrint&&Boolean(printing));
  let action=$state<'download'|'print'|'status'|null>(null);
  let menuOpen=$state(false);
  let trigger=$state<HTMLButtonElement|null>(null);
  $effect(()=>{
    if(disabled){menuOpen=false;action=null;}
    if(action==='download'&&!labels)action=null;
    if((action==='print'||action==='status')&&!printable)action=null;
  });
  function select(next:'download'|'print'|'status'){
    if(disabled||(next==='download'?!labels:!printable))return;
    action=next;
  }
  function restoreFocus(event?:Event){event?.preventDefault();trigger?.focus();}
</script>
{#if labels||printable}
  <DropdownMenu.Root bind:open={menuOpen}>
    <DropdownMenu.Trigger bind:ref={trigger}>{#snippet child({props})}<Button {...props} variant="outline" size="icon" {disabled} aria-label={t('web.Printing.moreActions')}><Ellipsis/></Button>{/snippet}</DropdownMenu.Trigger>
    <DropdownMenu.Content onCloseAutoFocus={event=>{if(action)event.preventDefault();}}>
      {#if labels}<DropdownMenu.Item {disabled} onSelect={()=>select('download')}>{t('labels.web.options')}</DropdownMenu.Item>{/if}
      {#if printable}<DropdownMenu.Item {disabled} onSelect={()=>select('print')}>{t('web.Printing.printLabel')}</DropdownMenu.Item>{/if}
      {#if printable&&recentJobId}<DropdownMenu.Item {disabled} onSelect={()=>select('status')}>{t('web.Printing.viewCreatedJob')}</DropdownMenu.Item>{/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
  {#if labels}
    <WorkspaceTaskSheet open={action==='download'} title={t('labels.web.options')} onOpenChange={open=>{if(!open)action=null;}} onCloseAutoFocus={restoreFocus}>
      {#if action==='download'}{#key labels}<LabelOptions workspace={labels} {scope}/>{/key}{/if}
    </WorkspaceTaskSheet>
  {/if}
  {#if (action==='print'||action==='status')&&printable&&printing}
    <AssetPrintDialog scope={{tenantId:scope.tenantId,inventoryId:scope.inventoryId}} assetId={scope.assetId} initialJobId={action==='status'?recentJobId:undefined} repository={printing.repository} intents={printing.intents} onRestoreFocus={restoreFocus} onClose={()=>{action=null;}}/>
  {/if}
{/if}
