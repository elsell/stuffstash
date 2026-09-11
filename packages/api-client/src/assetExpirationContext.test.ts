import { expect, it } from 'vitest';
import { StuffStashClient } from './stuffStashClient';

it('preserves personal detail context and does not invent it for older responses', async () => {
 const context = {state:'expired', trackingEnabled:false, advanceDays:14, timezone:'America/New_York'};
 let includeContext = true;
 const client = new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'token',fetch:async()=>Response.json({data:{id:'bottle',tenantId:'home',inventoryId:'main',kind:'item',title:'Bottle',description:'',parentAssetId:null,lifecycleState:'active',customFields:{},tags:[],expiration:{date:'2001-01',precision:'month'},...(includeContext?{expirationContext:context}:{})},meta:{}})});
 expect(await client.getAsset('home','main','bottle')).toMatchObject({expirationContext:context});
 includeContext=false;
 expect(await client.getAsset('home','main','bottle')).not.toHaveProperty('expirationContext');
});
