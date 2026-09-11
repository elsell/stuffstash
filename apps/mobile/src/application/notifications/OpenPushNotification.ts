import type { CurrentPrincipalRepository } from '../settings/SettingsQuery';
import type { NotificationInboxQueries } from './NotificationInboxQueries';
import type { SelectInventoryCommand } from '../home/SelectInventoryCommand';
import { normalizeInstanceUrl } from '../onboarding/ConnectionProfile';
import { assertReadActive, type ReadRequest } from '../shared/ReadRequest';
import { NotificationFailure } from './NotificationFailure';

export class OpenPushNotification {
 constructor(private readonly serverId:string, private readonly principals:CurrentPrincipalRepository,
  private readonly inbox:Pick<NotificationInboxQueries,'open'>, private readonly select:Pick<SelectInventoryCommand,'execute'>) {}
 async execute(payload:unknown,request:ReadRequest={}):Promise<string>{
  assertReadActive(request.signal);
  const target=parseTarget(payload);
  if(target.serverId!==normalizeInstanceUrl(this.serverId)) throw new NotificationFailure('wrong-recipient');
  const principal=await this.principals.getCurrentPrincipal(request);assertReadActive(request.signal);
  if(principal.id!==target.principalId) throw new NotificationFailure('wrong-recipient');
  const assetId=await this.inbox.open(target.tenantId,target.inventoryId,target.notificationId,request);assertReadActive(request.signal);
  await this.select.execute(target.inventoryId,request);assertReadActive(request.signal);
  return assetId;
 }
}
function parseTarget(payload:unknown){
 if(!payload || typeof payload!=='object' || Array.isArray(payload))throw new NotificationFailure('invalid');
 const value=payload as Record<string,unknown>;
 const read=(key:string)=>{const field=value[key];if(typeof field!=='string'||field.length>2048||!field.length||field.trim()!==field||/[\x00-\x1f\x7f]/.test(field))throw new NotificationFailure('invalid');return field;};
 const serverId=read('serverId');
 try {const url=new URL(serverId);if(!['http:','https:'].includes(url.protocol)||url.username||url.password||url.search||url.hash)throw new Error();}
 catch {throw new NotificationFailure('invalid');}
 return {serverId:normalizeInstanceUrl(serverId),principalId:read('principalId'),tenantId:read('tenantId'),inventoryId:read('inventoryId'),notificationId:read('notificationId')};
}
