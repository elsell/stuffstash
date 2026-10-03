import {afterEach,expect,it} from 'vitest';
import {flushSync,mount,tick,unmount} from 'svelte';
import LabelOptions from './LabelOptions.svelte';
import type {LabelRepository,LabelWorkspace} from '$lib/ports/labels';
import type {LabelArtifact,LabelChoice,LabelScope} from '$lib/domain/label';
const scope={tenantId:'tenant',inventoryId:'inventory',assetId:'asset'};
class LabelRepositoryFake implements LabelRepository {
  attempts:{choice:LabelChoice;signal:AbortSignal;complete:(value:LabelArtifact)=>void}[]=[];
  parse(){return {instanceId:'instance',labelId:'label'};}
  async resolve(){return {...scope,archived:false};}
  async catalog(){return {profiles:[{id:'profile',name:'29 × 90 mm'}],templates:[{id:'qr-title',version:1,name:'QR with title'},{id:'qr-only',version:1,name:'QR only'}]};}
  async render(_scope:LabelScope,choice:LabelChoice,_format:'png'|'pdf',signal:AbortSignal){return new Promise<LabelArtifact>(complete=>{this.attempts.push({choice,signal,complete});});}
}
let component:ReturnType<typeof mount>|undefined;
const createURL=URL.createObjectURL,revokeURL=URL.revokeObjectURL;
afterEach(async()=>{if(component)await unmount(component);component=undefined;document.body.innerHTML='';URL.createObjectURL=createURL;URL.revokeObjectURL=revokeURL;});
async function settle(){await tick();await tick();await tick();}
it('invalidates stale previews and cancels artifact delivery after leaving the task',async()=>{
  let sequence=0;const revoked:string[]=[];URL.createObjectURL=()=>`blob:label-${++sequence}`;URL.revokeObjectURL=value=>{revoked.push(value);};
  const repository=new LabelRepositoryFake();const saved:Blob[]=[];
  const workspace:LabelWorkspace={repository,camera:{async start(){}},files:{save(content){saved.push(content);},preparePrint(){return {show(){},close(){}};}}};
  component=mount(LabelOptions,{target:document.body,props:{workspace,scope}});await settle();
  const layout=document.querySelector<HTMLButtonElement>('#label-layout')!;
  layout.focus();layout.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));flushSync();
  await expect.poll(()=>layout.getAttribute('aria-expanded')).toBe('true');
  const option=()=>Array.from(document.querySelectorAll<HTMLElement>('[role="option"]')).find(item=>item.textContent?.trim()==='QR only');
  await expect.poll(()=>Boolean(option())).toBe(true);
  option()!.dispatchEvent(new PointerEvent('pointerup',{pointerType:'mouse',bubbles:true,cancelable:true}));flushSync();await settle();
  expect(repository.attempts[0].signal.aborted).toBe(true);
  const artifact={content:new Blob(['png'],{type:'image/png'}),displayRotation:270,width:306,height:991};
  repository.attempts[0].complete(artifact);await settle();expect(document.querySelector('img')).toBeNull();
  repository.attempts[1].complete(artifact);await settle();expect(document.querySelector('img')?.getAttribute('src')).toBe('blob:label-1');
  Array.from(document.querySelectorAll('button')).find(button=>button.textContent==='Download PNG')!.click();await settle();
  await unmount(component);component=undefined;repository.attempts[2].complete(artifact);await settle();
  expect(saved).toEqual([]);expect(revoked).toContain('blob:label-1');
});
