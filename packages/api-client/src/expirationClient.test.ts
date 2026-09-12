import { expect, it } from 'vitest';
import { StuffStashClient, StuffStashAPIError } from './stuffStashClient';

it('transports scoped expiry filters and preserves complete counts, precision and pagination', async () => {
 let request!: Request;
 const data={items:[{id:'item',expiration:{date:'2028-02',precision:'month'},ancestorPath:[{id:'cabinet',title:'Cabinet'}]}],counts:{soon:2,expired:1,all:300},timezone:'America/New_York'};
 const client=new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'access',fetch:async(input,init)=>{request=new Request(input,init);return Response.json({data,meta:{pagination:{hasMore:true,nextCursor:'next',limit:1}}});}});
 const page=await client.expiration.list('tenant','inventory',{mode:'all',q:'lens',tagIds:['a','b'],locationId:'cabinet',fromDate:'2028-02-01',throughDate:'2028-02-29',limit:1});
 expect(page).toEqual({...data,pagination:{hasMore:true,nextCursor:'next',limit:1}});
 expect(request.headers.get('Authorization')).toBe('Bearer access');
 const url=new URL(request.url);expect(url.pathname).toBe('/tenants/tenant/inventories/inventory/expiration-assets');
 expect(url.searchParams.get('locationId')).toBe('cabinet');expect(url.searchParams.getAll('tagIds')).toEqual(['a','b']);
});
it('does not turn an expiration server failure into an empty inventory',async()=>{
 const client=new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'access',fetch:async()=>Response.json({error:{code:'forbidden',message:'Denied'}},{status:403})});
 await expect(client.expiration.list('tenant','inventory')).rejects.toBeInstanceOf(StuffStashAPIError);
});

it.each([
 {counts:{all:1,soon:-1,expired:0},pagination:{limit:30,hasMore:false,nextCursor:null}},
 {counts:{all:1,soon:0,expired:2},pagination:{limit:30,hasMore:false,nextCursor:null}},
 {counts:{all:1,soon:1,expired:0},pagination:{limit:30,hasMore:true,nextCursor:null}},
 {counts:{all:1,soon:1,expired:0},pagination:{limit:30}},
])('rejects incomplete count/pagination contracts',async ({counts,pagination})=>{
 const client=new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'token',fetch:async()=>Response.json({data:{items:[],counts,timezone:'UTC'},meta:{pagination}})});
 await expect(client.expiration.list('tenant','inventory')).rejects.toThrow();
});
