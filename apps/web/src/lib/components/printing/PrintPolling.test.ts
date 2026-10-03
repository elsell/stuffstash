import {afterEach,expect,it} from 'vitest';
import {mount,tick,unmount} from 'svelte';
import InventoryPrintingSettings from './InventoryPrintingSettings.svelte';
import AssetPrintDialog from './AssetPrintDialog.svelte';
import {SessionPrintIntents} from '$lib/application/printing/manualPrint';
import {FakePrintingRepository} from '$lib/fakes/printingRepository';
import {FakePrintPollingRuntime} from '$lib/fakes/printPollingRuntime';
let component:ReturnType<typeof mount>|undefined;
afterEach(async()=>{if(component)await unmount(component);component=undefined;document.body.innerHTML='';});
async function settle(){for(let i=0;i<14;i++){await Promise.resolve();await tick();}}
class JobRepository extends FakePrintingRepository {
 reads=0;
 override async job(...args:Parameters<FakePrintingRepository['job']>){this.reads++;return super.job(...args);}
}
function fixture(){const repository=new JobRepository();repository.queued.set('job',{id:'job',assetId:'asset',printerId:'printer',status:'queued',revision:1,copies:1,completedCopies:0,reason:'',createdAt:'2026-10-03T12:00:00Z'});return repository;}
it('updates a visible job automatically and stops at terminal state and teardown',async()=>{
 const repository=fixture(),runtime=new FakePrintPollingRuntime(),intents=new SessionPrintIntents(repository,()=> 'intent');
 component=mount(AssetPrintDialog,{target:document.body,props:{scope:repository.scope,assetId:'asset',initialJobId:'job',repository,intents,pollingRuntime:runtime,onClose:()=>{}}});await settle();expect(repository.reads).toBe(1);
 Object.assign(repository.queued.get('job')!,{status:'completed',revision:2,completedCopies:1});runtime.advance(2000);await settle();expect(repository.reads).toBe(2);expect(document.body.textContent).toContain('Printed');runtime.advance(60000);await settle();expect(repository.reads).toBe(2);expect(runtime.listenerCount).toBe(0);
});
it('refreshes status without replacing an unsaved inventory default and pauses while hidden',async()=>{
 const repository=fixture(),runtime=new FakePrintPollingRuntime();component=mount(InventoryPrintingSettings,{target:document.body,props:{scope:repository.scope,repository,canConfigure:true,canPrint:true,pollingRuntime:runtime}});await settle();
 const input=document.getElementById('default-auto-print') as HTMLInputElement;input.click();await settle();expect(input.checked).toBe(true);
 runtime.setVisible(false);Object.assign(repository.queued.get('job')!,{status:'printing',revision:2});runtime.advance(60000);await settle();expect(document.body.textContent).not.toContain('Printing label');runtime.setVisible(true);await settle();expect(input.checked).toBe(true);expect(document.body.textContent).toContain('Printing');expect(repository.defaults.printOnCreateDefault).toBe(false);
 await unmount(component);component=undefined;expect(runtime.listenerCount).toBe(0);
});
