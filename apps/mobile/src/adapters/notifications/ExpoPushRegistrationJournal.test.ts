import { expect,it } from 'vitest';
import { ExpoPushRegistrationJournal } from './ExpoPushRegistrationJournal';
const scope={serverId:'https://api.test',principalId:'user',tenantId:'tenant',inventoryId:'inventory'};
class Storage {
 readonly WHEN_UNLOCKED_THIS_DEVICE_ONLY=7;
 options:Record<string,unknown>|undefined;
 value:string|null=null;
 fail=false;
 async getItemAsync(_key:string,options?:Record<string,unknown>){this.options=options;return this.value;}
 async setItemAsync(_key:string,value:string,options?:Record<string,unknown>){this.options=options;if(this.fail)throw new Error('private storage error');this.value=value;}
}
it('keeps installation identity across restart and serializes concurrent scoped changes',async()=>{
 const store=new Storage();let ids=0;
 const journal=new ExpoPushRegistrationJournal(store,()=>`installation-${++ids}`);
 expect(await Promise.all([journal.installationId(),journal.installationId()])).toEqual(['installation-1','installation-1']);
 await Promise.all([journal.remember({...scope,token:'must-not-save'} as typeof scope),journal.remember({...scope,inventoryId:'other'})]);
 expect(store.value).not.toContain('must-not-save');
 expect(store.options).toEqual({keychainService:'stuffstash.mobile.push',keychainAccessible:7});
 const restarted=new ExpoPushRegistrationJournal(store,()=>`installation-${++ids}`);
 expect(await restarted.installationId()).toBe('installation-1');
 expect(await restarted.list()).toHaveLength(2);
 await restarted.forget(scope);
 expect(await restarted.list()).toEqual([{...scope,inventoryId:'other'}]);
});
it('preserves data after write failure or malformed storage',async()=>{
 const store=new Storage();const journal=new ExpoPushRegistrationJournal(store,()=> 'installation');
 await journal.remember(scope);const saved=store.value;store.fail=true;
 await expect(journal.forget(scope)).rejects.toMatchObject({kind:'unavailable'});
 expect(store.value).toBe(saved);store.fail=false;
 expect(await journal.list()).toEqual([scope]);
 store.value='broken document';
 await expect(journal.installationId()).rejects.toMatchObject({kind:'unavailable'});
 expect(store.value).toBe('broken document');
});
