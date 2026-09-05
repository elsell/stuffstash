import { describe, expect, it } from 'vitest';
import { CaseAPIRepository } from './caseRepository';

const definition = {
  title: 'Find baby clothes', utterance: 'Where are my baby clothes?',
  assets: [{ id: 'loft', title: 'Loft', kind: 'location' },
    { id: 'clothes', title: '3–6 months', kind: 'item', description: 'Winter', parentId: 'loft', tagNames: ['baby', 'clothes'] }],
  expectations: { kind: 'answer', referencedAssets: ['clothes'], locations: [{ assetId: 'clothes', ancestorId: 'loft' }], forbiddenOperations: ['create'] }
};
const revision = { id: 'revision', caseId: 'case', number: 1, authorId: 'owner', createdAt: '2026-09-05T12:00:00Z', definition };
function repository(data: unknown, requests: Request[] = []) {
  return new CaseAPIRepository('https://api.example.test', () => 'session', async (input, init) => {
    requests.push(new Request(input, init));
    return Response.json({ data, meta: { tenantId: 'home' } });
  });
}
describe('evaluation case API repository', () => {
  it('preserves fixtures, tags, containment and expected locations across read and append', async () => {
    const requests: Request[] = [];
    const repo = repository(revision, requests);
    const saved = await repo.get('home', 'case', 'revision');
    expect(saved.definition.assets[0]).toEqual({ id: 'loft', title: 'Loft', kind: 'location', description: '', parentId: '', tagNames: [] });
    expect(saved.definition.assets[1]).toEqual(definition.assets[1]);
    expect(saved.definition.expectations).toEqual({ ...definition.expectations, proposals: [] });
    await repo.append('home', 'case', 1, saved.definition);
    expect(requests[1].url).toBe('https://api.example.test/tenants/home/conversation-evaluation-cases/case/revisions');
    expect(await requests[1].json()).toEqual({ expectedRevision: 1, definition: saved.definition });
  });
  it('preserves every expected proposal field on creation', async () => {
    const proposal = { operation: 'create', targetId: '', destinationId: 'loft', newKind: 'container', newTitle: 'Baby clothes', details: '' };
    const value = { ...revision, definition: { ...definition, expectations: { kind: 'proposal', proposals: [proposal] } } };
    const requests: Request[] = [];
    const repo = repository(value, requests);
    const saved = await repo.get('home', 'case');
    await repo.create('home', saved.definition);
    expect((await requests[1].json()).definition.expectations.proposals).toEqual([proposal]);
  });
  it('rejects responses for another case or revision', async () => {
    await expect(repository({ ...revision, caseId: 'other' }).get('home', 'case')).rejects.toMatchObject({ kind: 'invalid' });
    await expect(repository({ ...revision, id: 'other' }).get('home', 'case', 'revision')).rejects.toMatchObject({ kind: 'invalid' });
  });
  it('rejects unsupported outcome and fixture kinds', async () => {
    await expect(repository({ ...revision, definition: { ...definition, expectations: { kind: 'unknown' } } }).get('home', 'case')).rejects.toMatchObject({ kind: 'invalid' });
    await expect(repository({ ...revision, definition: { ...definition, assets: [{ id: 'x', title: 'x', kind: 'unknown' }] } }).get('home', 'case')).rejects.toMatchObject({ kind: 'invalid' });
  });
});
