import { expect, it } from 'vitest';
import { PushSetup } from './PushSetup';
import { NotificationFailure } from './NotificationFailure';
const scope={serverId:'https://api.test',principalId:'user',tenantId:'tenant',inventoryId:'inventory'};
function fixture() {
 const calls:string[]=[];
 let permission=true, active=true, revision=1, storedToken='old', nativeToken='old';
 const setup=new PushSetup({async permissionGranted(){return permission;},async requestPermission(){throw new Error('must not prompt');},async nativeToken(){return {transport:'apns',token:nativeToken};}}, {
  async installationId(){return 'phone';},async remember(){throw new Error('already journaled');},async list(){return [scope,{...scope,principalId:'other'},{...scope,serverId:'https://other.test'}];},async forget(){throw new Error('must retain intent');}
 }, {
  async getByInstallation(){calls.push('lookup');return {id:'device',installationId:'phone',transport:'apns',revision,active};},
  async register(tenant,inventory,input){
   expect([tenant,inventory]).toEqual(['tenant','inventory']); calls.push(`register:${input.revision}`);
   if(!(active&&storedToken===input.token&&input.revision===0)) {
    if(input.revision!==revision) throw new NotificationFailure('conflict');
    revision++;storedToken=input.token;active=true;
   }
   return {id:'device',installationId:'phone',transport:'apns',revision,active};
  },
  async revoke(_tenant,_inventory,_id,expected){expect(expected).toBe(revision);calls.push('revoke');active=false;revision++;return {id:'device',installationId:'phone',transport:'apns',revision,active};}
 },{record(){}});
 return {setup,calls,setPermission(value:boolean){permission=value;},setToken(value:string){nativeToken=value;},revision:()=>revision};
}
it('preserves unchanged registrations and refreshes rotated tokens only within the signed-in scope',async()=>{
 const f=fixture();await f.setup.reconcile(scope.serverId,scope.principalId);
 expect(f.calls).toEqual(['register:0']);expect(f.revision()).toBe(1);
 f.calls.length=0;f.setToken('new');await f.setup.reconcile(scope.serverId,scope.principalId);
 expect(f.calls).toEqual(['register:0','lookup','register:1']);expect(f.revision()).toBe(2);
});
it('revokes permission loss without forgetting prior device setup and restores registration when permission returns',async()=>{
 const f=fixture();f.setPermission(false);await f.setup.reconcile(scope.serverId,scope.principalId);
 expect(f.calls).toEqual(['lookup','revoke']);
 f.calls.length=0;f.setPermission(true);await f.setup.reconcile(scope.serverId,scope.principalId);
 expect(f.calls).toEqual(['register:0','lookup','register:2']);expect(f.revision()).toBe(3);
});
