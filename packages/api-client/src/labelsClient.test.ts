import {expect,it} from 'vitest';
import {LabelsClient} from './labelsClient';
import {parseLabelLink} from './labelLink';
it('resolves migrated identities only on configured authenticated API; refuses unsafe responses', async () => {
  const requests:Request[]=[];
  const client = new LabelsClient({baseUrl:'https://api.example',tokenProvider:()=> 'session',fetch:async request => {
    const r=new Request(request); requests.push(r);
    return new Response(JSON.stringify({data:{assetId:'asset',tenantId:'tenant',inventoryId:'inventory',lifecycleState:'archived'}}),{status:200,headers:{'content-type':'application/json'}});
  }});
  const ref=parseLabelLink('https://obsolete.example/l/v1/01ARZ3NDEKTSV4RRFFQ69G5FAV/01ARZ3NDEKTSV4RRFFQ69G5FAW');
  expect((await client.resolve(ref.instanceId,ref.labelId)).lifecycleState).toBe('archived');
  expect(requests[0].url).toBe(`https://api.example/labels/v1/${ref.instanceId}/${ref.labelId}`);
  expect(requests[0].headers.get('authorization')).toBe('Bearer session');
  expect(requests[0].redirect).toBe('error');
});
it.each([401,403,404])('does not return asset or artifact data after access is rejected (%s)', async status=>{
  const client=new LabelsClient({baseUrl:'https://api.example',tokenProvider:()=> 'expired',fetch:async()=>new Response(JSON.stringify({error:{code:'forbidden'}}),{status,headers:{'content-type':'application/json'}})});
  await expect(client.resolve('instance','label')).rejects.toMatchObject({status});
  await expect(client.content('tenant','inventory','render')).rejects.toMatchObject({status});
});
