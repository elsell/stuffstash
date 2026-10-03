import { createAuthenticatedTransport } from './authenticatedTransport';
import { StuffStashAPIError, type StuffStashClientOptions } from './stuffStashClient';
import type { components } from './generated/schema';
export interface PrintScope {tenantId:string;inventoryId:string}
type Schema=components['schemas'];
export class PrintingClient {
 private readonly transport;
 constructor(options:StuffStashClientOptions){this.transport=createAuthenticatedTransport(options);}
 async printers(scope:PrintScope,cursor?:string){return page(await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/printers',{params:{path:scope,query:{limit:100,cursor}}}));}
 async connectors(scope:PrintScope,cursor?:string){return page(await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/print-connectors',{params:{path:scope,query:{limit:100,cursor}}}));}
 async templates(scope:PrintScope){return unwrap(await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/label-templates',{params:{path:scope}}))??[];}
 async settings(scope:PrintScope){return unwrap(await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/print-settings',{params:{path:scope}}));}
 async saveSettings(scope:PrintScope,input:Schema['InventoryPrintSettings']){return unwrap(await this.transport.PUT('/tenants/{tenantId}/inventories/{inventoryId}/print-settings',{params:{path:scope},body:input}));}
 async updatePrinter(scope:PrintScope,printerId:string,input:Schema['UpdatePrinterBody']){return unwrap(await this.transport.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/printers/{printerId}',{params:{path:{...scope,printerId}},body:input}));}
 async render(scope:PrintScope,assetId:string,input:Schema['RenderInputBody']){return unwrap(await this.transport.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/label-renders',{params:{path:{...scope,assetId}},body:input}));}
 async content(scope:PrintScope,renderId:string){
  const {data,error,response}=await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/label-renders/{renderId}/content',{params:{path:{...scope,renderId}},parseAs:'blob'});
  if(!response.ok||error||!data)throw failure(response.status,error);
  if(response.headers.get('Content-Type')?.split(';')[0]!=='image/png')throw new StuffStashAPIError(502,'invalid_label','Invalid label content');
  return data;
 }
 async createJob(scope:PrintScope,assetId:string,input:Schema['PrintJobSelection'],key:string){return unwrap(await this.transport.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/print-jobs',{params:{path:{...scope,assetId},header:{'Idempotency-Key':key}},body:input}));}
 async job(scope:PrintScope,jobId:string){return unwrap(await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/print-jobs/{jobId}',{params:{path:{...scope,jobId}}}));}
 async jobs(scope:PrintScope,cursor?:string){return page(await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/print-jobs',{params:{path:scope,query:{limit:25,cursor}}}));}
 async cancel(scope:PrintScope,jobId:string,revision:number){return unwrap(await this.transport.POST('/tenants/{tenantId}/inventories/{inventoryId}/print-jobs/{jobId}/cancellation',{params:{path:{...scope,jobId}},body:{revision}}));}
}
function failure(status:number,error?:{error:{code:string;message:string}}){return new StuffStashAPIError(status,error?.error.code??'printing_unavailable',error?.error.message??'Printing request failed.');}
function unwrap<T>(result:{data?:{data:T};error?:{error:{code:string;message:string}};response:Response}):T{if(!result.response.ok||result.error||!result.data)throw failure(result.response.status,result.error);return result.data.data;}
function page<T>(result:{data?:{data:T[]|null;meta:{pagination?:{nextCursor?:string|null}}};error?:{error:{code:string;message:string}};response:Response}){return {items:unwrap(result)??[],nextCursor:result.data?.meta.pagination?.nextCursor??undefined};}
