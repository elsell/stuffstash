import {afterEach,expect,it} from 'vitest';
import {mount,tick,unmount} from 'svelte';
import LabelScanner from './LabelScanner.svelte';
import type {LabelRepository,LabelWorkspace} from '$lib/ports/labels';
import type {LabelDestination} from '$lib/domain/label';
import {parseLabelLink} from '@stuff-stash/api-client';
const code='https://old.example/l/v1/01ARZ3NDEKTSV4RRFFQ69G5FAV/01ARZ3NDEKTSV4RRFFQ69G5FAW';
let component:ReturnType<typeof mount>|undefined;
afterEach(async()=>{if(component)await unmount(component);component=undefined;document.body.innerHTML='';});
class LabelRepositoryFake implements LabelRepository {
  requests=0; authorized=true; release:((value:LabelDestination)=>void)|undefined;
  parse=parseLabelLink;
  async resolve(){this.requests++;if(!this.authorized)throw new Error('denied');return new Promise<LabelDestination>(resolve=>{this.release=resolve;});}
  async catalog(){return {profiles:[],templates:[]};}
  async render():Promise<never>{throw new Error('unavailable');}
}
function setup(repository:LabelRepositoryFake,navigate:(value:LabelDestination)=>void){
  const workspace:LabelWorkspace={repository,camera:{async start(){throw new Error('denied');}},files:{save(){},preparePrint(){return {show(){},close(){}};}}};
  component=mount(LabelScanner,{target:document.body,props:{workspace,onResolved:navigate}});
}
async function submit(value:string){const input=document.querySelector('input')!;input.value=value;input.dispatchEvent(new Event('input',{bubbles:true}));await tick();document.querySelector('form')!.dispatchEvent(new Event('submit',{bubbles:true,cancelable:true}));await tick();}
it('rejects foreign executable data before resolution and preserves paste after camera denial',async()=>{
  const repository=new LabelRepositoryFake();const navigation:LabelDestination[]=[];setup(repository,value=>navigation.push(value));
  await submit('javascript:alert(1)');expect(repository.requests).toBe(0);expect(document.body.textContent).toContain('not a supported');
  (Array.from(document.querySelectorAll('button')).find(button=>button.textContent?.includes('Use camera'))!).click();await tick();await tick();
  expect(document.body.textContent).toContain('Camera unavailable');expect(document.querySelector('input')).not.toBeNull();expect(navigation).toEqual([]);
});
it('suppresses duplicate submissions and late resolution after cancellation',async()=>{
  const repository=new LabelRepositoryFake();const navigation:LabelDestination[]=[];setup(repository,value=>navigation.push(value));
  await submit(code);await submit(code);expect(repository.requests).toBe(1);
  await unmount(component!);component=undefined;repository.release!({tenantId:'t',inventoryId:'i',assetId:'a',archived:true});await tick();await tick();expect(navigation).toEqual([]);
});
it('does not navigate after permission loss',async()=>{
  const repository=new LabelRepositoryFake();repository.authorized=false;const navigation:LabelDestination[]=[];setup(repository,value=>navigation.push(value));
  await submit(code);await tick();expect(document.body.textContent).toContain('Label unavailable');expect(navigation).toEqual([]);
});
it('navigates authorized archived labels using resolved scope',async()=>{
  const repository=new LabelRepositoryFake();const navigation:LabelDestination[]=[];setup(repository,value=>navigation.push(value));await submit(code);
  const destination={tenantId:'t2',inventoryId:'i2',assetId:'archived',archived:true};repository.release!(destination);await tick();await tick();expect(navigation).toEqual([destination]);
});
