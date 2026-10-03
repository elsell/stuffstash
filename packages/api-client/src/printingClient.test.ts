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
