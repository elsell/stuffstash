<script lang="ts">
 import type {AssetExpiration} from '$lib/domain/inventory';
 import {formatAssetExpiration} from '$lib/application/expirationPresentation';
 let {expiration,locale}:{expiration?:AssetExpiration;locale?:string}=$props();
</script>
{#if expiration}<span class="expiration-label">Expiration: {formatAssetExpiration(expiration,locale)}</span>{/if}
<style>
 .expiration-label {display:block;font-size:var(--text-metadata-size);line-height:var(--text-metadata-line-height);color:var(--muted-foreground);}
</style>
