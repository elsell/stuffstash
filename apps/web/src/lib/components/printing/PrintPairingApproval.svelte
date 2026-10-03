<script lang="ts">
  import { onMount, tick } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import PairingChoice from './PairingChoice.svelte';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Input } from '$lib/components/ui/input';
  import AuthBrand from '$lib/components/auth/AuthBrand.svelte';
  import { t } from '$lib/presentation/localization';
  import type { PrintPairingRepository } from '$lib/ports/printPairingRepository';
  import { PairingFailure, type PairingInventory, type PairingReview, type PairingSetup, type PairingSelection, type PairingFailureKind } from '$lib/domain/printPairing';
  import { approvePairingSelection } from '$lib/application/printing/pairingApproval';

  let { pairingId, repository, onSignIn }: { pairingId:string; repository:PrintPairingRepository; onSignIn:()=>Promise<void> } = $props();
  let inventories=$state<PairingInventory[]>([]);
  let inventoryIndex=$state('');
  let code=$state('');
  let review=$state<PairingReview|null>(null);
  let setup=$state<PairingSetup>({printers:[],media:[]});
  let selections=$state<PairingSelection[]>([]);
  let busy=$state(true);
  let initialized=$state(false);
  let failure=$state<PairingFailureKind|null>(null);
  let success=$state(false);
  let reviewHeading=$state<HTMLHeadingElement>();
  const scope=$derived(inventories[Number(inventoryIndex)]);
  const canReview=$derived(inventoryIndex!==''&&Boolean(scope)&&code.trim().length===8);
  const hasSelection=$derived(selections.some(s=>s.destination!==''));
  const errorKey=$derived(failure==='authentication_required'?'web.PrintPairing.authRequired':failure==='denied'?'web.PrintPairing.denied':failure==='invalid'?'web.PrintPairing.invalid':'web.PrintPairing.unavailable');

  onMount(()=>{void loadInventories();});
  async function loadInventories(){busy=true;failure=null;try{inventories=await repository.inventories();initialized=true;}catch(error){handle(error);}finally{busy=false;}}
  function handle(error:unknown){failure=error instanceof PairingFailure?error.kind:'unavailable';}
  function clearReview(){if(busy)return;review=null;selections=[];setup={printers:[],media:[]};failure=null;}
  async function loadReview(event:SubmitEvent){
    event.preventDefault();if(busy||!canReview)return;busy=true;failure=null;review=null;
    try{
      const reviewed=await repository.review(pairingId,scope,code.trim().toUpperCase());
      if(reviewed.id!==pairingId||reviewed.rotation===true)throw new PairingFailure('invalid');
      setup=await repository.setup(scope);review=reviewed;
      selections=reviewed.candidates.map(c=>({candidateId:c.id,destination:'',name:c.name,mediaKey:'',idempotencyKey:crypto.randomUUID()}));
      await tick();reviewHeading?.focus();
    }catch(error){handle(error);}finally{busy=false;}
  }
  async function approve(event:SubmitEvent){
    event.preventDefault();if(busy||!review||!canReview)return;busy=true;failure=null;
    try{await approvePairingSelection(repository,pairingId,scope,code.trim().toUpperCase(),review,setup,selections);success=true;code='';}
    catch(error){handle(error);}finally{busy=false;}
  }
</script>

