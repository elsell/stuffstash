import { expect, it } from 'vitest';
import { FakePrintingRepository } from '$lib/fakes/printingRepository';
import { ManualPrintRequest } from './manualPrint';
it('preserves an immutable request after lost response and retries without another print job', async () => {
    const repository = new FakePrintingRepository();
    const printer = repository.destinations[0];
    const request = new ManualPrintRequest(repository, repository.scope, 'asset', 'request-key');
    const selection = { printerId: printer.id, expectedMediaFingerprint: printer.mediaFingerprint, templateId: 'qr-title', templateVersion: 1, showReference: true, copies: 1 };
    await request.preview(selection, printer.media);
    repository.loseNextJobResponse = true;
    await expect(request.submit()).rejects.toThrow();
    expect(repository.queued.size).toBe(1);
    expect(request.locked).toBe(true);
    await expect(request.preview({ ...selection, copies: 2 }, printer.media)).rejects.toThrow();
    const job = await request.submit();
    expect(job.status).toBe('queued');
    expect(repository.queued.size).toBe(1);
});
it('requires fresh preview for changed selection and rejects revoked print permission', async () => {
    const repository = new FakePrintingRepository();
    const request = new ManualPrintRequest(repository, repository.scope, 'asset', 'request');
    await expect(request.submit()).rejects.toThrow();
    const printer = repository.destinations[0];
    await request.preview({ printerId: printer.id, expectedMediaFingerprint: printer.mediaFingerprint, templateId: 'qr-title', templateVersion: 1, showReference: true, copies: 1 }, printer.media);
    request.invalidate();
    await expect(request.submit()).rejects.toThrow();
    repository.canPrint = false;
    await expect(request.preview({ printerId: printer.id, expectedMediaFingerprint: printer.mediaFingerprint, templateId: 'qr-title', templateVersion: 1, showReference: true, copies: 1 }, printer.media)).rejects.toThrow();
});
import { SessionPrintIntents } from './manualPrint';
it('retains an ambiguous request across dialog and route lifetimes in the same session', async () => {
    const repo = new FakePrintingRepository();
    let sequence = 0;
    const intents = new SessionPrintIntents(repo, () => `request-${++sequence}`);
    const first = intents.forAsset(repo.scope, 'asset');
    const printer = repo.destinations[0];
    await first.preview({ printerId: printer.id, expectedMediaFingerprint: printer.mediaFingerprint, templateId: 'qr-title', templateVersion: 1, showReference: true, copies: 1 }, printer.media);
    repo.loseNextJobResponse = true;
    await expect(first.submit()).rejects.toThrow();
    const reopened = intents.forAsset(repo.scope, 'asset');
    expect(reopened).toBe(first);
    expect(() => intents.startAnother(repo.scope, 'asset')).toThrow();
    await reopened.submit();
    expect(repo.queued.size).toBe(1);
    expect(sequence).toBe(1);
});
it('allows correcting a definite rejection but keeps a previously ambiguous request immutable', async () => {
    const repo = new FakePrintingRepository();
    const printer = repo.destinations[0];
    const selection = { printerId: printer.id, expectedMediaFingerprint: printer.mediaFingerprint, templateId: 'qr-title', templateVersion: 1, showReference: true, copies: 1 };
    const request = new ManualPrintRequest(repo, repo.scope, 'asset', 'intent');
    await request.preview(selection, printer.media);
    repo.canPrint = false;
    await expect(request.submit()).rejects.toThrow();
    expect(request.locked).toBe(false);
    repo.canPrint = true;
    repo.loseNextJobResponse = true;
    await expect(request.submit()).rejects.toThrow();
    repo.canPrint = false;
    await expect(request.submit()).rejects.toThrow();
    expect(request.locked).toBe(true);
    expect(repo.queued.size).toBe(1);
});


class DelayedPreviewRepository extends FakePrintingRepository {
 readonly deliveries: Array<()=>void>=[];
 override async preview(...args:Parameters<FakePrintingRepository['preview']>){
  const rendered=await super.preview(...args);
  await new Promise<void>(resolve=>{this.deliveries.push(resolve);});
  return rendered;
 }
}
it('fences a late preview from a dismissed dialog or older selection',async()=>{
 const repo=new DelayedPreviewRepository();const intents=new SessionPrintIntents(repo,()=> 'intent');const request=intents.forAsset(repo.scope,'asset');const printer=repo.destinations[0];
 const selection={printerId:printer.id,expectedMediaFingerprint:printer.mediaFingerprint,templateId:'qr-title',templateVersion:1,showReference:true,copies:1};
 const old=request.preview(selection,printer.media);const rejected=expect(old).rejects.toMatchObject({kind:'conflict'});await Promise.resolve();
 const reopened=intents.forAsset(repo.scope,'asset');const current=reopened.preview({...selection,copies:2},printer.media);await Promise.resolve();repo.deliveries[1]();await current;
 repo.deliveries[0]();await rejected;const job=await reopened.submit();expect(job.copies).toBe(2);expect(repo.queued.size).toBe(1);
});
it('does not retain a preview whose dialog was canceled',async()=>{
 const repo=new DelayedPreviewRepository();const request=new ManualPrintRequest(repo,repo.scope,'asset','intent');const printer=repo.destinations[0];const controller=new AbortController();
 const pending=request.preview({printerId:printer.id,expectedMediaFingerprint:printer.mediaFingerprint,templateId:'qr-title',templateVersion:1,showReference:true,copies:1},printer.media,controller.signal);
 const rejected=expect(pending).rejects.toMatchObject({kind:'conflict'});await Promise.resolve();controller.abort();repo.deliveries[0]();await rejected;await expect(request.submit()).rejects.toMatchObject({kind:'invalid'});expect(repo.queued.size).toBe(0);
});
it('retains a linked reprint after dismissal and lost response without duplicating output',async()=>{
 const repo=new FakePrintingRepository();let key=0;const intents=new SessionPrintIntents(repo,()=>`key-${++key}`);const printer=repo.destinations[0];
 const selection={printerId:printer.id,expectedMediaFingerprint:printer.mediaFingerprint,templateId:'qr-title',templateVersion:1,showReference:true,copies:1};
 const original=intents.forAsset(repo.scope,'asset');await original.preview(selection,printer.media);const job=await original.submit();repo.queued.get(job.id)!.status='completed';
 const request=intents.forReprint(repo.scope,'asset',job.id);await request.preview(selection,printer.media);repo.loseNextJobResponse=true;await expect(request.submit()).rejects.toThrow();
 const reopened=intents.forReprint(repo.scope,'asset',job.id);expect(reopened).toBe(request);expect(reopened.locked).toBe(true);const successor=await reopened.submit();expect(successor.predecessor).toBe(job.id);expect(repo.queued.size).toBe(2);expect(key).toBe(2);
 repo.canPrint=false;const denied=intents.forReprint(repo.scope,'asset',successor.id);await expect(denied.preview(selection,printer.media)).rejects.toMatchObject({kind:'denied'});
});
