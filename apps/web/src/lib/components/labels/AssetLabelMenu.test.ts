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
async function start(printing=new FakePrintingRepository()){const repository=new LabelRepositoryFake();const workspace:LabelWorkspace={repository,files:{save(){throw new Error('No completed render');},preparePrint(){throw new Error('No printer in fixture');}},camera:{async start(){throw new Error('No camera in fixture');}}};component=mount(Harness,{target:document.body,props:{labelWorkspace:workspace,printingWorkspace:{apiIdentity:'fixture',repository:printing,intents:new SessionPrintIntents(printing,()=> 'intent')}}});await settle();return repository;}
function button(label:string){return [...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()===label)!;}
function more(){return document.querySelector<HTMLButtonElement>('button[aria-label="More actions"]')!;}
async function openMenu(){more().focus();more().dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));flushSync();await settle();}
async function select(label:string){const item=[...document.querySelectorAll<HTMLElement>('[role="menuitem"]')].find(item=>item.textContent?.trim()===label)!;item.focus();item.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));flushSync();await settle();}
it('has one contextual menu with viewer downloads and editor print/status commands',async()=>{await start();expect(document.querySelectorAll('button[aria-label="More actions"]')).toHaveLength(1);await openMenu();expect([...document.querySelectorAll('[role="menuitem"]')].map(n=>n.textContent?.trim())).toEqual(['Label options','Print label','View label job']);button('Viewer').click();await settle();expect([...document.querySelectorAll('[role="menuitem"]')].map(n=>n.textContent?.trim())).toEqual(['Label options']);});
it('aborts the label task on scope teardown and disables the single menu during saving',async()=>{const repository=await start();await openMenu();await select('Label options');expect(repository.reads).toHaveLength(1);button('Change asset').click();await settle();expect(repository.reads[0].aborted).toBe(true);expect(document.querySelector('[role="dialog"]')).toBeNull();button('Saving').click();await settle();expect(more().disabled).toBe(true);});
it('closes downloads when their workspace is removed and print tasks when edit permission is lost',async()=>{const repository=await start();await openMenu();await select('Label options');button('Clear labels').click();await settle();expect(repository.reads[0].aborted).toBe(true);await openMenu();await select('Print label');expect(document.querySelector('[role="dialog"]')).not.toBeNull();button('Viewer').click();await settle();expect(document.querySelector('[role="dialog"]')).toBeNull();expect(document.querySelector('button[aria-label="More actions"]')).toBeNull();});

it('prints one default copy without preview and reopens a lost request without another job',async()=>{
 const printing=new FakePrintingRepository();printing.loseNextJobResponse=true;await start(printing);
 await openMenu();await select('Print label');expect(printing.queued.size).toBe(1);expect([...printing.queued.values()][0].copies).toBe(1);
 expect(button('Preview label')).toBeUndefined();expect(button('Retry the same print request').disabled).toBe(false);
 document.querySelector<HTMLButtonElement>('[data-slot="dialog-close"]')!.click();await settle();
 await openMenu();await select('Print label');expect(printing.queued.size).toBe(1);
 button('Retry the same print request').click();await settle();expect(printing.queued.size).toBe(1);expect(document.body.textContent).toContain('Queued');
});
it('opens explicit printer options without printing and requires a preview before submission',async()=>{
 const printing=new FakePrintingRepository();await start(printing);await openMenu();await select('Label options');
 button('Printer options').click();await settle();expect(printing.queued.size).toBe(0);expect(document.querySelectorAll('[role="dialog"]')).toHaveLength(1);
 expect(button('Print label').disabled).toBe(true);button('Preview label').click();await settle();expect(button('Print label').disabled).toBe(false);
 const copies=document.querySelector<HTMLInputElement>('#label-copies')!;copies.value='2';copies.dispatchEvent(new Event('input',{bubbles:true}));await settle();expect(button('Print label').disabled).toBe(true);
 button('Preview label').click();await settle();button('Print label').click();await settle();expect(printing.queued.size).toBe(1);expect([...printing.queued.values()][0].copies).toBe(2);
});
it.each(['missing','retired','template'])('falls back to options for a %s default without silently printing',async(kind)=>{
 const printing=new FakePrintingRepository();if(kind==='missing')printing.defaults.defaultPrinterId=null;if(kind==='retired')printing.destinations[0].retired=true;if(kind==='template')printing.defaults.templateId='unavailable';
 await start(printing);await openMenu();await select('Print label');expect(printing.queued.size).toBe(0);expect(button('Preview label')).toBeDefined();
});
class DelayedSettingsRepository extends FakePrintingRepository {
 private releaseRead: (()=>void)|undefined;
 async settings(scope:Parameters<FakePrintingRepository['settings']>[0]){await new Promise<void>(resolve=>{this.releaseRead=resolve;});return super.settings(scope);}
 release(){this.releaseRead?.();}
}
it.each(['Viewer','Change asset','Saving'])('never starts a deferred default print after %s',async(action)=>{
 const printing=new DelayedSettingsRepository();await start(printing);await openMenu();await select('Print label');button(action).click();await settle();printing.release();await settle();expect(printing.queued.size).toBe(0);expect(document.querySelector('[role="dialog"]')).toBeNull();
});
