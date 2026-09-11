import type { NotificationObservability } from './NotificationObservability';
import type { ReadRequest } from '../shared/ReadRequest';
export interface PushReconciliationEvents {
 subscribe(listener: () => void): () => void;
}
export class PushReconciliationController {
 constructor(private readonly events: PushReconciliationEvents, private readonly session: {reconcile(request: ReadRequest): Promise<void>}, private readonly observer: NotificationObservability) {}
 start(): () => void {
  const controller=new AbortController();
  let running=false, requested=false;
  const trigger=()=>{
   if(controller.signal.aborted) return;
   requested=true;
   if(running) return;
   running=true;
   void (async()=>{
    try {
     while(requested && !controller.signal.aborted){
      requested=false;
      try {await this.session.reconcile({signal:controller.signal});}
      catch { if(!controller.signal.aborted) this.observer.record({operation:'push-reconcile',outcome:'failed'}); }
     }
    } finally {running=false;}
   })();
  };
  const unsubscribe=this.events.subscribe(trigger);
  trigger();
  return ()=>{controller.abort();unsubscribe();};
 }
}
