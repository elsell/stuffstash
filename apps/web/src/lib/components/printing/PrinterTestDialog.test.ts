import {afterEach,expect,it} from 'vitest';
import {mount,tick,unmount} from 'svelte';
import InventoryPrintingSettings from './InventoryPrintingSettings.svelte';
import {FakePrintingRepository} from '$lib/fakes/printingRepository';
import {SessionPrintIntents} from '$lib/application/printing/manualPrint';
let component:ReturnType<typeof mount>|undefined;
afterEach(async()=>{if(component)await unmount(component);component=undefined;document.body.innerHTML='';});
async function settle(){for(let i=0;i<10;i++){await Promise.resolve();await tick();}}
function button(text:string){return [...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()===text);}
async function start(repository:FakePrintingRepository,canPrint=true){const intents=new SessionPrintIntents(repository,()=>crypto.randomUUID());component=mount(InventoryPrintingSettings,{target:document.body,props:{scope:repository.scope,repository,intents,canConfigure:false,canPrint}});await settle();}
it('prints only after explicit submission and resumes the same diagnostic request after dismissal',async()=>{
 const repository=new FakePrintingRepository();await start(repository);button('Print test label')!.click();await settle();expect(repository.queued.size).toBe(0);expect(document.body.textContent).toContain('does not create an inventory item');
 repository.loseNextJobResponse=true;button('Print one test label')!.click();await settle();expect(repository.queued.size).toBe(1);
 document.querySelector<HTMLButtonElement>('[data-slot="dialog-close"]')!.click();await settle();button('Print test label')!.click();await settle();button('Retry the same print request')!.click();await settle();expect(repository.queued.size).toBe(1);const job=[...repository.queued.values()][0];expect(job.assetId).toBeUndefined();expect(job.copies).toBe(1);expect(document.body.textContent).toContain('Print job queued');
});
it('withholds diagnostic printing for viewers and retired printers',async()=>{const repository=new FakePrintingRepository();await start(repository,false);expect(button('Print test label')).toBeUndefined();await unmount(component!);component=undefined;repository.destinations[0].retired=true;await start(repository,true);expect(button('Print test label')).toBeUndefined();expect(repository.queued.size).toBe(0);});
