<script lang="ts">
 import * as Button from '$lib/components/ui/button/index.js';
 import type {ExpirationNotification} from '$lib/domain/notification';
 let {segments=[],incomplete=false,disabled=false,onOpen}:{segments?:ExpirationNotification['parentTrail'];incomplete?:boolean;disabled?:boolean;onOpen:(id:string)=>void}=$props();
 function revealParent(node:HTMLElement,_segments:unknown){
  const reveal=()=>{node.scrollLeft=node.scrollWidth;};reveal();
  const observer=typeof ResizeObserver==='undefined'?undefined:new ResizeObserver(reveal);observer?.observe(node);
  window.addEventListener('resize',reveal);
  return {update(){reveal();},destroy(){observer?.disconnect();window.removeEventListener('resize',reveal);}};
 }
</script>
{#if segments.length || incomplete}
 <nav aria-label="Item location" use:revealParent={segments}>
  {#if incomplete}<span aria-label={segments.length?'Partial location path':'Location unavailable'}>{segments.length?'…':'Location unavailable'}</span>{/if}
  {#each segments as segment,index (segment.assetId)}
   {#if index>0}<span aria-hidden="true">/</span>{/if}
   <Button.Root variant="ghost" {disabled} aria-label={`Open ${segment.title}`} onclick={()=>onOpen(segment.assetId)}>{segment.title}</Button.Root>
  {/each}
 </nav>
{/if}
<style>
 nav {display:flex;align-items:center;gap:var(--space-1);overflow-x:auto;min-width:0;max-width:100%;white-space:nowrap;}
 nav :global(button) {flex-shrink:0;}
</style>
