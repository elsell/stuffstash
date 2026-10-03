import {afterEach,expect,it} from 'vitest';
import {flushSync,mount,tick,unmount} from 'svelte';
import Harness from './AssetLabelMenu.test-harness.svelte';
import {FakePrintingRepository} from '$lib/fakes/printingRepository';
import {SessionPrintIntents} from '$lib/application/printing/manualPrint';
import type {LabelRepository,LabelWorkspace} from '$lib/ports/labels';
import type {LabelArtifact,LabelScope} from '$lib/domain/label';
class LabelRepositoryFake implements LabelRepository {
 readonly reads:AbortSignal[]=[];
 parse(){return {instanceId:'instance',labelId:'label'};}
 async resolve(){return {tenantId:'tenant',inventoryId:'inventory',assetId:'asset',archived:false};}
 async catalog(scope:LabelScope){if(scope.inventoryId!=='inventory')throw new Error('Unknown inventory');return {profiles:[{id:'profile',name:'29 × 90 mm'}],templates:[{id:'qr-title',version:1,name:'QR with title'}]};}
 async render(_scope:LabelScope,_choice:unknown,_format:unknown,signal:AbortSignal):Promise<LabelArtifact>{this.reads.push(signal);return new Promise((_,reject)=>{signal.addEventListener('abort',()=>reject(new DOMException('Aborted','AbortError')),{once:true});});}
}
let component:ReturnType<typeof mount>|undefined;
afterEach(async()=>{if(component)await unmount(component);component=undefined;document.body.innerHTML='';});
async function settle(){for(let i=0;i<6;i++){await Promise.resolve();await tick();}}
async function start(){const repository=new LabelRepositoryFake();const printing=new FakePrintingRepository();const workspace:LabelWorkspace={repository,files:{save(){throw new Error('No completed render');},preparePrint(){throw new Error('No printer in fixture');}},camera:{async start(){throw new Error('No camera in fixture');}}};component=mount(Harness,{target:document.body,props:{labelWorkspace:workspace,printingWorkspace:{apiIdentity:'fixture',repository:printing,intents:new SessionPrintIntents(printing,()=> 'intent')}}});await settle();return repository;}
function button(label:string){return [...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()===label)!;}
function more(){return document.querySelector<HTMLButtonElement>('button[aria-label="More actions"]')!;}
async function openMenu(){more().focus();more().dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));flushSync();await settle();}
async function select(label:string){const item=[...document.querySelectorAll<HTMLElement>('[role="menuitem"]')].find(item=>item.textContent?.trim()===label)!;item.focus();item.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));flushSync();await settle();}
it('has one contextual menu with viewer downloads and editor print/status commands',async()=>{await start();expect(document.querySelectorAll('button[aria-label="More actions"]')).toHaveLength(1);await openMenu();expect([...document.querySelectorAll('[role="menuitem"]')].map(n=>n.textContent?.trim())).toEqual(['Label options','Print label','View label job']);button('Viewer').click();await settle();expect([...document.querySelectorAll('[role="menuitem"]')].map(n=>n.textContent?.trim())).toEqual(['Label options']);});
it('aborts the label task on scope teardown and disables the single menu during saving',async()=>{const repository=await start();await openMenu();await select('Label options');expect(repository.reads).toHaveLength(1);button('Change asset').click();await settle();expect(repository.reads[0].aborted).toBe(true);expect(document.querySelector('[role="dialog"]')).toBeNull();button('Saving').click();await settle();expect(more().disabled).toBe(true);});
it('closes downloads when their workspace is removed and print tasks when edit permission is lost',async()=>{const repository=await start();await openMenu();await select('Label options');button('Clear labels').click();await settle();expect(repository.reads[0].aborted).toBe(true);await openMenu();await select('Print label');expect(document.querySelector('[role="dialog"]')).not.toBeNull();button('Viewer').click();await settle();expect(document.querySelector('[role="dialog"]')).toBeNull();expect(document.querySelector('button[aria-label="More actions"]')).toBeNull();});
