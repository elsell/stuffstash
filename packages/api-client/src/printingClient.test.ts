import { expect,it } from 'vitest';
import { PrintingClient } from './printingClient';
class PrintingHTTPFake {
 readonly scope={tenantId:'tenant',inventoryId:'inventory'};
 settings={revision:0,defaultPrinterId:null as string|null,template:{id:'qr-title',version:1,options:{showReference:true}},printOnCreateDefault:false};
 fetch:typeof fetch=async(input,init)=>{
  const request=new Request(input,init);const path=new URL(request.url).pathname;
  const denied=(status:number)=>Response.json({error:{code:'denied',message:'Denied'}},{status});
  if(!request.headers.get('Authorization'))return denied(401);
  if(!path.startsWith('/tenants/tenant/inventories/inventory/'))return denied(404);
  if(request.method==='PUT'){
   if(request.headers.get('Authorization')!=='Bearer owner')return denied(403);
   const next=await request.json() as typeof this.settings;
   if(next.revision!==this.settings.revision)return denied(409);
   this.settings={...next,revision:next.revision+1};
  }
  return Response.json({data:this.settings,meta:{}});
 };
}
it('preserves scoped settings authorization and optimistic revision through the generated transport',async()=>{
 const fake=new PrintingHTTPFake();let token:string|null=null;
 const client=new PrintingClient({baseUrl:'https://stash.example',tokenProvider:()=>token,fetch:fake.fetch});
 await expect(client.settings(fake.scope)).rejects.toMatchObject({status:401});
 token='viewer';const settings=await client.settings(fake.scope);
 await expect(client.saveSettings(fake.scope,settings)).rejects.toMatchObject({status:403});
 token='owner';await expect(client.saveSettings({...fake.scope,inventoryId:'foreign'},settings)).rejects.toMatchObject({status:404});
 const saved=await client.saveSettings(fake.scope,{...settings,defaultPrinterId:'printer',printOnCreateDefault:true});
 expect(saved.revision).toBe(1);expect(saved.printOnCreateDefault).toBe(true);
 await expect(client.saveSettings(fake.scope,settings)).rejects.toMatchObject({status:409});
});

class RecoveryHTTPFake {
 state='uncertain';revision=3;idle=false;physicalOutcome='uncertain';report:string|undefined;
 readonly requests=new Map<string,{fingerprint:string;id:string;predecessor?:string}>();
 fetch:typeof fetch=async(input,init)=>{
  const request=new Request(input,init);const path=new URL(request.url).pathname;const denied=(status:number)=>Response.json({error:{code:'denied',message:'Denied'}},{status});
  const principal=request.headers.get('Authorization');if(!principal)return denied(401);if(principal!=='Bearer editor')return denied(403);
  if(!path.startsWith('/tenants/tenant/inventories/inventory/'))return denied(404);
  const body=await request.json();
  if(path.endsWith('/resolution')){
   if(!body.acknowledgeUncertainty)return denied(400);
   if(this.report){if(this.report!==body.reportedOutcome)return denied(409);}
   else{if(!this.idle||this.state!=='uncertain'||body.revision!==this.revision)return denied(409);this.report=body.reportedOutcome;this.state='failed';this.revision++;}
   return Response.json({data:{id:'job',status:this.state,revision:this.revision,attempts:[{outcome:this.physicalOutcome}],resolution:{reportedOutcome:this.report,resolvedBy:'editor'}}});
  }
  const key=request.headers.get('Idempotency-Key');if(!key)return denied(400);
  const fingerprint=JSON.stringify({path,body});const prior=this.requests.get(key);if(prior&&prior.fingerprint!==fingerprint)return denied(409);
  if(!prior){if(path.endsWith('/reprints')&&this.state!=='failed')return denied(409);if(path.endsWith('/test-jobs')&&(body.copies!==1||body.printerId!=='printer'))return denied(400);this.requests.set(key,{fingerprint,id:`new-${this.requests.size+1}`,...(path.endsWith('/reprints')?{predecessor:'job'}:{})});}
  return Response.json({data:this.requests.get(key)});
 };
}
it('keeps resolution scoped, revision guarded, and physically uncertain through retry',async()=>{
 const fake=new RecoveryHTTPFake();let token:string|null=null;const client=new PrintingClient({baseUrl:'https://stash.example',tokenProvider:()=>token,fetch:fake.fetch});const scope={tenantId:'tenant',inventoryId:'inventory'};const choice={revision:3,acknowledgeUncertainty:true,reportedOutcome:'printed' as const};
 await expect(client.resolve(scope,'job',choice)).rejects.toMatchObject({status:401});token='viewer';await expect(client.resolve(scope,'job',choice)).rejects.toMatchObject({status:403});token='editor';await expect(client.resolve({...scope,inventoryId:'other'},'job',choice)).rejects.toMatchObject({status:404});await expect(client.resolve(scope,'job',choice)).rejects.toMatchObject({status:409});fake.idle=true;const result=await client.resolve(scope,'job',choice);expect(result.status).toBe('failed');expect(result.attempts?.[0].outcome).toBe('uncertain');await client.resolve(scope,'job',choice);expect(fake.revision).toBe(4);
});
it('preserves explicit reprint and test-label idempotency and predecessor identity',async()=>{
 const fake=new RecoveryHTTPFake();fake.state='failed';const client=new PrintingClient({baseUrl:'https://stash.example',tokenProvider:()=> 'editor',fetch:fake.fetch});const scope={tenantId:'tenant',inventoryId:'inventory'};const selection={printerId:'printer',expectedMediaFingerprint:'media',templateId:'qr-title',templateVersion:1,templateOptions:{showReference:true},copies:1};
 const first=await client.reprint(scope,'job',selection,'reprint-key');const retry=await client.reprint(scope,'job',selection,'reprint-key');expect(first.id).toBe(retry.id);expect(first.predecessor).toBe('job');await expect(client.reprint(scope,'job',{...selection,copies:2},'reprint-key')).rejects.toMatchObject({status:409});await client.testJob(scope,'printer',selection,'test-key');await client.testJob(scope,'printer',selection,'test-key');expect(fake.requests.size).toBe(2);
});
