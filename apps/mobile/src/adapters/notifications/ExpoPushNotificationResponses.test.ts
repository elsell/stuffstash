import {expect,it} from 'vitest';
import {ExpoPushNotificationResponses,type NativePushResponse} from './ExpoPushNotificationResponses';
const response=(id:string)=>({actionIdentifier:'default',notification:{request:{identifier:id,content:{data:{notificationId:id}}}}});
it('deduplicates launch/live responses and retains consumption across subscriptions',async()=>{
 let listener=(_value:NativePushResponse)=>{};const values:unknown[]=[];
 const source=new ExpoPushNotificationResponses({async getLastNotificationResponseAsync(){return response('a');},addNotificationResponseReceivedListener(value){listener=value;return {remove(){}};}},'default');
 const stop=source.subscribe(value=>{values.push(value);},()=>{throw new Error('unexpected');});await Promise.resolve();listener(response('a'));expect(values).toHaveLength(1);stop();
 const stopAgain=source.subscribe(value=>{values.push(value);},()=>{});await Promise.resolve();expect(values).toHaveLength(1);listener(response('b'));expect(values).toHaveLength(2);stopAgain();
});
it('ignores delayed launch reads after a newer live tap and ignores callbacks after disposal',async()=>{
 let resolve!:(value:ReturnType<typeof response>)=>void;let listener=(_value:NativePushResponse)=>{};const values:unknown[]=[];
 const source=new ExpoPushNotificationResponses({getLastNotificationResponseAsync(){return new Promise(done=>{resolve=done;});},addNotificationResponseReceivedListener(value){listener=value;return {remove(){}};}},'default');
 const stop=source.subscribe(value=>{values.push(value);},()=>{});listener(response('new'));resolve(response('old'));await Promise.resolve();stop();listener(response('later'));expect(values).toEqual([{notificationId:'new'}]);
});
it('allows a later explicit tap on the same notification after its previous handling finishes',async()=>{
 let listener=(_value:NativePushResponse)=>{};let calls=0;
 const source=new ExpoPushNotificationResponses({async getLastNotificationResponseAsync(){return null;},addNotificationResponseReceivedListener(value){listener=value;return {remove(){}};}},'default');
 const stop=source.subscribe(()=>{calls++;},()=>{});listener(response('a'));await new Promise(resolve=>setTimeout(resolve,0));listener(response('a'));expect(calls).toBe(2);stop();
});
