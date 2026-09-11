import type { PushRegistrationJournal, PushRegistrationScope } from '../../application/notifications/PushDevicePort';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';

type SecureStorage = {
 getItemAsync(key:string,options?:Record<string,unknown>):Promise<string|null>;
 setItemAsync(key:string,value:string,options?:Record<string,unknown>):Promise<void>;
 readonly WHEN_UNLOCKED_THIS_DEVICE_ONLY?:number;
};
type Journal = { version:1; installationId:string; scopes:PushRegistrationScope[] };
const storageKey='stuffstash.mobile.push.registrations';
export class ExpoPushRegistrationJournal implements PushRegistrationJournal {
 private tail:Promise<void>=Promise.resolve();
 constructor(private readonly storage:SecureStorage,private readonly generateId:()=>string){}
 installationId():Promise<string>{return this.run(async()=>{
  const current=await this.read();if(current)return current.installationId;
  const created=this.empty();await this.write(created);return created.installationId;
 });}
 remember(scope:PushRegistrationScope):Promise<void>{return this.run(async()=>{
  const clean=copyScope(scope);const current=await this.read()??this.empty();
  if(current.scopes.some(value=>sameScope(value,clean)))return;
  await this.write({...current,scopes:[...current.scopes,clean]});
 });}
 list():Promise<readonly PushRegistrationScope[]>{return this.run(async()=>{
  const current=await this.read();return current?.scopes.map(copyScope)??[];
 });}
 forget(scope:PushRegistrationScope):Promise<void>{return this.run(async()=>{
  const clean=copyScope(scope);const current=await this.read();if(!current)return;
  const scopes=current.scopes.filter(value=>!sameScope(value,clean));
  if(scopes.length!==current.scopes.length)await this.write({...current,scopes});
 });}
 private empty():Journal {const installationId=this.generateId();if(!validID(installationId))throw new Error('Invalid installation identity');return {version:1,installationId,scopes:[]};}
 private async read():Promise<Journal|null>{
  const raw=await this.storage.getItemAsync(storageKey,this.options());if(raw===null)return null;
  const value:unknown=JSON.parse(raw);
  if(!isRecord(value)||value.version!==1||!validID(value.installationId)||!Array.isArray(value.scopes))throw new Error('Invalid registration journal');
  return {version:1,installationId:value.installationId,scopes:value.scopes.map(copyScope)};
 }
 private write(value:Journal){return this.storage.setItemAsync(storageKey,JSON.stringify(value),this.options());}
 private options():Record<string,unknown>{return {keychainService:'stuffstash.mobile.push',...(this.storage.WHEN_UNLOCKED_THIS_DEVICE_ONLY!==undefined?{keychainAccessible:this.storage.WHEN_UNLOCKED_THIS_DEVICE_ONLY}:{})};}
 private run<T>(operation:()=>Promise<T>):Promise<T>{
  const work=this.tail.then(operation);
  this.tail=work.then(()=>undefined,()=>undefined);
  return work.catch(()=>{throw new NotificationFailure('unavailable');});
 }
}
function validID(value:unknown):value is string{return typeof value==='string'&&value.trim().length>0&&value.length<=128;}
function isRecord(value:unknown):value is Record<string,unknown>{return typeof value==='object'&&value!==null&&!Array.isArray(value);}
function copyScope(value:unknown):PushRegistrationScope{
 if(!isRecord(value)||typeof value.serverId!=='string'||!value.serverId.trim()||value.serverId.length>2048||!validID(value.principalId)||!validID(value.tenantId)||!validID(value.inventoryId))throw new Error('Invalid registration scope');
 return {serverId:value.serverId,principalId:value.principalId,tenantId:value.tenantId,inventoryId:value.inventoryId};
}
function sameScope(left:PushRegistrationScope,right:PushRegistrationScope){return left.serverId===right.serverId&&left.principalId===right.principalId&&left.tenantId===right.tenantId&&left.inventoryId===right.inventoryId;}
