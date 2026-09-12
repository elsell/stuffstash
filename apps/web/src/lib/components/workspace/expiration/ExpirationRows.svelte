<script lang="ts">
 import type { ExpirationItem } from '$lib/ports/expirationRepository';
 import type { Asset } from '$lib/domain/inventory';
 import { workspaceRouteHref } from '$lib/application/workspaceRoute';
 import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
 import CheckoutBadge from '../CheckoutBadge.svelte';
 import AssetThumb from '../AssetThumb.svelte';
 import AssetExpirationLabel from '../AssetExpirationLabel.svelte';
 let { items, onOpenAsset, grouped = true }: { items: ExpirationItem[]; onOpenAsset: (asset: Asset) => void; grouped?: boolean } = $props();
 function group(item: ExpirationItem) { return `${item.expirationContext?.state === 'expired' ? 'Expired · ' : ''}${new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(new Date(`${item.expiration!.date.slice(0,7)}-01T12:00:00Z`))}`; }
 function open(event: MouseEvent, item: Asset) { if (shouldHandleWorkspaceLinkClick(event)) { event.preventDefault(); onOpenAsset(item); } }
</script>
<div class="expiry-rows">
 {#each items as item, index (item.id)}
  {#if grouped && (index === 0 || group(items[index - 1]) !== group(item))}<h2 class="month">{group(item)}</h2>{/if}
  <a id={`expiration-${item.id}`} class="expiry-row" href={workspaceRouteHref({mode:'asset',assetId:item.id},item.tenantId,item.inventoryId)} onclick={event => open(event,item)}>
   <AssetThumb asset={item} size="md" />
   <span class="details"><strong>{item.title}</strong>{#if item.currentCheckout}<CheckoutBadge checkout={item.currentCheckout} />{/if}<AssetExpirationLabel expiration={item.expiration} context={item.expirationContext} /><span class="placement">{item.ancestorPath.map(parent => parent.title).join(' / ') || 'No location'}</span></span>
   <span aria-hidden="true" class="arrow">›</span>
  </a>
 {/each}
</div>
<style>
 .expiry-rows {min-width:0;} .month {font-size:var(--text-body-size);font-weight:600;margin:1.75rem 0 .5rem;color:var(--muted-foreground);}
 .expiry-row {display:flex;align-items:center;gap:1rem;min-height:5.5rem;padding:1rem 0;border-bottom:1px solid var(--border);color:var(--foreground);text-decoration:none;}
 .expiry-row:hover {background:var(--muted);} .expiry-row:focus-visible {outline:2px solid var(--ring);outline-offset:4px;border-radius:.5rem;}
 .details {display:grid;gap:.35rem;flex:1;min-width:0;} strong {font-size:var(--text-body-size);overflow-wrap:anywhere;} .placement {font-size:var(--text-metadata-size);color:var(--muted-foreground);overflow-wrap:anywhere;} .arrow {font-size:1.5rem;color:var(--muted-foreground);}
</style>
