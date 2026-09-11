import {expect,it} from 'vitest';
import {OpenPushNotification} from './OpenPushNotification';
const payload={serverId:'https://api.test',principalId:'user',tenantId:'tenant',inventoryId:'inventory',notificationId:'notice',assetId:'forged'};
function fixture(){
 const calls:string[]=[];
 const command=new OpenPushNotification('https://api.test',{async getCurrentPrincipal(){calls.push('principal');return {id:'user'};}},{async open(tenant,inventory,id){calls.push([tenant,inventory,id].join(':'));return 'verified-item';}},{async execute(id){calls.push(`select:${id}`);return {selectedInventoryId:id};}});
 return {command,calls};
}
it('resolves the authorized item and selects its inventory, ignoring payload asset IDs',async()=>{
 const f=fixture();await expect(f.command.execute(payload)).resolves.toBe('verified-item');expect(f.calls).toEqual(['principal','tenant:inventory:notice','select:inventory']);
});
it('rejects wrong servers, accounts, and malformed routing hints before inbox access',async()=>{
 for(const input of [null,{}, {...payload,serverId:'https://other.test'},{...payload,principalId:'other'},{...payload,inventoryId:['inventory']},{...payload,notificationId:'\n'}]){
  const f=fixture();await expect(f.command.execute(input)).rejects.toThrow();expect(f.calls.filter(call=>call!=='principal')).toEqual([]);
 }
});
it('does not select an inventory after its notification read is cancelled',async()=>{
 const controller=new AbortController();let selected=false;
 const command=new OpenPushNotification('https://api.test',{async getCurrentPrincipal(){return {id:'user'};}},{async open(){controller.abort();return 'item';}},{async execute(id){selected=true;return {selectedInventoryId:id};}});
 await expect(command.execute(payload,{signal:controller.signal})).rejects.toThrow();expect(selected).toBe(false);
});
it('does not select an inventory for a notification the server rejects',async()=>{
 const command=new OpenPushNotification('https://api.test',{async getCurrentPrincipal(){return {id:'user'};}},{async open(){throw new Error('not found');}},{async execute(){throw new Error('must not select');}});
 await expect(command.execute(payload)).rejects.toThrow('not found');
});
