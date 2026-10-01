import assert from 'node:assert/strict';
import test from 'node:test';
import { publishPlayNotes } from './play-store-notes.mjs';

const target = { packageName: 'com.example.app', track: 'internal', versionCode: '42', language: 'en-US', notes: '- Shared fix' };
function fixture(options = {}) {
  let live = { track: 'internal', releases: [
    { name: '1.2.3', versionCodes: ['42'], status: 'inProgress', userFraction: 0.2, countryTargeting: {countries:['US']}, releaseNotes:[{language:'fr-FR',text:'Bonjour'}] },
    { name: 'old', versionCodes: ['41'], status: 'completed' },
  ] };
  if (options.wrong) live.releases[0].versionCodes = ['43'];
  if (options.multiple) live.releases[0].versionCodes.push('44');
  if (options.same) live.releases[0].releaseNotes.push({language:'en-US',text:target.notes});
  const original = structuredClone(live);
  const edits = new Map(); let next = 0; let commits = 0; let updates = 0;
  const client = { async request(path, {method='GET',body}={}) {
    assert.ok(path.startsWith('/applications/com.example.app/edits'));
    if (method === 'POST' && path.endsWith('/edits')) { const id=String(++next); edits.set(id,structuredClone(live)); return {id}; }
    const id = path.split('/edits/')[1].split(/[/:]/)[0];
    if (method === 'DELETE') { edits.delete(id); return null; }
    if (method === 'POST' && path.endsWith(':commit')) {
      commits++; live=edits.get(id); edits.delete(id);
      if (options.badReadback) live.releases[0].releaseNotes=[];
      return {};
    }
    if (method === 'PUT') { updates++; edits.set(id,structuredClone(body)); return structuredClone(body); }
    if (options.denied) throw new Error('permission denied');
    return structuredClone(edits.get(id));
  }};
  return {client, original, live:()=>live, edits, counts:()=>({commits,updates})};
}
test('notes-only publication preserves all other track state and verifies committed notes', async () => {
  const f=fixture(); await publishPlayNotes(f.client,target);
  const expected=structuredClone(f.original); expected.releases[0].releaseNotes.push({language:'en-US',text:target.notes});
  assert.deepEqual(f.live(),expected); assert.deepEqual(f.counts(),{commits:1,updates:1}); assert.equal(f.edits.size,0);
});
test('already matching notes are idempotent without a commit', async () => {
  const f=fixture({same:true}); await publishPlayNotes(f.client,target); assert.deepEqual(f.counts(),{commits:0,updates:0}); assert.equal(f.edits.size,0);
});
for (const options of [{wrong:true},{multiple:true},{denied:true},{badReadback:true}]) test(`fails closed: ${JSON.stringify(options)}`, async () => {
 const f=fixture(options); await assert.rejects(publishPlayNotes(f.client,target)); assert.equal(f.edits.size,0);
 if (!options.badReadback) assert.equal(f.counts().commits,0);
});
test('unsafe targets are rejected before API calls', async () => {
 for (const changes of [{versionCode:'latest'},{packageName:'../other'},{track:'../production'},{notes:'x'.repeat(501)},{language:'bad/locale'}]) {
  const client={request(){assert.fail('unexpected API call');}};
  await assert.rejects(publishPlayNotes(client,{...target,...changes}));
 }
});

test('malformed version-code representation cannot select or mutate a release', async () => {
 let writes = 0;
 const client = { async request(path, {method='GET'}={}) {
  if(method === 'POST' && path.endsWith('/edits')) return {id:'edit'};
  if(method === 'DELETE') return null;
  if(method !== 'GET') writes++;
  return {track:'internal',releases:[{versionCodes:'4',status:'completed'}]};
 }};
 await assert.rejects(publishPlayNotes(client,{...target,versionCode:'4'}));
 assert.equal(writes,0);
});
