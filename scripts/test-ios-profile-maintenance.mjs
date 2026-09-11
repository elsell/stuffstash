import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { test } from 'node:test';
import { inspectProfileRepair, repairPushProfile } from './ios-profile-maintenance.mjs';
const origin = 'https://api.appstoreconnect.apple.com';
const certificate = Buffer.from('existing public certificate');
const digest = createHash('sha256').update(certificate).digest('hex');
const expected = { teamId: 'TEAM123456', bundleId: 'org.stuffstash.mobile' };
const profile = { ...expected, certificateSha256: [digest] };
function fixture() {
  const calls = [];
  const responses = {
    '/v1/bundleIds?filter%5Bidentifier%5D=org.stuffstash.mobile&limit=200': {data:[{id:'bundle',attributes:{identifier:expected.bundleId}}]},
    '/v1/certificates?limit=200': {data:[{id:'cert',attributes:{certificateType:'DISTRIBUTION',certificateContent:certificate.toString('base64'),expirationDate:'2027-01-01T00:00:00Z'}}]},
    '/v1/bundleIds/bundle/bundleIdCapabilities?limit=200': {data:[]}
  };
  return { calls, responses, api: async (path, method='GET', body) => {
    calls.push({path,method,body});
    if (method==='POST') return {data:{id:'created',attributes:{profileContent:'profile'}}};
    assert.ok(responses[path], `Unexpected path ${path}`); return responses[path];
  }};
}
const now = new Date('2026-09-11T00:00:00Z');
test('inspection resolves the existing bundle and certificate without writes', async () => {
  const f=fixture(); const plan=await inspectProfileRepair(f.api,profile,expected,now);
  assert.deepEqual(plan,{bundleId:'bundle',certificateId:'cert',pushEnabled:false});
  assert.ok(f.calls.every(c=>c.method==='GET'));
});
test('rejects another team or bundle before requesting Apple data',async()=>{
  for(const field of ['teamId','bundleId']) {const f=fixture(); await assert.rejects(inspectProfileRepair(f.api,{...profile,[field]:'other'},expected,now));assert.equal(f.calls.length,0);}
});
test('rejects missing, expired or development certificates',async()=>{
  for(const attributes of [{certificateContent:'b3RoZXI='},{expirationDate:'2020-01-01T00:00:00Z'},{certificateType:'DEVELOPMENT'}]) {
    const f=fixture(); Object.assign(f.responses['/v1/certificates?limit=200'].data[0].attributes,attributes);
    await assert.rejects(inspectProfileRepair(f.api,profile,expected,now));
  }
});
test('rejects foreign pagination and ambiguous bundle matches',async()=>{
  const f=fixture(); f.responses['/v1/certificates?limit=200'].links={next:'https://example.com/steal'};
  await assert.rejects(inspectProfileRepair(f.api,profile,expected,now));
  const other=fixture(); other.responses[Object.keys(other.responses)[0]].data.push({id:'second',attributes:{identifier:expected.bundleId}});
  await assert.rejects(inspectProfileRepair(other.api,profile,expected,now));
});
test('repair only enables push and creates a profile referencing the existing certificate',async()=>{
  const f=fixture(); const plan=await inspectProfileRepair(f.api,profile,expected,now);
  await repairPushProfile(f.api,plan,'Stuff Stash push profile');
  const writes=f.calls.filter(c=>c.method==='POST');assert.equal(writes.length,2);
  assert.equal(writes[0].body.data.attributes.capabilityType,'PUSH_NOTIFICATIONS');
  assert.equal(writes[1].path,'/v1/profiles');
  assert.equal(writes[1].body.data.attributes.profileType,'IOS_APP_STORE');
  assert.deepEqual(writes[1].body.data.relationships.certificates.data,[{type:'certificates',id:'cert'}]);
});
test('already enabled push is preserved and failed creation is not retried',async()=>{
  let calls=0;await assert.rejects(repairPushProfile(async(path)=>{calls++;assert.equal(path,'/v1/profiles');throw new Error('timeout');},{bundleId:'bundle',certificateId:'cert',pushEnabled:true},'profile'));
  assert.equal(calls,1);
});
