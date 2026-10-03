<script lang="ts">
  import {setContext,untrack} from 'svelte';
  import {Button} from '$lib/components/ui/button/index.js';
  import AssetLabelMenu from './AssetLabelMenu.svelte';
  import {labelWorkspaceContext,type LabelWorkspace} from '$lib/ports/labels';
  import {printingWorkspaceContext,type PrintingWorkspace} from '$lib/ports/printingRepository';
  let {labelWorkspace,printingWorkspace}:{labelWorkspace:LabelWorkspace;printingWorkspace:PrintingWorkspace}=$props();
  let labels=$state<LabelWorkspace|null>(untrack(()=>labelWorkspace));
  let canPrint=$state(true),disabled=$state(false),assetId=$state('asset');
  setContext(labelWorkspaceContext,()=>labels);
  setContext(printingWorkspaceContext,untrack(()=>printingWorkspace));
</script>
<Button onclick={()=>{canPrint=false;}}>Viewer</Button>
<Button onclick={()=>{disabled=true;}}>Saving</Button>
<Button onclick={()=>{assetId='other-asset';}}>Change asset</Button>
<Button onclick={()=>{labels=null;}}>Clear labels</Button>
{#key assetId}<AssetLabelMenu scope={{tenantId:'tenant',inventoryId:'inventory',assetId}} {canPrint} {disabled} recentJobId="job"/>{/key}
