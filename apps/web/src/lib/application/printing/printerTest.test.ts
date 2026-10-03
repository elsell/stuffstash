import {expect,it} from 'vitest';
import {FakePrintingRepository} from '$lib/fakes/printingRepository';
import {SessionPrintIntents} from './manualPrint';
it('retains exactly one asset-free diagnostic job across a lost response and forbids ambiguous replacement',async()=>{
 const repo=new FakePrintingRepository();let sequence=0;const intents=new SessionPrintIntents(repo,()=>`key-${++sequence}`);const p=repo.destinations[0];
 const request=intents.forPrinterTest(repo.scope,p.id);const selection={printerId:p.id,expectedMediaFingerprint:p.mediaFingerprint,templateId:'qr-title',templateVersion:1,showReference:true,copies:1};
 repo.loseNextJobResponse=true;await expect(request.submit(selection)).rejects.toThrow();expect(repo.queued.size).toBe(1);
 expect(()=>intents.startAnotherPrinterTest(repo.scope,p.id,{...[...repo.queued.values()][0],status:'completed'})).toThrow();
 const reopened=intents.forPrinterTest(repo.scope,p.id);expect(reopened).toBe(request);await expect(reopened.submit({...selection,copies:2})).rejects.toThrow();
 const job=await reopened.submit(selection);expect(job.assetId).toBeUndefined();expect(job.copies).toBe(1);expect(repo.queued.size).toBe(1);expect(sequence).toBe(1);
 repo.canPrint=false;const denied=intents.startAnotherPrinterTest(repo.scope,p.id,{...job,status:'completed'});await expect(denied.submit(selection)).rejects.toMatchObject({kind:'denied'});expect(denied.locked).toBe(false);
});
