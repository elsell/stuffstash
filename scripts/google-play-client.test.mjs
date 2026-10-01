import assert from 'node:assert/strict';
import test from 'node:test';
import { generateKeyPairSync, verify } from 'node:crypto';
import { GooglePlayClient, serviceAccountToken } from './google-play-client.mjs';

const { privateKey, publicKey } = generateKeyPairSync('rsa', { modulusLength: 2048 });
const credentials = { type: 'service_account', client_email: 'publisher@example.iam.gserviceaccount.com', private_key: privateKey.export({format:'pem',type:'pkcs8'}) };
test('Google OAuth assertion has the reviewed audience, scope and bounded lifetime', async () => {
 const token=await serviceAccountToken(credentials,{now:()=>1000,fetcher:async (url,options)=>{
  assert.equal(url,'https://oauth2.googleapis.com/token'); assert.equal(options.redirect,'error');
  const form=new URLSearchParams(options.body); const [header,body,signature]=form.get('assertion').split('.');
  assert.equal(JSON.parse(Buffer.from(header,'base64url')).alg,'RS256');
  assert.deepEqual(JSON.parse(Buffer.from(body,'base64url')), {iss:credentials.client_email,scope:'https://www.googleapis.com/auth/androidpublisher',aud:url,iat:1000,exp:1300});
  assert.ok(verify('RSA-SHA256',Buffer.from(header+'.'+body),publicKey,Buffer.from(signature,'base64url')));
  return Response.json({token_type:'Bearer',access_token:'access'});
 }});
 assert.equal(token,'access');
});
test('credential redirection and malformed keys are rejected before sending', async () => {
 for(const change of [{token_uri:'https://evil.example/token'},{type:'authorized_user'},{private_key:'SECRET'},{client_email:'other@example.com'}]) {
  await assert.rejects(serviceAccountToken({...credentials,...change},{fetcher:()=>assert.fail('credential leaked')}));
 }
});
test('API transport is origin constrained, authenticated and does not follow redirects', async () => {
 let calls=0;
 const client=new GooglePlayClient('access',{fetcher:async (url,options)=>{
  calls++; assert.equal(url,'https://androidpublisher.googleapis.com/androidpublisher/v3/applications/com.example.app/edits');
  assert.equal(options.headers.Authorization,'Bearer access'); assert.equal(options.redirect,'error');
  return Response.json({id:'123'});
 }});
 await client.request('/applications/com.example.app/edits',{method:'POST',body:{}});
 for(const path of ['https://evil.example','//evil.example','/applications/../other/edits','/applications/com.example.app/edits?redirect=evil']) assert.throws(()=>client.request(path));
 assert.equal(calls,1);
});
test('permission, transient and redirect failures are sanitized and writes are not retried', async () => {
 for(const status of [302,401,403,429,500]) {
  let calls=0;
  const client=new GooglePlayClient('SECRET',{fetcher:async ()=>{calls++; return new Response('SECRET',{status});}});
  await assert.rejects(client.request('/applications/com.example.app/edits',{method:'POST',body:{}}),error=>!error.message.includes('SECRET'));
  assert.equal(calls,1);
 }
 const client=new GooglePlayClient('access',{fetcher:async ()=>{throw new Error('SECRET');}});
 await assert.rejects(client.request('/applications/com.example.app/edits'),{message:'Google Play connection failed'});
});
test('malformed OAuth success response fails closed', async () => {
 await assert.rejects(serviceAccountToken(credentials,{fetcher:async ()=>Response.json({access_token:'SECRET'})}),{message:'Invalid Google Play token'});
});
