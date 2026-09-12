import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
import type { ExpirationRepository, ExpirationPage } from '$lib/ports/expirationRepository';
export interface ExpirationHomeState {page?:ExpirationPage;error:string;}
export class ExpirationHomeQuery {
 state:ExpirationHomeState={error:''}; private key=''; private controller?:AbortController;
 constructor(private readonly repository:Pick<ExpirationRepository,'list'>,private readonly changed:(state:ExpirationHomeState)=>void,private readonly observer?:WorkspaceObserver){}
 dispose(){this.controller?.abort();}
 async load(tenantId:string,inventoryId:string):Promise<boolean>{
  this.controller?.abort();const controller=new AbortController();this.controller=controller;
  const key=JSON.stringify([tenantId,inventoryId]);if(key!==this.key){this.key=key;this.update({error:''});}
  try{
   const [expired,soon]=await Promise.all([this.repository.list(tenantId,inventoryId,{mode:'expired'},{limit:2,signal:controller.signal}),this.repository.list(tenantId,inventoryId,{mode:'soon'},{limit:2,signal:controller.signal})]);
   if(controller.signal.aborted)return false;
   this.observer?.record('workspace.expiration_loaded');
   this.update({page:{...expired,items:[expired.items[0],soon.items[0],expired.items[1],soon.items[1]].filter(Boolean).slice(0,3)},error:''});return true;
  }catch(error){
   if(controller.signal.aborted)return false;
   this.observer?.record('workspace.expiration_load_failed');
   const denied=typeof error==='object'&&error!==null&&'status'in error&&[401,403,404].includes(Number(error.status));
   this.update({page:denied?undefined:this.state.page,error:'Expiration could not be refreshed.'});return false;
  }
 }
 private update(state:ExpirationHomeState){this.state=state;this.changed(state);}
}
