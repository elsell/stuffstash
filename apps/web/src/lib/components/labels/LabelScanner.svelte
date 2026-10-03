<script lang="ts">
  import {onDestroy,onMount} from 'svelte';
  import {t} from '$lib/presentation/localization';
  import type {LabelWorkspace} from '$lib/ports/labels';
  import type {LabelDestination} from '$lib/domain/label';
  import * as Button from '$lib/components/ui/button/index.js';
  import {Input} from '$lib/components/ui/input/index.js';
  let {workspace,onResolved}:{workspace:LabelWorkspace;onResolved:(destination:LabelDestination)=>void}=$props();
  let signInHref=$state('');
  let value=$state(''),error=$state(''),busy=$state(false),video:HTMLVideoElement;
  let camera:AbortController|undefined,resolve:AbortController|undefined;
  onMount(()=>{void startCamera();});
  onDestroy(()=>{camera?.abort();resolve?.abort();});
  async function open(code:string){
    if(busy)return;error='';signInHref='';let reference;
    try{reference=workspace.repository.parse(code);}catch{error=t('labels.web.invalid');return;}
    busy=true;camera?.abort();const current=new AbortController();resolve=current;
    try{const destination=await workspace.repository.resolve(reference,current.signal);if(!current.signal.aborted)onResolved(destination);}
    catch(cause){if(!current.signal.aborted){error=t(cause instanceof Error&&cause.message==='wrong_instance'?'labels.web.wrongInstance':'labels.web.unavailable');if(typeof cause==='object'&&cause!==null&&'status' in cause&&cause.status===401)signInHref=`/l/v1/${reference.instanceId}/${reference.labelId}`;}}
    finally{if(!current.signal.aborted)busy=false;}
  }
  async function startCamera(){
    camera?.abort();const current=new AbortController();camera=current;error='';
    try{await workspace.camera.start(video,code=>void open(code),current.signal);}catch{if(!current.signal.aborted&&!error)error=t('labels.web.cameraDenied');}
  }
</script>
<video bind:this={video!} muted playsinline aria-label={t('labels.web.scan')} class="w-full rounded-md"></video>
<Button.Root variant="outline" disabled={busy} onclick={()=>void startCamera()}>{t('labels.web.camera')}</Button.Root>
<form onsubmit={event=>{event.preventDefault();void open(value);}} class="grid gap-3">
  <label for="label-link">{t('labels.web.paste')}</label><Input id="label-link" bind:value autocomplete="off" autocapitalize="off" spellcheck="false" maxlength={4096}/>
  <Button.Root type="submit" disabled={busy||!value}>{t('labels.web.open')}</Button.Root>
</form>
{#if busy}<p role="status">{t('labels.web.resolving')}</p>{/if}
{#if error}<p role="alert">{error}</p>{/if}

{#if signInHref}<Button.Root href={signInHref}>{t('labels.web.signIn')}</Button.Root>{/if}
