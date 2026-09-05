import { describe, expect, it } from 'vitest';
import type { CaseDefinition } from '$lib/domain/conversationCase';
import { prepareCaseDraft } from './caseDraftValidation';
const base: CaseDefinition = { title: ' Baby clothes ', utterance: ' Where are my baby clothes? ', assets: [
  { id: 'loft', title: ' Loft ', kind: 'location', parentId: '', description: '', tagNames: [] },
  { id: 'clothes', title: '3–6 months', kind: 'item', parentId: 'loft', description: '', tagNames: [' baby ', '', 'Baby', 'clothes, winter'] }
], expectations: { kind: 'answer', referencedAssets: ['clothes'], locations: [{ assetId: 'clothes', ancestorId: 'loft' }], proposals: [], forbiddenOperations: ['create'] } };
describe('case draft preparation', () => {
  it('points to duplicate location expectations instead of silently deleting them', () => {
    const draft = structuredClone(base); draft.expectations.locations.push({ ...draft.expectations.locations[0] });
    const result = prepareCaseDraft(draft);
    expect(result.issues.map(issue => issue.field)).toContain('location-parent-1');
    expect(result.definition.expectations.locations).toHaveLength(2);
  });
  it('normalizes a valid tagged location case without mutating its draft', () => {
    const result = prepareCaseDraft(base);
    expect(result.issues).toEqual([]); expect(result.definition.title).toBe('Baby clothes');
    expect(result.definition.assets[1].tagNames).toEqual(['baby', 'clothes, winter']);
    expect(base.assets[1].tagNames).toContain(' baby ');
  });
  it('identifies broken containment and location expectations', () => {
    const draft = structuredClone(base); draft.assets[1].parentId = 'missing';
    const fields = prepareCaseDraft(draft).issues.map(issue => issue.field);
    expect(fields).toContain('fixture-parent-clothes'); expect(fields).toContain('location-parent-0');
  });
  it('rejects contradictory creation expectations and blank titles', () => {
    const draft = structuredClone(base); draft.expectations.kind = 'proposal';
    draft.expectations.proposals = [{ operation: 'create', targetId: '', destinationId: 'loft', newKind: 'item', newTitle: '', details: '' }];
    const fields = prepareCaseDraft(draft).issues.map(issue => issue.field);
    expect(fields).toContain('proposal-title-0'); expect(fields).toContain('proposal-operation-0');
  });
  it('checks UTF-8 byte limits without silently truncating names', () => {
    const draft = structuredClone(base); draft.assets[1].title = '衣'.repeat(54);
    const result = prepareCaseDraft(draft);
    expect(result.issues.map(issue => issue.field)).toContain('fixture-title-clothes');
    expect(result.definition.assets[1].title).toBe(draft.assets[1].title);
  });
});
