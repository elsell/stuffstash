import {ExpoPushDevice} from './ExpoPushDevice';
import {PushReconciliationController} from '../../application/notifications/PushReconciliationController';
import {expect,it} from 'vitest';
import {ExpoPushReconciliationEvents} from './ExpoPushReconciliationEvents';
it('triggers on foreground and token changes and removes both subscriptions',()=>{
 let onState=(value:string)=>{};let onToken=(_token:{type:string;data:unknown})=>{};let removed=0,count=0;
 const events=new ExpoPushReconciliationEvents({addEventListener(_event,listener){onState=listener;return {remove(){removed++;}};}},{addPushTokenListener(listener){onToken=listener;return {remove(){removed++;}};}},{acceptNativeToken(){return true;},invalidateNativeToken(){}});
 const stop=events.subscribe(()=>{count++;});onState('background');expect(count).toBe(0);onState('active');onToken({type:'ios',data:'token'});expect(count).toBe(2);stop();expect(removed).toBe(2);
});
it('does not fetch indefinitely when native token retrieval emits a token event',async()=>{
 let tokenListener=(_token:{type:string;data:unknown})=>{};let fetches=0;
 const api={async getPermissionsAsync(){return {granted:true};},async requestPermissionsAsync(){return {granted:true};},async setNotificationChannelAsync(){},async getDevicePushTokenAsync(){fetches++;const token={type:'ios',data:'native-token'};if(fetches<4)tokenListener(token);return token;},addPushTokenListener(listener:(token:{type:string;data:unknown})=>void){tokenListener=listener;return {remove(){}};}};
 const device=new ExpoPushDevice(api,'ios',3);
 const events=new ExpoPushReconciliationEvents({addEventListener(){return {remove(){}};}},api,device);
 const controller=new PushReconciliationController(events,{async reconcile(){await device.nativeToken();}},{record(){}});
 const stop=controller.start();await new Promise(resolve=>setTimeout(resolve,0));stop();expect(fetches).toBe(1);
});
