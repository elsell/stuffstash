<script lang="ts">
  import {page} from '$app/state';
  import {goto} from '$app/navigation';
  import {LabelsClient} from '@stuff-stash/api-client';
  import {ApiLabelRepository} from '$lib/adapters/labels/ApiLabelRepository';
  import {labelDestinationHref} from '$lib/application/labels/labelNavigation';
  import {loadRuntimeConfig,type RuntimeConfig} from '$lib/runtimeConfig';
  import {getStoredSession,startSignIn,signOut} from '$lib/auth';
  import {isAuthenticationRequiredError} from '$lib/application/authenticationRequired';
  import type {LabelReference} from '$lib/domain/label';
  import {t} from '$lib/presentation/localization';
  import * as Button from '$lib/components/ui/button/index.js';
  let config=$state<RuntimeConfig|null>(null),reference=$state<LabelReference|null>(null),error=$state(''),busy=$state(true),signedOut=$state(false);
  let controller:AbortController|undefined;
  $effect(()=>{void initialize(page.url.href);return()=>controller?.abort();});
  async function initialize(source=page.url.href){
    controller?.abort();const current=new AbortController();controller=current;busy=true;error='';reference=null;signedOut=false;
    try{
      const loaded=await loadRuntimeConfig();if(current.signal.aborted)return;config=loaded;
      const repository=new ApiLabelRepository(new LabelsClient({baseUrl:loaded.apiBaseUrl,tokenProvider:()=>getStoredSession()?.idToken??null}));
      reference=repository.parse(source);signedOut=!getStoredSession();if(signedOut)return;
      const destination=await repository.resolve(reference,current.signal);if(!current.signal.aborted)await goto(labelDestinationHref(destination),{replaceState:true});
    }catch(cause){if(current.signal.aborted)return;
      if(isAuthenticationRequiredError(cause)||(typeof cause==='object'&&cause!==null&&'status' in cause&&cause.status===401)){signOut();signedOut=true;}
      else error=t(cause instanceof Error&&cause.message==='wrong_instance'?'labels.web.wrongInstance':'labels.web.unavailable');
    }finally{if(!current.signal.aborted)busy=false;}
  }
  async function signIn(){if(!config)return;try{await startSignIn(config);}catch{error=t('labels.web.unavailable');}}
</script>
<svelte:head><title>{t('labels.web.landing')}</title></svelte:head>
<main class="mx-auto grid max-w-lg gap-6 p-8">
  <h1 class="text-2xl font-semibold">{t('labels.web.landing')}</h1>
  {#if busy}<p role="status">{t('labels.web.resolving')}</p>{/if}
  {#if signedOut}<Button.Root onclick={()=>void signIn()}>{t('labels.web.signIn')}</Button.Root>{/if}
  {#if error}<p role="alert">{error}</p><Button.Root onclick={()=>void initialize()}>{t('labels.web.retry')}</Button.Root>{/if}
  {#if reference}<Button.Root variant="outline" href={`stuffstash://labels/v1/${reference.instanceId}/${reference.labelId}`}>{t('labels.web.app')}</Button.Root>{/if}
</main>
