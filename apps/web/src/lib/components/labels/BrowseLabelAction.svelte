<script lang="ts">
  import {getContext,tick} from 'svelte';
  import {goto} from '$app/navigation';
  import {t} from '$lib/presentation/localization';
  import {labelWorkspaceContext,type LabelWorkspace} from '$lib/ports/labels';
  import type {LabelDestination} from '$lib/domain/label';
  import {labelDestinationHref} from '$lib/application/labels/labelNavigation';
  import * as Button from '$lib/components/ui/button/index.js';
  import ScanLine from '@lucide/svelte/icons/scan-line';
  import WorkspaceTaskSheet from '../workspace/action-surface/WorkspaceTaskSheet.svelte';
  import LabelScanner from './LabelScanner.svelte';
  const workspace=getContext<()=>LabelWorkspace|null>(labelWorkspaceContext);
  let open=$state(false);
  async function resolved(destination:LabelDestination){open=false;await tick();await goto(labelDestinationHref(destination));}
</script>
{#if workspace?.()}
  <Button.Root variant="outline" onclick={()=>{open=true;}}><ScanLine/>{t('labels.web.scan')}</Button.Root>
  <WorkspaceTaskSheet {open} title={t('labels.web.scan')} onOpenChange={value=>{open=value;}}>{#if open}<LabelScanner workspace={workspace()!} onResolved={destination=>void resolved(destination)}/>{/if}</WorkspaceTaskSheet>
{/if}
