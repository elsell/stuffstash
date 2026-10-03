<script lang="ts">
import { onMount } from 'svelte';
import * as Button from '$lib/components/ui/button/index.js';
import * as Card from '$lib/components/ui/card/index.js';
import { Input } from '$lib/components/ui/input';
import { Label } from '$lib/components/ui/label/index.js';
import AuthBrand from '$lib/components/auth/AuthBrand.svelte';
import { t } from '$lib/presentation/localization';
import { PairingFailure, type PairingInventory, type PairingReview, type PairingFailureKind } from '$lib/domain/printPairing';
import type { PrintRotationRepository, RotationTarget, RotationConnector } from '$lib/ports/printRotationRepository';
let { pairingId, target, repository, onSignIn }: {
    pairingId: string;
    target: RotationTarget;
    repository: PrintRotationRepository;
    onSignIn: () => Promise<void>;
} = $props();
let inventory = $state<PairingInventory | null>(null), connector = $state<RotationConnector | null>(null), review = $state<PairingReview | null>(null);
let code = $state(''), busy = $state(true), success = $state(false), failure = $state<PairingFailureKind | null>(null);
const errorKey = $derived(failure === 'authentication_required' ? 'web.PrintPairing.authRequired' : failure === 'denied' ? 'web.PrintPairing.denied' : failure === 'invalid' ? 'web.PrintPairing.invalid' : 'web.PrintPairing.unavailable');
function handle(error: unknown) { failure = error instanceof PairingFailure ? error.kind : 'unavailable'; }
async function initialize() { busy = true; failure = null; inventory = null; connector = null; try {
    const inventories = await repository.inventories();
    const selected = inventories.find(i => i.tenantId === target.tenantId && i.inventoryId === target.inventoryId);
    if (!selected)
        throw new PairingFailure('denied');
    const found = await repository.connector(selected, target.connectorId);
    if (found.id !== target.connectorId)
        throw new PairingFailure('invalid');
    inventory = selected;
    connector = found;
}
catch (error) {
    handle(error);
}
finally {
    busy = false;
} }
onMount(() => { void initialize(); });
async function inspect(event: SubmitEvent) { event.preventDefault(); if (busy || !inventory || code.trim().length !== 8)
    return; busy = true; failure = null; review = null; try {
    const next = await repository.review(pairingId, inventory, code.trim().toUpperCase());
    if (next.id !== pairingId || next.rotation !== true)
        throw new PairingFailure('invalid');
    const current = await repository.connector(inventory, target.connectorId);
    if (current.id !== target.connectorId)
        throw new PairingFailure('invalid');
    connector = current;
    review = next;
}
catch (error) {
    handle(error);
}
finally {
    busy = false;
} }
async function approve() { if (busy || !inventory || !connector || !review)
    return; busy = true; failure = null; try {
    await repository.rotate(pairingId, inventory, code.trim().toUpperCase(), connector.id, connector.generation);
    success = true;
    code = '';
}
catch (error) {
    review = null;
    handle(error);
}
finally {
    busy = false;
} }
</script>
<main class="rotation-shell"><Card.Root class="rotation-card" aria-busy={busy}>
 <Card.Header><AuthBrand/><Card.Title role="heading" aria-level={1}>{t('web.PrintRotation.title')}</Card.Title></Card.Header>
 <Card.Content>
 {#if success}<p role="status">{t('web.PrintRotation.success')}</p><Button.Root href={`/tenants/${encodeURIComponent(target.tenantId)}/inventories/${encodeURIComponent(target.inventoryId)}`}>{t('web.PrintPairing.openInventory')}</Button.Root>
 {:else}
  {#if failure}<p role="alert">{t(errorKey)}</p>{#if failure==='authentication_required'}<Button.Root onclick={()=>void onSignIn()}>{t('web.PrintPairing.continueSignIn')}</Button.Root>{/if}{/if}
  {#if inventory&&connector}
   <h2>{connector.name}</h2><p>{inventory.tenantName} / {inventory.name}</p><p>{t('web.PrintRotation.help')}</p>
   <form onsubmit={inspect} class="fields"><Label for="rotation-code">{t('web.PrintPairing.code')}</Label><Input id="rotation-code" bind:value={code} oninput={()=>{review=null;}} disabled={busy} maxlength={8} autocomplete="one-time-code" autocapitalize="characters" spellcheck={false} required/>{#if !review}<Button.Root type="submit" disabled={busy||code.trim().length!==8}>{t('web.PrintPairing.review')}</Button.Root>{/if}</form>
   {#if review}<details><summary>{t('web.PrintPairing.fingerprint')}</summary><p class="fingerprint">{review.fingerprint}</p></details><Button.Root disabled={busy} onclick={()=>void approve()}>{t('web.PrintRotation.approve')}</Button.Root>{/if}
  {:else if busy}<p role="status">{t('web.PrintPairing.loading')}</p>{:else}<Button.Root onclick={()=>void initialize()}>{t('web.PrintPairing.retry')}</Button.Root>{/if}
  <Button.Root href="/" variant="outline" class="rotation-cancel" disabled={busy}>{t('web.PrintPairing.cancel')}</Button.Root>
 {/if}
 </Card.Content>
</Card.Root></main>
<style>.rotation-shell{display:grid;min-height:100svh;place-items:center;padding:24px}:global(.rotation-card){width:100%;max-width:36rem;min-width:0}:global(.rotation-cancel){display:flex;width:fit-content;margin-top:16px}.fields{display:grid;gap:10px}p,details{margin-block:16px}h2,p{overflow-wrap:anywhere}.fingerprint{font-family:monospace}</style>
