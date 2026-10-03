<script lang="ts">
  import {getContext} from 'svelte';
  import {t} from '$lib/presentation/localization';
  import {labelWorkspaceContext,type LabelWorkspace} from '$lib/ports/labels';
  import type {LabelScope} from '$lib/domain/label';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import Ellipsis from '@lucide/svelte/icons/ellipsis';
  import WorkspaceTaskSheet from '../workspace/action-surface/WorkspaceTaskSheet.svelte';
  import LabelOptions from './LabelOptions.svelte';
  let {scope,disabled=false}:{scope:LabelScope;disabled?:boolean}=$props();
  const workspace=getContext<()=>LabelWorkspace|null>(labelWorkspaceContext);
  let open=$state(false);
</script>
{#if workspace?.()}
  <DropdownMenu.Root><DropdownMenu.Trigger>{#snippet child({props})}<Button.Root {...props} variant="outline" {disabled} aria-label={t('labels.web.more')}><Ellipsis/></Button.Root>{/snippet}</DropdownMenu.Trigger>
    <DropdownMenu.Content><DropdownMenu.Item onclick={()=>{open=true;}}>{t('labels.web.options')}</DropdownMenu.Item></DropdownMenu.Content>
  </DropdownMenu.Root>
  <WorkspaceTaskSheet {open} title={t('labels.web.options')} onOpenChange={value=>{open=value;}}>
    {#if open}{#key workspace()}{#key `${scope.tenantId}:${scope.inventoryId}:${scope.assetId}`}<LabelOptions workspace={workspace()!} {scope}/>{/key}{/key}{/if}
  </WorkspaceTaskSheet>
{/if}
