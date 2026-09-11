import { test } from 'node:test';
import assert from 'node:assert/strict';
import { generateKeyPairSync, verify } from 'node:crypto';
import { appleMaintenanceClient } from './apple-maintenance-client.mjs';
const { privateKey, publicKey } = generateKeyPairSync('ec',{namedCurve:'prime256v1'});
const options={privateKey,keyId:'ABCDEFGHIJ',issuerId:'00000000-0000-0000-0000-000000000001',now:()=>new Date('2026-09-11T00:00:00Z')};
test('signs bounded Apple JWT and sends authenticated request without redirects',async()=>{
 const api=appleMaintenanceClient({...options,fetch:async(url,init)=>{
  assert.equal(url,'https://api.appstoreconnect.apple.com/v1/profiles');assert.equal(init.redirect,'error');
  const token=init.headers.Authorization.slice(7);const [header,payload,sig]=token.split('.');
  assert.equal(JSON.parse(Buffer.from(header,'base64url')).alg,'ES256');
  const claims=JSON.parse(Buffer.from(payload,'base64url'));assert.equal(claims.exp-claims.iat,300);
  assert.equal(verify('sha256',Buffer.from(`${header}.${payload}`),{key:publicKey,dsaEncoding:'ieee-p1363'},Buffer.from(sig,'base64url')),true);
  return Response.json({data:[]});
 }});await api('/v1/profiles');
});
test('rejects foreign URLs without sending credentials',async()=>{
 const api=appleMaintenanceClient({...options,fetch:async()=>{assert.fail('must not fetch');}});
 await assert.rejects(api('https://example.com/v1/profiles'));
});
test('errors expose status without response secrets or retries',async()=>{
 let calls=0;const api=appleMaintenanceClient({...options,fetch:async()=>{calls++;return new Response('secret-response',{status:403});}});
 await assert.rejects(api('/v1/profiles'),{message:'Apple maintenance request failed (HTTP 403)'});assert.equal(calls,1);
});
