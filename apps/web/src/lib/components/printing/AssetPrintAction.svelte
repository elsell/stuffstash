<script lang="ts">
import { getContext } from 'svelte';
import Ellipsis from '@lucide/svelte/icons/ellipsis';
import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
import { Button } from '$lib/components/ui/button/index.js';
import { printingWorkspaceContext, type PrintingWorkspace } from '$lib/ports/printingRepository';
import type { PrintScope } from '$lib/domain/printing';
import { t } from '$lib/presentation/localization';
import AssetPrintDialog from './AssetPrintDialog.svelte';
let { scope, assetId }: {
    scope: PrintScope;
    assetId: string;
} = $props();
const printing = getContext<PrintingWorkspace | undefined>(printingWorkspaceContext);
let open = $state(false);
let trigger = $state<HTMLButtonElement | null>(null);
</script>
{#if printing}
 <DropdownMenu.Root><DropdownMenu.Trigger bind:ref={trigger}>{#snippet child({props})}<Button variant="outline" size="icon" aria-label={t('web.Printing.moreActions')} {...props}><Ellipsis/></Button>{/snippet}</DropdownMenu.Trigger><DropdownMenu.Content><DropdownMenu.Item onSelect={()=>{open=true;}}>{t('web.Printing.printLabel')}</DropdownMenu.Item></DropdownMenu.Content></DropdownMenu.Root>
 {#if open}{#key `${scope.tenantId}/${scope.inventoryId}/${assetId}`}<AssetPrintDialog {scope} {assetId} repository={printing.repository} intents={printing.intents} onRestoreFocus={()=>trigger?.focus()} onClose={()=>{open=false;}}/>{/key}{/if}
{/if}
