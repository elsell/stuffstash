import { expect,it } from 'vitest';
import { PushSetup } from './PushSetup';
import { NotificationFailure } from './NotificationFailure';
const scope={serverId:'https://api.test',principalId:'user',tenantId:'tenant',inventoryId:'inventory'};
it('leaves preferences and registration untouched after permission denial',async()=>{
 const calls:string[]=[];
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){return false;},async nativeToken(){throw new Error('not allowed');}}, {async installationId(){return 'phone';},async remember(){calls.push('remember');},async list(){return [];},async forget(){calls.push('forget');}}, {async getByInstallation(){throw new Error('not allowed');},async register(){throw new Error('not allowed');},async revoke(){throw new Error('not allowed');}}, {record(){}});
 expect(await setup.enable(scope,{async refresh(){calls.push('refresh');},async savePushEnabled(){calls.push('enable');}})).toBe('denied');
 expect(calls).toEqual([]);
});
it('journals before registration and preserves cleanup metadata after an uncertain failure',async()=>{
 const calls:string[]=[];
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){return true;},async nativeToken(){return {transport:'apns',token:'abab'};}}, {async installationId(){return 'phone';},async remember(){calls.push('remember');},async list(){return [];},async forget(){calls.push('forget');}}, {async getByInstallation(){throw new NotificationFailure('not-found');},async register(){calls.push('register');throw new NotificationFailure('unavailable');},async revoke(){throw new Error('not allowed');}}, {record(){}});
 await expect(setup.enable(scope,{async refresh(){calls.push('refresh');},async savePushEnabled(){calls.push('enable');}})).rejects.toMatchObject({kind:'unavailable'});
 expect(calls).toEqual(['remember','register']);
});
it('refreshes revision on cleanup conflicts and leaves other users records alone',async()=>{
 const calls:string[]=[];
 let revision=1;
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){return false;},async nativeToken(){throw new Error('unused');}}, {
  async installationId(){return 'phone';},async remember(){},async list(){return [scope,{...scope,principalId:'other'}];},async forget(value){calls.push(`forget:${value.principalId}`);}
 },{
  async register(){throw new Error('unused');},async getByInstallation(){calls.push(`read:${revision}`);return {id:'device',installationId:'phone',transport:'apns',revision,active:true};},
  async revoke(_tenant,_inventory,_device,current){calls.push(`revoke:${current}`);if(revision===1){revision=2;throw new NotificationFailure('conflict');}return {id:'device',installationId:'phone',transport:'apns',revision:3,active:false};}
 },{record(){}});
 await setup.cleanup(scope.serverId,scope.principalId);
 expect(calls).toEqual(['read:1','revoke:1','read:2','revoke:2','forget:user']);
});
it('enables delivery only after journaling and registration succeed',async()=>{
 const calls:string[]=[];
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){return true;},async nativeToken(){return {transport:'apns',token:'abab'};}}, {
  async installationId(){return 'phone';},async remember(){calls.push('remember');},async list(){return [];},async forget(){}
 },{
  async getByInstallation(){throw new NotificationFailure('not-found');},async register(){calls.push('register');return {id:'device',installationId:'phone',transport:'apns',revision:1,active:true};},async revoke(){throw new Error('unused');}
 },{record(){}});
 await expect(setup.enable(scope,{async refresh(){calls.push('refresh');},async savePushEnabled(value){calls.push(`push:${value}`);}})).resolves.toBe('enabled');
 expect(calls).toEqual(['remember','register','refresh','push:true']);
});
