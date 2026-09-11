<script lang="ts">
 import type {AssetExpiration, AssetExpirationContext} from '$lib/domain/inventory';
 import {formatAssetExpiration} from '$lib/application/expirationPresentation';
 import Clock from '@lucide/svelte/icons/clock';
 import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
 let {expiration,context,locale}:{expiration?:AssetExpiration;context?:AssetExpirationContext;locale?:string}=$props();
 let warning=$derived(!!context?.trackingEnabled && context.state !== 'current');
 let expired=$derived(context?.state === 'expired');
 let label=$derived(context && !context.trackingEnabled ? 'Tracking disabled' : warning ? expired ? 'Expired' : 'Expires soon' : 'Expiration');
</script>
{#if expiration}<span class="expiration-label" class:warning class:expired={warning && expired}>{#if warning}{#if expired}<TriangleAlert size={14} aria-hidden="true" />{:else}<Clock size={14} aria-hidden="true" />{/if}{/if}<span>{label}: {formatAssetExpiration(expiration,locale)}</span></span>{/if}
<style>
 .expiration-label {display:flex;align-items:center;gap:0.25rem;font-size:var(--text-metadata-size);line-height:var(--text-metadata-line-height);color:var(--muted-foreground);}
 .warning {color:var(--warning-foreground, var(--foreground));} .expired {color:var(--destructive);}
</style>
