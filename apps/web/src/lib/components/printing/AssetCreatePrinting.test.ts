import {afterEach,expect,it} from 'vitest';
import {mount,unmount,tick} from 'svelte';
import AddAssetTray from '../workspace/AddAssetTray.svelte';
import {printingWorkspaceContext} from '$lib/ports/printingRepository';
import {FakePrintingRepository} from '$lib/fakes/printingRepository';
import {SessionPrintIntents} from '$lib/application/printing/manualPrint';
import type {AddAssetSubmission,AddAssetSaveResult} from '$lib/domain/inventory';
let component:ReturnType<typeof mount>|undefined;
afterEach(()=>{if(component)unmount(component);component=undefined;document.body.innerHTML='';});
class CreateDrafts {
 readonly submissions:AddAssetSubmission[]=[];
 async save(draft:AddAssetSubmission):Promise<AddAssetSaveResult>{this.submissions.push(draft);return {saved:false};}
}
async function settle(){for(let i=0;i<8;i++){await Promise.resolve();await tick();}}
async function start(repository:FakePrintingRepository,drafts:CreateDrafts,pendingPrintDraft?:AddAssetSubmission){component=mount(AddAssetTray,{target:document.body,context:new Map([[printingWorkspaceContext,{apiIdentity:'fixture',repository,intents:new SessionPrintIntents(repository,()=> 'intent')}]]),props:{open:true,printScope:repository.scope,pendingPrintDraft,closeHref:'/settings',parentTargets:[],mediaPolicy:{supportedContentTypes:['image/png'],maxBytes:1024},customAssetTypes:[],customFieldDefinitions:[],saving:false,onClose:()=>{},onSave:drafts.save.bind(drafts)}});await settle();}
it('initializes the print choice once and honors an explicit unchecked create',async()=>{const repository=new FakePrintingRepository();repository.defaults.printOnCreateDefault=true;const drafts=new CreateDrafts();await start(repository,drafts);const checkbox=document.getElementById('create-print-label') as HTMLInputElement;expect(checkbox.checked).toBe(true);checkbox.click();repository.defaults.printOnCreateDefault=true;await settle();expect(checkbox.checked).toBe(false);const title=document.getElementById('asset-title') as HTMLInputElement;title.value='Drill';title.dispatchEvent(new Event('input',{bubbles:true}));await settle();[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()==='Save item')!.click();await settle();expect(drafts.submissions[0].printLabel).toBeUndefined();});
it('restores the exact pending create selection and freezes corrections while ambiguous',async()=>{const repository=new FakePrintingRepository();const drafts=new CreateDrafts();const pending:AddAssetSubmission={kind:'item',title:'Original drill',description:'',parentAssetId:null,photos:[],creationKey:'original-key',printLabel:{printerId:'printer',expectedMediaFingerprint:'old-media',templateId:'qr-title',templateVersion:1,showReference:false,copies:1}};await start(repository,drafts,pending);expect((document.getElementById('asset-title') as HTMLInputElement).value).toBe('Original drill');expect(document.getElementById('asset-title')?.matches(':disabled')).toBe(true);[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()==='Retry create and print')!.click();await settle();expect(drafts.submissions[0]).toBe(pending);expect(drafts.submissions[0].creationKey).toBe('original-key');});

it('requires an explicit no-label choice when print defaults cannot be loaded',async()=>{
 class UnavailableSettings extends FakePrintingRepository {override async settings():Promise<Awaited<ReturnType<FakePrintingRepository['settings']>>>{throw new Error('Controlled service outage');}}
 const repository=new UnavailableSettings();const drafts=new CreateDrafts();await start(repository,drafts);const title=document.getElementById('asset-title') as HTMLInputElement;title.value='Drill';title.dispatchEvent(new Event('input',{bubbles:true}));await settle();
 const save=[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()==='Save item')!;expect(save.disabled).toBe(true);
 [...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()==='Continue without a label')!.click();await settle();expect(save.disabled).toBe(false);save.click();await settle();expect(drafts.submissions[0].printLabel).toBeUndefined();
});
