import {afterEach,expect,it} from 'vitest';
import {mount,tick,unmount} from 'svelte';
import InventoryPrintingSettings from './InventoryPrintingSettings.svelte';
import {FakePrintingRepository} from '$lib/fakes/printingRepository';
let component:ReturnType<typeof mount>|undefined;
afterEach(async()=>{if(component)await unmount(component);component=undefined;document.body.innerHTML='';});
async function settle(){for(let i=0;i<10;i++){await Promise.resolve();await tick();}}
function button(text:string){return [...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()===text);}
async function start(repository:FakePrintingRepository,canConfigure=true){component=mount(InventoryPrintingSettings,{target:document.body,props:{scope:repository.scope,repository,canConfigure,canPrint:true}});await settle();}
it('shows authoritative availability independently from printer attention and derives unnamed media dimensions',async()=>{
 const repo=new FakePrintingRepository();repo.registrations[0].availability='online';repo.destinations[0].media.name='';repo.destinations[0].readiness='error';repo.destinations[0].readinessReason='paper_empty';repo.destinations[0].reportedAt='2026-10-03T12:00:00Z';await start(repo,false);
 expect(document.body.textContent).toContain('Computer online');expect(document.body.textContent).toContain('Load a label roll');expect(document.body.textContent).toContain('29 × 90 mm');expect(button('Edit printer')).toBeUndefined();
});
it('keeps a stale printer draft until explicit reload and then saves the current supported media revision',async()=>{
 const repo=new FakePrintingRepository();await start(repo);button('Edit printer')!.click();await settle();
 const name=document.getElementById('edit-printer-name') as HTMLInputElement;name.value='My printer';name.dispatchEvent(new Event('input',{bubbles:true}));repo.destinations[0].revision++;repo.destinations[0].name='Renamed elsewhere';
 document.querySelector('[data-slot="dialog-content"] form')!.dispatchEvent(new Event('submit',{bubbles:true,cancelable:true}));await settle();expect(name.value).toBe('My printer');expect(button('Save printer')!.disabled).toBe(true);expect(repo.destinations[0].name).toBe('Renamed elsewhere');
 button('Reload printer')!.click();await settle();expect((document.getElementById('edit-printer-name') as HTMLInputElement).value).toBe('Renamed elsewhere');
 document.querySelector('[data-slot="dialog-content"] form')!.dispatchEvent(new Event('submit',{bubbles:true,cancelable:true}));await settle();expect(repo.destinations[0].revision).toBe(3);expect(repo.destinations[0].media.presetId).toBe('brother-ql800-29x90');
});

class ControlledPrinterRepository extends FakePrintingRepository {
 mediaRead?:Promise<void>;
 saveWrite?:Promise<void>;
 override async mediaProfiles(...args:Parameters<FakePrintingRepository['mediaProfiles']>){await this.mediaRead;return super.mediaProfiles(...args);}
 override async updatePrinter(...args:Parameters<FakePrintingRepository['updatePrinter']>){await this.saveWrite;return super.updatePrinter(...args);}
}
it('allows closing a pending media read and ignores its late result',async()=>{
 const repo=new ControlledPrinterRepository();await start(repo);
 let finish!:()=>void;repo.mediaRead=new Promise<void>(resolve=>{finish=resolve;});
 button('Edit printer')!.click();await settle();expect(button('Cancel')!.disabled).toBe(false);
 button('Cancel')!.click();await settle();expect(document.getElementById('edit-printer-name')).toBeNull();
 finish();await settle();expect(document.getElementById('edit-printer-name')).toBeNull();expect(repo.destinations[0].revision).toBe(1);
});
it('keeps a pending save locked until the write completes',async()=>{
 const repo=new ControlledPrinterRepository();await start(repo);button('Edit printer')!.click();await settle();
 let finish!:()=>void;repo.saveWrite=new Promise<void>(resolve=>{finish=resolve;});
 document.querySelector('[data-slot="dialog-content"] form')!.dispatchEvent(new Event('submit',{bubbles:true,cancelable:true}));await settle();
 expect(button('Cancel')!.disabled).toBe(true);expect(document.getElementById('edit-printer-name')).not.toBeNull();
 finish();await settle();expect(document.getElementById('edit-printer-name')).toBeNull();expect(repo.destinations[0].revision).toBe(2);
});
