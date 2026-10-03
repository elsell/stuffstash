<script lang="ts">
 import { onMount } from 'svelte';
 import { page } from '$app/state';
 import { getStoredSession, startSignIn } from '$lib/auth';
 import { loadRuntimeConfig, type RuntimeConfig } from '$lib/runtimeConfig';
 import { ApiPrintPairingRepository } from '$lib/adapters/api/printPairingRepository';
 import type { PrintPairingRepository } from '$lib/ports/printPairingRepository';
 import PrintPairingApproval from '$lib/components/printing/PrintPairingApproval.svelte';
 import AuthBrand from '$lib/components/auth/AuthBrand.svelte';
 import * as Button from '$lib/components/ui/button/index.js';
 import * as Card from '$lib/components/ui/card/index.js';
 import { t } from '$lib/presentation/localization';

 let config=$state<RuntimeConfig|null>(null);
 let repository=$state<PrintPairingRepository|null>(null);
 let loading=$state(true);
 let failed=$state(false);
 let signingIn=$state(false);
 const pairingId=$derived(page.params.pairingId??'');
 onMount(()=>{void initialize();});
 async function initialize(){
  loading=true;failed=false;
  try{config=await loadRuntimeConfig();if(getStoredSession())repository=new ApiPrintPairingRepository(config.apiBaseUrl,()=>getStoredSession()?.idToken??null);}
  catch{failed=true;}finally{loading=false;}
 }
 async function signIn(){
  if(!config||signingIn)return;signingIn=true;
  try{await startSignIn(config,window.location,window.sessionStorage,window.history,`/print-connectors/pair/${encodeURIComponent(pairingId)}`);}
  catch{failed=true;}finally{signingIn=false;}
 }
</script>
<svelte:head><title>{t('web.PrintPairing.pageTitle')}</title></svelte:head>
{#if repository}
 {#key pairingId}<PrintPairingApproval {pairingId} {repository} onSignIn={signIn}/>{/key}
{:else}
 <main class="pairing-entry"><Card.Root class="pairing-entry-card">
  <Card.Header><AuthBrand/><Card.Title role="heading" aria-level={1}>{t('web.PrintPairing.title')}</Card.Title></Card.Header>
  <Card.Content>
   {#if loading}<p role="status">{t('web.PrintPairing.loading')}</p>
   {:else if failed}<p role="alert">{t('web.PrintPairing.unavailable')}</p><Button.Root onclick={()=>void initialize()}>{t('web.PrintPairing.retry')}</Button.Root>
   {:else}<p>{t('web.PrintPairing.signIn')}</p><Button.Root disabled={signingIn} onclick={()=>void signIn()}>{t('web.PrintPairing.continueSignIn')}</Button.Root>{/if}
  </Card.Content>
 </Card.Root></main>
{/if}
<style>.pairing-entry{display:grid;min-height:100svh;place-items:center;padding:24px}:global(.pairing-entry-card){width:100%;max-width:36rem}p{margin-bottom:16px}</style>
