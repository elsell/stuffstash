import type {PushNotificationResponses} from '../../application/notifications/PushNotificationResponses';
export type NativePushResponse={actionIdentifier:string;notification:{request:{identifier:string;content:{data?:unknown}}}};
type NativeResponses={
 getLastNotificationResponseAsync():Promise<NativePushResponse|null>;
 addNotificationResponseReceivedListener(listener:(response:NativePushResponse)=>void):{remove():void};
};
export class ExpoPushNotificationResponses implements PushNotificationResponses {
 private lastHandled?:string;
 constructor(private readonly native:NativeResponses,private readonly defaultAction:string){}
 subscribe(listener:(payload:unknown)=>void | Promise<void>,onFailure:()=>void):()=>void {
  let active=true,receivedLive=false;
  const pending=new Set<string>();
  const deliver=(response:NativePushResponse|null,initial=false)=>{
   if(!active||!response||response.actionIdentifier!==this.defaultAction)return;
   const key=response.notification.request.identifier;
   if((initial && key===this.lastHandled)||pending.has(key))return;
   this.lastHandled=key;pending.add(key);
   try {
    void Promise.resolve(listener(response.notification.request.content.data)).catch(()=>{if(active)onFailure();}).finally(()=>{pending.delete(key);});
   } catch {pending.delete(key);if(active)onFailure();}
  };
  const subscription=this.native.addNotificationResponseReceivedListener(response=>{receivedLive=true;deliver(response);});
  void this.native.getLastNotificationResponseAsync().then(response=>{if(!receivedLive)deliver(response,true);}).catch(()=>{if(active)onFailure();});
  return ()=>{active=false;subscription.remove();};
 }
}
