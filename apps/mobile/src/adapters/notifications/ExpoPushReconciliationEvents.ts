import type { ExpoPushDevice } from './ExpoPushDevice';
import type { PushReconciliationEvents } from '../../application/notifications/PushReconciliationController';
type Subscription={remove():void};
export class ExpoPushReconciliationEvents implements PushReconciliationEvents {
 constructor(
  private readonly appState:{addEventListener(event:'change',listener:(state:string)=>void):Subscription},
  private readonly notifications:{addPushTokenListener(listener:(token:{type:string;data:unknown})=>void):Subscription},
  private readonly device: Pick<ExpoPushDevice,'acceptNativeToken'|'invalidateNativeToken'>
 ) {}
 subscribe(listener:()=>void):()=>void {
  const state=this.appState.addEventListener('change',value=>{if(value==='active'){this.device.invalidateNativeToken();listener();}});
  const token=this.notifications.addPushTokenListener(token=>{if(this.device.acceptNativeToken(token))listener();});
  return ()=>{state.remove();token.remove();};
 }
}
