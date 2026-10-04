import {afterEach,expect,it} from 'vitest';
import {flushSync,mount,tick,unmount} from 'svelte';
import InventoryPrintingSettings from './InventoryPrintingSettings.svelte';
import {FakePrintingRepository} from '$lib/fakes/printingRepository';
import type {PrintJob} from '$lib/domain/printing';
let component:ReturnType<typeof mount>|undefined;
afterEach(async()=>{if(component)await unmount(component);component=undefined;document.body.innerHTML='';});
async function settle(){for(let i=0;i<8;i++){await Promise.resolve();await tick();}}
function button(text:string){return [...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.trim()===text);}
function uncertain():PrintJob{return {id:'job',printerId:'printer',assetId:'asset',status:'uncertain',revision:3,copies:1,completedCopies:0,reason:'response_lost',attemptOutcome:'uncertain',createdAt:'2026-10-03T12:00:00Z'};}
async function start(repository:FakePrintingRepository,canPrint=true){component=mount(InventoryPrintingSettings,{target:document.body,props:{view:'history',scope:repository.scope,repository,canConfigure:false,canPrint}});await settle();}
async function selectOutcome(){const trigger=document.getElementById('resolution-outcome-job')!;trigger.focus();trigger.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));flushSync();await settle();const option=[...document.querySelectorAll<HTMLElement>('[role="option"]')].find(n=>n.textContent?.trim()==='I saw the label print')!;option.dispatchEvent(new PointerEvent('pointerup',{pointerType:'mouse',bubbles:true,cancelable:true}));flushSync();await settle();}
it('waits for current-attempt idle evidence and keeps viewers from resolving',async()=>{const repository=new FakePrintingRepository();repository.queued.set('job',uncertain());await start(repository,false);expect(button('Resolve this job')).toBeUndefined();expect(document.body.textContent).toContain('must confirm the printer is idle');repository.queued.get('job')!.idleConfirmedAt='2026-10-03T12:01:00Z';button('Refresh status')!.click();await settle();expect(button('Resolve this job')).toBeUndefined();expect(repository.queued.get('job')!.status).toBe('uncertain');});
it('requires acknowledgement and retries a lost resolution without rewriting physical evidence or printing',async()=>{const repository=new FakePrintingRepository();repository.queued.set('job',{...uncertain(),idleConfirmedAt:'2026-10-03T12:01:00Z'});repository.loseNextResolutionResponse=true;await start(repository);expect(button('Resolve this job')!.disabled).toBe(true);await selectOutcome();expect(button('Resolve this job')!.disabled).toBe(true);(document.getElementById('resolution-ack-job') as HTMLInputElement).click();await settle();button('Resolve this job')!.click();await settle();expect(repository.queued.get('job')!.resolution?.reportedOutcome).toBe('printed');expect(document.querySelector('[role="alert"]')).not.toBeNull();expect((document.getElementById('resolution-outcome-job') as HTMLButtonElement).disabled).toBe(true);button('Retry the same resolution')!.click();await settle();expect(document.body.textContent).toContain('Reported printed');expect(document.body.textContent).toContain('the printer did not confirm the outcome.');expect(document.body.textContent).toContain('User report: I saw the label print');expect(repository.queued.get('job')!.attemptOutcome).toBe('uncertain');expect(repository.queued.get('job')!.completedCopies).toBe(0);expect(repository.queued.size).toBe(1);});
it('requires fresh observation after a conflict and explicit refresh reveals a later uncertain attempt',async()=>{
  const repository=new FakePrintingRepository();
  repository.queued.set('job',{...uncertain(),idleConfirmedAt:'2026-10-03T12:01:00Z'});
  await start(repository);
  await selectOutcome();
  (document.getElementById('resolution-ack-job') as HTMLInputElement).click();await settle();
  // The server reconciles the old attempt without output, requeues it, then a
  // later attempt loses its result and subsequently confirms idle again.
  Object.assign(repository.queued.get('job')!,{status:'queued',revision:4,idleConfirmedAt:undefined,attemptOutcome:'not_printed'});
  Object.assign(repository.queued.get('job')!,{status:'printing',revision:5});
  Object.assign(repository.queued.get('job')!,{status:'uncertain',revision:6,attemptOutcome:'uncertain',idleConfirmedAt:'2026-10-03T12:03:00Z'});
  button('Resolve this job')!.click();await settle();
  expect(repository.queued.get('job')!.resolution).toBeUndefined();
  expect(button('Retry the same resolution')).toBeDefined();
  expect((document.getElementById('resolution-outcome-job') as HTMLButtonElement).disabled).toBe(true);
  button('Refresh status')!.click();await settle();
  expect(button('Resolve this job')!.disabled).toBe(true);
  expect((document.getElementById('resolution-outcome-job') as HTMLButtonElement).disabled).toBe(false);
  expect((document.getElementById('resolution-ack-job') as HTMLInputElement).checked).toBe(false);
  await selectOutcome();expect(button('Resolve this job')!.disabled).toBe(true);
  (document.getElementById('resolution-ack-job') as HTMLInputElement).click();await settle();
  button('Resolve this job')!.click();await settle();
  expect(repository.queued.get('job')!.revision).toBe(7);
  expect(repository.queued.get('job')!.resolution?.reportedOutcome).toBe('printed');
  expect(repository.queued.size).toBe(1);
});
