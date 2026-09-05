import type { CaseDefinition, CaseFixtureAsset } from '$lib/domain/conversationCase';
export function fixtureParentChoices(assets: readonly CaseFixtureAsset[], id: string): CaseFixtureAsset[] {
  const byId = new Map(assets.map(asset => [asset.id, asset]));
  return assets.filter(candidate => {
    if (candidate.kind === 'item') return false;
    const seen = new Set<string>();
    let current: CaseFixtureAsset | undefined = candidate;
    while (current) {
      if (current.id === id || seen.has(current.id)) return false;
      seen.add(current.id);
      current = byId.get(current.parentId);
    }
    return true;
  });
}
export function fixtureRemovalBlocked(definition: CaseDefinition, id: string): boolean {
  const expectations = definition.expectations;
  return definition.assets.some(asset => asset.parentId === id) || expectations.referencedAssets.includes(id) ||
    expectations.locations.some(location => location.assetId === id || location.ancestorId === id) ||
    expectations.proposals.some(proposal => proposal.targetId === id || proposal.destinationId === id);
}
export function nextFixtureId(assets: readonly CaseFixtureAsset[]): string {
  const used = new Set(assets.map(asset => asset.id));
  let number = 1;
  while (used.has(`asset-${number}`)) number++;
  return `asset-${number}`;
}
