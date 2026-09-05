export type CaseAssetKind = 'item' | 'container' | 'location';
export type CaseOutcome = 'answer' | 'clarification' | 'proposal' | 'failure';
export type CaseOperation = 'locate' | 'exists' | 'list_inventory' | 'list_contents' | 'detail' | 'checkout_status' |
  'asset_history' | 'checkout_history' | 'create' | 'move' | 'archive' | 'restore' | 'checkout' | 'return' | 'unsupported';
export interface CaseFixtureAsset {
  id: string; title: string; kind: CaseAssetKind; description: string; parentId: string; tagNames: string[];
}
export interface CaseLocation { assetId: string; ancestorId: string }
export interface CaseProposal {
  operation: CaseOperation; targetId: string; destinationId: string; newKind: CaseAssetKind | ''; newTitle: string; details: string;
}
export interface CaseExpectations {
  kind: CaseOutcome; referencedAssets: string[]; locations: CaseLocation[]; proposals: CaseProposal[]; forbiddenOperations: CaseOperation[];
}
export interface CaseDefinition { title: string; utterance: string; assets: CaseFixtureAsset[]; expectations: CaseExpectations }
export interface CaseRevision { id: string; caseId: string; number: number; authorId: string; createdAt: string; definition: CaseDefinition }
export interface CaseHead { id: string; title: string; latestRevision: number; latestRevisionId: string; createdAt: string; updatedAt: string }
