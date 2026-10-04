import type {Page} from '@playwright/test';
import {installAuthenticatedWorkspace} from './workspace-fixture';

// Controlled transport implementation: printer edits enforce revisions and persist.
// It cannot contact a connector or produce physical output.
export async function installPrintingSettings(page:Page, waitForProfiles?:()=>Promise<void>|undefined){
 await installAuthenticatedWorkspace(page);
 const media={name:'29 × 90 mm (DK-11201)',presetId:'brother-ql800-29x90',version:1,widthMicrometers:29000,heightMicrometers:89800,marginsMicrometers:{left:1524,right:1524,top:2963,bottom:2963},resolutionDpi:300,rasterWidth:306,rasterHeight:991,orientation:'feed',colorMode:'monochrome',cutPolicy:'after_label',displayRotation:270};
 const printer={id:'printer',name:'Garage Brother',adapterId:'brother-ql800',revision:1,retired:false,media:{...media,name:''},mediaFingerprint:'media-v1',readiness:'ready',reportedAt:'2026-10-04T02:00:00Z'};
 const scope='/tenants/tenant-home/inventories/inventory-household';
 const endpoints=new Set(['printers','printer-profiles','print-connectors','print-settings','label-templates','print-jobs']);
 await page.route('http://127.0.0.1:18080/**',async route=>{
  const path=new URL(route.request().url()).pathname;
  if(!path.startsWith(`${scope}/`))return route.fallback();
  const operation=path.slice(scope.length+1);if(!endpoints.has(operation)&&operation!=='printers/printer')return route.fallback();
  let data:unknown;
  if(operation==='printers/printer'&&route.request().method()==='PATCH'){
   const input=route.request().postDataJSON();if(input.revision!==printer.revision)return route.fulfill({status:409,json:{error:{code:'conflict',message:'Printer changed'}}});
   Object.assign(printer,{name:input.name,retired:input.retired,revision:printer.revision+1});data=printer;
  }else if(route.request().method()!=='GET')return route.fulfill({status:405});
  else if(operation==='printers')data=[printer];
  else if(operation==='printer-profiles'){await waitForProfiles?.();data=[{adapterId:'brother-ql800',media:[media]}];}
  else if(operation==='print-connectors')data=[{id:'computer',name:'Garage computer',state:'active',authorizationPending:false,availability:'online',printerIds:['printer'],lastSeenAt:'2026-10-04T02:00:00Z'}];
  else if(operation==='print-settings')data={revision:1,defaultPrinterId:'printer',template:{id:'qr-title',version:1,options:{showReference:true}},printOnCreateDefault:false};
  else if(operation==='label-templates')data=[{id:'qr-title',version:1,name:'QR and title',defaults:{show_reference:true}}];
  else if(operation==='print-jobs')data=[{id:'reported',printerId:'printer',status:'failed',revision:4,copies:1,createdAt:'2026-10-04T01:00:00Z',attempts:[{outcome:'uncertain',completedCopies:0}],resolution:{reportedOutcome:'printed',resolvedBy:'oidc_'+('long-subject-'.repeat(9)),resolvedAt:'2026-10-04T01:05:00Z'}},{id:'completed',printerId:'printer',status:'completed',revision:3,copies:1,createdAt:'2026-10-04T00:00:00Z',attempts:[{outcome:'completed',completedCopies:1}]}];
  return route.fulfill({status:200,json:{data,meta:{pagination:{}}}});
 });
 return printer;
}
