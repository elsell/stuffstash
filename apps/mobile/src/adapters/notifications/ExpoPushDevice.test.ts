import { expect, it } from 'vitest';
import { ExpoPushDevice } from './ExpoPushDevice';
class Notifications {
 events:string[]=[];
 granted=false;
 allow=true;
 token: {type:string;data:unknown}={type:'ios',data:'abcdef'};
 async setNotificationChannelAsync(id:string, options:{name:string;importance:number}) {this.events.push(`channel:${id}:${options.importance}`);return null;}
 async getPermissionsAsync(){this.events.push('check');return {granted:this.granted};}
 async requestPermissionsAsync(){this.events.push('request');return {granted:this.allow};}
 async getDevicePushTokenAsync(){this.events.push('token');return this.token;}
}
it('creates the Android channel before permission and maps native tokens',async()=>{
 const api=new Notifications();api.token={type:'android',data:'opaque-fcm-token'};
 const device=new ExpoPushDevice(api,'android',3);
 expect(await device.requestPermission()).toBe(true);
 expect(api.events).toEqual(['channel:expiration:3','check','request']);
 expect(await device.nativeToken()).toEqual({transport:'fcm',token:'opaque-fcm-token'});
});
it('does not prompt again for grants and preserves denial',async()=>{
 const api=new Notifications();api.granted=true;
 const device=new ExpoPushDevice(api,'ios',3);
 expect(await device.requestPermission()).toBe(true);
 expect(api.events).toEqual(['check']);
 expect(await device.nativeToken()).toEqual({transport:'apns',token:'abcdef'});
 api.granted=false;api.allow=false;
 expect(await device.requestPermission()).toBe(false);
});
it('rejects unsupported platforms and malformed or mismatched tokens without exposing data',async()=>{
 const api=new Notifications();
 await expect(new ExpoPushDevice(api,'web',3).requestPermission()).rejects.toMatchObject({kind:'unavailable'});
 expect(api.events).toEqual([]);
 const device=new ExpoPushDevice(api,'ios',3);
 for(const token of [{type:'android',data:'secret'}, {type:'ios',data:123}, {type:'ios',data:'secret\n'}, {type:'ios',data:''}]){
  api.token=token;
  await expect(device.nativeToken()).rejects.toMatchObject({kind:'unavailable'});
 }
});
it('hides native provider failures',async()=>{
 const api=new Notifications();
 api.getDevicePushTokenAsync=async()=>{throw new Error('private provider token');};
 await expect(new ExpoPushDevice(api,'ios',3).nativeToken()).rejects.toMatchObject({kind:'unavailable',message:'Notifications could not be updated. Try again.'});
});
it('checks permission silently without prompting or fetching a token',async()=>{
 const device=new ExpoPushDevice({async getPermissionsAsync(){return {granted:false};},async requestPermissionsAsync(){throw new Error('must not prompt');},async getDevicePushTokenAsync(){throw new Error('must not fetch');},async setNotificationChannelAsync(){throw new Error('must not alter');}},'ios',3);
 await expect(device.permissionGranted()).resolves.toBe(false);
});
it('keeps a newer token event when an older native read finishes later',async()=>{
 let complete!:(value:{type:string;data:unknown})=>void;
 const device=new ExpoPushDevice({async getPermissionsAsync(){return {granted:true};},async requestPermissionsAsync(){return {granted:true};},async setNotificationChannelAsync(){},getDevicePushTokenAsync(){return new Promise(resolve=>{complete=resolve;});}},'ios',3);
 const reading=device.nativeToken();expect(device.acceptNativeToken({type:'ios',data:'new-token'})).toBe(true);
 complete({type:'ios',data:'old-token'});
 await expect(reading).resolves.toEqual({transport:'apns',token:'new-token'});
 await expect(device.nativeToken()).resolves.toEqual({transport:'apns',token:'new-token'});
});
it('discards a native read invalidated by a later foreground transition',async()=>{
 let complete!:(value:{type:string;data:unknown})=>void;
 const device=new ExpoPushDevice({async getPermissionsAsync(){return {granted:true};},async requestPermissionsAsync(){return {granted:true};},async setNotificationChannelAsync(){},getDevicePushTokenAsync(){return new Promise(resolve=>{complete=resolve;});}},'ios',3);
 const reading=device.nativeToken();device.invalidateNativeToken();complete({type:'ios',data:'old-token'});
 await expect(reading).rejects.toMatchObject({kind:'unavailable'});
});
