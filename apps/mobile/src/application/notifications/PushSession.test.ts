import { expect, it } from 'vitest';
import { PushSession } from './PushSession';
import type { PushRegistrationScope } from './PushDevicePort';
import { PushSetup } from './PushSetup';
import { NotificationFailure } from './NotificationFailure';
it('binds registration to the authenticated principal and cleans it before disconnecting',async()=>{
 const events:string[]=[];
 let remembered:PushRegistrationScope[]=[];
 const device={id:'device',installationId:'phone',transport:'apns' as const,revision:1,active:true};
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){return true;},async nativeToken(){return {transport:'apns',token:'abcd'};}},{
  async installationId(){return 'phone';},async remember(scope){remembered.push(scope);},async list(){return remembered;},async forget(){remembered=[];}
 },{async register(){events.push('register');return device;},async getByInstallation(){return device;},async revoke(){events.push('revoke');return {...device,active:false,revision:2};}},{record(){}});
 const session=new PushSession('https://api.test', {async getCurrentPrincipal(){return {id:'signed-in-user'};}},setup,()=>({async refresh(){},async savePushEnabled(){events.push('enable');}}));
 expect(await session.enable('tenant','inventory')).toBe('enabled');
 expect(remembered).toEqual([{serverId:'https://api.test',principalId:'signed-in-user',tenantId:'tenant',inventoryId:'inventory'}]);
 await session.disconnect(async()=>{events.push('disconnect');});
 expect(events).toEqual(['register','enable','revoke','disconnect']);
});
it('does not disconnect when cleanup fails or mutate after identity resolution is cancelled',async()=>{
 let disconnected=false;
 const controller=new AbortController();
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){throw new Error('unexpected');},async nativeToken(){throw new Error('unexpected');}}, {
  async installationId(){throw new NotificationFailure('unavailable');},async remember(){},async list(){throw new NotificationFailure('unavailable');},async forget(){}
 },{async register(){throw new Error('unexpected');},async getByInstallation(){throw new Error('unexpected');},async revoke(){throw new Error('unexpected');}},{record(){}});
 const session=new PushSession('https://api.test',{async getCurrentPrincipal(){controller.abort();return {id:'user'};}},setup,()=>({async refresh(){},async savePushEnabled(){}}));
 await expect(session.enable('tenant','inventory',{signal:controller.signal})).rejects.toBeDefined();
 await expect(session.disconnect(async()=>{disconnected=true;})).rejects.toMatchObject({kind:'unavailable'});
 expect(disconnected).toBe(false);
});

it('allows disconnect without network identity lookup when this server has no registrations',async()=>{
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){return false;},async nativeToken(){throw new Error('unused');}}, {
  async installationId(){throw new Error('unused');},async remember(){},async list(){return [];},async forget(){}
 },{async register(){throw new Error('unused');},async getByInstallation(){throw new Error('unused');},async revoke(){throw new Error('unused');}},{record(){}});
 const session=new PushSession('https://api.test',{async getCurrentPrincipal(){throw new Error('offline');}},setup,()=>({async refresh(){},async savePushEnabled(){}}));
 let disconnected=false;
 await session.disconnect(async()=>{disconnected=true;});
 expect(disconnected).toBe(true);
 await expect(session.enable('tenant','inventory')).rejects.toMatchObject({kind:'conflict'});
});
it('keeps setup excluded until the disconnect action finishes',async()=>{
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){return false;},async nativeToken(){throw new Error('unused');}}, {
  async installationId(){throw new Error('unused');},async remember(){},async list(){return [];},async forget(){}
 },{async register(){throw new Error('unused');},async getByInstallation(){throw new Error('unused');},async revoke(){throw new Error('unused');}},{record(){}});
 const session=new PushSession('https://api.test',{async getCurrentPrincipal(){throw new Error('unexpected');}},setup,()=>({async refresh(){},async savePushEnabled(){}}));
 let finish!:()=>void;
 const held=new Promise<void>(resolve=>{finish=resolve;});
 const disconnect=session.disconnect(()=>held);
 await expect(session.enable('tenant','inventory')).rejects.toMatchObject({kind:'conflict'});
 finish(); await disconnect;
});
it('cancels and joins background reconciliation before disconnecting',async()=>{
 const events:string[]=[];
 let started!:()=>void;
 const ready=new Promise<void>(resolve=>{started=resolve;});
 const setup=new PushSetup({async permissionGranted(){return true;},async requestPermission(){return true;},async nativeToken(){return {transport:'apns',token:'token'};}}, {
  async installationId(){return 'phone';},async remember(){},async list(){return [{serverId:'https://api.test',principalId:'user',tenantId:'tenant',inventoryId:'inventory'}];},async forget(){events.push('forgot');}
 },{
  async register(_tenant,_inventory,_input,signal){started();await new Promise<void>(resolve=>signal!.addEventListener('abort',()=>resolve(),{once:true}));events.push('cancelled');throw new Error('aborted');},
  async getByInstallation(){throw new NotificationFailure('not-found');},async revoke(){throw new Error('unexpected');}
 },{record(){}});
 const session=new PushSession('https://api.test',{async getCurrentPrincipal(){return {id:'user'};}},setup,()=>({async refresh(){},async savePushEnabled(){}}));
 const background=session.reconcile().catch(()=>{});await ready;
 await session.disconnect(async()=>{events.push('logout');});await background;
 expect(events).toEqual(['cancelled','forgot','logout']);
});