<main class="pairing-shell">
 <Card.Root class="pairing-card" aria-busy={busy}>
  <Card.Header><AuthBrand/><Card.Title role="heading" aria-level={1}>{t(success?'web.PrintPairing.success':'web.PrintPairing.title')}</Card.Title></Card.Header>
  <Card.Content>
   {#if success}
    <p role="status">{t('web.PrintPairing.successHelp')}</p>
    <Button.Root href={`/tenants/${encodeURIComponent(scope.tenantId)}/inventories/${encodeURIComponent(scope.inventoryId)}`}>{t('web.PrintPairing.openInventory')}</Button.Root>
   {:else}
    {#if failure}<p role="alert">{t(errorKey)}</p>{/if}
    {#if failure==='authentication_required'}<Button.Root onclick={()=>void onSignIn()}>{t('web.PrintPairing.continueSignIn')}</Button.Root>{/if}
    {#if !initialized}
      {#if busy}<p role="status">{t('web.PrintPairing.loading')}</p>{:else}<Button.Root onclick={()=>void loadInventories()}>{t('web.PrintPairing.retry')}</Button.Root>{/if}
    {:else if inventories.length===0}<p>{t('web.PrintPairing.empty')}</p>
    {:else}
     <form onsubmit={loadReview} class="fields">
      <PairingChoice id="pair-inventory" label={t('web.PrintPairing.inventory')} value={inventoryIndex} disabled={busy} options={[{value:'',label:t('web.PrintPairing.chooseInventory')},...inventories.map((inventory,index)=>({value:String(index),label:`${inventory.tenantName} / ${inventory.name}`}))]} onChange={value=>{inventoryIndex=value;clearReview();}}/>
      <Label for="pair-code">{t('web.PrintPairing.code')}</Label>
      <Input id="pair-code" bind:value={code} oninput={clearReview} disabled={busy} maxlength={8} autocomplete="one-time-code" autocapitalize="characters" spellcheck={false} required/>
      {#if !review}<Button.Root type="submit" disabled={busy||!canReview}>{t(busy?'web.PrintPairing.busy':'web.PrintPairing.review')}</Button.Root>{/if}
     </form>
     {#if review}
      <form onsubmit={approve} class="review fields">
       <h2 tabindex="-1" bind:this={reviewHeading}>{review.name}</h2>
       <details><summary>{t('web.PrintPairing.fingerprint')}</summary><p class="fingerprint">{review.fingerprint}</p><p>{t('web.PrintPairing.fingerprintHelp')}</p></details>
       <p>{t('web.PrintPairing.selectionHelp')}</p>
       {#each review.candidates as candidate,index}
        <fieldset disabled={busy||Boolean(selections[index].registrationStarted)}>
         <legend>{candidate.name}</legend>
         <PairingChoice id={`destination-${index}`} label={t('web.PrintPairing.destination')} value={selections[index].destination} disabled={busy||Boolean(selections[index].registrationStarted)} options={[{value:'',label:t('web.PrintPairing.skip')},{value:'new',label:t('web.PrintPairing.create')},...setup.printers.filter(p=>p.adapterId===candidate.adapterId).map(p=>({value:p.id,label:`${p.name} — ${p.mediaName}`}))]} onChange={value=>{selections[index].destination=value;}}/>
         {#if selections[index].destination==='new'}
          <Label for={`name-${index}`}>{t('web.PrintPairing.name')}</Label><Input id={`name-${index}`} bind:value={selections[index].name} maxlength={100} required/>
          <PairingChoice id={`media-${index}`} label={t('web.PrintPairing.media')} value={selections[index].mediaKey} disabled={busy||Boolean(selections[index].registrationStarted)} options={[{value:'',label:t('web.PrintPairing.chooseMedia')},...setup.media.filter(m=>m.adapterId===candidate.adapterId).map(m=>({value:m.key,label:m.name}))]} onChange={value=>{selections[index].mediaKey=value;}}/>
         {/if}
        </fieldset>
        {#if selections[index].createdPrinterId}<p role="status">{t('web.PrintPairing.savedDestination')}</p>{/if}
       {/each}
       <Button.Root type="submit" disabled={busy||!hasSelection}>{t(busy?'web.PrintPairing.busy':'web.PrintPairing.approve')}</Button.Root>
      </form>
     {/if}
    {/if}
    <Button.Root href="/" variant="outline" class="cancel" disabled={busy}>{t('web.PrintPairing.cancel')}</Button.Root>
   {/if}
  </Card.Content>
 </Card.Root>
</main>
<style>
 .pairing-shell{display:grid;min-height:100svh;place-items:center;padding:24px;background:var(--background)}
 :global(.pairing-card){width:100%;max-width:36rem;min-width:0}
 .fields{display:grid;gap:10px}.review{margin-top:24px}legend{font-weight:600}fieldset{display:grid;gap:10px;border:1px solid var(--border);padding:16px;border-radius:var(--radius-control);min-width:0}.fingerprint,h2,legend{overflow-wrap:anywhere}.fingerprint{font-family:monospace}p{margin-block:12px}:global(.cancel){margin-top:20px}h2{font-size:var(--text-section-size);font-weight:600}
</style>
