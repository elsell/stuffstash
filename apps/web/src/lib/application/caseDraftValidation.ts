import { t } from '$lib/presentation/localization';
import type { CaseDefinition } from '$lib/domain/conversationCase';
import { fixtureParentChoices } from './caseFixtureEditing';
export interface CaseDraftIssue { field: string; message: string }
const bytes = (value: string) => new TextEncoder().encode(value).length;
export function prepareCaseDraft(input: CaseDefinition): { definition: CaseDefinition; issues: CaseDraftIssue[] } {
  const definition = structuredClone(input);
  definition.title = definition.title.trim(); definition.utterance = definition.utterance.trim();
  for (const asset of definition.assets) {
    asset.title = asset.title.trim(); asset.description = asset.description.trim();
    const seen = new Set<string>();
    asset.tagNames = asset.tagNames.map(tag => tag.trim()).filter(tag => {
      const key = tag.toLowerCase(); if (!tag || seen.has(key)) return false; seen.add(key); return true;
    });
  }
  for (const proposal of definition.expectations.proposals) { proposal.newTitle = proposal.newTitle.trim(); proposal.details = proposal.details.trim(); }
  const issues: CaseDraftIssue[] = [];
  const issue = (field: string, message: string) => { issues.push({ field, message }); };
  const text = (field: string, value: string, max: number, required = true, byteLimit = false) => {
    if ((required && !value) || (byteLimit ? bytes(value) : [...value].length) > max) issue(field, t(`web.caseDraftValidation.${required ? 'required' : 'optional'}${byteLimit ? 'Bytes' : 'Characters'}`, { max }));
  };
  text('case-title', definition.title, 160); text('case-utterance', definition.utterance, 4000);
  if (definition.assets.length > 100) issue('case-fixtures-title', t('web.caseDraftValidation.useNoMoreThan100Fixtures'));
  const assets = new Map(definition.assets.map(asset => [asset.id, asset]));
  for (const asset of definition.assets) {
    text(`fixture-title-${asset.id}`, asset.title, 160, true, true);
    text(`fixture-description-${asset.id}`, asset.description, 2000, false);
    if (asset.tagNames.length > 32 || asset.tagNames.some(tag => bytes(tag) > 80)) issue(`fixture-tags-${asset.id}`, t('web.caseDraftValidation.useUpTo32TagsEachNoLongerThan'));
    const seen = new Set([asset.id]); let parent = asset.parentId;
    while (parent) {
      const ancestor = assets.get(parent);
      if (!ancestor || ancestor.kind === 'item' || seen.has(parent) || seen.size >= 32) { issue(`fixture-parent-${asset.id}`, t('web.caseDraftValidation.chooseAValidParentWithoutCircularOrOverlyDeep')); break; }
      seen.add(parent); parent = ancestor.parentId;
    }
  }
  const expected = definition.expectations;
  const changes = new Set(['create', 'move', 'archive', 'restore', 'checkout', 'return']);
  if (expected.proposals.length > 100 || expected.locations.length > 100 || expected.referencedAssets.length > 100) issue('case-expectations-title', t('web.caseDraftValidation.useNoMoreThan100ExpectationsOfEachKind'));
  if (new Set(expected.referencedAssets).size !== expected.referencedAssets.length || new Set(expected.forbiddenOperations).size !== expected.forbiddenOperations.length) issue('case-expectations-title', t('web.caseDraftValidation.removeRepeatedReferencesOrForbiddenChanges'));
  if (expected.forbiddenOperations.some(operation => !changes.has(operation))) issue('case-expectations-title', t('web.caseDraftValidation.onlyInventoryChangesCanBeForbidden'));
  if ((expected.kind === 'proposal') !== (expected.proposals.length > 0)) issue('expected-outcome', t('web.caseDraftValidation.proposedChangesRequireTheProposalOutcomeWithAtLeast'));
  if (expected.referencedAssets.some(id => !assets.has(id))) issue('case-expectations-title', t('web.caseDraftValidation.chooseExistingFixturesForRequiredReferences'));
  const locations = new Set<string>();
  expected.locations.forEach((location, index) => {
    const identity = JSON.stringify([location.assetId, location.ancestorId]);
    if (locations.has(identity)) issue(`location-parent-${index}`, t('web.caseDraftValidation.removeThisRepeatedLocationExpectation'));
    locations.add(identity);
    let parent = assets.get(location.assetId)?.parentId; const seen = new Set<string>(); let matched = false;
    while (parent && !seen.has(parent)) { seen.add(parent); if (parent === location.ancestorId && assets.has(parent)) matched = true; parent = assets.get(parent)?.parentId; }
    if (!matched) issue(`location-parent-${index}`, t('web.caseDraftValidation.chooseAnAncestorOfTheSelectedFixture'));
  });
  const proposals = new Set<string>();
  expected.proposals.forEach((proposal, index) => {
    const identity = JSON.stringify([proposal.operation, proposal.targetId, proposal.destinationId, proposal.newKind, proposal.newTitle, proposal.details]);
    if (proposals.has(identity)) issue(`proposal-operation-${index}`, t('web.caseDraftValidation.removeThisRepeatedProposedChange'));
    proposals.add(identity);
    if (!changes.has(proposal.operation)) issue(`proposal-operation-${index}`, t('web.caseDraftValidation.chooseASupportedInventoryChange'));
    if (expected.forbiddenOperations.includes(proposal.operation)) issue(`proposal-operation-${index}`, t('web.caseDraftValidation.thisChangeIsAlsoForbiddenUpdateOneOfThe'));
    if (proposal.operation === 'create') {
      text(`proposal-title-${index}`, proposal.newTitle, 160, true, true);
      if (!['item', 'container', 'location'].includes(proposal.newKind)) issue(`proposal-kind-${index}`, t('web.caseDraftValidation.chooseAKindForTheNewItem'));
      if (proposal.targetId) issue(`proposal-operation-${index}`, t('web.caseDraftValidation.creationCannotTargetAnExistingFixture'));
    } else {
      if (!assets.has(proposal.targetId)) issue(`proposal-target-${index}`, t('web.caseDraftValidation.chooseTheExistingFixtureToChange'));
      if (proposal.newKind || proposal.newTitle) issue(`proposal-operation-${index}`, t('web.caseDraftValidation.onlyCreationCanSpecifyANewNameAndKind'));
    }
    if (proposal.destinationId) {
      const eligible = fixtureParentChoices(definition.assets, proposal.operation === 'move' ? proposal.targetId : '');
      if (!['create', 'move'].includes(proposal.operation) || !eligible.some(asset => asset.id === proposal.destinationId)) issue(`proposal-destination-${index}`, t('web.caseDraftValidation.chooseAValidContainerOrLocationDestination'));
    }
    text(`proposal-details-${index}`, proposal.details, 500, false);
    if (proposal.details && !['checkout', 'return'].includes(proposal.operation)) issue(`proposal-operation-${index}`, t('web.caseDraftValidation.detailsApplyOnlyToCheckoutAndReturn'));
  });
  return { definition, issues };
}
