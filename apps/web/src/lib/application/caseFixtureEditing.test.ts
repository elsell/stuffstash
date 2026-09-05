import { describe, expect, it } from 'vitest';
import type { CaseDefinition } from '$lib/domain/conversationCase';
import { fixtureParentChoices, fixtureRemovalBlocked, nextFixtureId } from './caseFixtureEditing';
const assets: CaseDefinition['assets'] = [
  { id: 'asset-1', title: 'Loft', kind: 'location', parentId: '', description: '', tagNames: [] },
  { id: 'bin', title: 'Clothes', kind: 'container', parentId: 'asset-1', description: '', tagNames: [] },
  { id: 'item', title: '3–6 months', kind: 'item', parentId: 'bin', description: '', tagNames: ['baby', 'clothes'] }
];
const definition: CaseDefinition = { title: 'Clothes', utterance: 'Where are my baby clothes?', assets,
  expectations: { kind: 'answer', referencedAssets: ['item'], locations: [{ assetId: 'item', ancestorId: 'asset-1' }], proposals: [], forbiddenOperations: [] } };
describe('case fixture editing', () => {
  it('excludes self, descendants and items from parent choices', () => {
    expect(fixtureParentChoices(assets, 'asset-1')).toEqual([]);
    expect(fixtureParentChoices(assets, 'bin').map(asset => asset.id)).toEqual(['asset-1']);
    expect(fixtureParentChoices(assets, 'item').map(asset => asset.id)).toEqual(['asset-1', 'bin']);
  });
  it('protects fixtures referenced by containment, answers and proposals', () => {
    expect(fixtureRemovalBlocked(definition, 'asset-1')).toBe(true);
    expect(fixtureRemovalBlocked(definition, 'item')).toBe(true);
    const proposalCase = { ...definition, expectations: { ...definition.expectations, referencedAssets: [], locations: [],
      proposals: [{ operation: 'move' as const, targetId: 'item', destinationId: 'bin', newKind: '' as const, newTitle: '', details: '' }] } };
    expect(fixtureRemovalBlocked(proposalCase, 'item')).toBe(true);
    expect(fixtureRemovalBlocked({ ...definition, expectations: { ...definition.expectations, referencedAssets: [], locations: [] } }, 'item')).toBe(false);
  });
  it('chooses a collision-free local fixture identifier', () => { expect(nextFixtureId(assets)).toBe('asset-2'); });
});
