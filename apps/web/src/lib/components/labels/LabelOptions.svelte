<script lang="ts">
  import {onMount} from 'svelte';
  import {t} from '$lib/presentation/localization';
  import type {LabelWorkspace} from '$lib/ports/labels';
  import type {LabelScope,LabelCatalog,LabelChoice,LabelArtifact} from '$lib/domain/label';
  import * as Button from '$lib/components/ui/button/index.js';
  let {workspace,scope} : {workspace:LabelWorkspace;scope:LabelScope}=$props();
  let catalog=$state<LabelCatalog|null>(null),choice=$state<LabelChoice|null>(null),artifact=$state<LabelArtifact|null>(null);
  let error=$state(''),busy=$state(false),preview=$state('');
  let controller:AbortController|undefined;
  const previewTransform=$derived(`rotate(${artifact?.displayRotation??0}deg)`);
  onMount(()=>{void initialize();return()=>{controller?.abort();if(preview)URL.revokeObjectURL(preview);};});
  async function initialize(){
    controller?.abort();const current=new AbortController();controller=current;busy=true;error='';
    try{catalog=await workspace.repository.catalog(scope,current.signal);if(current.signal.aborted)return;
      const profile=catalog.profiles[0],template=catalog.templates.find(v=>v.id==='qr-title')??catalog.templates[0];
      if(!profile||!template){error=t('labels.web.noMedia');return;}
      choice={profileId:profile.id,templateId:template.id,templateVersion:template.version,showReference:true};
      await render('png');
    }catch{if(!current.signal.aborted)error=t('labels.web.unavailable');}finally{if(controller===current)busy=false;}
  }
  async function render(format:'png'|'pdf',delivery?:'download'|'print'){
    if(!choice)return;controller?.abort();const current=new AbortController();controller=current;
    let printTarget:ReturnType<LabelWorkspace['files']['preparePrint']>|undefined;
    try{if(delivery==='print')printTarget=workspace.files.preparePrint();}catch{error=t('labels.web.unavailable');return;}
    busy=true;error='';artifact=null;if(preview){URL.revokeObjectURL(preview);preview='';}
    try{const result=await workspace.repository.render(scope,{...choice},format,current.signal);if(current.signal.aborted){printTarget?.close();return;}
      if(delivery==='download')workspace.files.save(result.content,format);
      else if(delivery==='print')printTarget?.show(result.content);
      else {artifact=result;preview=URL.createObjectURL(result.content);}
    }catch{printTarget?.close();if(!current.signal.aborted)error=t('labels.web.unavailable');}finally{if(controller===current)busy=false;}
  }
  function selectTemplate(value:string){if(!choice||!catalog)return;const template=catalog.templates.find(v=>v.id===value);if(template){choice={...choice,templateId:template.id,templateVersion:template.version};void render('png');}}
</script>
{#if catalog && choice}
  <label>{t('labels.web.size')}<select value={choice.profileId} onchange={event=>{if(choice){choice={...choice,profileId:event.currentTarget.value};void render('png');}}} disabled={catalog.profiles.length===1}>{#each catalog.profiles as profile}<option value={profile.id}>{profile.name}</option>{/each}</select></label>
  <label>{t('labels.web.layout')}<select value={choice.templateId} onchange={event=>selectTemplate(event.currentTarget.value)}>{#each catalog.templates as template}<option value={template.id}>{template.name}</option>{/each}</select></label>
  <label><input type="checkbox" checked={choice.showReference} onchange={event=>{if(choice){choice={...choice,showReference:event.currentTarget.checked};void render('png');}}}/>{t('labels.web.reference')}</label>
{/if}
{#if busy}<p role="status">{t('labels.web.loading')}</p>{/if}
{#if error}<p role="alert">{error}</p><Button.Root onclick={()=>void initialize()}>{t('labels.web.retry')}</Button.Root>{/if}
{#if artifact && preview}<div class="label-preview"><img src={preview} alt={t('labels.web.preview')} style:transform={previewTransform}/></div>{/if}
<div class="flex flex-wrap gap-2">
  <Button.Root disabled={!choice||busy} onclick={()=>void render('png','download')}>{t('labels.web.png')}</Button.Root>
  <Button.Root variant="outline" disabled={!choice||busy} onclick={()=>void render('pdf','download')}>{t('labels.web.pdf')}</Button.Root>
  <Button.Root variant="outline" disabled={!choice||busy} onclick={()=>void render('pdf','print')}>{t('labels.web.print')}</Button.Root>
</div><p class="text-sm text-muted-foreground">{t('labels.web.actualSize')}</p>
<style>.label-preview{display:grid;place-items:center;overflow:hidden;height:180px;background:white;border:1px solid #ddd}.label-preview img{height:300px;max-width:100%;image-rendering:pixelated}select{display:block;width:100%;padding:.5rem;border:1px solid currentColor;border-radius:.4rem}</style>
