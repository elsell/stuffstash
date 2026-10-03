import {expect,it} from 'vitest';
import {ApiPrintingRepository} from './printingRepository';
import {fakePrintMedia} from '$lib/fakes/printingRepository';
class PrinterServer {
 readonly printer={id:'printer',adapterId:'brother-ql800',name:'Garage',revision:4,retired:false,media:{...fakePrintMedia,name:''},mediaFingerprint:'media-v1',readiness:'error',readinessReason:'paper_empty',reportedAt:'2026-10-03T12:00:00Z'};
 async fetch(input:RequestInfo|URL,init?:RequestInit){
  const request=input instanceof Request?input:new Request(input,init),url=new URL(request.url);
  const failure=(status:number)=>Response.json({error:{code:'rejected',message:'Rejected'}},{status});
  if(url.origin!=='https://api.example'||request.headers.get('Authorization')!=='Bearer owner')return failure(401);
  const base='/tenants/tenant/inventories/inventory';
  if(request.method==='GET'&&url.pathname===`${base}/printers`)return Response.json({data:[this.printer],meta:{pagination:{}}});
  if(request.method==='GET'&&url.pathname===`${base}/printer-profiles`)return Response.json({data:[{adapterId:this.printer.adapterId,name:'Brother',media:[fakePrintMedia],supportedPlatforms:['linux'],transport:'usb',physicallyVerified:false}]});
  if(request.method==='GET'&&url.pathname===`${base}/print-connectors`)return Response.json({data:[{id:'connector',name:'Host',generation:1,state:'active',availability:'offline',authorizationPending:false,printerIds:['printer'],lastSeenAt:'2026-10-03T12:00:00Z'}],meta:{pagination:{}}});
  if(request.method==='PATCH'&&url.pathname===`${base}/printers/printer`){
   const body=await request.json();if(body.revision!==this.printer.revision)return failure(409);
   if(Object.keys(body).sort().join(',')!=='name,presetId,presetVersion,retired,revision'||body.presetId!==fakePrintMedia.presetId||body.presetVersion!==fakePrintMedia.version)return failure(422);
   Object.assign(this.printer,{name:body.name,retired:body.retired,revision:this.printer.revision+1,media:{...fakePrintMedia}});return Response.json({data:this.printer});
  }
  return failure(404);
 }
}
it('maps authoritative health and sends only scoped revisioned printer configuration',async()=>{
 const server=new PrinterServer(),scope={tenantId:'tenant',inventoryId:'inventory'};const repo=new ApiPrintingRepository('https://api.example',()=> 'owner',server.fetch.bind(server));
 const printer=(await repo.printers(scope))[0];expect(printer.readinessReason).toBe('paper_empty');expect(printer.reportedAt).toBe(server.printer.reportedAt);expect((await repo.connectors(scope))[0].availability).toBe('offline');
 const profile=(await repo.mediaProfiles(scope))[0];const updated=await repo.updatePrinter(scope,printer,'Updated',false,profile.media);expect(updated.revision).toBe(5);
 await expect(repo.updatePrinter(scope,printer,'Stale',false,profile.media)).rejects.toMatchObject({kind:'conflict'});expect(server.printer.name).toBe('Updated');
 await expect(repo.updatePrinter({...scope,inventoryId:'other'},updated,'Foreign',false,profile.media)).rejects.toMatchObject({kind:'invalid'});
 const revoked=new ApiPrintingRepository('https://api.example',()=> 'expired',server.fetch.bind(server));await expect(revoked.updatePrinter(scope,updated,'Denied',false,profile.media)).rejects.toMatchObject({kind:'authentication_required'});
});
