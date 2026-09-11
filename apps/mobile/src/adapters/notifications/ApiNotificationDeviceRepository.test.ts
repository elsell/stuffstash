import { expect,it } from 'vitest';
import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiNotificationDeviceRepository } from './ApiNotificationDeviceRepository';
it('maps token-free metadata and safe conflict errors',async()=>{
 let fail=false;
 const repository=new ApiNotificationDeviceRepository(new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'access',fetch:async()=>fail?Response.json({error:{code:'conflict',message:'sensitive provider body'}},{status:409}):Response.json({data:{id:'device',installationId:'phone',transport:'apns',revision:2,active:false,token:'must-not-escape'},meta:{}})}));
 await expect(repository.getByInstallation('tenant','inventory','phone')).resolves.toEqual({id:'device',installationId:'phone',transport:'apns',revision:2,active:false});
 fail=true;
 await expect(repository.revoke('tenant','inventory','device',2)).rejects.toMatchObject({kind:'conflict'});
});
it('does not accept a response after cancellation',async()=>{
 const controller=new AbortController();
 const repository=new ApiNotificationDeviceRepository(new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'access',fetch:async()=>{controller.abort();return Response.json({data:{id:'device',installationId:'phone',transport:'apns',revision:1,active:true},meta:{}});}}));
 await expect(repository.register('tenant','inventory',{installationId:'phone',transport:'apns',token:'abab',revision:0},controller.signal)).rejects.toMatchObject({name:'AbortError'});
});
