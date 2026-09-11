import type { NotificationObservability, NotificationEvent } from './NotificationObservability';
import { assertReadActive, type ReadRequest } from '../shared/ReadRequest';
import type { NotificationDeviceRepository } from './NotificationDeviceRepository';
import { NotificationFailure } from './NotificationFailure';
import type { PushDevicePort, PushRegistrationJournal, PushRegistrationScope } from './PushDevicePort';

type PushPreferences = {
 refresh(request?: ReadRequest): Promise<unknown>;
 savePushEnabled(enabled: boolean, request?: ReadRequest): Promise<unknown>;
};
export class PushSetup {
 private pending = false;
 constructor(private readonly device: PushDevicePort, private readonly journal: PushRegistrationJournal, private readonly repository: NotificationDeviceRepository, private readonly observer: NotificationObservability) {}
 enable(scope: PushRegistrationScope, preferences: PushPreferences, request: ReadRequest = {}): Promise<'enabled' | 'denied'> {
  return this.run('push-setup', request, async () => {
   validateScope(scope);
   const granted = await this.device.requestPermission(); assertReadActive(request.signal);
   if (!granted) return 'denied';
   const token = await this.device.nativeToken(); assertReadActive(request.signal);
   const installationId = await this.journal.installationId(); assertReadActive(request.signal);
   await this.journal.remember(scope); assertReadActive(request.signal);
   const current = await this.lookup(scope, installationId, request);
   const registration = await this.repository.register(scope.tenantId,scope.inventoryId,{installationId,...token,revision:current?.revision ?? 0},request.signal);
   assertReadActive(request.signal);
   if (!registration.active || registration.installationId !== installationId) throw new NotificationFailure('unavailable');
   await preferences.refresh(request); assertReadActive(request.signal);
   await preferences.savePushEnabled(true,request); assertReadActive(request.signal);
   return 'enabled';
  });
 }
 async hasRegistrations(serverId: string, request: ReadRequest = {}): Promise<boolean> {
  assertReadActive(request.signal);
  const records = await this.journal.list();
  assertReadActive(request.signal);
  return records.some(scope => scope.serverId === serverId);
 }
 cleanup(serverId: string, principalId: string, request: ReadRequest = {}): Promise<void> {
  return this.run('push-cleanup',request,async()=>{
   if (!serverId || !principalId) throw new NotificationFailure('invalid');
   const installationId = await this.journal.installationId(); assertReadActive(request.signal);
   const records = await this.journal.list(); assertReadActive(request.signal);
   for (const scope of records) {
    if (scope.serverId !== serverId || scope.principalId !== principalId) continue;
    validateScope(scope);
    await this.revoke(scope,installationId,request); assertReadActive(request.signal);
    await this.journal.forget(scope); assertReadActive(request.signal);
   }
  });
 }
 private async lookup(scope: PushRegistrationScope,installationId: string,request: ReadRequest) {
  assertReadActive(request.signal);
  try { const value = await this.repository.getByInstallation(scope.tenantId,scope.inventoryId,installationId,request.signal); assertReadActive(request.signal); return value; }
  catch(error){ assertReadActive(request.signal); if(error instanceof NotificationFailure && error.kind === 'not-found') return null; throw error; }
 }
 private async revoke(scope: PushRegistrationScope,installationId: string,request: ReadRequest) {
  for(let attempt=0;attempt<2;attempt++){
   const current=await this.lookup(scope,installationId,request);
   if(!current?.active) return;
   try {
    await this.repository.revoke(scope.tenantId,scope.inventoryId,current.id,current.revision,request.signal); assertReadActive(request.signal); return;
   } catch(error) {
    assertReadActive(request.signal);
    if(error instanceof NotificationFailure && error.kind === 'not-found') return;
    if(attempt===0 && error instanceof NotificationFailure && error.kind === 'conflict') continue;
    throw error;
   }
  }
 }
 private async run<T>(operationName: NotificationEvent['operation'], request: ReadRequest,operation:()=>Promise<T>):Promise<T>{
  assertReadActive(request.signal);
  if(this.pending) throw new NotificationFailure('conflict');
  this.pending=true;
  try { const value = await operation(); this.observer.record({operation:operationName,outcome:'succeeded'}); return value; } catch(error) { this.observer.record({operation:operationName,outcome:'failed'}); throw error; } finally{this.pending=false;}
 }
}
function validateScope(scope: PushRegistrationScope){if(!scope.serverId || !scope.principalId || !scope.tenantId || !scope.inventoryId)throw new NotificationFailure('invalid');}
